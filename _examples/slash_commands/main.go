package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/wumpkit-go/wumpkit"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	appIDStr := os.Getenv("DISCORD_APP_ID")

	if token == "" || appIDStr == "" {
		log.Fatal("Both DISCORD_TOKEN and DISCORD_APP_ID environment variables are required")
	}

	appIDNum, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid DISCORD_APP_ID: %v", err)
	}
	appID := wumpkit.Snowflake(appIDNum)

	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds,
		}),
	)

	// register & Synchronize Global Application Commands when bot is ready
	bot.OnReady(func(ctx context.Context, event wumpkit.ReadyEvent) {
		log.Printf("Logged in as %s (ID: %s)", event.User.Username, event.User.ID)

		// 1. fluent Slash Command: /ping
		pingCmd := wumpkit.NewSlashCommandBuilder("ping", "Check bot gateway latency").Build()

		// 2. fluent Slash Command: /info with options and choices
		infoCmd := wumpkit.NewSlashCommandBuilder("info", "Get server or user information").
			AddUserOption("target", "Target user to inspect", false).
			AddStringOption("detail", "Detail level", false,
				wumpkit.NewStringChoice("Standard", "standard"),
				wumpkit.NewStringChoice("Extended", "extended"),
			).
			Build()

		// 3. fluent Context Menu Command: User Info (Right-click User -> Apps -> User Profile)
		userProfileCmd := wumpkit.NewUserCommandBuilder("User Profile").Build()

		// atomically synchronize all application commands with Discord in a single bulk overwrite call
		commands := []wumpkit.ApplicationCommand{pingCmd, infoCmd, userProfileCmd}
		synced, err := bot.BulkOverwriteGlobalApplicationCommands(ctx, appID, commands)
		if err != nil {
			log.Fatalf("Failed synchronizing application commands: %v", err)
		}
		log.Printf("Successfully synchronized %d application commands with Discord", len(synced))
	})

	// handle Slash Command and Context Menu Interactions
	bot.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {

		switch in := event.Interaction.(type) {
		case *wumpkit.ChatInputCommandInteraction:
			switch in.Data.Name {
			case "ping":
				latency := bot.ShardManager().AverageLatency()
				msg := fmt.Sprintf("Pong. Gateway Latency: %v", latency)
				_ = bot.RespondMessage(ctx, in, msg)

			case "info":
				// defer response to show Discord loading state
				_ = bot.DeferResponse(ctx, in, false)

				// simulate async background work
				time.Sleep(200 * time.Millisecond)

				user := in.Author()

				embed := wumpkit.NewEmbedBuilder().
					SetTitle("User Information").
					SetDescription(fmt.Sprintf(
						"%s\n%s: %s\n%s: %s\n%s: %v",
						wumpkit.Bold("Target Details"),
						wumpkit.InlineCode("Username"), user.Username,
						wumpkit.InlineCode("User ID"), user.ID,
						wumpkit.InlineCode("Is Bot"), user.Bot,
					)).
					SetThumbnail(user.AvatarURL()).
					SetColor(wumpkit.ColorSummerSky).
					SetTimestamp(time.Now()).
					Build()

				_, _ = bot.EditOriginalInteractionResponse(ctx, appID, in.GetToken(), wumpkit.EditFollowupMessageOptions{
					Embeds: &[]wumpkit.Embed{embed},
				})
			}

		case *wumpkit.UserContextMenuCommandInteraction:
			if in.Data.Name == "User Profile" {
				targetUser := in.Data.Resolved.Users[in.Data.TargetID]
				msg := fmt.Sprintf("Context menu profile lookup for: %s (ID: %s)", targetUser.Username, targetUser.ID)
				_ = bot.RespondMessage(ctx, in, msg, wumpkit.RespondMessageOptions{
					Ephemeral: true,
				})
			}
		}
	})

	log.Println("Starting Slash Commands Bot...")
	if err := bot.StartGatewayWithSignals(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
