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
	"fmt"
	"net/http"
	"net/url"
)

// CreateWebhookOptions contains parameters for creating a webhook.
//
// Reference: https://discord.com/developers/docs/resources/webhook#create-webhook
type CreateWebhookOptions struct {
	// Name is the name of the webhook (1-80 characters).
	Name string `json:"name"`

	// Avatar is the image data URI for the default webhook avatar.
	Avatar *Base64Image `json:"avatar,omitempty"`

	// Reason is the audit log reason.
	Reason string `json:"-"`
}

// ModifyWebhookOptions contains parameters for modifying a webhook.
//
// Reference: https://discord.com/developers/docs/resources/webhook#modify-webhook
type ModifyWebhookOptions struct {
	// Name is the default name of the webhook (1-80 characters). Zero value means not provided.
	Name string `json:"name,omitempty"`

	// Avatar is the image data URI for the default webhook avatar. Pointer allows setting nil to clear.
	Avatar *Base64Image `json:"avatar,omitempty"`

	// ChannelID is the new channel ID to move the webhook to.
	ChannelID Snowflake `json:"channel_id,omitempty"`

	// Reason is the audit log reason.
	Reason string `json:"-"`
}

// ExecuteWebhookOptions contains parameters for executing a webhook.
//
// Reference: https://discord.com/developers/docs/resources/webhook#execute-webhook
type ExecuteWebhookOptions struct {
	// Content is the message contents (up to 2000 characters).
	Content string `json:"content,omitempty"`

	// Username overrides the default username of the webhook.
	Username string `json:"username,omitempty"`

	// AvatarURL overrides the default avatar of the webhook.
	AvatarURL string `json:"avatar_url,omitempty"`

	// TTS specifies whether this is a text-to-speech message.
	TTS bool `json:"tts,omitempty"`

	// Embeds is embedded rich content (up to 10 embeds).
	Embeds []Embed `json:"embeds,omitempty"`

	// AllowedMentions specifies allowed mentions for the message.
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`

	// Components is interactive components to include with the message.
	Components []LayoutComponent `json:"components,omitempty"`

	// Files is a slice of files to upload.
	Files []File `json:"-"`

	// Attachments is attachment objects with filename and description for uploads.
	Attachments []Attachment `json:"attachments,omitempty"`

	// Flags is message flags combined as a bitfield (e.g. SuppressEmbeds, SuppressNotifications).
	Flags MessageFlags `json:"flags,omitempty"`

	// ThreadName specifies the name of thread to create (in forum/media channels).
	ThreadName string `json:"thread_name,omitempty"`

	// AppliedTags is tag IDs to apply to the thread (in forum/media channels).
	AppliedTags []Snowflake `json:"applied_tags,omitempty"`

	// Poll contains poll creation parameters.
	Poll *PollCreateOptions `json:"poll,omitempty"`

	// Wait waits for server confirmation of message send before returning, and returns the created message.
	Wait bool `json:"-"`

	// ThreadID sends the message to the specified thread within the webhook's channel.
	ThreadID Snowflake `json:"-"`
}

// EditWebhookMessageOptions contains parameters for editing a message sent by a webhook.
//
// Reference: https://discord.com/developers/docs/resources/webhook#edit-webhook-message
type EditWebhookMessageOptions struct {
	// Content is the new message contents (up to 2000 characters).
	Content *string `json:"content,omitempty"`

	// Embeds is the new array of embeds (up to 10).
	Embeds *[]Embed `json:"embeds,omitempty"`

	// Components is the new array of interactive components.
	Components *[]LayoutComponent `json:"components,omitempty"`

	// AllowedMentions specifies allowed mentions for the message.
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`

	// Files is new files to upload.
	Files []File `json:"-"`

	// Attachments is retained and new attachments to keep.
	Attachments *[]Attachment `json:"attachments,omitempty"`

	// ThreadID specifies the thread where the message is located.
	ThreadID Snowflake `json:"-"`
}

// CreateWebhook creates a new webhook in the specified channel.
//
// Reference: https://discord.com/developers/docs/resources/webhook#create-webhook
func (r *requester) CreateWebhook(
	ctx context.Context,
	channelID Snowflake,
	opts CreateWebhookOptions,
) (Webhook, error) {
	payload, err := json.Marshal(opts)
	if err != nil {
		return Webhook{}, fmt.Errorf("failed to marshal CreateWebhookOptions: %w", err)
	}

	return requestJSON[Webhook](ctx, r, Request{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/channels/%s/webhooks", channelID),
		Body:   payload,
		Reason: opts.Reason,
	})
}

// FetchChannelWebhooks returns webhooks for a channel.
//
// Reference: https://discord.com/developers/docs/resources/webhook#get-channel-webhooks
func (r *requester) FetchChannelWebhooks(
	ctx context.Context,
	channelID Snowflake,
) ([]Webhook, error) {
	return requestJSON[[]Webhook](ctx, r, Request{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/channels/%s/webhooks", channelID),
	})
}

