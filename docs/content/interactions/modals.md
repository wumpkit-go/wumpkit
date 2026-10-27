---
title: "Modals & Text Inputs"
weight: 40
---

# Modals & Text Inputs

Modals are popup forms that prompt users to fill out text fields. They can be triggered in response to a slash command or a button click.

Each text input inside a modal must be placed in its own Action Row. A modal supports up to 5 text inputs.

## Text Input Styles

Discord supports two styles of text inputs:

* `wumpkit.TextInputStyleShort`: Single-line text input (e.g. names, emails, titles).
* `wumpkit.TextInputStyleParagraph`: Multi-line text area (e.g. feedback, descriptions, bug reports).

## Opening a Modal

To show a modal, respond to an incoming interaction with `InteractionCallbackTypeModal`:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func showFeedbackModal(ctx context.Context, client *wumpkit.Client, in wumpkit.Interaction) error {
	subjectInput := wumpkit.NewTextInputBuilder().
		SetCustomID("feedback_subject").
		SetLabel("Subject").
		SetStyle(wumpkit.TextInputStyleShort).
		SetPlaceholder("Brief summary of your feedback").
		SetMinLength(5).
		SetMaxLength(100).
		SetRequired(true).
		Build()

	detailsInput := wumpkit.NewTextInputBuilder().
		SetCustomID("feedback_details").
		SetLabel("Details").
		SetStyle(wumpkit.TextInputStyleParagraph).
		SetPlaceholder("Explain what happened in detail...").
		SetMinLength(10).
		SetMaxLength(1000).
		SetRequired(true).
		Build()

	row1 := wumpkit.NewActionRowBuilder().AddComponents(subjectInput).Build()
	row2 := wumpkit.NewActionRowBuilder().AddComponents(detailsInput).Build()

	_, err := client.CreateInteractionResponse(ctx, in.GetID(), in.GetToken(), wumpkit.CreateInteractionResponseOptions{
		Response: wumpkit.InteractionResponse{
			Type: wumpkit.InteractionCallbackTypeModal,
			Data: wumpkit.InteractionCallbackDataModal{
				CustomID:   "modal_feedback",
				Title:      "Submit Server Feedback",
				Components: []wumpkit.Component{row1, row2},
			},
		},
	})
	return err
}
```

## Handling Modal Submissions

When the user submits the form, Discord delivers an interaction of type `ModalSubmitInteraction`. Cast the interaction and read the submitted fields:

```go
client.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
	modal, ok := event.Interaction.(*wumpkit.ModalSubmitInteraction)
	if !ok {
		return
	}

	if modal.Data.CustomID == "modal_feedback" {
		var subject, details string

		for _, rowComp := range modal.Data.Components {
			if row, ok := rowComp.(*wumpkit.ActionRowComponent); ok {
				for _, comp := range row.Components {
					if input, ok := comp.(*wumpkit.TextInputComponent); ok {
						switch input.CustomID {
						case "feedback_subject":
							subject = input.Value
						case "feedback_details":
							details = input.Value
						}
					}
				}
			}
		}

		_ = client.RespondMessage(ctx, modal, "Thank you! Your feedback has been received.", wumpkit.RespondMessageOptions{
			Ephemeral: true,
		})

		log.Printf("Received feedback: [%s] %s", subject, details)
	}
})
```
