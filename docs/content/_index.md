---
title: "Wumpkit"
description: "A lightweight and high-performance Discord API library for Go"
---

# Wumpkit

Wumpkit is a lightweight, high-performance Discord API library written in Go.

## Documentation Overview

* [Getting Started](/getting-started/) — Installation, setup, and first bot.
* [Core Concepts](/core-concepts/) — Client configuration, Gateway lifecycle, and REST.
* [Interactions](/interactions/) — Slash commands, buttons, select menus, and modals.
* [Router](/router/) — Command routing and middleware architecture.

## Quick Example

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wumpkit-go/wumpkit"
)

func main() {
	client := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(os.Getenv("DISCORD_TOKEN")),
	)

	client.OnReady(func(ctx context.Context, e wumpkit.ReadyEvent) {
		fmt.Printf("Logged in as %s!\n", e.User.Username)
	})

	if err := client.StartGatewayWithSignals(); err != nil {
		panic(err)
	}
}
```
