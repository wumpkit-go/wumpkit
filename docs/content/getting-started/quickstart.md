---
title: "Quickstart"
weight: 2
---

# Quickstart

This guide shows you how to set up a Discord application, get your bot token, and run a working ping-pong bot in Go.

## 1. Create Your Discord Bot

Before writing code, you need a bot token from Discord:

1. Open the [Discord Developer Portal](https://discord.com/developers/applications).
2. Click **New Application**, give it a name, and create it.
3. In the left menu, click **Bot**.
4. Click **Reset Token** and copy your token to a safe place. Never share this token or commit it into git.
5. Scroll down to **Privileged Gateway Intents** and enable **Message Content Intent**. This lets your bot read message text like `!ping`. Click **Save Changes**.

To invite the bot to your test server:
1. Click **OAuth2** in the left menu, then click **URL Generator**.
2. Under Scopes, select `bot`.
3. Under Bot Permissions, select `Send Messages` and `Read Message History`.
4. Copy the generated URL at the bottom and open it in your browser to invite the bot.

## 2. Write the Bot Code

Set your token in your terminal:

```bash
export DISCORD_TOKEN="your_bot_token_here"
```

Create a `main.go` file with the following code:

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
		slog.Error("DISCORD_TOKEN environment variable is required")
		os.Exit(1)
	}

	// create the bot client
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
		}),
	)

	// listen for when the bot is fully connected
	bot.OnReady(func(ctx context.Context, event wumpkit.ReadyEvent) {
		slog.Info("Bot is ready",
			"username", event.User.Username,
			"guilds", len(event.Guilds),
		)
	})

	// listen for incoming messages
	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		// ignore messages from bots, including ourselves
		if event.Message.Author.Bot {
			return
		}

		if event.Message.Content == "!ping" {
			latency := bot.ShardManager().AverageLatency()

			_, err := bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
				Content: fmt.Sprintf("Pong! Heartbeat latency is %v", latency),
			})
			if err != nil {
				slog.Error("failed to send message", "error", err)
			}
		}
	})

	// start the gateway connection and wait for interrupt signals (ctrl+c)
	slog.Info("Starting bot...")
	if err := bot.StartGatewayWithSignals(); err != nil {
		slog.Error("bot stopped with error", "error", err)
	}
}
```

## 3. Run Your Bot

Run the bot from your terminal:

```bash
go run main.go
```

You should see log output showing your bot has connected. Open Discord, head to a channel where your bot is present, and type `!ping`. The bot will respond with its heartbeat latency.

Press `Ctrl+C` in your terminal to shut the bot down cleanly.
