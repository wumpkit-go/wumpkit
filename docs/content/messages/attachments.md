---
title: "File Uploads"
weight: 30
---

# File Uploads

Discord lets you upload files like images, logs, text files, and audio clips alongside your messages.

wumpkit handles multipart form encoding for you automatically when you pass files in `CreateMessageOptions`.

## Uploading From Disk

Use `wumpkit.NewFileFromDisk` to open and attach a file stored on your machine. You can pass an optional alt text description for accessibility:

```go
package main

import (
	"context"
	"log"

	"github.com/wumpkit/wumpkit"
)

func uploadReport(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) {
	file, err := wumpkit.NewFileFromDisk("logs/report.txt", "Daily server report")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}

	_, err = client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Content: "Here is the latest daily report:",
		Files:   []wumpkit.File{file},
	})
	if err != nil {
		log.Printf("failed to send file: %v", err)
	}
}
```

## Uploading From Memory

If you generate files dynamically in your program (such as JSON data, generated graphs, or CSV sheets), use `wumpkit.NewFile` with any `io.Reader`:

```go
package main

import (
	"bytes"
	"context"

	"github.com/wumpkit/wumpkit"
)

func sendExportData(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) {
	data := []byte("id,username,score\n1,alice,95\n2,bob,88\n")
	buf := bytes.NewReader(data)

	csvFile := wumpkit.NewFile("leaderboard.csv", buf, "User leaderboard export")

	_, _ = client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Content: "Export complete. Download your CSV file below:",
		Files:   []wumpkit.File{csvFile},
	})
}
```

## Referencing Uploaded Images in Embeds

You can upload an image file and reference it inside an embed using the special `attachment://` URL scheme:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func sendImageEmbed(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) {
	badgeFile, err := wumpkit.NewFileFromDisk("assets/badge.png", "Winner badge")
	if err != nil {
		return
	}

	embed := wumpkit.NewEmbedBuilder().
		SetTitle("Tournament Winner!").
		SetDescription("Congratulations to our champion!").
		SetColor(wumpkit.ColorGold).
		SetImage("attachment://badge.png").
		Build()

	_, _ = client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Embeds: []wumpkit.Embed{embed},
		Files:  []wumpkit.File{badgeFile},
	})
}
```

## Downloading Files From Discord

wumpkit provides a helper function `wumpkit.DownloadFile` to download files from Discord CDN URLs directly into a directory on disk:

```go
filePath, err := wumpkit.DownloadFile(ctx, attachment.URL, "user_upload", "./downloads")
if err != nil {
	log.Printf("download failed: %v", err)
	return
}
log.Printf("file saved to: %s", filePath)
```
