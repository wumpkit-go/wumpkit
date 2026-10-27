package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wumpkit/wumpkit"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	// configure a multi-shard bot with explicit shard IDs (e.g. Shards 0 and 1 of 4 total shards)
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			TotalShards: 4,
			ShardIDs:    []int{0, 1},
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
			Compression: wumpkit.GatewayCompressionZlibStream,
		}),
	)

	// shard-aware message handler
	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		if event.Message.Author.Bot {
			return
		}

		if event.Message.Content == "!shard" {
			// calculate shard routing for this guild
			guildShardID := event.Message.GuildID.GuildShardID(bot.TotalShards())
			avgLatency := bot.ShardManager().AverageLatency()

			embed := wumpkit.NewEmbedBuilder().
				SetTitle("Shard Diagnostics").
				SetDescription(fmt.Sprintf(
					"Guild Shard ID: %d\n"+
						"Total Shards: %d\n"+
						"Average Latency: %v\n"+
						"Active Shards: %d",
					guildShardID,
					bot.TotalShards(),
					avgLatency,
					bot.ShardManager().ShardCount(),
				)).
				SetColor(wumpkit.ColorSummerSky).
				SetTimestamp(time.Now()).
				Build()

			_, _ = bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
				Embeds: []wumpkit.Embed{embed},
			})
		}
	})

	// periodic shard health monitor
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if sm := bot.ShardManager(); sm != nil {
				for _, shard := range sm.Shards() {
					log.Printf(
						"[Shard %d/%d] Connected: %v | Latency: %v | Seq: %d",
						shard.ShardID(),
						shard.TotalShards(),
						shard.IsConnected(),
						shard.Latency(),
						shard.Sequence(),
					)
				}
			}
		}
	}()

	log.Println("Starting Multi-Sharded Bot...")
	if err := bot.StartGatewayWithSignals(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
