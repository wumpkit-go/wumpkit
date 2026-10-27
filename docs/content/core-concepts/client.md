---
title: "Client & Options"
weight: 1
---

# Client & Options

The `wumpkit.Client` is the main entry point for your application. It bundles the Discord WebSocket Gateway, the REST API requester, in-memory or Redis caching, and the event dispatcher into a single struct.

## Creating the Client

Create a client with `wumpkit.New()`. You pass a root context and configuration options:

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/wumpkit/wumpkit"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(os.Getenv("DISCORD_TOKEN")),
		wumpkit.WithLogger(logger),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds | wumpkit.GatewayIntentGuildMessages,
		}),
	)

	if err := bot.StartGatewayWithSignals(); err != nil {
		logger.Error("bot exited", "error", err)
	}
}
```

## Available Options

Here are the primary options you can pass to `wumpkit.New()`:

* `wumpkit.WithToken(token)`:
  Sets the bot authentication token. If your token includes the "Bot " prefix, Wumpkit strips it automatically. If the token is empty, the program logs an error and exits.

* `wumpkit.WithGatewayConfig(config)`:
  Configures the ShardManager and Gateway connections. Use this to select intents, set total shards, or pick specific shard IDs. You can also use its alias `wumpkit.WithShardManagerConfig(config)`.

* `wumpkit.WithLogger(logger)`:
  Supplies a custom logger. If omitted, Wumpkit uses Go standard `slog.Default()` at the Info level.

* `wumpkit.WithCacheManager(manager)`:
  Swaps the default in-memory cache with another implementation, such as the distributed Redis cache manager from `github.com/wumpkit/wumpkit/extra/redis`.

* `wumpkit.WithRESTConfig(config)`:
  Customizes HTTP client behavior, including request timeouts, retry counts, or custom rate limit stores. You can also use its alias `wumpkit.WithRequesterConfig(config)`.

## Starting and Stopping

Wumpkit offers two ways to run the Gateway:

1. `bot.StartGatewayWithSignals()`:
   The easiest and recommended method. It connects to the Gateway and listens in the background for operating system interrupt signals like `SIGINT` (Ctrl+C) and `SIGTERM`. When a signal arrives, it closes all WebSocket shards cleanly before exiting.

2. `bot.StartGateway()`:
   Connects to the Gateway and blocks until your parent context is cancelled or a fatal connection error occurs. Use this if you already have custom shutdown logic in your service.

To shut down the client manually from code, call `bot.Close()`.

## Cache-Aside Data Helpers

The client includes helper methods that look for entities in the cache first. If an entity is not cached, the client fetches it from Discord REST API and saves it in the cache for future calls:

```go
// checks cache first, falls back to rest api if missing
guild, err := bot.Guild(ctx, guildID)
if err != nil {
	slog.Error("could not get guild", "error", err)
	return
}
slog.Info("found guild", "name", guild.Name)

// also available for channels, members, and users
channel, err := bot.Channel(ctx, channelID)
member, err := bot.Member(ctx, guildID, userID)
user, err := bot.User(ctx, userID)
```
