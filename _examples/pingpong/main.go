package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/wumpkit/wumpkit"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		slog.Error("DISCORD_TOKEN environment variable is required")
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
		}),
		wumpkit.WithLogger(logger),
	)

	bot.OnReady(func(ctx context.Context, event wumpkit.ReadyEvent) {
		logger.Info("Bot connected and ready",
			"username", event.User.Username,
			"guilds_count", len(event.Guilds),
		)
	})

	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		if event.Message.Author.Bot {
			return
		}

		if event.Message.Content == "!ping" {
			latency := bot.ShardManager().AverageLatency()

			embed := wumpkit.NewEmbedBuilder().
				SetTitle("Pong").
				SetDescription(fmt.Sprintf("Heartbeat Latency: **%v**", latency)).
				SetColor(wumpkit.ColorSummerSky).
				SetTimestamp(time.Now()).
				Build()

			button := wumpkit.NewButtonBuilder().
				SetCustomID("ping_again").
				SetLabel("Refresh Ping").
				SetStyle(wumpkit.ButtonStylePrimary).
				Build()

			row := wumpkit.NewActionRowBuilder().
				SetComponent(button).
				Build()

			_, err := bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
				Embeds:     []wumpkit.Embed{embed},
				Components: []wumpkit.LayoutComponent{row},
			})
			if err != nil {
				logger.Error("failed sending pong message", "error", err)
			}
		}
	})

	bot.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
		if in, ok := event.Interaction.(*wumpkit.ComponentInteraction); ok {
			latency := bot.ShardManager().AverageLatency()

			embed := wumpkit.NewEmbedBuilder().
				SetTitle("Pong (Refreshed)").
				SetDescription(fmt.Sprintf("Heartbeat Latency: **%v**", latency)).
				SetColor(wumpkit.ColorMediumSeaGreen).
				SetTimestamp(time.Now()).
				Build()

			_ = bot.UpdateComponentMessage(ctx, in, wumpkit.RespondMessageOptions{
				Embeds: []wumpkit.Embed{embed},
			})
		}
	})

	logger.Info("Starting bot...")
	if err := bot.StartGatewayWithSignals(); err != nil {
		logger.Error("bot stopped with error", "error", err)
	}
}
