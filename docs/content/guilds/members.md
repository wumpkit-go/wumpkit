---
title: "Members"
weight: 20
---

# Members

Members represent users inside a specific Discord server. A member carries server-specific state such as their nickname, assigned roles, join date, voice status, and moderation flags.

## Fetching a Member

To look up a member by their user ID:

```go
package main

import (
	"context"
	"log"

	"github.com/wumpkit/wumpkit"
)

func inspectMember(ctx context.Context, client *wumpkit.Client, guildID, userID wumpkit.Snowflake) {
	member, err := client.FetchMember(ctx, guildID, userID)
	if err != nil {
		log.Printf("member not found: %v", err)
		return
	}

	log.Printf("Username: %s", member.User.Username)
	if member.Nick != nil {
		log.Printf("Server Nickname: %s", *member.Nick)
	}
	log.Printf("Joined Server: %s", member.JoinedAt)
	log.Printf("Role Count: %d", len(member.Roles))
}
```

## Searching and Listing Members

wumpkit allows you to list members in bulk or search by username/nickname prefix:

```go
// list up to 100 members starting after a given user ID
members, err := client.ListMembers(ctx, guildID, wumpkit.ListMembersOptions{
	Limit: 100,
})

// search members matching a query prefix
results, err := client.SearchMembers(ctx, guildID, wumpkit.SearchMembersOptions{
	Query: "alice",
	Limit: 10,
})
```

## Changing Nicknames and Properties

Use `ModifyMember` to update a member's server nickname, voice state, or roles:

```go
newNick := "Moderator Alice"

updated, err := client.ModifyMember(ctx, guildID, userID, wumpkit.ModifyMemberOptions{
	Nick:   &newNick,
	Mute:   wumpkit.Ptr(false),
	Deaf:   wumpkit.Ptr(false),
	Reason: "Standardizing staff display names",
})
if err != nil {
	log.Printf("failed to modify member: %v", err)
	return
}

log.Printf("updated nickname: %s", *updated.Nick)
```

## Changing the Bot's Own Nickname

If your bot wants to rename itself in a specific server without needing full member management permissions, use `ModifyCurrentMember`:

```go
botNick := "Helper Bot [Active]"
_, err := client.ModifyCurrentMember(ctx, guildID, wumpkit.ModifyCurrentMemberOptions{
	Nick:   &botNick,
	Reason: "Setting online status tag",
})
```
