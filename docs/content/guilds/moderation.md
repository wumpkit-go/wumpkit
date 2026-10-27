---
title: "Moderation & Bans"
weight: 30
---

# Moderation & Bans

wumpkit provides full support for Discord server moderation, including kicking, banning, bulk banning, and member timeouts.

Every moderation action supports an audit log reason string so your actions are properly documented in the server audit log.

## Kicking a Member

Kicking removes a user from the server, but allows them to rejoin if they have a valid invite:

```go
package main

import (
	"context"
	"log"

	"github.com/wumpkit/wumpkit"
)

func kickTroublemaker(ctx context.Context, client *wumpkit.Client, guildID, userID wumpkit.Snowflake) {
	err := client.KickMember(ctx, guildID, userID, wumpkit.KickMemberOptions{
		Reason: "Repeated spam in general chat",
	})
	if err != nil {
		log.Printf("failed to kick member: %v", err)
		return
	}

	log.Printf("kicked user %s successfully", userID)
}
```

## Banning a Member

Banning permanently removes a user from the server and prevents rejoining. You can also purge recent messages by specifying `DeleteMessageSeconds` (up to 7 days, or 604800 seconds):

```go
func banMaliciousUser(ctx context.Context, client *wumpkit.Client, guildID, userID wumpkit.Snowflake) {
	err := client.BanMember(ctx, guildID, userID, wumpkit.BanMemberOptions{
		DeleteMessageSeconds: 86400, // delete past 24 hours of messages
		Reason:               "Account compromised / raid participant",
	})
	if err != nil {
		log.Printf("failed to ban user: %v", err)
		return
	}

	log.Printf("banned user %s", userID)
}
```

## Unbanning a Member

Use `UnbanMember` to lift a ban:

```go
err := client.UnbanMember(ctx, guildID, userID, wumpkit.UnbanMemberOptions{
	Reason: "Ban appeal accepted",
})
```

## Bulk Banning Users

When handling server raids or automated spam attacks, you can ban up to 200 users in a single HTTP request using `BulkBanMembers`:

```go
spammers := []wumpkit.Snowflake{123456789, 234567890, 345678901}

result, err := client.BulkBanMembers(ctx, guildID, wumpkit.BulkBanMembersOptions{
	UserIDs:              spammers,
	DeleteMessageSeconds: 3600, // delete past 1 hour of messages
	Reason:               "Raid prevention bulk ban",
})
if err != nil {
	log.Printf("bulk ban failed: %v", err)
	return
}

log.Printf("banned %d users, failed on %d users", len(result.BannedUsers), len(result.FailedUsers))
```

## Timing Out a Member

Timeouts (communication disable) prevent a member from chatting, sending reactions, or speaking in voice channels for a temporary period (up to 28 days in the future):

```go
func timeoutMember(ctx context.Context, client *wumpkit.Client, guildID, userID wumpkit.Snowflake, duration time.Duration) {
	until := time.Now().Add(duration)

	_, err := client.ModifyMember(ctx, guildID, userID, wumpkit.ModifyMemberOptions{
		CommunicationDisabledUntil: &until,
		Reason:                     "Excessive caps and rule violation",
	})
	if err != nil {
		log.Printf("failed to timeout member: %v", err)
		return
	}

	log.Printf("user %s timed out until %s", userID, until.Format(time.RFC3339))
}
```

To remove a timeout immediately, set `CommunicationDisabledUntil` to `nil` or a time in the past:

```go
past := time.Now().Add(-1 * time.Minute)
_, _ = client.ModifyMember(ctx, guildID, userID, wumpkit.ModifyMemberOptions{
	CommunicationDisabledUntil: &past,
	Reason:                     "Timeout lifted early by moderator",
})
```
