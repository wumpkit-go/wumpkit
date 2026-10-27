---
title: "Making REST Requests"
weight: 1
---

# Making REST Requests

All Discord REST API endpoints are exposed as methods on your `*wumpkit.Client`. Wumpkit uses Discord API v10 by default.

## Basic API Requests

Every REST call in Wumpkit accepts a `context.Context` as its first parameter:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wumpkit/wumpkit"
)

func fetchBotUser(client *wumpkit.Client) error {
	// create a context with a 5-second deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// fetch current authenticated bot user profile (@me)
	me, err := client.FetchCurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("failed fetching bot user: %w", err)
	}

	fmt.Printf("Bot Username: %s (ID: %s)\n", me.Username, me.ID)
	return nil
}
```

## Configuring Request Settings

You can customize timeout durations, retry counts, and custom headers when initializing your client:

```go
bot := wumpkit.New(
	ctx,
	wumpkit.WithToken(token),
	wumpkit.WithRequesterConfig(wumpkit.RequesterConfig{
		BaseURL:    "https://discord.com/api/v10",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}),
)
```

## Error Handling

When Discord returns an error (such as `403 Forbidden` due to missing bot permissions or `404 Not Found` for deleted channels), Wumpkit parses the response body into structured error details:

```go
msg, err := client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
	Content: "Hello!",
})
if err != nil {
	// inspect error and handle accordingly
	log.Printf("request failed: %v\n", err)
	return err
}
```
