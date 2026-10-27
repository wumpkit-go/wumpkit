---
title: "Listening to Events"
weight: 3
---

# Listening to Events

Wumpkit dispatches real-time Discord events directly to typed callback functions registered on your `*Client`.

## Registering Event Callbacks

Register callbacks before or after connecting to the Gateway. Each callback receives a `context.Context` and a typed event struct containing the event payload:

```go
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wumpkit/wumpkit"
)

func registerEvents(bot *wumpkit.Client) {
	// fired when the bot finishes identifying with discord
	bot.OnReady(func(ctx context.Context, evt wumpkit.ReadyEvent) {
		slog.Info("Connected to Discord",
			"user", evt.User.Username,
			"guilds_count", len(evt.Guilds),
			"shard_id", evt.ShardID,
		)
	})

	// fired when a guild member joins
	bot.OnGuildMemberAdd(func(ctx context.Context, evt wumpkit.GuildMemberAddEvent) {
		fmt.Printf("User %s joined guild %s\n", evt.Member.User.Username, evt.GuildID)
	})

	// fired when an interaction (slash command, button, modal) arrives
	bot.OnInteractionCreate(func(ctx context.Context, evt wumpkit.InteractionCreateEvent) {
		switch in := evt.Interaction.(type) {
		case *wumpkit.ChatInputCommandInteraction:
			fmt.Printf("Received slash command /%s\n", in.Data.Name)
		case *wumpkit.ComponentInteraction:
			fmt.Printf("Received component click: %s\n", in.Data.CustomID)
		}
	})
}
```

## Frequently Used Event Listeners

| Registration Method | Event Payload Struct | When Fired |
| :--- | :--- | :--- |
| `bot.OnReady` | `ReadyEvent` | Initial gateway handshake finished |
| `bot.OnResumed` | `ResumedEvent` | Gateway session resumed after a disconnect |
| `bot.OnMessageCreate` | `MessageCreateEvent` | A message was created in a channel or DM |
| `bot.OnMessageUpdate` | `MessageUpdateEvent` | A message was edited or pinned |
| `bot.OnMessageDelete` | `MessageDeleteEvent` | A single message was deleted |
| `bot.OnInteractionCreate` | `InteractionCreateEvent` | Slash command or button/select/modal interaction |
| `bot.OnGuildCreate` | `GuildCreateEvent` | Bot joined a server or server became available |
| `bot.OnGuildDelete` | `GuildDeleteEvent` | Bot was kicked, banned, or left a server |
| `bot.OnGuildMemberAdd` | `GuildMemberAddEvent` | New member joined a server |
| `bot.OnGuildMemberRemove` | `GuildMemberRemoveEvent` | Member left, was kicked, or banned |
| `bot.OnVoiceStateUpdate` | `VoiceStateUpdateEvent` | Member joined, left, muted, or moved voice channels |

## Dispatch Execution Modes

By default, Wumpkit processes event handlers sequentially on each shard connection to guarantee message order and prevent race conditions on shared memory.

If your bot processes high event volume and you want each incoming event to spawn its own goroutine, set `DispatcherExecutionModeConcurrent`:

```go
bot := wumpkit.New(
	ctx,
	wumpkit.WithToken(token),
	wumpkit.WithDispatcherConfig(wumpkit.DispatcherConfig{
		ExecutionMode: wumpkit.DispatcherExecutionModeConcurrent,
	}),
)
```

In concurrent mode, each registered handler executes in a new goroutine as soon as the event payload is unmarshaled.
