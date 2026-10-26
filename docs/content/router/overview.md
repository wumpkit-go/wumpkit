---
title: "Overview & Features"
weight: 1
---

# Router Overview & Features

Wumpkit Router is a standalone, high-performance interaction router designed for zero allocations on routing fast paths.

```go
package main

import (
	"github.com/wumpkit-go/router"
	"github.com/wumpkit-go/wumpkit"
)

type PingCommand struct {
	router.ChatCommandConfig
}

func (p *PingCommand) Run(ctx router.Context) error {
	return ctx.ReplyEphemeral("Pong!")
}
```
