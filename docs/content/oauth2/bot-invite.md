---
title: "Bot Invites & Permissions"
weight: 20
---

# Bot Invites & Permissions

To add your bot to a Discord server, users must follow an OAuth2 authorization link that grants the bot access and specifies its server permissions.

wumpkit provides `GenerateBotInviteURL` to build clean, compliant invite links.

## Generating an Invite Link

Call `wumpkit.GenerateBotInviteURL` with your application ID and the permissions your bot requires:

```go
package main

import (
	"fmt"

	"github.com/wumpkit/wumpkit"
)

func createInviteURL(appID wumpkit.Snowflake) string {
	// combine the permissions your bot needs
	perms := wumpkit.PermissionSendMessages |
		wumpkit.PermissionEmbedLinks |
		wumpkit.PermissionAttachFiles |
		wumpkit.PermissionReadMessageHistory

	// generate invite link with bot and slash command scopes
	inviteURL := wumpkit.GenerateBotInviteURL(
		appID,
		perms,
		wumpkit.OAuth2ScopeBot,
		wumpkit.OAuth2ScopeApplicationsCommands,
	)

	return inviteURL
}
```

## Adding Guild or User Installation Scopes

If your application supports user-installed apps (apps that can be invoked in direct messages or any server without requiring the server admin to install the bot), configure installation scopes:

```go
opts := wumpkit.OAuth2AuthorizeURLOptions{
	ClientID: appID,
	Scopes: []wumpkit.OAuth2Scope{
		wumpkit.OAuth2ScopeBot,
		wumpkit.OAuth2ScopeApplicationsCommands,
	},
	Permissions: wumpkit.PermissionSendMessages | wumpkit.PermissionViewChannel,
	IntegrationType: wumpkit.ApplicationIntegrationTypeGuildInstall,
}

inviteURL := wumpkit.GenerateOAuth2AuthorizeURL(opts)
```

## Permission Best Practices

1. **Principle of Least Privilege**: Only request the permissions your bot actually uses. Asking for `PermissionAdministrator` causes users to hesitate when inviting your bot.
2. **Channel Overwrites**: Even if a bot has a permission globally, channel-specific permission overwrites can deny access. Check permissions in the channel before sending messages.
