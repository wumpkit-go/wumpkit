---
title: "Cache-Aside Pattern"
weight: 3
---

# Cache-Aside Pattern

Wumpkit implements the **Cache-Aside** design pattern on top of its in-memory storage. This ensures you never have to manually write boilerplate code to check your cache, make a REST fallback call on a cache miss, and save the result.

## The Three Getter Types Explained

| Getter Family | Example Method | Network Behavior | Returns |
| :--- | :--- | :--- | :--- |
| **Cache-Aside** | `client.User(ctx, id)` | Checks memory first. Calls REST only if missing. | `(User, error)` |
| **Direct Cache** | `client.GetUser(ctx, id)` | Memory only. Never makes network requests. | `(User, bool)` |
| **Direct REST** | `client.FetchUser(ctx, id)` | Always makes an HTTP request to Discord. | `(User, error)` |

## 1. Cache-Aside Methods (Recommended Default)

Methods without a prefix (such as `client.User`, `client.Guild`, `client.Channel`, `client.Member`, `client.Role`) implement cache-aside.

When you call `client.User(ctx, userID)`:
1. It queries the local in-memory cache for `userID`.
2. If the user is present in memory, it returns the cached struct immediately (0 milliseconds network latency).
3. If the user is missing from memory (a cache miss), it automatically sends an HTTP request to Discord's `/users/{id}` endpoint.
4. It stores the freshly fetched user into the cache for future calls and returns it to your caller.

```go
package main

import (
	"context"
	"fmt"

	"github.com/wumpkit/wumpkit"
)

func getUserInfo(ctx context.Context, client *wumpkit.Client, userID wumpkit.Snowflake) error {
	// transparently uses memory if present, or fetches via rest if absent
	user, err := client.User(ctx, userID)
	if err != nil {
		return err
	}

	fmt.Printf("User: %s (Bot: %t)\n", user.Username, user.Bot)
	return nil
}
```

## 2. When to Use Direct Cache Lookups (`Get*`)

Use `Get*` methods when:
- You are handling hot code paths (e.g. processing 10,000 chat messages a second) and cannot afford any HTTP latency or rate limit usage.
- You only want to know if an entity is currently loaded in memory.

```go
// checks cache only, zero network traffic
if member, ok := client.GetMember(ctx, guildID, userID); ok {
	// member is cached
	_ = member
} else {
	// member is not in memory; handle accordingly without blocking
}
```

## 3. When to Use Direct REST Calls (`Fetch*`)

Use `Fetch*` methods when:
- You need guaranteed fresh data (for example, fetching member counts on a guild before running a billing or stats report).
- You are updating or modifying properties through Discord REST endpoints.

```go
// always queries discord api to get exact counts
guild, err := client.FetchGuild(ctx, guildID, wumpkit.FetchGuildOptions{
	WithCounts: true,
})
```
