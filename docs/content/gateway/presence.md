---
title: "Bot Presence & Status"
weight: 4
---

# Bot Presence & Status

You can customize your bot's online status (online, idle, do not disturb, invisible) and set rich activities (playing a game, watching a video, listening to music, or a custom text status).

## Setting Initial Presence

When connecting to the Gateway, you can define the bot's initial presence in your `ShardManagerConfig`:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func main() {
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken("my_bot_token"),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Presence: &wumpkit.GatewayPresenceUpdate{
				Status: wumpkit.UserStatusOnline,
				Activities: []wumpkit.Activity{
					{
						Name: "with Go code",
						Type: wumpkit.ActivityTypeGame,
					},
				},
			},
		}),
	)

	_ = bot
}
```

## Updating Presence at Runtime

You can change your bot's activity or status at any time during execution by calling `UpdatePresence`. This sends an Opcode 3 Presence Update across all active shards:

```go
func setWatchingStatus(client *wumpkit.Client, serverCount int) error {
	return client.UpdatePresence(wumpkit.GatewayPresenceUpdate{
		Status: wumpkit.UserStatusOnline,
		Activities: []wumpkit.Activity{
			{
				Name: fmt.Sprintf("%d servers", serverCount),
				Type: wumpkit.ActivityTypeWatching,
			},
		},
	})
}
```

## Status Types

| Constant | Description | Visual Appearance |
| :--- | :--- | :--- |
| `UserStatusOnline` | Active and available | Green dot |
| `UserStatusIdle` | Away or idle | Yellow crescent |
| `UserStatusDoNotDisturb` | Busy / DND | Red circle with minus sign |
| `UserStatusInvisible` | Appears offline | Grey circle |

## Activity Types

| Constant | Example Text |
| :--- | :--- |
| `ActivityTypeGame` | Playing **Minecraft** |
| `ActivityTypeStreaming` | Streaming **Twitch** |
| `ActivityTypeListening` | Listening to **Spotify** |
| `ActivityTypeWatching` | Watching **YouTube** |
| `ActivityTypeCompeting` | Competing in **Tournament** |
| `ActivityTypeCustom` | Custom text status |
