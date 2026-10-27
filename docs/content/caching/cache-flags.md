---
title: "Controlling Cached Entities"
weight: 2
---

# Controlling Cached Entities

By default, Wumpkit caches guilds, channels, roles, members, users, and voice states. However, as your bot joins hundreds or thousands of servers, keeping every entity in memory can consume significant RAM.

Wumpkit allows you to explicitly control which Discord entities are stored in memory using `CacheFlags`.

## How to Configure Cache Flags

Pass a configured `InMemoryCacheManager` with your desired flags using `WithCacheManager`:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func main() {
	// example: only cache guilds, channels, and roles to save memory
	myCache := wumpkit.NewInMemoryCacheManager(
		wumpkit.CacheFlagGuilds|
			wumpkit.CacheFlagChannels|
			wumpkit.CacheFlagRoles,
		wumpkit.DefaultCacheConfig(),
	)

	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken("my_bot_token"),
		wumpkit.WithCacheManager(myCache),
	)

	_ = bot
}
```

## Available Cache Flags

| Flag Constant | Stored Data | Memory Impact |
| :--- | :--- | :--- |
| `CacheFlagGuilds` | Server names, owners, features, and settings | Low |
| `CacheFlagChannels` | Text, voice, category, and thread channel states | Low |
| `CacheFlagRoles` | Roles, colors, positions, and permission bitmasks | Low |
| `CacheFlagUsers` | User profiles, usernames, and avatars | Moderate |
| `CacheFlagMembers` | Server member profiles, nicknames, and role lists | High (scales with member count) |
| `CacheFlagVoiceStates` | Active voice channel states (muted, deafened, channel) | Low to Moderate |
| `CacheFlagEmojis` | Custom server emojis | Low |
| `CacheFlagStickers` | Guild stickers | Low |
| `CacheFlagMessages` | In-memory message history (disabled by default) | High |
| `CacheFlagGuildScheduledEvents` | Server scheduled events | Low |
| `CacheFlagAutoModerationRules` | Server auto moderation rule definitions | Low |
| `CacheFlagsAll` | Enables caching for every supported entity | Highest |
| `CacheFlagsNone` | Disables in-memory caching completely | Lowest (zero cache overhead) |

## Common Configuration Presets

### 1. Minimal Memory Preset (Zero Cache)
For serverless bots or microservices that handle webhooks or stateless commands, you can disable caching entirely to run with minimal memory footprints:

```go
wumpkit.WithCacheManager(wumpkit.NewInMemoryCacheManager(
	wumpkit.CacheFlagsNone,
	wumpkit.DefaultCacheConfig(),
))
```

### 2. Standard Bot Preset (Fast Commands)
Caches server structure, channels, and roles, but avoids caching every individual member:

```go
wumpkit.WithCacheManager(wumpkit.NewInMemoryCacheManager(
	wumpkit.CacheFlagGuilds|
		wumpkit.CacheFlagChannels|
		wumpkit.CacheFlagRoles,
	wumpkit.DefaultCacheConfig(),
))
```

### 3. Moderation Bot Preset (Full Member Data)
Caches member join dates, assigned roles, and voice states to provide instant moderation checks:

```go
wumpkit.WithCacheManager(wumpkit.NewInMemoryCacheManager(
	wumpkit.CacheFlagGuilds|
		wumpkit.CacheFlagChannels|
		wumpkit.CacheFlagRoles|
		wumpkit.CacheFlagMembers|
		wumpkit.CacheFlagVoiceStates,
	wumpkit.DefaultCacheConfig(),
))
```
