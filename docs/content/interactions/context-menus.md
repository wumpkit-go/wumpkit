---
title: "Context Menus"
weight: 50
---

# Context Menus

Context menu commands appear when users right-click (or tap and hold on mobile) on a user or message in Discord.

Discord supports two types of context menus:
* **User Commands**: Run by right-clicking a user (under Apps).
* **Message Commands**: Run by right-clicking a message (under Apps).

## Defining Context Menus

wumpkit provides `NewUserCommandBuilder` and `NewMessageCommandBuilder` to build context menus:

```go
package main

import (
	"github.com/wumpkit/wumpkit"
)

// user context menu command
userProfileCmd := wumpkit.NewUserCommandBuilder("User Profile").Build()

// message context menu command
reportMessageCmd := wumpkit.NewMessageCommandBuilder("Report Message").Build()
```

## Registering Context Menus

Register your context menu commands alongside slash commands using `BulkOverwriteGlobalApplicationCommands`:

```go
commands := []wumpkit.ApplicationCommand{
	pingCmd,
	userProfileCmd,
	reportMessageCmd,
}

_, err := client.BulkOverwriteGlobalApplicationCommands(ctx, appID, commands)
```

## Handling User Commands

When someone runs a user context menu command, Discord sends an interaction of type `UserContextMenuCommandInteraction`. Cast the event to read the target user:

```go
client.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
	switch in := event.Interaction.(type) {
	case *wumpkit.UserContextMenuCommandInteraction:
		if in.Data.Name == "User Profile" {
			targetID := in.Data.TargetID
			targetUser := in.Data.Resolved.Users[targetID]

			msg := fmt.Sprintf("Selected User: %s (ID: %s)", targetUser.Username, targetID)
			_ = client.RespondMessage(ctx, in, msg, wumpkit.RespondMessageOptions{
				Ephemeral: true,
			})
		}
	}
})
```

## Handling Message Commands

When someone runs a message context menu command, cast the event to `*wumpkit.MessageContextMenuCommandInteraction`:

```go
case *wumpkit.MessageContextMenuCommandInteraction:
	if in.Data.Name == "Report Message" {
		targetID := in.Data.TargetID
		targetMsg := in.Data.Resolved.Messages[targetID]

		_ = client.RespondMessage(ctx, in, "Thank you. This message has been flagged for moderator review.", wumpkit.RespondMessageOptions{
			Ephemeral: true,
		})

		log.Printf("Reported message content: %s (Author: %s)", targetMsg.Content, targetMsg.Author.Username)
	}
```
