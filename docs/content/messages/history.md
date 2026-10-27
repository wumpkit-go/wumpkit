---
title: "Message History & Pagination"
weight: 60
---

# Message History & Pagination

Discord limits message history queries to 100 messages per HTTP call. When you need to read older messages, audit channel logs, or scan history, you must paginate across multiple requests.

wumpkit includes a built-in `MessagePaginator` that tracks cursors for you automatically.

## Using the Message Paginator

Create a paginator using `client.NewMessagePaginator(channelID)`:

```go
package main

import (
	"context"
	"log"

	"github.com/wumpkit/wumpkit"
)

func fetchChannelArchive(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) {
	paginator := client.NewMessagePaginator(channelID)

	for {
		// fetch up to 100 messages at a time
		batch, err := paginator.Next(ctx, 100)
		if err != nil {
			log.Printf("failed to fetch page: %v", err)
			break
		}

		// an empty batch means you reached the beginning of the channel history
		if len(batch) == 0 {
			break
		}

		for _, msg := range batch {
			log.Printf("[%s] %s: %s", msg.Timestamp, msg.Author.Username, msg.Content)
		}
	}
}
```

## Setting an Initial Cursor

You can tell the paginator where to start reading history:

```go
// fetch messages older than a specific message ID
paginator := client.NewMessagePaginator(channelID).Before(targetMessageID)

// fetch messages newer than a specific message ID
paginator := client.NewMessagePaginator(channelID).After(targetMessageID)
```

## Collecting Multiple Pages Automatically

If you just want a set number of messages (for instance, the last 500 messages in a channel), call `.Collect()`:

```go
// automatically runs sequential requests until 500 messages are gathered
messages, err := client.NewMessagePaginator(channelID).Collect(ctx, 500)
if err != nil {
	log.Printf("failed to collect messages: %v", err)
	return
}

log.Printf("retrieved %d messages from history", len(messages))
```
