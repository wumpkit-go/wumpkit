---
title: "Slash Commands"
weight: 10
---

# Slash Commands

Slash commands let users run commands by typing `/` in Discord. wumpkit provides a fluent command builder and handles interaction responses with simple helper methods.

## Defining Commands

Use `wumpkit.NewSlashCommandBuilder` to construct slash commands:

```go
package main

import (
	"github.com/wumpkit/wumpkit"
)

// basic ping command
pingCmd := wumpkit.NewSlashCommandBuilder("ping", "Check bot latency").Build()

// command with options and choices
infoCmd := wumpkit.NewSlashCommandBuilder("info", "Look up server or user information").
	AddUserOption("target", "Target user to inspect", false).
	AddStringOption("detail", "Detail level", false,
		wumpkit.NewStringChoice("Standard", "standard"),
		wumpkit.NewStringChoice("Extended", "extended"),
	).
	Build()
```

You can add options for strings, integers, numbers, booleans, users, channels, roles, mentionables, and attachments.

## Registering Commands

Discord requires you to register application commands over REST before they appear in the client. The cleanest pattern is using bulk overwrite in your `OnReady` event:

```go
bot.OnReady(func(ctx context.Context, event wumpkit.ReadyEvent) {
	commands := []wumpkit.ApplicationCommand{pingCmd, infoCmd}

	// register global commands across all servers
	appID := wumpkit.Snowflake(123456789012345678)
	synced, err := bot.BulkOverwriteGlobalApplicationCommands(ctx, appID, commands)
	if err != nil {
		log.Printf("failed syncing commands: %v", err)
		return
	}

	log.Printf("synced %d commands successfully", len(synced))
})
```

Tip: During development, you can register commands to a single server using `bot.BulkOverwriteGuildApplicationCommands()`. Guild commands update instantly, whereas global commands can take a few minutes to roll out across all Discord regions.

## Handling Command Invocations

When a user runs a slash command, Discord fires an `EventInteractionCreate` event. You handle this with `bot.OnInteractionCreate`:

```go
bot.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
	// check if this interaction is a slash command
	in, ok := event.Interaction.(*wumpkit.ChatInputCommandInteraction)
	if !ok {
		return
	}

	switch in.Data.Name {
	case "ping":
		latency := bot.ShardManager().AverageLatency()
		_ = bot.RespondMessage(ctx, in, fmt.Sprintf("Pong! Latency is %v", latency))

	case "info":
		// read option values from in.Data.Options
	}
})
```

## Immediate and Ephemeral Responses

Discord gives bots 3 seconds to respond to an interaction. If your response is ready right away, use `bot.RespondMessage`:

```go
// public response visible to everyone in the channel
_ = bot.RespondMessage(ctx, in, "Operation completed successfully.")

// ephemeral response visible only to the person who ran the command
_ = bot.RespondMessage(ctx, in, "This message is private to you.", wumpkit.RespondMessageOptions{
	Ephemeral: true,
})
```

## Deferring Long-Running Tasks

If your command needs to perform work that takes longer than 3 seconds (such as querying a database or an external API), defer the response first. This shows a "Bot is thinking..." loading state in Discord:

```go
case "info":
	// 1. acknowledge immediately and show the loading indicator
	if err := bot.DeferResponse(ctx, in, false); err != nil {
		log.Printf("failed deferring response: %v", err)
		return
	}

	// 2. perform background work
	time.Sleep(1 * time.Second)

	// 3. update the original response with your final content
	result := "Here are your query results!"
	_, err := bot.EditOriginalInteractionResponse(ctx, appID, in.GetToken(), wumpkit.EditFollowupMessageOptions{
		Content: &result,
	})
	if err != nil {
		log.Printf("failed updating response: %v", err)
	}
```

## Handling Autocomplete

If an option has autocomplete enabled, Discord sends an interaction when the user types in the option input. Respond using `bot.RespondAutocomplete`:

```go
if auto, ok := event.Interaction.(*wumpkit.AutocompleteInteraction); ok {
	choices := []wumpkit.ApplicationCommandOptionChoice{
		wumpkit.NewStringChoice("First Suggestion", "choice_1"),
		wumpkit.NewStringChoice("Second Suggestion", "choice_2"),
	}
	_ = bot.RespondAutocomplete(ctx, auto, choices)
}
```
