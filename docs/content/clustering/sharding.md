---
title: "Manual & Auto Sharding"
weight: 1
---

# Manual & Auto Sharding

Sharding splits your bot connection into multiple independent WebSocket connections called shards. Each shard is responsible for a subset of your bot servers.

Discord requires sharding when your bot reaches 2,500 servers.

## Automatic Sharding

For a single bot process, Wumpkit can configure sharding automatically. If you leave `TotalShards` as zero, Wumpkit calls Discord `/gateway/bot` endpoint, retrieves the recommended shard count for your bot, and connects all shards:

```go
bot := wumpkit.New(
	context.Background(),
	wumpkit.WithToken(token),
	wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
		Intents: wumpkit.GatewayIntentGuilds | wumpkit.GatewayIntentGuildMessages,
	}),
)
```

## Manual Sharding (Multi-Process)

When scaling across multiple servers, Docker containers, or Kubernetes pods, you run multiple instances of your bot program. Each process handles a specific list of shard IDs:

```go
package main

import (
	"context"
	"os"

	"github.com/wumpkit-go/wumpkit"
)

func main() {
	// example: 4 total shards across your entire bot
	// this specific process runs shard 0 and shard 1
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(os.Getenv("DISCORD_TOKEN")),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			TotalShards: 4,
			ShardIDs:    []int{0, 1},
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
		}),
	)

	_ = bot.StartGatewayWithSignals()
}
```

A second process can then be started with `ShardIDs: []int{2, 3}` and `TotalShards: 4`.

## Calculating Guild Shard Routing

Discord assigns guilds to shards using this formula:

`(guild_id >> 22) % total_shards`

In Wumpkit, the `wumpkit.Snowflake` type provides a built-in helper method:

```go
bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
	// calculate which shard handles this guild
	guildShardID := event.Message.GuildID.GuildShardID(bot.TotalShards())

	slog.Info("message received",
		"guild_id", event.Message.GuildID,
		"handled_by_shard", guildShardID,
	)
})
```

## Inspecting Shard Health and Latency

You can monitor your shards at runtime using the `ShardManager`:

```go
sm := bot.ShardManager()
if sm != nil {
	slog.Info("shard statistics",
		"active_shards_in_process", sm.ShardCount(),
		"average_latency", sm.AverageLatency(),
	)

	// iterate over each shard individually
	for _, shard := range sm.Shards() {
		slog.Info("shard status",
			"shard_id", shard.ShardID(),
			"connected", shard.IsConnected(),
			"latency", shard.Latency(),
			"sequence", shard.Sequence(),
		)
	}
}
```
