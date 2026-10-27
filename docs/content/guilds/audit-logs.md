---
title: "Audit Logs"
weight: 50
---

# Audit Logs

The server audit log records administrative and moderation events, such as kicked members, created roles, modified channels, and purged messages.

wumpkit lets you fetch and filter audit log entries to build custom security monitors, logging bots, or moderation dashboards.

## Fetching Audit Logs

Use `FetchGuildAuditLog` on your client:

```go
package main

import (
	"context"
	"log"

	"github.com/wumpkit/wumpkit"
)

func printRecentAuditEvents(ctx context.Context, client *wumpkit.Client, guildID wumpkit.Snowflake) {
	auditLog, err := client.FetchGuildAuditLog(ctx, guildID, wumpkit.FetchAuditLogOptions{
		Limit: 20,
	})
	if err != nil {
		log.Printf("failed to fetch audit log: %v", err)
		return
	}

	for _, entry := range auditLog.AuditLogEntries {
		log.Printf("Action: %v | Moderator: %s | Target: %s | Reason: %s",
			entry.ActionType,
			entry.UserID,
			entry.TargetID,
			entry.Reason,
		)
	}
}
```

## Filtering by Action Type or Moderator

You can narrow your queries to specific event types or specific users:

```go
// fetch only recent member kicks
kickLogs, err := client.FetchGuildAuditLog(ctx, guildID, wumpkit.FetchAuditLogOptions{
	ActionType: wumpkit.AuditLogEventMemberKick,
	Limit:      10,
})

// fetch actions performed by a specific moderator
moderatorLogs, err := client.FetchGuildAuditLog(ctx, guildID, wumpkit.FetchAuditLogOptions{
	UserID: targetModeratorID,
	Limit:  25,
})
```

## Common Audit Log Events

* `wumpkit.AuditLogEventMemberKick`: A member was kicked.
* `wumpkit.AuditLogEventMemberBanAdd`: A member was banned.
* `wumpkit.AuditLogEventMemberBanRemove`: A ban was lifted.
* `wumpkit.AuditLogEventMemberUpdate`: A member's nickname, timeout, or roles changed.
* `wumpkit.AuditLogEventChannelCreate`: A channel was created.
* `wumpkit.AuditLogEventChannelDelete`: A channel was deleted.
* `wumpkit.AuditLogEventRoleCreate`: A role was created.
* `wumpkit.AuditLogEventRoleDelete`: A role was deleted.
* `wumpkit.AuditLogEventMessageDelete`: Messages were deleted by a moderator.

## Inspecting Changes

Entries often include a list of changes detailing what was modified before and after the action:

```go
for _, change := range entry.Changes {
	log.Printf("Key changed: %s | Old: %v | New: %v", change.Key, change.OldValue, change.NewValue)
}
```
