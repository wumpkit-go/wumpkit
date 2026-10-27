---
title: "REST API & Messages"
weight: 3
---

# REST API & Messages

Wumpkit includes a full Discord REST client built into `Client`. You can use it to send messages, construct embeds, upload files, modify channels, and manage server roles.

## Sending Messages

To send a message to a Discord channel, call `bot.CreateMessage()`:

```go
msg, err := bot.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
	Content: "Hello from Wumpkit!",
})
if err != nil {
	slog.Error("failed to send message", "error", err)
	return
}
slog.Info("sent message", "id", msg.ID)
```

## Building Rich Embeds

Wumpkit provides a fluent builder called `wumpkit.NewEmbedBuilder()` to create Discord embeds:

```go
import (
	"time"
	"github.com/wumpkit/wumpkit"
)

embed := wumpkit.NewEmbedBuilder().
	SetTitle("Server Status").
	SetDescription("All systems operational.").
	SetColor(wumpkit.ColorSummerSky).
	SetThumbnail("https://example.com/logo.png").
	AddField("Uptime", "99.98%", true).
	AddField("Memory", "24 MB", true).
	SetFooter("Bot v1.0.0", "").
	SetTimestamp(time.Now()).
	Build()

_, err := bot.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
	Embeds: []wumpkit.Embed{embed},
})
```

You can attach up to 10 embeds to a single message.

## Uploading Files

To attach files to a message, pass a slice of `wumpkit.File`:

```go
// open file from disk
file, err := wumpkit.NewFileFromDisk("chart.png", "Monthly activity chart")
if err != nil {
	slog.Error("cannot open file", "error", err)
	return
}

_, err = bot.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
	Content: "Here is the latest report:",
	Files:   []wumpkit.File{file},
})
```

You can also create an in-memory file from any `io.Reader`:

```go
file := wumpkit.NewFile("log.txt", strings.NewReader("log content here"), "Log dump")
```

## Editing and Deleting Messages

Edit an existing message by channel ID and message ID:

```go
newContent := "Updated message text."
_, err := bot.EditMessage(ctx, channelID, messageID, wumpkit.EditMessageOptions{
	Content: &newContent,
})
```

Delete a message:

```go
err := bot.DeleteMessage(ctx, channelID, messageID)
```

## Handling Discord REST Errors

When Discord rejects a request, Wumpkit returns a structured `*wumpkit.DiscordRestError`. You can inspect the HTTP status code, Discord internal error code, and error message:

```go
import (
	"errors"
	"github.com/wumpkit/wumpkit"
)

_, err := bot.CreateMessage(ctx, channelID, options)
if err != nil {
	var discordErr *wumpkit.DiscordRestError
	if errors.As(err, &discordErr) {
		slog.Error("discord rejected request",
			"status_code", discordErr.StatusCode,
			"discord_code", discordErr.Code,
			"message", discordErr.Message,
		)
		return
	}

	slog.Error("generic network error", "error", err)
}
```

## Automatic Rate Limiting

Discord enforces strict per-route rate limits. When you send multiple requests to the same route, Wumpkit handles rate limits automatically in the background. It reads Discord `X-RateLimit-*` headers, pauses the calling goroutine until the bucket resets, and retries the request transparently.
