package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wumpkit-go/wumpkit"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}
	if publicKey == "" {
		log.Fatal("DISCORD_PUBLIC_KEY environment variable is required (found in Discord Developer Portal -> General Information)")
	}

	// create client with Bot Token and Application Public Key
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithPublicKey(publicKey),
	)

	// 1. handle Webhook Events (APPLICATION_AUTHORIZED, APPLICATION_DEAUTHORIZED, ENTITLEMENT_CREATE, etc.)
	bot.OnApplicationAuthorized(func(ctx context.Context, event wumpkit.ApplicationAuthorizedEvent) {
		log.Printf("App authorized by user %s (ID: %s)", event.User.Username, event.User.ID)
	})

	bot.OnApplicationDeauthorized(func(ctx context.Context, event wumpkit.ApplicationDeauthorizedEvent) {
		log.Printf("App deauthorized by user %s (ID: %s)", event.User.Username, event.User.ID)
	})

	bot.OnEntitlementCreate(func(ctx context.Context, event wumpkit.EntitlementCreateEvent) {
		log.Printf("New entitlement created! SKU ID: %s, User ID: %s", event.Entitlement.SkuID, event.Entitlement.UserID)
	})

	// 2. handle Slash Commands & Component Interactions over HTTP
	bot.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
		switch cmd := event.Interaction.(type) {
		case *wumpkit.ChatInputCommandInteraction:
			switch cmd.Data.Name {
			case "ping":
				embed := wumpkit.NewEmbedBuilder().
					SetTitle("Pong! 🏓").
					SetDescription("Hello from Discord **Webhook Events**! No Gateway WebSocket connection needed.").
					SetColor(wumpkit.ColorMediumSeaGreen).
					SetTimestamp(time.Now()).
					Build()

				_ = bot.RespondEmbeds(ctx, cmd, []wumpkit.Embed{embed})

			case "serverless":
				_ = bot.RespondMessage(ctx, cmd, fmt.Sprintf(
					"Running Webhook Events server on port :%s with ED25519 signature verification.", port,
				))
			}
		}
	})

	log.Printf("Starting Discord Webhook Events Server on :%s...", port)
	log.Printf("Configure your Webhook Events URL in the Discord Developer Portal to: https://your-domain.com/events")

	// Starting modes:
	//
	// option A: Start as a standalone Webhook Events server with graceful OS signal shutdown
	//           err := bot.StartWebhookEventsWithSignals(":" + port, wumpkit.WithWebhookEventsPath("/events"))
	//
	// option B: Mount into your existing router (net/http, gin, chi, echo, etc.)
	//           http.Handle("/events", bot.WebhookEventsHandler())
	//           http.ListenAndServe(":"+port, nil)
	//
	// option C: Start as a traditional WebSocket Gateway bot instead
	//           err := bot.StartGatewayWithSignals()

	if err := bot.StartWebhookEventsWithSignals(":"+port, wumpkit.WithWebhookEventsPath("/events")); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
