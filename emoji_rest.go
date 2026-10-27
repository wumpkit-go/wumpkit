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

// CreateGuildEmojiOptions contains parameters for creating a new guild emoji.
//
// Reference: https://discord.com/developers/docs/resources/emoji#create-guild-emoji
type CreateGuildEmojiOptions struct {
	// Name is the name of the emoji.
	Name string `json:"name"`

	// Image is the emoji image data.
	// Emojis and animated emojis have a maximum file size of 256 KiB.
	Image Base64Image `json:"image"`

	// Roles are the roles allowed to use this emoji.
	Roles []Snowflake `json:"roles,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// ModifyGuildEmojiOptions contains parameters for modifying an existing guild emoji.
//
// Reference: https://discord.com/developers/docs/resources/emoji#modify-guild-emoji
type ModifyGuildEmojiOptions struct {
	// Name is the name of the emoji. Zero value means not provided.
	Name string `json:"name,omitempty"`

	// Roles are the roles allowed to use this emoji.
	Roles []Snowflake `json:"roles,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// CreateApplicationEmojiOptions contains parameters for creating a new application emoji.
//
// Reference: https://discord.com/developers/docs/resources/emoji#create-application-emoji
type CreateApplicationEmojiOptions struct {
	// Name is the name of the emoji.
	Name string `json:"name"`

	// Image is the emoji image data.
	// Emojis and animated emojis have a maximum file size of 256 KiB.
	Image Base64Image `json:"image"`
}

// ModifyApplicationEmojiOptions contains parameters for updating an existing application emoji.
//
// Reference: https://discord.com/developers/docs/resources/emoji#modify-application-emoji
type ModifyApplicationEmojiOptions struct {
	// Name is the name of the emoji. Zero value means not provided.
	Name string `json:"name,omitempty"`
}

type applicationEmojisResponse struct {
	Items []Emoji `json:"items"`
}

// ListGuildEmojis retrieves a list of emoji objects for the given guild.
//
// Includes user fields if the bot has the PermissionCreateGuildExpressions or PermissionManageGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/emoji#list-guild-emojis
func (r *requester) ListGuildEmojis(ctx context.Context, guildID Snowflake) ([]Emoji, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/emojis",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var emojis []Emoji
	if err := json.NewDecoder(body).Decode(&emojis); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/emojis",
				"error", err.Error(),
			)
		return nil, err
	}
	return emojis, nil
}

// FetchGuildEmoji retrieves an emoji object for the given guild and emoji IDs.
//
// Includes the user field if the bot has the PermissionManageGuildExpressions permission,
// or if the bot created the emoji and has the PermissionCreateGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/emoji#get-guild-emoji
func (r *requester) FetchGuildEmoji(ctx context.Context, guildID, emojiID Snowflake) (Emoji, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/emojis/" + emojiID.String(),
	})
	if err != nil {
		return Emoji{}, err
	}
	defer body.Close()

	var emoji Emoji
	if err := json.NewDecoder(body).Decode(&emoji); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/emojis/{id}",
				"error", err.Error(),
			)
		return Emoji{}, err
	}
	return emoji, nil
}

// CreateGuildEmoji creates a new emoji for the guild.
//
// Requires the PermissionCreateGuildExpressions permission.
//
// Emojis and animated emojis have a maximum file size of 256 KiB.
// Attempting to upload an emoji larger than this limit will fail with an error.
//
// Reference: https://discord.com/developers/docs/resources/emoji#create-guild-emoji
func (r *requester) CreateGuildEmoji(ctx context.Context, guildID Snowflake, opts CreateGuildEmojiOptions) (Emoji, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/" + guildID.String() + "/emojis",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return Emoji{}, err
	}
	defer body.Close()

	var emoji Emoji
	if err := json.NewDecoder(body).Decode(&emoji); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/guilds/{id}/emojis",
				"error", err.Error(),
			)
		return Emoji{}, err
	}
	return emoji, nil
}

