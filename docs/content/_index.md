---
title: "Wumpkit Documentation"
---

# Wumpkit

Wumpkit is a lightweight Discord API library written in Go. It connects to Discord Gateway v10, makes REST API requests, handles interactions, and supports multi-process sharding with Redis.

## Documentation Sections

* [Getting Started](/getting-started/): Prerequisites, installation, and your first working bot.
* [Core Concepts](/core-concepts/): Client options, Gateway intents, event listeners, and REST calls.
* [Interactions](/interactions/): Slash commands, buttons, dropdown menus, and modal dialogs.
* [Sharding and Redis](/clustering/): Running multiple shards, tracking shard latency, and sharing rate limits with Redis.

## Quick Example

Here is a small bot that connects to the Gateway, listens for `!ping`, and replies with the current heartbeat latency:

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/wumpkit-go/wumpkit"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		slog.Error("DISCORD_TOKEN is required")
		os.Exit(1)
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
		slog.Info("Connected to Discord", "username", event.User.Username, "guilds", len(event.Guilds))
	})

	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		if event.Message.Author.Bot || event.Message.Content != "!ping" {
			return
		}

		latency := bot.ShardManager().AverageLatency()
		_, _ = bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
			Content: fmt.Sprintf("Pong! Latency: %v", latency),
		})
	})

	if err := bot.StartGatewayWithSignals(); err != nil {
		slog.Error("bot stopped with error", "error", err)
	}
}
```
