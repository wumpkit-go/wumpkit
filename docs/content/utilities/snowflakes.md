---
title: "Snowflakes & IDs"
weight: 30
---

# Snowflakes & IDs

Discord uses 64-bit unsigned integers called **Snowflakes** as unique identifiers for users, channels, guilds, messages, and roles.

Snowflakes are not random numbers. They encode the exact millisecond when the entity was created, along with internal Discord worker and process identifiers.

## Working with Snowflakes

wumpkit represents IDs using the `wumpkit.Snowflake` type:

```go
package main

import (
	"fmt"
	"log"

	"github.com/wumpkit/wumpkit"
)

func main() {
	// parse a snowflake from a string
	id, err := wumpkit.ParseSnowflake("123456789012345678")
	if err != nil {
		log.Fatalf("invalid snowflake: %v", err)
	}

	// format back to string
	str := id.String()
	fmt.Printf("Snowflake string: %s\n", str)

	// check if snowflake is unset (zero)
	if id.UnSet() {
		fmt.Println("ID is zero")
	}
}
```

## Extracting Creation Timestamps

Because snowflakes encode time relative to the Discord epoch (January 1, 2015 UTC), you can determine the exact creation time of any Discord account, message, or server without querying the API:

```go
func printCreationDate(id wumpkit.Snowflake) {
	createdAt := id.Timestamp()
	fmt.Printf("Created at: %s (UTC)\n", createdAt.UTC().Format("2006-01-02 15:04:05"))
}
```

## Shard Routing Calculation

To determine which shard handles a given guild snowflake across a cluster:

```go
totalShards := 16
shardID := guildID.GuildShardID(totalShards)
fmt.Printf("Guild %s belongs to shard #%d\n", guildID, shardID)
```

## Extracting Application ID from a Bot Token

Discord bot tokens encode the application ID in the first segment of the token. You can extract it using `ParseAppIDFromToken`:

```go
appID, err := wumpkit.ParseAppIDFromToken(botToken)
if err != nil {
	log.Fatalf("failed to decode bot token: %v", err)
}

fmt.Printf("Application ID: %s\n", appID)
```
