---
title: "Serverless & HTTP Events"
weight: 2
---

# Serverless & HTTP Interactions

Discord allows applications to receive slash commands and events via HTTP POST requests instead of maintaining a persistent WebSocket Gateway connection. This architecture is ideal for serverless environments (like AWS Lambda, Google Cloud Run, and Kubernetes autoscaling) where keeping long-lived connections is unnecessary.

Wumpkit includes built-in support for Discord HTTP interactions, complete with automatic Ed25519 cryptographic signature verification and ping-pong handshakes.

## How It Works

1. In the **Discord Developer Portal**, navigate to your application and copy your **Public Key** from the General Information page.
2. Provide an **Interactions Endpoint URL** (e.g. `https://your-domain.com/events`). Discord sends a test `PING` to verify your server.
3. Wumpkit verifies the `X-Signature-Ed25519` and `X-Signature-Timestamp` headers automatically on every request.

## Standalone Server Setup

You can run Wumpkit directly as an HTTP server:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/wumpkit/wumpkit"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithPublicKey(publicKey),
	)

	// handle incoming slash commands over http
	bot.OnInteractionCreate(func(ctx context.Context, evt wumpkit.InteractionCreateEvent) {
		cmd, ok := evt.Interaction.(*wumpkit.ChatInputCommandInteraction)
		if !ok {
			return
		}

		if cmd.Data.Name == "status" {
			_ = bot.RespondMessage(ctx, cmd, "Serverless worker online and responsive.")
		}
	})

	// handle developer portal events (app authorization, subscriptions, etc.)
	bot.OnApplicationAuthorized(func(ctx context.Context, evt wumpkit.ApplicationAuthorizedEvent) {
		log.Printf("Bot installed into server by %s\n", evt.User.Username)
	})

	log.Printf("Starting HTTP Webhook server on port :%s\n", port)
	if err := bot.StartWebhookEventsWithSignals(":"+port, wumpkit.WithWebhookEventsPath("/events")); err != nil {
		log.Fatal(err)
	}
}
```

## Mounting in Existing Go Routers

If your application already runs a web server (like `net/http`, Chi, Echo, or Gin), you do not need to replace your server. Simply mount Wumpkit's `WebhookEventsHandler()` on your existing route:

```go
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/wumpkit/wumpkit"
)

func main() {
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(os.Getenv("DISCORD_TOKEN")),
		wumpkit.WithPublicKey(os.Getenv("DISCORD_PUBLIC_KEY")),
	)

	bot.OnInteractionCreate(func(ctx context.Context, evt wumpkit.InteractionCreateEvent) {
		// handle interactions
	})

	mux := http.NewServeMux()

	// mount discord webhook handler at /api/discord/events
	mux.Handle("/api/discord/events", bot.WebhookEventsHandler())

	// standard health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.ListenAndServe(":8080", mux)
}
```
