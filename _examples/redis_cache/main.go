package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wumpkit/wumpkit"
	"github.com/wumpkit/wumpkit/extra/redis"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	// connect to Redis
	rdb := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis ping failed (%v). Ensure Redis is running on %s", err, redisAddr)
	}

	// configure distributed cache with fine-grained TTLs and sliding expiration
	redisCache := redis.NewRedisCacheManager(rdb, redis.WithRedisCacheConfig(redis.RedisCacheConfig{
		Prefix:     "mybot:cache:",
		Flags:      wumpkit.CacheFlagsAll,
		DefaultTTL: 24 * time.Hour,
		ResourceTTLs: redis.RedisCacheResourceTTLs{
			Users:    1 * time.Hour,
			Guilds:   24 * time.Hour,
			Members:  6 * time.Hour,
			Channels: 12 * time.Hour,
			Messages: 30 * time.Minute,
		},
		SlidingExpiration: redis.RedisCacheSlidingExpiration{
			Users:    true, // reset user TTL whenever accessed
			Channels: true,
		},
	}))

	bot := wumpkit.New(
		ctx,
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
		}),
		wumpkit.WithCacheManager(redisCache),
	)

	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		if event.Message.Author.Bot {
			return
		}

		if event.Message.Content == "!stats" {
			// cache counters querying Redis directly
			userCount := bot.CountUsers(ctx)
			guildCount := bot.CountGuilds(ctx)
			memberCount := bot.CountGuildMembers(ctx, event.Message.GuildID)

			// cache-Aside lookup: checks Redis first, falls back to REST if missing, and updates Redis
			guild, err := bot.Guild(ctx, event.Message.GuildID)
			guildName := "Unknown"
			if err == nil {
				guildName = guild.Name
			}

			embed := wumpkit.NewEmbedBuilder().
				SetTitle("Redis Distributed Cache Statistics").
				SetDescription(fmt.Sprintf(
					"Current Guild: %s\n"+
						"Cached Users: %d\n"+
						"Cached Guilds: %d\n"+
						"Cached Members in Guild: %d",
					guildName, userCount, guildCount, memberCount,
				)).
				SetColor(wumpkit.ColorMediumSeaGreen).
				SetTimestamp(time.Now()).
				Build()

			_, _ = bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
				Embeds: []wumpkit.Embed{embed},
			})
		}
	})

	log.Println("Starting Bot with Distributed Redis Cache...")
	if err := bot.StartGatewayWithSignals(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