// ModifyGuildEmoji updates the given emoji.
//
// For emojis created by the current user, requires either the PermissionCreateGuildExpressions or PermissionManageGuildExpressions permission.
// For other emojis, requires the PermissionManageGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/emoji#modify-guild-emoji
func (r *requester) ModifyGuildEmoji(ctx context.Context, guildID, emojiID Snowflake, opts ModifyGuildEmojiOptions) (Emoji, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/emojis/" + emojiID.String(),
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return Emoji{}, err
	}
	defer body.Close()

	var emoji Emoji
	if err := json.NewDecoder(body).Decode(&emoji); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/emojis/{id}",
				"error", err.Error(),
			)
		return Emoji{}, err
	}
	return emoji, nil
}

// DeleteGuildEmoji deletes the given emoji.
//
// For emojis created by the current user, requires either the PermissionCreateGuildExpressions or PermissionManageGuildExpressions permission.
// For other emojis, requires the PermissionManageGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/emoji#delete-guild-emoji
func (r *requester) DeleteGuildEmoji(ctx context.Context, guildID, emojiID Snowflake, reason string) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/guilds/" + guildID.String() + "/emojis/" + emojiID.String(),
		Reason: reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// ListApplicationEmojis retrieves all emojis for an application.
//
// Includes a user object for the team member that uploaded the emoji from the app's settings,
// or for the bot user if uploaded using the API.
//
// Reference: https://discord.com/developers/docs/resources/emoji#list-application-emojis
func (r *requester) ListApplicationEmojis(ctx context.Context, applicationID Snowflake) ([]Emoji, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/applications/" + applicationID.String() + "/emojis",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var response applicationEmojisResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/applications/{id}/emojis",
				"error", err.Error(),
			)
		return nil, err
	}
	return response.Items, nil
}

// FetchApplicationEmoji retrieves a specific emoji from an application.
//
// Includes the user field.
//
// Reference: https://discord.com/developers/docs/resources/emoji#get-application-emoji
func (r *requester) FetchApplicationEmoji(ctx context.Context, applicationID, emojiID Snowflake) (Emoji, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/applications/" + applicationID.String() + "/emojis/" + emojiID.String(),
	})
	if err != nil {
		return Emoji{}, err
	}
	defer body.Close()

	var emoji Emoji
	if err := json.NewDecoder(body).Decode(&emoji); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/applications/{id}/emojis/{id}",
				"error", err.Error(),
			)
		return Emoji{}, err
	}
	return emoji, nil
}

// CreateApplicationEmoji creates a new emoji for an application.
//
// Emojis and animated emojis have a maximum file size of 256 KiB.
// Attempting to upload an emoji larger than this limit will fail with an error.
//
// Reference: https://discord.com/developers/docs/resources/emoji#create-application-emoji
func (r *requester) CreateApplicationEmoji(ctx context.Context, applicationID Snowflake, opts CreateApplicationEmojiOptions) (Emoji, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/applications/" + applicationID.String() + "/emojis",
		Body:   reqBody,
	})
	if err != nil {
		return Emoji{}, err
	}
	defer body.Close()

	var emoji Emoji
	if err := json.NewDecoder(body).Decode(&emoji); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/applications/{id}/emojis",
				"error", err.Error(),
			)
		return Emoji{}, err
	}
	return emoji, nil
}

// ModifyApplicationEmoji updates an existing application emoji.
//
// Reference: https://discord.com/developers/docs/resources/emoji#modify-application-emoji
func (r *requester) ModifyApplicationEmoji(ctx context.Context, applicationID, emojiID Snowflake, opts ModifyApplicationEmojiOptions) (Emoji, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/applications/" + applicationID.String() + "/emojis/" + emojiID.String(),
		Body:   reqBody,
	})
	if err != nil {
		return Emoji{}, err
	}
	defer body.Close()

	var emoji Emoji
	if err := json.NewDecoder(body).Decode(&emoji); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/applications/{id}/emojis/{id}",
				"error", err.Error(),
			)
		return Emoji{}, err
	}
	return emoji, nil
}

// DeleteApplicationEmoji deletes an application emoji.
//
// Reference: https://discord.com/developers/docs/resources/emoji#delete-application-emoji
func (r *requester) DeleteApplicationEmoji(ctx context.Context, applicationID, emojiID Snowflake) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/applications/" + applicationID.String() + "/emojis/" + emojiID.String(),
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}
