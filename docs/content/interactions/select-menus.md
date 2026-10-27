---
title: "Select Menus"
weight: 30
---

# Select Menus

Select menus let users choose one or more options from a dropdown list.

Each select menu must be placed in its own Action Row. You can place up to 5 select menus per message across 5 action rows.

## Types of Select Menus

wumpkit supports all five Discord select menu types:

1. **String Select Menu**: Static list of text options you provide.
2. **User Select Menu**: Dropdown populated with guild users and bots.
3. **Role Select Menu**: Dropdown populated with server roles.
4. **Channel Select Menu**: Dropdown populated with channels (filterable by channel type).
5. **Mentionable Select Menu**: Dropdown containing both users and roles.

## String Select Menu

Use `wumpkit.NewStringSelectMenuBuilder` to build custom option lists:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func sendColorPicker(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) {
	menu := wumpkit.NewStringSelectMenuBuilder().
		SetCustomID("color_select").
		SetPlaceholder("Choose your favorite color...").
		SetMinValues(1).
		SetMaxValues(1).
		AddOptions(
			wumpkit.SelectOption{
				Label:       "Red",
				Value:       "color_red",
				Description: "Bright crimson hue",
				Emoji:       &wumpkit.PartialEmoji{Name: "🔴"},
			},
			wumpkit.SelectOption{
				Label:       "Blue",
				Value:       "color_blue",
				Description: "Deep ocean blue",
				Emoji:       &wumpkit.PartialEmoji{Name: "🔵"},
			},
			wumpkit.SelectOption{
				Label:       "Green",
				Value:       "color_green",
				Description: "Vibrant emerald green",
				Emoji:       &wumpkit.PartialEmoji{Name: "🟢"},
			},
		).
		Build()

	row := wumpkit.NewActionRowBuilder().
		AddComponents(menu).
		Build()

	_, _ = client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Content:    "Select your preferred profile color:",
		Components: []wumpkit.LayoutComponent{row},
	})
}
```

## User and Role Select Menus

When you want users to pick a user or a role without typing names or IDs manually:

```go
// user picker
userMenu := wumpkit.NewUserSelectMenuBuilder().
	SetCustomID("assignee_select").
	SetPlaceholder("Select a team member...").
	SetMinValues(1).
	SetMaxValues(3).
	Build()

// role picker
roleMenu := wumpkit.NewRoleSelectMenuBuilder().
	SetCustomID("role_select").
	SetPlaceholder("Pick server roles...").
	Build()
```

## Channel Select Menu

You can filter channel menus to specific types (such as only text channels or voice channels):

```go
channelMenu := wumpkit.NewChannelSelectMenuBuilder().
	SetCustomID("target_channel").
	SetPlaceholder("Select an announcement channel...").
	SetChannelTypes(wumpkit.ChannelTypeGuildText, wumpkit.ChannelTypeGuildAnnouncement).
	Build()
```

## Handling Selected Values

When a user makes a selection, handle the `*wumpkit.ComponentInteraction` and inspect `in.Data.Values`:

```go
client.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
	in, ok := event.Interaction.(*wumpkit.ComponentInteraction)
	if !ok {
		return
	}

	if in.Data.CustomID == "color_select" && len(in.Data.Values) > 0 {
		picked := in.Data.Values[0]
		_ = client.RespondMessage(ctx, in, "You selected: "+picked, wumpkit.RespondMessageOptions{
			Ephemeral: true,
		})
	}
})
```
