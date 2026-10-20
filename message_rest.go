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
	"strconv"
)

// AllowedMentionType represents the allowed mention types.
//
// Reference: https://discord.com/developers/docs/resources/message#allowed-mentions-object-allowed-mention-types
type AllowedMentionType string

const (
	// AllowedMentionTypeRoles controls role mentions.
	AllowedMentionTypeRoles AllowedMentionType = "roles"

	// AllowedMentionTypeUsers controls user mentions.
	AllowedMentionTypeUsers AllowedMentionType = "users"

	// AllowedMentionTypeEveryone controls @everyone and @here mentions.
	AllowedMentionTypeEveryone AllowedMentionType = "everyone"
)

// AllowedMentions allows for more granular control over mentions without various hacks.
//
// Reference: https://discord.com/developers/docs/resources/message#allowed-mentions-object
type AllowedMentions struct {
	// Parse is an array of allowed mention types to parse from the content.
	Parse []AllowedMentionType `json:"parse,omitempty"`

	// Roles is an array of role_ids to mention (max 100).
	Roles []Snowflake `json:"roles,omitempty"`

	// Users is an array of user_ids to mention (max 100).
	Users []Snowflake `json:"users,omitempty"`

	// RepliedUser specifies whether to mention the author of the message being replied to (default false).
	RepliedUser bool `json:"replied_user,omitempty"`
}

// Is returns true if this AllowedMentionType matches the provided one.
func (v AllowedMentionType) Is(other AllowedMentionType) bool {
	return v == other
}

// CreateMessageOptions contains options for creating a new message.
//
// Reference: https://discord.com/developers/docs/resources/message#create-message
type CreateMessageOptions struct {
	// Content is the message contents (up to 2000 characters).
	Content string `json:"content,omitempty"`

	// Nonce can be used to verify a message was sent (up to 25 characters).
	Nonce any `json:"nonce,omitempty"`

	// TTS specifies whether this is a text-to-speech message.
	TTS bool `json:"tts,omitempty"`

	// Embeds is up to 10 rich embeds (up to 6000 characters combined).
	Embeds []Embed `json:"embeds,omitempty"`

	// AllowedMentions specifies allowed mentions for the message.
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`

	// MessageReference includes source message data for replies.
	MessageReference *MessageReference `json:"message_reference,omitempty"`

	// Components is an array of components to include with the message.
	Components []LayoutComponent `json:"components,omitempty"`

	// StickerIDs is IDs of up to 3 stickers in the server to send in the message.
	StickerIDs []Snowflake `json:"sticker_ids,omitempty"`

	// Files is a slice of files to upload with the message.
	Files []File `json:"-"`

	// Attachments is attachment objects with filename and description for partial uploads.
	Attachments []Attachment `json:"attachments,omitempty"`

	// Flags is message flags combined as a bitfield (e.g. SuppressEmbeds, SuppressNotifications).
	Flags MessageFlags `json:"flags,omitempty"`

	// EnforceNonce specifies whether to enforce nonce validation for uniqueness.
	EnforceNonce bool `json:"enforce_nonce,omitempty"`

	// Poll contains poll creation parameters.
	Poll *PollCreateOptions `json:"poll,omitempty"`
}

// EditMessageOptions contains options for editing an existing message.
//
// Reference: https://discord.com/developers/docs/resources/message#edit-message
type EditMessageOptions struct {
	// Content is the new message contents (up to 2000 characters). Pointer allows setting empty string to clear.
	Content *string `json:"content,omitempty"`

	// Embeds is the new array of embeds (up to 10). Pointer allows passing empty slice to clear embeds.
	Embeds *[]Embed `json:"embeds,omitempty"`

	// Flags is message flags combined as a bitfield (only SuppressEmbeds can be set).
	Flags *MessageFlags `json:"flags,omitempty"`

	// AllowedMentions specifies allowed mentions for the message.
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`

	// Components is the new array of interactive components. Pointer allows passing empty slice to clear components.
	Components *[]LayoutComponent `json:"components,omitempty"`

	// Files is new files to upload.
	Files []File `json:"-"`

	// Attachments is retained and new attachments to keep. Pointer allows passing empty slice to clear attachments.
	Attachments *[]Attachment `json:"attachments,omitempty"`
}

// FetchMessagesOptions contains query options for fetching channel messages.
//
// Reference: https://discord.com/developers/docs/resources/message#get-channel-messages
type FetchMessagesOptions struct {
	// Around fetches messages around this message ID.
	Around Snowflake `json:"around,omitempty"`

	// Before fetches messages before this message ID.
	Before Snowflake `json:"before,omitempty"`

	// After fetches messages after this message ID.
	After Snowflake `json:"after,omitempty"`

	// Limit is the max number of messages to return (1-100, default 50).
	Limit int `json:"limit,omitempty"`
}

// BulkDeleteMessagesOptions specifies messages to delete in bulk.
//
// Reference: https://discord.com/developers/docs/resources/message#bulk-delete-messages
type BulkDeleteMessagesOptions struct {
	// Messages is a slice of 2 to 100 message IDs to delete.
	Messages []Snowflake `json:"messages"`

	// Reason is the audit log reason.
	Reason string `json:"-"`
}

