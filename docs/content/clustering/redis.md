---
title: "Distributed Redis Cache"
weight: 2
---

# Distributed Redis Cache & Rate Limiting

When you scale your bot across multiple separate machines or containers, each process runs independently. Wumpkit provides an official Redis extension under `github.com/wumpkit-go/wumpkit/extra/redis` to share state, rate limits, and caches across all nodes.

## The Clustering Challenges

Running multiple bot processes without a shared backend causes three common problems:

1. **REST Rate Limit Collisions**:
   Discord rate limits are enforced per-route across your entire bot token. If two separate processes hit the same channel route concurrently without knowing about each other, they can trigger HTTP 429 penalties.

2. **Gateway Identify Lockouts**:
   Discord allows only 1 Gateway Identify handshake every 5 seconds globally. If multiple cluster processes restart at the same time, their shards will spam Identify simultaneously, causing Discord to close the WebSocket with error code 4000 or 4004.

3. **Fragmented Caches**:
   In-memory caching only stores data for the guilds assigned to the local process. A command running on Shard 0 would not have cached user or guild data from Shard 3.

## Setting Up Redis Extensions

You can solve all three issues by plugging Redis into Wumpkit:

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/wumpkit-go/wumpkit"
	"github.com/wumpkit-go/wumpkit/extra/redis"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	// 1. connect to redis
	rdb := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("cannot reach redis", "error", err)
	}

	// 2. shared rest rate limit store
	sharedRateLimits := redis.NewRedisRateLimitStore(rdb, "mybot:ratelimit:")

	// 3. shared gateway identify rate limiter
	sharedIdentifyLimiter := redis.NewRedisShardsIdentifyRateLimiter(
		rdb,
		redis.WithIdentifyMaxConcurrency(1),
		redis.WithIdentifyInterval(5*time.Second),
		redis.WithIdentifyKey("mybot:identify:lock"),
	)

	// 4. shared distributed cache
	sharedCache := redis.NewRedisCacheManager(rdb, redis.WithRedisCacheConfig(redis.RedisCacheConfig{
		Prefix:     "mybot:cache:",
		Flags:      wumpkit.CacheFlagsAll,
		DefaultTTL: 12 * time.Hour,
		ResourceTTLs: redis.RedisCacheResourceTTLs{
			Users:    1 * time.Hour,
			Guilds:   24 * time.Hour,
			Members:  6 * time.Hour,
			Channels: 12 * time.Hour,
			Messages: 30 * time.Minute,
		},
		SlidingExpiration: redis.RedisCacheSlidingExpiration{
			Users:    true,
			Channels: true,
		},
	}))

	// 5. initialize the bot client with redis components
	bot := wumpkit.New(
		ctx,
		wumpkit.WithToken(os.Getenv("DISCORD_TOKEN")),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			TotalShards:     4,
			ShardIDs:        []int{0, 1},
			Intents:         wumpkit.GatewayIntentGuilds | wumpkit.GatewayIntentGuildMessages,
			IdentifyLimiter: sharedIdentifyLimiter,
		}),
		wumpkit.WithRESTConfig(wumpkit.RequesterConfig{
			RateLimitStore: sharedRateLimits,
		}),
		wumpkit.WithCacheManager(sharedCache),
	)

	if err := bot.StartGatewayWithSignals(); err != nil {
		slog.Error("bot stopped", "error", err)
	}
}
```

## Cache Queries

When using `sharedCache`, the helper methods on `bot` query Redis directly. All nodes in your cluster see the exact same cached data:

```go
// query total cached entities across all clusters
totalUsers := bot.CountUsers(ctx)
totalGuilds := bot.CountGuilds(ctx)
guildMembers := bot.CountGuildMembers(ctx, guildID)

slog.Info("cluster cache metrics",
	"cached_users", totalUsers,
	"cached_guilds", totalGuilds,
	"members_in_guild", guildMembers,
)
```
