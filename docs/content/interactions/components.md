---
title: "Buttons & Menus"
weight: 2
---

# Buttons & Menus

Message components let users interact directly with your messages using buttons and dropdown select menus.

## Component Layout Rules

Discord organizes message components into Action Rows:

* A message can contain up to 5 Action Rows.
* Each Action Row can hold up to 5 Buttons, OR 1 Select Menu.
* Buttons and Select Menus cannot be mixed in the same Action Row.

## 1. Creating Buttons

Use `wumpkit.NewButtonBuilder()` to construct buttons. Discord offers five button styles:

* `ButtonStylePrimary`: Blurple button for standard actions.
* `ButtonStyleSecondary`: Grey button for neutral actions.
* `ButtonStyleSuccess`: Green button for positive confirmations.
* `ButtonStyleDanger`: Red button for destructive actions.
* `ButtonStyleLink`: Opens an external website (does not send an interaction).

```go
package main

import (
	"github.com/wumpkit-go/wumpkit"
)

// confirmation button
btnAccept := wumpkit.NewButtonBuilder().
	SetCustomID("btn_accept").
	SetLabel("Accept").
	SetStyle(wumpkit.ButtonStyleSuccess).
	Build()

// cancellation button
btnDecline := wumpkit.NewButtonBuilder().
	SetCustomID("btn_decline").
	SetLabel("Decline").
	SetStyle(wumpkit.ButtonStyleDanger).
	Build()

// external link button
btnDocs := wumpkit.NewButtonBuilder().
	SetURL("https://github.com/wumpkit-go/wumpkit").
	SetLabel("Read Docs").
	SetStyle(wumpkit.ButtonStyleLink).
	Build()

// place buttons inside an action row
buttonRow := wumpkit.NewActionRowBuilder().
	SetComponent(btnAccept).
	SetComponent(btnDecline).
	SetComponent(btnDocs).
	Build()
```

## 2. Creating Select Menus

Use `wumpkit.NewStringSelectMenuBuilder()` to create a dropdown list:

```go
selectMenu := wumpkit.NewStringSelectMenuBuilder().
	SetCustomID("select_theme").
	SetPlaceholder("Choose a color theme...").
	SetOptions([]wumpkit.SelectOptionStructure{
		{
			Label:       "Sky Blue",
			Value:       "blue",
			Description: "Clean blue accent",
		},
		{
			Label:       "Forest Green",
			Value:       "green",
			Description: "Natural dark green",
		},
	}).
	Build()

selectRow := wumpkit.NewActionRowBuilder().
	SetComponent(selectMenu).
	Build()
```

## 3. Sending Components in a Message

Attach your action rows to a message using the `Components` option:

```go
_, err := bot.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
	Content:    "Please accept the guidelines and choose your theme:",
	Components: []wumpkit.LayoutComponent{buttonRow, selectRow},
})
```

## 4. Handling Component Clicks

When a user clicks a button or picks a dropdown option, Discord triggers `OnInteractionCreate`. Check for `*wumpkit.ComponentInteraction`:

```go
bot.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
	in, ok := event.Interaction.(*wumpkit.ComponentInteraction)
	if !ok {
		return
	}

	user := in.Author()

	switch in.Data.CustomID {
	case "btn_accept":
		// reply privately to the user
		_ = bot.RespondMessage(ctx, in, "Terms accepted. Welcome!", wumpkit.RespondMessageOptions{
			Ephemeral: true,
		})

	case "btn_decline":
		_ = bot.RespondMessage(ctx, in, "You declined the terms.", wumpkit.RespondMessageOptions{
			Ephemeral: true,
		})

	case "select_theme":
		// in.Data.Values contains the list of selected value strings
		selected := in.Data.Values[0]

		// update the original message in place, removing the buttons
		_ = bot.UpdateComponentMessage(ctx, in, wumpkit.RespondMessageOptions{
			Content: fmt.Sprintf("%s selected theme: %s", user.Username, selected),
		})
	}
})
```

Notice `bot.UpdateComponentMessage()`: instead of posting a new message, this replaces the content of the message that had the button or menu.
