---
title: "Buttons"
weight: 20
---

# Buttons

Buttons are interactive UI elements attached to messages. When clicked, they send an interaction to your bot.

Buttons must be placed inside an **Action Row**. A message can have up to 5 action rows, and each action row can hold up to 5 buttons.

## Button Styles

Discord offers five button styles:

* `wumpkit.ButtonStylePrimary`: Blurple button for main actions.
* `wumpkit.ButtonStyleSecondary`: Grey button for neutral actions.
* `wumpkit.ButtonStyleSuccess`: Green button for confirmations.
* `wumpkit.ButtonStyleDanger`: Red button for destructive actions like bans or deletions.
* `wumpkit.ButtonStyleLink`: Link button that opens a URL in the browser (does not trigger a bot interaction).

## Constructing Buttons

Use `wumpkit.NewButtonBuilder` to create buttons and `wumpkit.NewActionRowBuilder` to bundle them:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func sendConfirmationPrompt(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) {
	btnAccept := wumpkit.NewButtonBuilder().
		SetCustomID("terms_accept").
		SetLabel("Accept").
		SetStyle(wumpkit.ButtonStyleSuccess).
		Build()

	btnDecline := wumpkit.NewButtonBuilder().
		SetCustomID("terms_decline").
		SetLabel("Decline").
		SetStyle(wumpkit.ButtonStyleDanger).
		Build()

	btnDocs := wumpkit.NewButtonBuilder().
		SetLabel("Terms of Service").
		SetStyle(wumpkit.ButtonStyleLink).
		SetURL("https://example.com/terms").
		Build()

	row := wumpkit.NewActionRowBuilder().
		AddComponents(btnAccept, btnDecline, btnDocs).
		Build()

	_, _ = client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Content:    "Please review and accept our server terms:",
		Components: []wumpkit.LayoutComponent{row},
	})
}
```

## Handling Button Clicks

When a user clicks a button, handle `EventInteractionCreate` and cast the interaction to `*wumpkit.ComponentInteraction`:

```go
client.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
	in, ok := event.Interaction.(*wumpkit.ComponentInteraction)
	if !ok {
		return
	}

	switch in.Data.CustomID {
	case "terms_accept":
		// respond with a private confirmation message
		_ = client.RespondMessage(ctx, in, "Terms accepted. Welcome!", wumpkit.RespondMessageOptions{
			Ephemeral: true,
		})

	case "terms_decline":
		_ = client.RespondMessage(ctx, in, "You declined the terms.", wumpkit.RespondMessageOptions{
			Ephemeral: true,
		})
	}
})
```

## Updating the Original Message

If you want the button click to update the message itself (for example, disabling buttons or editing text), call `UpdateComponentMessage`:

```go
case "terms_accept":
	disabledRow := wumpkit.NewActionRowBuilder().
		AddComponents(
			wumpkit.NewButtonBuilder().SetCustomID("terms_accept").SetLabel("Accepted").SetStyle(wumpkit.ButtonStyleSuccess).SetDisabled(true).Build(),
		).
		Build()

	_ = client.UpdateComponentMessage(ctx, in, wumpkit.RespondMessageOptions{
		Content:    "Terms accepted! You now have access.",
		Components: []wumpkit.Component{disabledRow},
	})
```