// ReactionType indicates the type of reaction to fetch.
type ReactionType int

const (
	ReactionTypeNormal ReactionType = 0
	ReactionTypeBurst  ReactionType = 1
)

// Is returns true if this ReactionType matches the provided one.
func (v ReactionType) Is(other ReactionType) bool {
	return v == other
}

// FetchReactionsOptions contains query options for fetching reactions on a message.
//
// Reference: https://discord.com/developers/docs/resources/message#get-reactions
type FetchReactionsOptions struct {
	// After fetches users after this user ID.
	After Snowflake `json:"after,omitempty"`

	// Limit is the max number of users to return (1-100, default 25).
	Limit int `json:"limit,omitempty"`

	// Type is the reaction type (normal or burst).
	Type ReactionType `json:"type,omitempty"`
}

// CreateMessage sends a message to a guild text or DM channel.
//
// Reference: https://discord.com/developers/docs/resources/message#create-message
func (r *requester) CreateMessage(ctx context.Context, channelID Snowflake, opts CreateMessageOptions) (Message, error) {
	endpoint := fmt.Sprintf("/channels/%s/messages", channelID)
	if len(opts.Files) > 0 {
		payload, err := json.Marshal(opts)
		if err != nil {
			return Message{}, fmt.Errorf("failed to marshal CreateMessageOptions: %w", err)
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
		return Message{}, fmt.Errorf("failed to marshal CreateMessageOptions: %w", err)
	}

	return requestJSON[Message](ctx, r, Request{
		Method: http.MethodPost,
		URL:    endpoint,
		Body:   payload,
	})
}

// FetchChannelMessages returns the messages for a channel.
//
// Reference: https://discord.com/developers/docs/resources/message#get-channel-messages
func (r *requester) FetchChannelMessages(ctx context.Context, channelID Snowflake, opts ...FetchMessagesOptions) ([]Message, error) {
	var first FetchMessagesOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	query := url.Values{}
	if first.Around != 0 {
		query.Set("around", first.Around.String())
	}
	if first.Before != 0 {
		query.Set("before", first.Before.String())
	}
	if first.After != 0 {
		query.Set("after", first.After.String())
	}
	if first.Limit > 0 {
		query.Set("limit", strconv.Itoa(first.Limit))
	}

	endpoint := fmt.Sprintf("/channels/%s/messages", channelID)
	if q := query.Encode(); q != "" {
		endpoint += "?" + q
	}

	return requestJSON[[]Message](ctx, r, Request{
		Method: http.MethodGet,
		URL:    endpoint,
	})
}

// FetchChannelMessage returns a specific message in the channel.
//
// Reference: https://discord.com/developers/docs/resources/message#get-channel-message
func (r *requester) FetchChannelMessage(ctx context.Context, channelID, messageID Snowflake) (Message, error) {
	endpoint := fmt.Sprintf("/channels/%s/messages/%s", channelID, messageID)
	return requestJSON[Message](ctx, r, Request{
		Method: http.MethodGet,
		URL:    endpoint,
	})
}

// CrosspostMessage crossposts a message in an Announcement Channel to following channels.
//
// Reference: https://discord.com/developers/docs/resources/message#crosspost-message
func (r *requester) CrosspostMessage(ctx context.Context, channelID, messageID Snowflake) (Message, error) {
	endpoint := fmt.Sprintf("/channels/%s/messages/%s/crosspost", channelID, messageID)
	return requestJSON[Message](ctx, r, Request{
		Method: http.MethodPost,
		URL:    endpoint,
	})
}

// CreateReaction adds a reaction to a message with an emoji (e.g. "😀" or "name:123456789").
//
// Reference: https://discord.com/developers/docs/resources/message#create-reaction
func (r *requester) CreateReaction(ctx context.Context, channelID, messageID Snowflake, emoji string) error {
	endpoint := fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, url.PathEscape(emoji))
	return requestVoid(ctx, r, Request{
		Method: http.MethodPut,
		URL:    endpoint,
	})
}

// DeleteOwnReaction deletes a reaction the current user has made for the specified emoji.
//
// Reference: https://discord.com/developers/docs/resources/message#delete-own-reaction
func (r *requester) DeleteOwnReaction(ctx context.Context, channelID, messageID Snowflake, emoji string) error {
	endpoint := fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, url.PathEscape(emoji))
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    endpoint,
	})
}

// DeleteUserReaction deletes another user's reaction for the specified emoji.
//
// Reference: https://discord.com/developers/docs/resources/message#delete-user-reaction
func (r *requester) DeleteUserReaction(ctx context.Context, channelID, messageID, userID Snowflake, emoji string) error {
	endpoint := fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/%s", channelID, messageID, url.PathEscape(emoji), userID)
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    endpoint,
	})
}

