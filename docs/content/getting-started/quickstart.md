---
title: "Quickstart"
weight: 2
---

# Quickstart

Build and run your first Discord bot with Wumpkit.

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
