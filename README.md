# wumpkit

[![Go Reference](https://pkg.go.dev/badge/github.com/wumpkit-go/wumpkit.svg)](https://pkg.go.dev/github.com/wumpkit-go/wumpkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/wumpkit-go/wumpkit)](https://goreportcard.com/report/github.com/wumpkit-go/wumpkit)
[![Test](https://github.com/wumpkit-go/wumpkit/actions/workflows/test.yml/badge.svg)](https://github.com/wumpkit-go/wumpkit/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/wumpkit-go/wumpkit?include_prereleases)](https://github.com/wumpkit-go/wumpkit/releases)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)

A lightweight and idiomatic Discord API wrapper for Go.

---

## Features

- **Context-First**: Built entirely around Go's standard `context.Context` package.
- **Pluggable Components**: First-class support for swapping core components (like caching, rate limit stores, logging) with custom implementations.
- **Automatic Sharding**: Full support for automatic and manual sharding in an easy way.

---

## Installation

```bash
go get github.com/wumpkit-go/wumpkit
```

> Requires **Go 1.22+**.

---

## Quick Start

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
		slog.Info("Connected", "user", event.User.Username, "guilds", len(event.Guilds))
	})

	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		if event.Message.Author.Bot || event.Message.Content != "!ping" {
			return
		}

		latency := bot.ShardManager().AverageLatency()
		_, _ = bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
			Content: fmt.Sprintf("Pong! Heartbeat latency: %v", latency),
		})
	})

	if err := bot.StartGatewayWithSignals(); err != nil {
		slog.Error("bot stopped", "error", err)
	}
}
```

---

## Distributed Setup (Redis)

For clustered bots sharing rate limits and cache across multiple nodes:

```go
import (
	"github.com/wumpkit-go/wumpkit"
	"github.com/wumpkit-go/wumpkit/extra/redis"
	goredis "github.com/redis/go-redis/v9"
)

rdb := goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})

bot := wumpkit.New(
	ctx,
	wumpkit.WithToken(token),
	wumpkit.WithCacheManager(redis.NewRedisCacheManager(rdb, wumpkit.CacheFlagsAll)),
	wumpkit.WithRateLimitStore(redis.NewRedisRateLimitStore(rdb)),
)
```

---

## Examples

Runnable examples are in the [`_examples`](_examples/) directory:

- [`pingpong`](_examples/pingpong/) - Basic ping pong bot
- [`slash_commands`](_examples/slash_commands/) - Application slash commands
- [`components`](_examples/components/) - Buttons and select menus
- [`jointocreate`](_examples/jointocreate/) - Dynamic voice channel creator
- [`sharded_bot`](_examples/sharded_bot/) - Multi-sharded bot setup
- [`redis_cache`](_examples/redis_cache/) - Distributed Redis cache and rate limiting
- [`oauth2_server`](_examples/oauth2_server/) - Standalone OAuth2 web server

---

## Community & Governance

- [Changelog](CHANGELOG.md) - Release history and version notes following Keep a Changelog.
- [Contributing Guide](CONTRIBUTING.md) - Development workflow, code style, and PR guidelines.
- [Security Policy](.github/SECURITY.md) - Vulnerability disclosure and reporting procedure.
- [Code of Conduct](.github/CODE_OF_CONDUCT.md) - Contributor Covenant 2.1 standards.
- [Support Guide](.github/SUPPORT.md) - Getting help, reporting bugs, and asking questions.

---

## License

BSD 3-Clause License. See [LICENSE](LICENSE) for details.
