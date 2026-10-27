---
title: "Incoming Webhooks"
weight: 1
---

# Incoming Webhooks

Discord webhooks allow external services (such as GitHub, CI/CD pipelines, or game servers) to post messages directly to a channel without requiring an active bot user account.

## Creating a Webhook

Create a webhook on a channel:

```go
package main

import (
	"context"
	"fmt"

	"github.com/wumpkit/wumpkit"
)

func createChannelWebhook(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) (*wumpkit.Webhook, error) {
	webhook, err := client.CreateWebhook(ctx, channelID, wumpkit.CreateWebhookOptions{
		Name: "Deployment Alerts",
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Created webhook: %s (ID: %s)\n", webhook.Name, webhook.ID)
	// webhook.Token contains the secret token needed to execute it
	return &webhook, nil
}
```

## Executing a Webhook

To post a message through a webhook, call `ExecuteWebhook`. You can customize the displayed username and avatar per message without modifying the webhook configuration:

```go
func sendAlert(ctx context.Context, client *wumpkit.Client, webhookID wumpkit.Snowflake, token string) error {
	embed := wumpkit.NewEmbedBuilder().
		SetTitle("Deployment Successful").
		SetDescription("Production build v2.4.0 deployed to all edge nodes.").
		SetColor(wumpkit.ColorMediumSeaGreen).
		Build()

	// wait=true returns the created Message object from Discord
	_, err := client.ExecuteWebhook(ctx, webhookID, token, wumpkit.ExecuteWebhookOptions{
		Username:  "DeployBot",
		AvatarURL: "https://example.com/deploybot-avatar.png",
		Embeds:    []wumpkit.Embed{embed},
	}, true)

	return err
}
```

## Posting into Forum or Announcement Threads

To deliver a webhook message into a specific thread inside a forum or text channel, set the `ThreadID` option:

```go
func sendToThread(ctx context.Context, client *wumpkit.Client, webhookID wumpkit.Snowflake, token string, threadID wumpkit.Snowflake) error {
	_, err := client.ExecuteWebhook(ctx, webhookID, token, wumpkit.ExecuteWebhookOptions{
		Content:  "Log update posted directly into this thread.",
		ThreadID: threadID,
	})
	return err
}
```

## Deleting a Webhook

```go
err := client.DeleteWebhook(ctx, webhookID)
```
