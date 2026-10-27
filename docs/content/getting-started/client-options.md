---
title: "Client Options"
weight: 3
---

# Client Options

When you initialize a new client with `wumpkit.New()`, you can pass functional options to configure logging, gateway settings, HTTP behavior, and caching.

## Configuration Example

Here is a bot configured with custom logging, gateway intents, and HTTP timeout options:

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/wumpkit/wumpkit"
)

func main() {
	// custom structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	bot := wumpkit.New(
		context.Background(),
		// bot token (bot prefix is stripped automatically if present)
		wumpkit.WithToken(os.Getenv("DISCORD_TOKEN")),

		// custom logger
		wumpkit.WithLogger(logger),

		// gateway connection settings
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
		}),

		// http rest client settings
		wumpkit.WithRequesterConfig(wumpkit.RequesterConfig{
			BaseURL:    "https://discord.com/api/v10",
			Timeout:    10 * time.Second,
			MaxRetries: 3,
		}),
	)

	// start with graceful os signal handling (sigint, sigterm)
	if err := bot.StartGatewayWithSignals(); err != nil {
		logger.Error("bot stopped with error", "error", err)
	}
}
```

## Available Client Options

### `WithToken(token string)`
Sets your bot authentication token. Wumpkit verifies the token is not empty and automatically strips any leading `Bot ` prefix if you accidentally included it.

### `WithLogger(logger Logger)`
Sets a custom logger for all internal Gateway and REST logs. Accepts standard `*slog.Logger` or any type implementing Wumpkit's `Logger` interface (`Debug`, `Info`, `Warn`, `Error`). If omitted, `slog.Default()` is used.

### `WithGatewayConfig(config ShardManagerConfig)`
Configures Gateway intents, WebSocket compression (`CompressionZstd` or `CompressionZlib`), shard counts, and identify rate limiting.

### `WithRequesterConfig(config RequesterConfig)`
Configures the HTTP client used for Discord REST requests. You can set the base URL, custom HTTP client, request timeouts, and maximum retries for transient network errors.

### `WithCacheManager(manager CacheManager)`
Replaces the default cache manager with a customized in-memory cache or an external store.

### `WithPublicKey(key string)`
Sets the application public key used to verify incoming Ed25519 signatures when running a serverless HTTP bot.
