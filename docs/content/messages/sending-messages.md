---
title: "Sending Messages"
weight: 1
---

# Sending Messages

To send messages in a guild text channel, thread, or direct message channel, use the `CreateMessage` method on your client.

## Basic Text Messages

Pass the channel ID and a `CreateMessageOptions` struct with your content:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func sendAnnouncement(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) error {
	msg, err := client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Content: "Server maintenance scheduled for tonight at 02:00 UTC.",
	})
	if err != nil {
		return err
	}

	_ = msg.ID // newly created message ID
	return nil
}
```

## Replying to a Message

To reply to a specific message with an inline reference (the small quote indicator above the response), set `MessageReference`:

```go
func replyToUser(ctx context.Context, client *wumpkit.Client, channelID, sourceMessageID wumpkit.Snowflake) error {
	_, err := client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Content: "Got your report! We are investigating.",
		MessageReference: &wumpkit.MessageReference{
			MessageID: sourceMessageID,
		},
	})
	return err
}
```

## Controlling Mentions

By default, mentioning a user ID or role ID in message content pings that person or role. You can explicitly restrict or allow which mentions trigger notifications using `AllowedMentions`:

```go
_, err := client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
	Content: "Hello <@123456789>, this message will not trigger a notification ping.",
	AllowedMentions: &wumpkit.AllowedMentions{
		// explicitly empty: do not ping any users, roles, or @everyone
		Parse: []wumpkit.AllowedMentionType{},
		// set to false to avoid pinging the author when replying
		RepliedUser: false,
	},
})
```

## Silent Messages

If you want to send a message without triggering push notifications or sound alerts for any channel members, add the `MessageFlagSuppressNotifications` flag:

```go
_, err := client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
	Content: "Quiet update: system reboot completed successfully.",
	Flags:   wumpkit.MessageFlagSuppressNotifications,
})
```

## Editing and Deleting Messages

Edit an existing message using `EditMessage`:

```go
newContent := "Updated announcement text."
editedMsg, err := client.EditMessage(ctx, channelID, messageID, wumpkit.EditMessageOptions{
	Content: &newContent,
})
```

Delete a message using `DeleteMessage`:

```go
err := client.DeleteMessage(ctx, channelID, messageID)
```
