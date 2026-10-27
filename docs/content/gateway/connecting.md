---
title: "Connecting & Lifecycle"
weight: 1
---

# Gateway Lifecycle

When your bot connects to Discord, Wumpkit establishes a WebSocket connection to Discord's Gateway v10 endpoint, negotiates compression, sends an Identify payload, and starts the heartbeat timer.

## Starting the Connection

Wumpkit offers two methods to launch the Gateway connection:

### 1. With Automatic Signal Handling (`StartGatewayWithSignals`)
This is the recommended approach for standalone bots. It starts all configured shards, connects to Discord, and blocks until your process receives an operating system termination signal (`SIGINT`, `SIGTERM`, or `Ctrl+C`). It then closes the WebSocket cleanly and cancels internal contexts:

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/wumpkit/wumpkit"
)

func main() {
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(os.Getenv("DISCORD_TOKEN")),
	)

	// blocks until an interrupt signal is received
	if err := bot.StartGatewayWithSignals(); err != nil {
		slog.Error("gateway stopped with error", "error", err)
	}
}
```

### 2. Manual Lifecycle Management (`StartGateway` & `Close`)
If your application manages its own goroutines, web server, or background services, start and stop the Gateway manually:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

bot := wumpkit.New(ctx, wumpkit.WithToken(token))

// starts connection in the background
if err := bot.StartGateway(); err != nil {
	log.Fatal(err)
}

// when your application shuts down:
if err := bot.Close(); err != nil {
	log.Printf("error closing gateway: %v\n", err)
}
```

## Transport Compression

Wumpkit supports transparent Gateway payload compression. By default, it uses shared zlib stream compression (`GatewayCompressionZlibStream`), which reduces inbound network bandwidth by up to 80% without noticeable CPU overhead.

You can configure compression mode in your shard manager settings:

```go
bot := wumpkit.New(
	ctx,
	wumpkit.WithToken(token),
	wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
		// options: GatewayCompressionZlibStream, GatewayCompressionZstdStream, GatewayCompressionNone
		Compression: wumpkit.GatewayCompressionZlibStream,
	}),
)
```

## Session Resumption

If the network connection drops or Discord moves your shard to a different gateway server, Wumpkit automatically reconnects and sends a Resume payload containing your active session ID and last acknowledged sequence number.

When a session resumes successfully:
1. Missed events are replayed by Discord in order.
2. The `OnResumed` event handler fires on your client.
3. Your local cache remains intact without needing to re-fetch all guilds.
