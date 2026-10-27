---
title: "How Caching Works"
weight: 1
---

# How Caching Works

When building Discord bots, making HTTP requests for every command or event is slow and will quickly exhaust your Discord rate limits. Wumpkit solves this by maintaining an in-memory representation of server state.

## Automatic Gateway Synchronization

The cache is not static. When your bot connects to the Gateway, Wumpkit automatically populates and updates the in-memory cache as Discord events stream in:

1. **On Connect (`READY` & `GUILD_CREATE`)**: Wumpkit registers initial guild structures, roles, channels, and member information.
2. **On Changes (`CHANNEL_UPDATE`, `GUILD_ROLE_UPDATE`)**: Updated properties are applied directly to existing cached objects.
3. **On Member Joins/Leaves (`GUILD_MEMBER_ADD`, `GUILD_MEMBER_REMOVE`)**: The internal member collection adds or removes the corresponding user.
4. **On Deletions (`CHANNEL_DELETE`, `GUILD_ROLE_DELETE`)**: Stale entries are immediately purged from memory.

## Thread Safety & Sharded Maps

All cached data in Wumpkit is stored using internal concurrent sharded maps (`internal/shardedmap`). Instead of relying on a single global `sync.RWMutex` that creates contention across shards, keys are distributed across multiple stripes. This allows concurrent reads and writes across different shards without blocking each other.

## Direct In-Memory Lookups

If you need instant access to an object already stored in memory, use the `Get*` family of methods on your client. These methods never perform network calls:

```go
package main

import (
	"context"
	"fmt"

	"github.com/wumpkit/wumpkit"
)

func checkCachedData(ctx context.Context, client *wumpkit.Client, guildID, channelID, userID wumpkit.Snowflake) {
	// look up guild from memory
	if guild, ok := client.GetGuild(ctx, guildID); ok {
		fmt.Printf("Guild Name: %s (Owner: %s)\n", guild.Name, guild.OwnerID)
	}

	// look up channel from memory
	if channel, ok := client.GetChannel(ctx, channelID); ok {
		fmt.Printf("Channel Name: #%s\n", channel.Name)
	}

	// look up member in a specific guild
	if member, ok := client.GetMember(ctx, guildID, userID); ok {
		fmt.Printf("Member Nickname: %s\n", member.Nick)
	}
}
```
