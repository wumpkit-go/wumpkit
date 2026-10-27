---
title: "Rich Embeds"
weight: 2
---

# Rich Embeds

Discord allows messages to include up to 10 rich embeds. Embeds let you present structured information with colored borders, author icons, fields, timestamps, and footers.

## Creating an Embed

Wumpkit includes a fluent `EmbedBuilder` to construct embed payloads cleanly:

```go
package main

import (
	"context"
	"time"

	"github.com/wumpkit/wumpkit"
)

func sendAnnouncementEmbed(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) error {
	embed := wumpkit.NewEmbedBuilder().
		SetTitle("Server Update v2.5").
		SetDescription("The game servers have been upgraded with new community maps.").
		SetColor(wumpkit.ColorBlurple).
		SetAuthor("Server Admin", "", "https://example.com/admin.png").
		AddField("Downtime", "15 minutes", true).
		AddField("Patch Size", "450 MB", true).
		AddField("Status", "Online", true).
		SetFooter("Broadcast System", "").
		SetTimestamp(time.Now()).
		Build()

	_, err := client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Embeds: []wumpkit.Embed{embed},
	})
	return err
}
```

## Reusable Builders

All builders in Wumpkit implement the `ReusableBuilder` interface with a `Reset()` method. If you format multiple embeds inside loops, reuse the builder instance to eliminate unnecessary heap allocations:

```go
builder := wumpkit.NewEmbedBuilder()

for _, announcement := range announcements {
	builder.Reset()

	embed := builder.
		SetTitle(announcement.Title).
		SetDescription(announcement.Body).
		SetColor(wumpkit.ColorSummerSky).
		Build()

	// send embed...
}
```

## Available EmbedBuilder Methods

- `SetTitle(title string)`: Main title text (up to 256 characters)
- `SetDescription(desc string)`: Embed body content (up to 4096 characters)
- `SetColor(color Color)`: Colored vertical strip on the left edge
- `SetAuthor(name, url, iconURL string)`: Small author banner at the top
- `SetThumbnail(url string)`: Small preview image placed in the top-right corner
- `SetImage(url string)`: Full-width banner image placed at the bottom
- `AddField(name, value string, inline bool)`: Key-value data row
- `SetFooter(text, iconURL string)`: Small footer text at the bottom
- `SetTimestamp(time.Time)`: ISO8601 timestamp rendered next to the footer
