---
title: "Emoji Reactions"
weight: 50
---

# Emoji Reactions

Reactions let users and bots add emoji responses to existing messages.

## Adding a Reaction

Use `CreateReaction` to add an emoji to any message:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func addThumbsUp(ctx context.Context, client *wumpkit.Client, channelID, messageID wumpkit.Snowflake) {
	// standard Unicode emoji
	_ = client.CreateReaction(ctx, channelID, messageID, "👍")

	// custom server emoji uses the format "name:id"
	_ = client.CreateReaction(ctx, channelID, messageID, "wumpus:123456789012345678")
}
```

## Removing Reactions

wumpkit provides methods to delete reactions individually or in bulk:

```go
// delete your bot's own reaction
_ = client.DeleteOwnReaction(ctx, channelID, messageID, "👍")

// delete another user's reaction (requires Manage Messages permission)
_ = client.DeleteUserReaction(ctx, channelID, messageID, targetUserID, "👍")

// delete all reactions of a single emoji type
_ = client.DeleteAllReactionsForEmoji(ctx, channelID, messageID, "👍")

// remove every reaction from the message
_ = client.DeleteAllReactions(ctx, channelID, messageID)
```

## Inspecting Who Reacted

Use `FetchReactions` to get the list of users who reacted with an emoji:

```go
users, err := client.FetchReactions(ctx, channelID, messageID, "⭐", wumpkit.FetchReactionsOptions{
	Limit: 100,
})
if err != nil {
	log.Printf("failed to fetch reactions: %v", err)
	return
}

for _, u := range users {
	log.Printf("%s reacted with a star!", u.Username)
}
```

## Listening to Reaction Events

To respond when a user reacts to a message, register a handler for `EventMessageReactionAdd`:

```go
client.On(wumpkit.EventMessageReactionAdd, func(ctx context.Context, event *wumpkit.MessageReactionAddEvent) {
	// ignore bot reactions if needed
	if event.UserID == client.BotUser().ID {
		return
	}

	log.Printf("User %s added reaction %s on message %s",
		event.UserID,
		event.Emoji.Name,
		event.MessageID,
	)
})
```
