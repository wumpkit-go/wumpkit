/************************************************************************************
 *
 * wumpkit, A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD 3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package wumpkit

import (
	"context"
	"encoding/json"
)

// InteractionCallbackType represents the type of an interaction callback.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-interaction-callback-type
type InteractionCallbackType int

const (
	// InteractionCallbackTypePong ACK a Ping.
	InteractionCallbackTypePong InteractionCallbackType = 1

	// InteractionCallbackTypeChannelMessageWithSource respond to an interaction with a message.
	InteractionCallbackTypeChannelMessageWithSource InteractionCallbackType = 4

	// InteractionCallbackTypeDeferredChannelMessageWithSource ACK an interaction and edit a response later, the user sees a loading state.
	InteractionCallbackTypeDeferredChannelMessageWithSource InteractionCallbackType = 5

	// InteractionCallbackTypeDeferredUpdateMessage for components, ACK an interaction and edit the original message later; the user does not see a loading state.
	InteractionCallbackTypeDeferredUpdateMessage InteractionCallbackType = 6

	// InteractionCallbackTypeUpdateMessage for components, edit the message the component was attached to.
	InteractionCallbackTypeUpdateMessage InteractionCallbackType = 7

	// InteractionCallbackTypeApplicationCommandAutocompleteResult respond to an autocomplete interaction with suggested choices.
	InteractionCallbackTypeApplicationCommandAutocompleteResult InteractionCallbackType = 8

	// InteractionCallbackTypeModal respond to an interaction with a popup modal.
	InteractionCallbackTypeModal InteractionCallbackType = 9

	// InteractionCallbackTypePremiumRequired launch the Activity or respond with an upgrade prompt (deprecated).
	InteractionCallbackTypePremiumRequired InteractionCallbackType = 10

	// InteractionCallbackTypeLaunchActivity launch the Activity linked to the application.
	InteractionCallbackTypeLaunchActivity InteractionCallbackType = 12
)

// Is returns true if this InteractionCallbackType matches the provided one.
func (v InteractionCallbackType) Is(other InteractionCallbackType) bool {
	return v == other
}

// InteractionResponse represents an interaction response payload.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object
type InteractionResponse struct {
	// Type is the type of response.
	Type InteractionCallbackType `json:"type"`

	// Data is an optional response data object.
	Data any `json:"data,omitempty"`
}

// InteractionCallbackDataMessage represents the callback data for message responses.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-messages
type InteractionCallbackDataMessage struct {
	TTS             *bool              `json:"tts,omitempty"`
	Content         *string            `json:"content,omitempty"`
	Embeds          []Embed            `json:"embeds,omitempty"`
	AllowedMentions *AllowedMentions   `json:"allowed_mentions,omitempty"`
	Flags           *MessageFlags      `json:"flags,omitempty"`
	Components      []Component        `json:"components,omitempty"`
	Attachments     []Attachment       `json:"attachments,omitempty"`
	Poll            *PollCreateOptions `json:"poll,omitempty"`
	Files           []File             `json:"-"`
}

// ApplicationCommandOptionChoice represents a choice for autocomplete options.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-choice-structure
type ApplicationCommandOptionChoice struct {
	// Name is 1-100 character choice name.
	Name string `json:"name"`

	// NameLocalizations is localization dictionary for the name field.
	NameLocalizations map[Locale]string `json:"name_localizations,omitempty"`

	// Value is value of the choice, up to 100 characters if string.
	Value any `json:"value"`
}

// InteractionCallbackDataAutocomplete represents the callback data for autocomplete responses.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-autocomplete
type InteractionCallbackDataAutocomplete struct {
	Choices []ApplicationCommandOptionChoice `json:"choices"`
}

// InteractionCallbackDataModal represents the callback data for modal responses.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-modal
type InteractionCallbackDataModal struct {
	CustomID   string      `json:"custom_id"`
	Title      string      `json:"title"`
	Components []Component `json:"components"`
}

// CreateInteractionResponseOptions contains parameters for creating an interaction response.
type CreateInteractionResponseOptions struct {
	Response     InteractionResponse
	Files        []File
	WithResponse bool
}

// InteractionCallbackResource represents resource metadata returned in an interaction callback response.
type InteractionCallbackResource struct {
	ID       Snowflake       `json:"id"`
	Type     InteractionType `json:"type"`
	Activity any             `json:"activity,omitempty"`
	Message  *Message        `json:"message,omitempty"`
}

// InteractionCallbackResponse represents the response when with_response=true is requested.
type InteractionCallbackResponse struct {
	Interaction InteractionCallbackResource  `json:"interaction"`
	Resource    *InteractionCallbackResource `json:"resource,omitempty"`
}

// CreateFollowupMessageOptions contains parameters for creating an interaction followup message.
type CreateFollowupMessageOptions struct {
	Content         string             `json:"content,omitempty"`
	Username        string             `json:"username,omitempty"`
	AvatarURL       string             `json:"avatar_url,omitempty"`
	TTS             bool               `json:"tts,omitempty"`
	Embeds          []Embed            `json:"embeds,omitempty"`
	AllowedMentions *AllowedMentions   `json:"allowed_mentions,omitempty"`
	Components      []Component        `json:"components,omitempty"`
	Files           []File             `json:"-"`
	PayloadJSON     string             `json:"payload_json,omitempty"`
	Attachments     []Attachment       `json:"attachments,omitempty"`
	Flags           MessageFlags       `json:"flags,omitempty"`
	ThreadName      string             `json:"thread_name,omitempty"`
	AppliedTags     []Snowflake        `json:"applied_tags,omitempty"`
	Poll            *PollCreateOptions `json:"poll,omitempty"`
}

// EditFollowupMessageOptions contains parameters for editing an interaction followup message or original response.
type EditFollowupMessageOptions struct {
	Content         *string          `json:"content,omitempty"`
	Embeds          *[]Embed         `json:"embeds,omitempty"`
	Components      *[]Component     `json:"components,omitempty"`
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
	Attachments     *[]Attachment    `json:"attachments,omitempty"`
	Files           []File           `json:"-"`
	Flags           *MessageFlags    `json:"flags,omitempty"`
}

// CreateInteractionResponse creates a response to an Interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#create-interaction-response
func (r *requester) CreateInteractionResponse(
	ctx context.Context,
	interactionID Snowflake,
	interactionToken string,
	opts CreateInteractionResponseOptions,
) (*InteractionCallbackResponse, error) {
	endpoint := "/interactions/" + interactionID.String() + "/" + interactionToken + "/callback"
	if opts.WithResponse {
		endpoint += "?with_response=true"
	}

	if len(opts.Files) > 0 {
		payloadBytes, _ := json.Marshal(opts.Response)
		body, err := r.DoMultipartRequest(ctx, MultipartRequest{
			Method:      "POST",
			URL:         endpoint,
			PayloadJSON: payloadBytes,
			Files:       opts.Files,
		})
		if err != nil {
			return nil, err
		}
		defer body.Close()

		if opts.WithResponse {
			var resp InteractionCallbackResponse
			if err := json.NewDecoder(body).Decode(&resp); err != nil {
				return nil, err
			}
			return &resp, nil
		}
		return nil, nil
	}

	reqBody, _ := json.Marshal(opts.Response)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    endpoint,
		Body:   reqBody,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if opts.WithResponse {
		var resp InteractionCallbackResponse
		if err := json.NewDecoder(body).Decode(&resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}
	return nil, nil
}

// FetchOriginalInteractionResponse returns the initial Interaction response.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#get-original-interaction-response
func (r *requester) FetchOriginalInteractionResponse(
	ctx context.Context,
	applicationID Snowflake,
	interactionToken string,
) (Message, error) {
	endpoint := "/webhooks/" + applicationID.String() + "/" + interactionToken + "/messages/@original"
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
	})
	if err != nil {
		return Message{}, err
	}
	defer body.Close()

	var msg Message
	if err := json.NewDecoder(body).Decode(&msg); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/webhooks/{id}/{token}/messages/@original",
				"error", err.Error(),
			)
		return Message{}, err
	}
	return msg, nil
}

// EditOriginalInteractionResponse edits the initial Interaction response.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#edit-original-interaction-response
func (r *requester) EditOriginalInteractionResponse(
	ctx context.Context,
	applicationID Snowflake,
	interactionToken string,
	opts EditFollowupMessageOptions,
) (Message, error) {
	endpoint := "/webhooks/" + applicationID.String() + "/" + interactionToken + "/messages/@original"

	if len(opts.Files) > 0 {
		payloadBytes, _ := json.Marshal(opts)
		body, err := r.DoMultipartRequest(ctx, MultipartRequest{
			Method:      "PATCH",
			URL:         endpoint,
			PayloadJSON: payloadBytes,
			Files:       opts.Files,
		})
		if err != nil {
			return Message{}, err
		}
		defer body.Close()

		var msg Message
		if err := json.NewDecoder(body).Decode(&msg); err != nil {
			return Message{}, err
		}
		return msg, nil
	}

	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    endpoint,
		Body:   reqBody,
	})
	if err != nil {
		return Message{}, err
	}
	defer body.Close()

	var msg Message
	if err := json.NewDecoder(body).Decode(&msg); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/webhooks/{id}/{token}/messages/@original",
				"error", err.Error(),
			)
		return Message{}, err
	}
	return msg, nil
}

// DeleteOriginalInteractionResponse deletes the initial Interaction response.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#delete-original-interaction-response
func (r *requester) DeleteOriginalInteractionResponse(
	ctx context.Context,
	applicationID Snowflake,
	interactionToken string,
) error {
	endpoint := "/webhooks/" + applicationID.String() + "/" + interactionToken + "/messages/@original"
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    endpoint,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// CreateFollowupMessage creates a followup message for an Interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#create-followup-message
func (r *requester) CreateFollowupMessage(
	ctx context.Context,
	applicationID Snowflake,
	interactionToken string,
	opts CreateFollowupMessageOptions,
) (Message, error) {
	endpoint := "/webhooks/" + applicationID.String() + "/" + interactionToken

	if len(opts.Files) > 0 {
		payloadBytes, _ := json.Marshal(opts)
		body, err := r.DoMultipartRequest(ctx, MultipartRequest{
			Method:      "POST",
			URL:         endpoint,
			PayloadJSON: payloadBytes,
			Files:       opts.Files,
		})
		if err != nil {
			return Message{}, err
		}
		defer body.Close()

		var msg Message
		if err := json.NewDecoder(body).Decode(&msg); err != nil {
			return Message{}, err
		}
		return msg, nil
	}

	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    endpoint,
		Body:   reqBody,
	})
	if err != nil {
		return Message{}, err
	}
	defer body.Close()

	var msg Message
	if err := json.NewDecoder(body).Decode(&msg); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/webhooks/{id}/{token}",
				"error", err.Error(),
			)
		return Message{}, err
	}
	return msg, nil
}

// FetchFollowupMessage returns a followup message for an Interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#get-followup-message
func (r *requester) FetchFollowupMessage(
	ctx context.Context,
	applicationID Snowflake,
	interactionToken string,
	messageID Snowflake,
) (Message, error) {
	endpoint := "/webhooks/" + applicationID.String() + "/" + interactionToken + "/messages/" + messageID.String()
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
	})
	if err != nil {
		return Message{}, err
	}
	defer body.Close()

	var msg Message
	if err := json.NewDecoder(body).Decode(&msg); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/webhooks/{id}/{token}/messages/{id}",
				"error", err.Error(),
			)
		return Message{}, err
	}
	return msg, nil
}

// EditFollowupMessage edits a followup message for an Interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#edit-followup-message
func (r *requester) EditFollowupMessage(
	ctx context.Context,
	applicationID Snowflake,
	interactionToken string,
	messageID Snowflake,
	opts EditFollowupMessageOptions,
) (Message, error) {
	endpoint := "/webhooks/" + applicationID.String() + "/" + interactionToken + "/messages/" + messageID.String()

	if len(opts.Files) > 0 {
		payloadBytes, _ := json.Marshal(opts)
		body, err := r.DoMultipartRequest(ctx, MultipartRequest{
			Method:      "PATCH",
			URL:         endpoint,
			PayloadJSON: payloadBytes,
			Files:       opts.Files,
		})
		if err != nil {
			return Message{}, err
		}
		defer body.Close()

		var msg Message
		if err := json.NewDecoder(body).Decode(&msg); err != nil {
			return Message{}, err
		}
		return msg, nil
	}

	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    endpoint,
		Body:   reqBody,
	})
	if err != nil {
		return Message{}, err
	}
	defer body.Close()

	var msg Message
	if err := json.NewDecoder(body).Decode(&msg); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/webhooks/{id}/{token}/messages/{id}",
				"error", err.Error(),
			)
		return Message{}, err
	}
	return msg, nil
}

// DeleteFollowupMessage deletes a followup message for an Interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#delete-followup-message
func (r *requester) DeleteFollowupMessage(
	ctx context.Context,
	applicationID Snowflake,
	interactionToken string,
	messageID Snowflake,
) error {
	endpoint := "/webhooks/" + applicationID.String() + "/" + interactionToken + "/messages/" + messageID.String()
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    endpoint,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}