// FetchReactions returns a list of users that reacted with the specified emoji.
//
// Reference: https://discord.com/developers/docs/resources/message#get-reactions
func (r *requester) FetchReactions(ctx context.Context, channelID, messageID Snowflake, emoji string, opts ...FetchReactionsOptions) ([]User, error) {
	var first FetchReactionsOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	query := url.Values{}
	if first.After != 0 {
		query.Set("after", first.After.String())
	}
	if first.Limit > 0 {
		query.Set("limit", strconv.Itoa(first.Limit))
	}
	if first.Type != 0 {
		query.Set("type", strconv.Itoa(int(first.Type)))
	}

	endpoint := fmt.Sprintf("/channels/%s/messages/%s/reactions/%s", channelID, messageID, url.PathEscape(emoji))
	if q := query.Encode(); q != "" {
		endpoint += "?" + q
	}

	return requestJSON[[]User](ctx, r, Request{
		Method: http.MethodGet,
		URL:    endpoint,
	})
}

// DeleteAllReactions deletes all reactions on a message.
//
// Reference: https://discord.com/developers/docs/resources/message#delete-all-reactions
func (r *requester) DeleteAllReactions(ctx context.Context, channelID, messageID Snowflake) error {
	endpoint := fmt.Sprintf("/channels/%s/messages/%s/reactions", channelID, messageID)
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    endpoint,
	})
}

// DeleteAllReactionsForEmoji deletes all reactions for a specific emoji on a message.
//
// Reference: https://discord.com/developers/docs/resources/message#delete-all-reactions-for-emoji
func (r *requester) DeleteAllReactionsForEmoji(ctx context.Context, channelID, messageID Snowflake, emoji string) error {
	endpoint := fmt.Sprintf("/channels/%s/messages/%s/reactions/%s", channelID, messageID, url.PathEscape(emoji))
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    endpoint,
	})
}

// EditMessage edits a previously sent message.
//
// Reference: https://discord.com/developers/docs/resources/message#edit-message
func (r *requester) EditMessage(ctx context.Context, channelID, messageID Snowflake, opts EditMessageOptions) (Message, error) {
	endpoint := fmt.Sprintf("/channels/%s/messages/%s", channelID, messageID)
	if len(opts.Files) > 0 {
		payload, err := json.Marshal(opts)
		if err != nil {
			return Message{}, fmt.Errorf("failed to marshal EditMessageOptions: %w", err)
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
		return Message{}, fmt.Errorf("failed to marshal EditMessageOptions: %w", err)
	}

	return requestJSON[Message](ctx, r, Request{
		Method: http.MethodPatch,
		URL:    endpoint,
		Body:   payload,
	})
}

// DeleteMessage deletes a message in a channel.
//
// Reference: https://discord.com/developers/docs/resources/message#delete-message
func (r *requester) DeleteMessage(ctx context.Context, channelID, messageID Snowflake, reason ...string) error {
	auditReason := ""
	if len(reason) > 0 {
		auditReason = reason[0]
	}
	endpoint := fmt.Sprintf("/channels/%s/messages/%s", channelID, messageID)
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    endpoint,
		Reason: auditReason,
	})
}

// BulkDeleteMessages deletes multiple messages in a single request (2-100 messages, younger than 2 weeks).
//
// Reference: https://discord.com/developers/docs/resources/message#bulk-delete-messages
func (r *requester) BulkDeleteMessages(ctx context.Context, channelID Snowflake, opts BulkDeleteMessagesOptions) error {
	payload, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("failed to marshal BulkDeleteMessagesOptions: %w", err)
	}

	endpoint := fmt.Sprintf("/channels/%s/messages/bulk-delete", channelID)
	return requestVoid(ctx, r, Request{
		Method: http.MethodPost,
		URL:    endpoint,
		Body:   payload,
		Reason: opts.Reason,
	})
}

// FetchPinnedMessages returns all pinned messages in the channel.
//
// Reference: https://discord.com/developers/docs/resources/message#get-pinned-messages
func (r *requester) FetchPinnedMessages(ctx context.Context, channelID Snowflake) ([]Message, error) {
	endpoint := fmt.Sprintf("/channels/%s/pins", channelID)
	return requestJSON[[]Message](ctx, r, Request{
		Method: http.MethodGet,
		URL:    endpoint,
	})
}

// PinMessage pins a message in the channel.
//
// Reference: https://discord.com/developers/docs/resources/message#pin-message
func (r *requester) PinMessage(ctx context.Context, channelID, messageID Snowflake, reason ...string) error {
	auditReason := ""
	if len(reason) > 0 {
		auditReason = reason[0]
	}
	endpoint := fmt.Sprintf("/channels/%s/pins/%s", channelID, messageID)
	return requestVoid(ctx, r, Request{
		Method: http.MethodPut,
		URL:    endpoint,
		Reason: auditReason,
	})
}

// UnpinMessage unpins a message in the channel.
//
// Reference: https://discord.com/developers/docs/resources/message#unpin-message
func (r *requester) UnpinMessage(ctx context.Context, channelID, messageID Snowflake, reason ...string) error {
	auditReason := ""
	if len(reason) > 0 {
		auditReason = reason[0]
	}
	endpoint := fmt.Sprintf("/channels/%s/pins/%s", channelID, messageID)
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    endpoint,
		Reason: auditReason,
	})
}
