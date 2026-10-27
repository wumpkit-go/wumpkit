---
title: "Channels"
weight: 10
---

# Channels

Channels are the spaces where guild conversations and media sharing happen.

## Creating Channels

Use `CreateChannel` on your client to create a text channel, voice channel, category, or forum:

```go
package main

import (
	"context"
	"log"

	"github.com/wumpkit/wumpkit"
)

func createServerChannels(ctx context.Context, client *wumpkit.Client, guildID wumpkit.Snowflake) {
	// create a text channel
	textChannel, err := client.CreateChannel(ctx, guildID, wumpkit.CreateChannelOptions{
		Name:  "announcements",
		Type:  wumpkit.ChannelTypeGuildText,
		Topic: "Official server news and updates",
	})
	if err != nil {
		log.Printf("failed to create text channel: %v", err)
		return
	}

	log.Printf("created channel %s (ID: %s)", textChannel.Name(), textChannel.ID())
}
```

## Common Channel Types

* `wumpkit.ChannelTypeGuildText`: Standard text channel.
* `wumpkit.ChannelTypeGuildVoice`: Voice chat channel.
* `wumpkit.ChannelTypeGuildCategory`: Organizational category folder.
* `wumpkit.ChannelTypeGuildAnnouncement`: News channel that other servers can follow.
* `wumpkit.ChannelTypeGuildForum`: Forum channel organizing discussions into dedicated threads.

## Setting Permission Overwrites

You can lock down channels or hide them from specific roles by configuring permission overwrites upon creation:

```go
func createPrivateStaffChannel(ctx context.Context, client *wumpkit.Client, guildID, staffRoleID, everyoneRoleID wumpkit.Snowflake) {
	_, err := client.CreateChannel(ctx, guildID, wumpkit.CreateChannelOptions{
		Name: "staff-only",
		Type: wumpkit.ChannelTypeGuildText,
		PermissionOverwrites: []wumpkit.PermissionOverwrite{
			{
				ID:   everyoneRoleID,
				Type: wumpkit.PermissionOverwriteTypeRole,
				Deny: wumpkit.PermissionViewChannel,
			},
			{
				ID:    staffRoleID,
				Type:  wumpkit.PermissionOverwriteTypeRole,
				Allow: wumpkit.PermissionViewChannel | wumpkit.PermissionSendMessages,
			},
		},
	})
	if err != nil {
		log.Printf("failed to create staff channel: %v", err)
	}
}
```

## Modifying Existing Channels

Use `ModifyGuildChannel` to change a channel's name, topic, category parent, or slowmode:

```go
slowmode := 10 // 10 seconds between user messages

_, err := client.ModifyGuildChannel(ctx, channelID, wumpkit.ModifyGuildChannelOptions{
	Name:                 "general-chat",
	Topic:                "Hang out and talk about code",
	RateLimitPerUser:     &slowmode,
	Reason:               "Updated channel rules",
})
```

## Listing Guild Channels

Fetch every channel in a server using `FetchGuildChannels`:

```go
channels, err := client.FetchGuildChannels(ctx, guildID)
if err != nil {
	log.Printf("failed to list channels: %v", err)
	return
}

for _, ch := range channels {
	log.Printf("#%s (Type: %v)", ch.Name(), ch.Type())
}
```