// FetchGuildWebhooks returns all webhooks in a guild.
//
// Reference: https://discord.com/developers/docs/resources/webhook#get-guild-webhooks
func (r *requester) FetchGuildWebhooks(ctx context.Context, guildID Snowflake) ([]Webhook, error) {
	return requestJSON[[]Webhook](ctx, r, Request{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/guilds/%s/webhooks", guildID),
	})
}

// FetchWebhook returns a webhook by ID (requires Manage Webhooks permission).
//
// Reference: https://discord.com/developers/docs/resources/webhook#get-webhook
func (r *requester) FetchWebhook(ctx context.Context, webhookID Snowflake) (Webhook, error) {
	return requestJSON[Webhook](ctx, r, Request{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/webhooks/%s", webhookID),
	})
}

// FetchWebhookWithToken returns a webhook by ID and token (does not require authentication).
//
// Reference: https://discord.com/developers/docs/resources/webhook#get-webhook-with-token
func (r *requester) FetchWebhookWithToken(
	ctx context.Context,
	webhookID Snowflake,
	token string,
) (Webhook, error) {
	return requestJSON[Webhook](ctx, r, Request{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/webhooks/%s/%s", webhookID, token),
		NoAuth: true,
	})
}

// ModifyWebhook modifies a webhook (requires Manage Webhooks permission).
//
// Reference: https://discord.com/developers/docs/resources/webhook#modify-webhook
func (r *requester) ModifyWebhook(
	ctx context.Context,
	webhookID Snowflake,
	opts ModifyWebhookOptions,
) (Webhook, error) {
	payload, err := json.Marshal(opts)
	if err != nil {
		return Webhook{}, fmt.Errorf("failed to marshal ModifyWebhookOptions: %w", err)
	}

	return requestJSON[Webhook](ctx, r, Request{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/webhooks/%s", webhookID),
		Body:   payload,
		Reason: opts.Reason,
	})
}

// ModifyWebhookWithToken modifies a webhook using its token (does not require authentication).
//
// Reference: https://discord.com/developers/docs/resources/webhook#modify-webhook-with-token
func (r *requester) ModifyWebhookWithToken(
	ctx context.Context,
	webhookID Snowflake,
	token string,
	opts ModifyWebhookOptions,
) (Webhook, error) {
	payload, err := json.Marshal(opts)
	if err != nil {
		return Webhook{}, fmt.Errorf("failed to marshal ModifyWebhookOptions: %w", err)
	}

	return requestJSON[Webhook](ctx, r, Request{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/webhooks/%s/%s", webhookID, token),
		Body:   payload,
		Reason: opts.Reason,
		NoAuth: true,
	})
}

// DeleteWebhook deletes a webhook by ID (requires Manage Webhooks permission).
//
// Reference: https://discord.com/developers/docs/resources/webhook#delete-webhook
func (r *requester) DeleteWebhook(
	ctx context.Context,
	webhookID Snowflake,
	reason ...string,
) error {
	auditReason := ""
	if len(reason) > 0 {
		auditReason = reason[0]
	}
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/webhooks/%s", webhookID),
		Reason: auditReason,
	})
}

// DeleteWebhookWithToken deletes a webhook using its token.
//
// Reference: https://discord.com/developers/docs/resources/webhook#delete-webhook-with-token
func (r *requester) DeleteWebhookWithToken(
	ctx context.Context,
	webhookID Snowflake,
	token string,
	reason ...string,
) error {
	auditReason := ""
	if len(reason) > 0 {
		auditReason = reason[0]
	}
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/webhooks/%s/%s", webhookID, token),
		Reason: auditReason,
		NoAuth: true,
	})
}

// ExecuteWebhook executes a webhook with message data.
//
// Reference: https://discord.com/developers/docs/resources/webhook#execute-webhook
func (r *requester) ExecuteWebhook(
	ctx context.Context,
	webhookID Snowflake,
	token string,
	opts ExecuteWebhookOptions,
) (Message, error) {
	query := url.Values{}
	if opts.Wait {
		query.Set("wait", "true")
	}
	if opts.ThreadID != 0 {
		query.Set("thread_id", opts.ThreadID.String())
	}

	endpoint := fmt.Sprintf("/webhooks/%s/%s", webhookID, token)
	if q := query.Encode(); q != "" {
		endpoint += "?" + q
	}

	if len(opts.Files) > 0 {
		payload, err := json.Marshal(opts)
		if err != nil {
			return Message{}, fmt.Errorf("failed to marshal ExecuteWebhookOptions: %w", err)
		}
		return requestMultipartJSON[Message](ctx, r, MultipartRequest{
			Method:      http.MethodPost,
			URL:         endpoint,
			PayloadJSON: payload,
			Files:       opts.Files,
		})
	}

	payload, err := json.Marshal(opts)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal ExecuteWebhookOptions: %w", err)
	}

	if !opts.Wait {
		err := requestVoid(ctx, r, Request{
			Method: http.MethodPost,
			URL:    endpoint,
			Body:   payload,
			NoAuth: true,
		})
		return Message{}, err
	}

	return requestJSON[Message](ctx, r, Request{
		Method: http.MethodPost,
		URL:    endpoint,
		Body:   payload,
		NoAuth: true,
	})
}

