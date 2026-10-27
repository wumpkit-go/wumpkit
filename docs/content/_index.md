---
title: "Wumpkit Documentation"
---

# Wumpkit Documentation

wumpkit is a lightweight, modular Discord API library for Go. It connects to Discord Gateway v10, makes REST API requests with automatic rate limiting, handles slash commands and UI interactions, and supports multi-process sharding with Redis.

## Documentation Sections

* [Getting Started](getting-started/): Prerequisites, installation, quickstart guide, and client options.
* [Gateway Connection](gateway/): Connecting to the WebSocket gateway, managing intents, listening to events, and setting presence.
* [State & Caching](caching/): In-memory caching, cache flags, memory optimization, and automatic cache-aside queries.
* [REST Client](rest/): Direct HTTP API requests, multipart file uploads, route buckets, and rate limit handling.
* [Messages & Content](messages/): Sending messages, rich embeds, file uploads, native polls, reactions, and history pagination.
* [Interactions & UI](interactions/): Slash commands, buttons, select menus, popup modals, and context menus.
* [Guilds & Moderation](guilds/): Channels, member profiles, kicks, bans, bulk bans, timeouts, roles, and audit logs.
* [Sharding & Redis](clustering/): Multi-shard routing, latency tracking, and distributed state coordination with Redis.
* [Webhooks & HTTP Bots](webhooks/): Traditional incoming webhooks and serverless HTTP interactions with Ed25519 verification.
* [OAuth2 & Web Auth](oauth2/): Web authentication flows, token refreshes, and bot invitation URLs.
* [Utilities & Formatting](utilities/): Discord markdown helpers, dynamic timestamp formatting, colors, and snowflake calculations.

## Quick Example

Here is a small bot that connects to the Gateway, listens for `!ping`, and replies with the current heartbeat latency:

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
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
		}),
	)

	bot.OnReady(func(ctx context.Context, event wumpkit.ReadyEvent) {
		log.Printf("Logged in as %s across %d guilds", event.User.Username, len(event.Guilds))
	})

	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		if event.Message.Author.Bot || event.Message.Content != "!ping" {
			return
		}

		latency := bot.ShardManager().AverageLatency()
		_, _ = bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
			Content: fmt.Sprintf("Pong! Heartbeat latency is %v", latency),
		})
	})

	if err := bot.StartGatewayWithSignals(); err != nil {
		log.Printf("bot exited: %v", err)
	}
}
```
