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

// CreateGuildStickerOptions contains parameters for creating a new guild sticker.
//
// Reference: https://discord.com/developers/docs/resources/sticker#create-guild-sticker
type CreateGuildStickerOptions struct {
	// Name is the name of the sticker (2-30 characters).
	Name string

	// Description is the description of the sticker (empty or 2-100 characters).
	Description string

	// Tags is the autocomplete/suggestion tags for the sticker (max 200 characters).
	Tags string

	// File is the sticker file to upload (PNG, APNG, GIF, or Lottie JSON, max 512 KiB).
	File File

	// Reason is the reason shown in the audit log for this action.
	Reason string
}

// ModifyGuildStickerOptions contains parameters for modifying a guild sticker.
//
// Reference: https://discord.com/developers/docs/resources/sticker#modify-guild-sticker
type ModifyGuildStickerOptions struct {
	// Name is the name of the sticker (2-30 characters).
	Name string `json:"name,omitempty"`

	// Description is the description of the sticker (2-100 characters). Pointer allows setting nil to clear.
	Description *string `json:"description,omitempty"`

	// Tags is autocomplete/suggestion tags for the sticker (max 200 characters).
	Tags string `json:"tags,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// FetchSticker returns a sticker object for the given sticker ID.
//
// Reference: https://discord.com/developers/docs/resources/sticker#get-sticker
func (r *requester) FetchSticker(ctx context.Context, stickerID Snowflake) (Sticker, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/stickers/" + stickerID.String(),
	})
	if err != nil {
		return Sticker{}, err
	}
	defer body.Close()

	var sticker Sticker
	if err := json.NewDecoder(body).Decode(&sticker); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/stickers/{id}",
				"error", err.Error(),
			)
		return Sticker{}, err
	}
	return sticker, nil
}

// ListNitroStickerPacks returns the list of sticker packs available to Nitro subscribers.
//
// Reference: https://discord.com/developers/docs/resources/sticker#list-sticker-packs
func (r *requester) ListNitroStickerPacks(ctx context.Context) ([]StickerPack, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/sticker-packs",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var response NitroStickerPacksResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/sticker-packs",
			"error", err.Error(),
		)
		return nil, err
	}
	return response.StickerPacks, nil
}

// ListGuildStickers returns an array of sticker objects for the given guild.
//
// Reference: https://discord.com/developers/docs/resources/sticker#list-guild-stickers
func (r *requester) ListGuildStickers(ctx context.Context, guildID Snowflake) ([]Sticker, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/stickers",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var stickers []Sticker
	if err := json.NewDecoder(body).Decode(&stickers); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/stickers",
				"error", err.Error(),
			)
		return nil, err
	}
	for i := range stickers {
		stickers[i].GuildID = guildID
	}
	return stickers, nil
}

// FetchGuildSticker returns a sticker object for the given guild and sticker IDs.
//
// Reference: https://discord.com/developers/docs/resources/sticker#get-guild-sticker
func (r *requester) FetchGuildSticker(ctx context.Context, guildID, stickerID Snowflake) (Sticker, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/stickers/" + stickerID.String(),
	})
	if err != nil {
		return Sticker{}, err
	}
	defer body.Close()

	var sticker Sticker
	if err := json.NewDecoder(body).Decode(&sticker); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/stickers/{id}",
				"error", err.Error(),
			)
		return Sticker{}, err
	}
	sticker.GuildID = guildID
	return sticker, nil
}

// CreateGuildSticker creates a new sticker for the guild.
//
// Requires the PermissionCreateGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/sticker#create-guild-sticker
func (r *requester) CreateGuildSticker(ctx context.Context, guildID Snowflake, opts CreateGuildStickerOptions) (Sticker, error) {
	extra := map[string]string{
		"name": opts.Name,
		"tags": opts.Tags,
	}
	if opts.Description != "" {
		extra["description"] = opts.Description
	}

	body, err := r.DoMultipartRequest(ctx, MultipartRequest{
		Method:      "POST",
		URL:         "/guilds/" + guildID.String() + "/stickers",
		ExtraFields: extra,
		Files:       []File{opts.File},
		Reason:      opts.Reason,
	})
	if err != nil {
		return Sticker{}, err
	}
	defer body.Close()

	var sticker Sticker
	if err := json.NewDecoder(body).Decode(&sticker); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/guilds/{id}/stickers",
				"error", err.Error(),
			)
		return Sticker{}, err
	}
	sticker.GuildID = guildID
	return sticker, nil
}

// ModifyGuildSticker modifies the given sticker in a guild.
//
// Requires the PermissionManageGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/sticker#modify-guild-sticker
func (r *requester) ModifyGuildSticker(ctx context.Context, guildID, stickerID Snowflake, opts ModifyGuildStickerOptions) (Sticker, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/stickers/" + stickerID.String(),
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return Sticker{}, err
	}
	defer body.Close()

	var sticker Sticker
	if err := json.NewDecoder(body).Decode(&sticker); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/stickers/{id}",
				"error", err.Error(),
			)
		return Sticker{}, err
	}
	sticker.GuildID = guildID
	return sticker, nil
}

// DeleteGuildSticker deletes the given sticker from a guild.
//
// Requires the PermissionManageGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/sticker#delete-guild-sticker
func (r *requester) DeleteGuildSticker(ctx context.Context, guildID, stickerID Snowflake, reason string) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/guilds/" + guildID.String() + "/stickers/" + stickerID.String(),
		Reason: reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}
