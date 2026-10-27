package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/wumpkit/wumpkit"
	"github.com/wumpkit/wumpkit/extra/redis"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// read cluster ID and topology from environment
	// example: 4 total clusters managing 16 total shards
	clusterID, _ := strconv.Atoi(os.Getenv("CLUSTER_ID"))         // 0, 1, 2, or 3
	totalClusters, _ := strconv.Atoi(os.Getenv("TOTAL_CLUSTERS")) // e.g. 4
	totalShards, _ := strconv.Atoi(os.Getenv("TOTAL_SHARDS"))     // e.g. 16

	if totalClusters <= 0 {
		totalClusters = 1
	}
	if totalShards <= 0 {
		totalShards = 4
	}

	// calculate which shards this specific cluster process is responsible for
	shardsPerCluster := totalShards / totalClusters
	startShardID := clusterID * shardsPerCluster
	endShardID := startShardID + shardsPerCluster
	if clusterID == totalClusters-1 {
		endShardID = totalShards // handle any remainder
	}

	var assignedShards []int
	for s := startShardID; s < endShardID; s++ {
		assignedShards = append(assignedShards, s)
	}

	log.Printf("[Cluster %d/%d] Managing Shards %v of %d total shards",
		clusterID, totalClusters, assignedShards, totalShards)

	// connect to Redis for shared cluster infrastructure
	rdb := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis ping failed (%v). Ensure Redis is running on %s", err, redisAddr)
	}

	// Clustering Best Practices & Architecture Requirements:
	//
	// 1. MUST USE: Shared REST Rate Limit Store (Redis)
	//    discord rate limits are enforced per-route across all bot processes
	//    if each cluster uses an isolated in-memory rate limiter, requests made
	//    concurrently from different clusters will collide and trigger HTTP 429
	//    penalties or Cloudflare IP bans
	//    => use: redis.NewRedisRateLimitStore(rdb)
	//
	// 2. STRONGLY RECOMMENDED: Shared Gateway Shard Identify Rate Limiter (Redis)
	//    discord allows only `max_concurrency` (default 1) Identify per 5 seconds
	//    globally for your bot token. If multiple clusters boot or reconnect at
	//    the same time without a shared identify rate limiter, shards will spam
	//    identify payloads simultaneously, leading to immediate Gateway session
	//    closures (Close Code 4000/4004) and identify rate limit lockouts
	//    => use: redis.NewRedisShardsIdentifyRateLimiter(rdb)
	//
	// 3. STRONGLY RECOMMENDED: Shared Distributed Cache (Redis)
	//    while in-memory caching will technically work locally, each cluster will
	//    only have partial cache data for the guilds connected to its shards
	//    using the Redis cache manager ensures all clusters share guild, channel,
	//    and user data with automatic sliding expiration and cross-process eviction
	//    => use: redis.NewRedisCacheManager(rdb, ...)

	sharedRateLimitStore := redis.NewRedisRateLimitStore(rdb, "mybot:ratelimit:")
	sharedIdentifyLimiter := redis.NewRedisShardsIdentifyRateLimiter(
		rdb,
		redis.WithIdentifyMaxConcurrency(1), // adjust if your bot has large-bot sharding max_concurrency > 1
		redis.WithIdentifyInterval(5*time.Second),
		redis.WithIdentifyKey("mybot:identify:slots"),
	)
	sharedCache := redis.NewRedisCacheManager(rdb, redis.WithRedisCacheConfig(redis.RedisCacheConfig{
		Prefix:     "mybot:cache:",
		Flags:      wumpkit.CacheFlagsAll,
		DefaultTTL: 12 * time.Hour,
	}))

	bot := wumpkit.New(
		ctx,
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			TotalShards:     totalShards,
			ShardIDs:        assignedShards,
			Intents:         wumpkit.GatewayIntentGuilds | wumpkit.GatewayIntentGuildMessages | wumpkit.GatewayIntentMessageContent,
			IdentifyLimiter: sharedIdentifyLimiter,
		}),
		wumpkit.WithRESTConfig(wumpkit.RequesterConfig{
			RateLimitStore: sharedRateLimitStore,
		}),
		wumpkit.WithCacheManager(sharedCache),
	)

	bot.OnReady(func(ctx context.Context, event wumpkit.ReadyEvent) {
		log.Printf("[Cluster %d] Shard %d READY (SessionID: %s)",
			clusterID, event.ShardID, event.SessionID)
	})

	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		if event.Message.Author.Bot {
			return
		}

		if event.Message.Content == "!cluster" {
			guildShardID := event.Message.GuildID.GuildShardID(totalShards)
			embed := wumpkit.NewEmbedBuilder().
				SetTitle("Cluster Diagnostics").
				SetDescription(fmt.Sprintf(
					"Cluster ID: %d / %d\n"+
						"Managing Shards: %v\n"+
						"Current Guild Shard: %d\n"+
						"Shared Rate Limiter: Redis (Active)\n"+
						"Shared Identify Limiter: Redis (Active)\n"+
						"Shared Cache: Redis (Active)",
					clusterID, totalClusters, assignedShards, guildShardID,
				)).
				SetColor(wumpkit.ColorSummerSky).
				SetTimestamp(time.Now()).
				Build()

			_, _ = bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
				Embeds: []wumpkit.Embed{embed},
			})
		}
	})

	log.Printf("[Cluster %d] Starting Cluster...", clusterID)
	if err := bot.StartGatewayWithSignals(); err != nil {
		log.Fatalf("[Cluster %d] Bot fatal error: %v", clusterID, err)
	}
}