// FetchWebhookMessage returns a previously-sent webhook message from the same token.
//
// Reference: https://discord.com/developers/docs/resources/webhook#get-webhook-message
func (r *requester) FetchWebhookMessage(
	ctx context.Context,
	webhookID Snowflake,
	token string,
	messageID Snowflake,
	threadID ...Snowflake,
) (Message, error) {
	endpoint := fmt.Sprintf("/webhooks/%s/%s/messages/%s", webhookID, token, messageID)
	if len(threadID) > 0 && threadID[0] != 0 {
		endpoint += fmt.Sprintf("?thread_id=%s", threadID[0])
	}
	return requestJSON[Message](ctx, r, Request{
		Method: http.MethodGet,
		URL:    endpoint,
		NoAuth: true,
	})
}

// EditWebhookMessage edits a previously-sent webhook message from the same token.
//
// Reference: https://discord.com/developers/docs/resources/webhook#edit-webhook-message
func (r *requester) EditWebhookMessage(
	ctx context.Context,
	webhookID Snowflake,
	token string,
	messageID Snowflake,
	opts EditWebhookMessageOptions,
) (Message, error) {
	endpoint := fmt.Sprintf("/webhooks/%s/%s/messages/%s", webhookID, token, messageID)
	if opts.ThreadID != 0 {
		endpoint += fmt.Sprintf("?thread_id=%s", opts.ThreadID)
	}

	if len(opts.Files) > 0 {
		payload, err := json.Marshal(opts)
		if err != nil {
			return Message{}, fmt.Errorf("failed to marshal EditWebhookMessageOptions: %w", err)
		}
		return requestMultipartJSON[Message](ctx, r, MultipartRequest{
			Method:      http.MethodPatch,
			URL:         endpoint,
			PayloadJSON: payload,
			Files:       opts.Files,
		})
	}

	payload, err := json.Marshal(opts)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal EditWebhookMessageOptions: %w", err)
	}

	return requestJSON[Message](ctx, r, Request{
		Method: http.MethodPatch,
		URL:    endpoint,
		Body:   payload,
		NoAuth: true,
	})
}

// DeleteWebhookMessage deletes a message that was created by the webhook.
//
// Reference: https://discord.com/developers/docs/resources/webhook#delete-webhook-message
func (r *requester) DeleteWebhookMessage(
	ctx context.Context,
	webhookID Snowflake,
	token string,
	messageID Snowflake,
	threadID ...Snowflake,
) error {
	endpoint := fmt.Sprintf("/webhooks/%s/%s/messages/%s", webhookID, token, messageID)
	if len(threadID) > 0 && threadID[0] != 0 {
		endpoint += fmt.Sprintf("?thread_id=%s", threadID[0])
	}
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    endpoint,
		NoAuth: true,
	})
}

// ExecuteSlackWebhook executes a Slack-compatible webhook.
//
// Reference: https://discord.com/developers/docs/resources/webhook#execute-slackcompatible-webhook
func (r *requester) ExecuteSlackWebhook(
	ctx context.Context,
	webhookID Snowflake,
	token string,
	payload []byte,
	wait bool,
	threadID ...Snowflake,
) error {
	query := url.Values{}
	if wait {
		query.Set("wait", "true")
	}
	if len(threadID) > 0 && threadID[0] != 0 {
		query.Set("thread_id", threadID[0].String())
	}
	endpoint := fmt.Sprintf("/webhooks/%s/%s/slack", webhookID, token)
	if q := query.Encode(); q != "" {
		endpoint += "?" + q
	}
	return requestVoid(ctx, r, Request{
		Method: http.MethodPost,
		URL:    endpoint,
		Body:   payload,
		NoAuth: true,
	})
}

// ExecuteGitHubWebhook executes a GitHub-compatible webhook.
//
// Reference: https://discord.com/developers/docs/resources/webhook#execute-githubcompatible-webhook
func (r *requester) ExecuteGitHubWebhook(
	ctx context.Context,
	webhookID Snowflake,
	token string,
	payload []byte,
	wait bool,
	threadID ...Snowflake,
) error {
	query := url.Values{}
	if wait {
		query.Set("wait", "true")
	}
	if len(threadID) > 0 && threadID[0] != 0 {
		query.Set("thread_id", threadID[0].String())
	}
	endpoint := fmt.Sprintf("/webhooks/%s/%s/github", webhookID, token)
	if q := query.Encode(); q != "" {
		endpoint += "?" + q
	}
	return requestVoid(ctx, r, Request{
		Method: http.MethodPost,
		URL:    endpoint,
		Body:   payload,
		NoAuth: true,
	})
}
