---
title: "Gateway & Events"
weight: 2
---

# Gateway & Events

Discord uses a real-time WebSocket connection called the Gateway to send live events to your bot, such as new messages, member joins, and role updates.

## Gateway Intents

Intents tell Discord which events your bot wants to receive. Discord will not send events that your bot did not subscribe to.

You configure intents when setting up your client:

```go
bot := wumpkit.New(
	context.Background(),
	wumpkit.WithToken(token),
	wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
		Intents: wumpkit.GatewayIntentGuilds |
			wumpkit.GatewayIntentGuildMessages |
			wumpkit.GatewayIntentGuildMessageReactions |
			wumpkit.GatewayIntentMessageContent,
	}),
)
```

### Common Intents

Here are the most frequently used intents:

* `GatewayIntentGuilds`: Server joins, leaves, channel creation, and role updates.
* `GatewayIntentGuildMessages`: Receiving message notifications in server channels.
* `GatewayIntentGuildMessageReactions`: Reaction adds and removals on messages.
* `GatewayIntentDirectMessages`: Receiving direct messages sent directly to your bot.

### Privileged Intents

Discord restricts three intents. You must turn these on manually in the Discord Developer Portal under the Bot tab:

* `GatewayIntentMessageContent`: Required to read the actual text content and attachments in messages sent by regular users.
* `GatewayIntentGuildMembers`: Required to receive member join, leave, and chunking events.
* `GatewayIntentGuildPresences`: Required to track user online statuses and gaming activities.

## Registering Event Handlers

Wumpkit event handlers are strongly typed. Every handler receives a standard `context.Context` and a typed event struct.

### Ready Event

The `READY` event fires when a shard finishes connecting and Discord sends initial information:

```go
bot.OnReady(func(ctx context.Context, event wumpkit.ReadyEvent) {
	slog.Info("Shard is connected",
		"shard_id", event.ShardID,
		"username", event.User.Username,
		"total_guilds", len(event.Guilds),
	)
})
```

### Message Create Event

The `MESSAGE_CREATE` event triggers when a message is sent in a channel visible to your bot:

```go
bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
	// always ignore other bots to avoid infinite reply loops
	if event.Message.Author.Bot {
		return
	}

	if event.Message.Content == "!help" {
		_, _ = bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
			Content: "Available commands: !help, !ping",
		})
	}
})
```

### Guild Member Add Event

Triggered when a new user joins a server (requires `GatewayIntentGuildMembers`):

```go
bot.OnGuildMemberAdd(func(ctx context.Context, event wumpkit.GuildMemberAddEvent) {
	slog.Info("New member joined",
		"guild_id", event.GuildID,
		"username", event.Member.User.Username,
	)
})
```

### Interactions Event

Triggered when a user runs a slash command, clicks a button, or submits a modal:

```go
bot.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
	switch in := event.Interaction.(type) {
	case *wumpkit.ChatInputCommandInteraction:
		_ = bot.RespondMessage(ctx, in, "Command received!")
	case *wumpkit.ComponentInteraction:
		_ = bot.RespondMessage(ctx, in, "Button clicked!")
	}
})
```

## Monitoring Latency

Each Gateway shard sends periodic heartbeats to Discord to verify the connection is healthy. You can check the current average heartbeat ping across all shards at any time:

```go
latency := bot.ShardManager().AverageLatency()
slog.Info("Gateway ping", "latency", latency)
```
