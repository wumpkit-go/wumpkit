---
title: "Gateway Intents"
weight: 2
---

# Gateway Intents

Gateway Intents tell Discord which real-time events your bot wants to receive. Discord filters out all unrequested event traffic before it reaches your server, saving CPU cycles, network bandwidth, and memory.

## Combining Intents

Combine intent flags using the bitwise OR operator (`|`):

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
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
		}),
	)

	_ = bot
}
```

## Privileged Intents

Discord classifies certain sensitive data streams as **Privileged Intents**. You must explicitly enable them in the Discord Developer Portal under your application's **Bot** tab before your bot can connect with them:

### 1. Message Content Intent (`GatewayIntentMessageContent`)
Required if your bot reads the text body, embeds, or attachments of messages in text channels (for example, prefix commands like `!ping`). Without this intent, the `Content` field of incoming messages will be empty for regular messages.

*Note: Slash commands and button interactions do not require this intent.*

### 2. Guild Members Intent (`GatewayIntentGuildMembers`)
Required if your bot tracks member join and leave events (`OnGuildMemberAdd`, `OnGuildMemberRemove`), or updates roles on users joining.

### 3. Presence Intent (`GatewayIntentGuildPresences`)
Required if your bot monitors user online status, game activities, or custom statuses. Most moderation and utility bots do not need this intent.

## Common Intent Combinations

### Slash Command & Interaction Bots
Bots that only use slash commands, buttons, select menus, and modals:

```go
Intents: wumpkit.GatewayIntentGuilds,
```

### General Chat & Prefix Command Bots
Bots that read messages in server channels and respond to text commands:

```go
Intents: wumpkit.GatewayIntentGuilds |
	wumpkit.GatewayIntentGuildMessages |
	wumpkit.GatewayIntentMessageContent,
```

### Moderation & Member Tracking Bots
Bots that manage roles, welcome new members, and log bans:

```go
Intents: wumpkit.GatewayIntentGuilds |
	wumpkit.GatewayIntentGuildMembers |
	wumpkit.GatewayIntentGuildBans |
	wumpkit.GatewayIntentGuildMessages,
```
