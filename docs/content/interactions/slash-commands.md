---
title: "Slash Commands"
weight: 1
---

# Slash Commands

Slash commands let users run commands by typing `/` in Discord. Wumpkit provides a fluent command builder and handles interaction responses.

## 1. Defining Slash Commands

Use `wumpkit.NewSlashCommandBuilder()` to define your commands:

```go
package main

import (
	"github.com/wumpkit/wumpkit"
)

// simple ping command
pingCmd := wumpkit.NewSlashCommandBuilder("ping", "Check bot latency").Build()

// info command with user option and string choices
infoCmd := wumpkit.NewSlashCommandBuilder("info", "Look up server or user information").
	AddUserOption("target", "Target user to inspect", false).
	AddStringOption("detail", "Detail level", false,
		wumpkit.NewStringChoice("Standard", "standard"),
		wumpkit.NewStringChoice("Extended", "extended"),
	).
	Build()
```

You can add options for strings, integers, floats, booleans, users, channels, roles, mentionables, and attachments.

## 2. Registering Commands with Discord

Discord requires you to register your commands via REST before they appear in the client. The cleanest way is using bulk overwrite inside your `OnReady` event:

```go
bot.OnReady(func(ctx context.Context, event wumpkit.ReadyEvent) {
	commands := []wumpkit.ApplicationCommand{pingCmd, infoCmd}

	// sync global commands across all guilds
	appID := wumpkit.Snowflake(123456789012345678)
	synced, err := bot.BulkOverwriteGlobalApplicationCommands(ctx, appID, commands)
	if err != nil {
		slog.Error("failed syncing commands", "error", err)
		return
	}

	slog.Info("commands synced successfully", "count", len(synced))
})
```

Tip: During development, you can register commands to a single guild using `bot.BulkOverwriteGuildApplicationCommands()`. Guild commands update immediately, whereas global commands can take a few minutes to cache across Discord servers.

## 3. Handling Command Invocations

When a user runs a slash command, Discord fires an `INTERACTION_CREATE` event. You handle this with `bot.OnInteractionCreate`:

```go
bot.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
	// check if this interaction is a chat input (slash) command
	in, ok := event.Interaction.(*wumpkit.ChatInputCommandInteraction)
	if !ok {
		return
	}

	switch in.Data.Name {
	case "ping":
		latency := bot.ShardManager().AverageLatency()
		_ = bot.RespondMessage(ctx, in, fmt.Sprintf("Pong! Latency is %v", latency))

	case "info":
		// handle info command
	}
})
```

## 4. Immediate and Ephemeral Responses

Discord gives bots 3 seconds to respond to an interaction. If your response is ready right away, use `bot.RespondMessage`:

```go
// public response visible to everyone in the channel
_ = bot.RespondMessage(ctx, in, "Operation successful.")

// ephemeral response visible only to the person who ran the command
_ = bot.RespondMessage(ctx, in, "This message is private.", wumpkit.RespondMessageOptions{
	Ephemeral: true,
})
```

## 5. Deferring Long-Running Tasks

If your bot needs to perform work that takes longer than 3 seconds (such as querying a database or an external API), you must defer the response first. This shows a "Bot is thinking..." loading state in Discord:

```go
case "info":
	// 1. acknowledge the command immediately and show loading indicator
	if err := bot.DeferResponse(ctx, in, false); err != nil {
		slog.Error("failed deferring response", "error", err)
		return
	}

	// 2. perform background work
	time.Sleep(500 * time.Millisecond)

	// 3. update the original response with your final content
	result := "Here are the query results!"
	_, err := bot.EditOriginalInteractionResponse(ctx, appID, in.GetToken(), wumpkit.EditFollowupMessageOptions{
		Content: &result,
	})
	if err != nil {
		slog.Error("failed updating response", "error", err)
	}
```
