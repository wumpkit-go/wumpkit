/************************************************************************************
 *
 * wumpkit, A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD-3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package wumpkit

import (
	"context"
)

// RespondMessageOptions configures high-level interaction responses.
type RespondMessageOptions struct {
	// Content is the text content of the message.
	Content string

	// Flags specifies message bitfield flags (e.g. MessageFlagEphemeral, MessageFlagSuppressEmbeds).
	Flags MessageFlags

	// Ephemeral specifies whether the message is visible only to the invoking user.
	Ephemeral bool

	// TTS specifies whether this is a text-to-speech message.
	TTS bool

	// Embeds contains up to 10 embeds to include with the response.
	Embeds []Embed

	// Components contains message components (buttons, select menus).
	Components []Component

	// Files contains files to upload with the response.
	Files []File

	// AllowedMentions controls mention suppression.
	AllowedMentions *AllowedMentions
}

// RespondMessage sends an immediate message response to an interaction.
// This simplifies the multi-layered CreateInteractionResponse structure into a clean, single-line call.
func (c *Client) RespondMessage(
	ctx context.Context,
	interaction Interaction,
	content string,
	opts ...RespondMessageOptions,
) error {
	var opt RespondMessageOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	var flags *MessageFlags
	if opt.Flags != 0 {
		f := opt.Flags
		if opt.Ephemeral {
			f |= MessageFlagEphemeral
		}
		flags = &f
	} else if opt.Ephemeral {
		f := MessageFlagEphemeral
		flags = &f
	}

	var cnt *string
	if content != "" {
		cnt = &content
	} else if opt.Content != "" {
		cnt = &opt.Content
	}

	var tts *bool
	if opt.TTS {
		t := true
		tts = &t
	}

	_, err := c.CreateInteractionResponse(
		ctx,
		interaction.GetID(),
		interaction.GetToken(),
		CreateInteractionResponseOptions{
			Response: InteractionResponse{
				Type: InteractionCallbackTypeChannelMessageWithSource,
				Data: InteractionCallbackDataMessage{
					Content:         cnt,
					TTS:             tts,
					Embeds:          opt.Embeds,
					Components:      opt.Components,
					AllowedMentions: opt.AllowedMentions,
					Flags:           flags,
				},
			},
			Files: opt.Files,
		},
	)
	return err
}

// RespondEmbeds sends an immediate message response containing embeds to an interaction.
func (c *Client) RespondEmbeds(
	ctx context.Context,
	interaction Interaction,
	embeds []Embed,
	opts ...RespondMessageOptions,
) error {
	var opt RespondMessageOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	opt.Embeds = embeds
	return c.RespondMessage(ctx, interaction, "", opt)
}

// DeferResponse acknowledges an interaction with a deferred response (showing a loading state).
// If ephemeral is true, the follow-up message will only be visible to the user who invoked the command.
func (c *Client) DeferResponse(ctx context.Context, interaction Interaction, ephemeral bool) error {
	var flags *MessageFlags
	if ephemeral {
		f := MessageFlagEphemeral
		flags = &f
	}

	_, err := c.CreateInteractionResponse(
		ctx,
		interaction.GetID(),
		interaction.GetToken(),
		CreateInteractionResponseOptions{
			Response: InteractionResponse{
				Type: InteractionCallbackTypeDeferredChannelMessageWithSource,
				Data: InteractionCallbackDataMessage{
					Flags: flags,
				},
			},
		},
	)
	return err
}

// DeferComponentUpdate acknowledges a component interaction without showing a loading state,
// allowing you to edit the message at a later time.
func (c *Client) DeferComponentUpdate(ctx context.Context, interaction Interaction) error {
	_, err := c.CreateInteractionResponse(
		ctx,
		interaction.GetID(),
		interaction.GetToken(),
		CreateInteractionResponseOptions{
			Response: InteractionResponse{
				Type: InteractionCallbackTypeDeferredUpdateMessage,
			},
		},
	)
	return err
}

// UpdateComponentMessage responds to a component interaction by immediately updating the original message.
func (c *Client) UpdateComponentMessage(
	ctx context.Context,
	interaction Interaction,
	opts ...RespondMessageOptions,
) error {
	var opt RespondMessageOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	var cnt *string
	if opt.Content != "" {
		cnt = &opt.Content
	}

	_, err := c.CreateInteractionResponse(
		ctx,
		interaction.GetID(),
		interaction.GetToken(),
		CreateInteractionResponseOptions{
			Response: InteractionResponse{
				Type: InteractionCallbackTypeUpdateMessage,
				Data: InteractionCallbackDataMessage{
					Content:         cnt,
					Embeds:          opt.Embeds,
					Components:      opt.Components,
					AllowedMentions: opt.AllowedMentions,
				},
			},
			Files: opt.Files,
		},
	)
	return err
}

// RespondAutocomplete responds to an autocomplete interaction with suggested choices.
func (c *Client) RespondAutocomplete(
	ctx context.Context,
	interaction Interaction,
	choices []ApplicationCommandOptionChoice,
) error {
	_, err := c.CreateInteractionResponse(
		ctx,
		interaction.GetID(),
		interaction.GetToken(),
		CreateInteractionResponseOptions{
			Response: InteractionResponse{
				Type: InteractionCallbackTypeApplicationCommandAutocompleteResult,
				Data: InteractionCallbackDataAutocomplete{
					Choices: choices,
				},
			},
		},
	)
	return err
}
