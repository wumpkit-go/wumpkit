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

// CreateGuildTemplateOptions contains parameters for creating a guild template.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#create-guild-template
type CreateGuildTemplateOptions struct {
	// Name is the name of the template (1-100 characters).
	Name string `json:"name"`

	// Description is the description for the template (0-120 characters).
	Description *string `json:"description,omitempty"`
}

// ModifyGuildTemplateOptions contains parameters for modifying a guild template.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#modify-guild-template
type ModifyGuildTemplateOptions struct {
	// Name is the name of the template (1-100 characters). Zero value means not provided.
	Name string `json:"name,omitempty"`

	// Description is the description for the template (0-120 characters). Pointer allows setting nil to clear.
	Description *string `json:"description,omitempty"`
}

// CreateGuildFromTemplateOptions contains parameters for creating a guild from a template.
type CreateGuildFromTemplateOptions struct {
	// Name is the name of the guild (2-100 characters).
	Name string `json:"name"`

	// Icon is the 128x128 1024x1024 png/jpeg image for the guild icon.
	Icon *Base64Image `json:"icon,omitempty"`
}

// FetchGuildTemplate returns a guild template object for the given code.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#get-guild-template
func (r *requester) FetchGuildTemplate(
	ctx context.Context,
	templateCode string,
) (GuildTemplate, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/templates/" + templateCode,
	})
	if err != nil {
		return GuildTemplate{}, err
	}
	defer body.Close()

	var template GuildTemplate
	if err := json.NewDecoder(body).Decode(&template); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/templates/{code}",
				"error", err.Error(),
			)
		return GuildTemplate{}, err
	}
	return template, nil
}

// CreateGuildFromTemplate creates a new guild based on a template.
//
// Note:
//   - This endpoint can be used only by bots in less than 10 guilds.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#create-guild-from-guild-template
func (r *requester) CreateGuildFromTemplate(
	ctx context.Context,
	templateCode string,
	opts CreateGuildFromTemplateOptions,
) (RestGuild, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/templates/" + templateCode,
		Body:   reqBody,
	})
	if err != nil {
		return RestGuild{}, err
	}
	defer body.Close()

	var guild RestGuild
	if err := json.NewDecoder(body).Decode(&guild); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/guilds/templates/{code}",
				"error", err.Error(),
			)
		return RestGuild{}, err
	}
	return guild, nil
}

// FetchGuildTemplates returns an array of guild template objects for a guild.
//
// Requires the PermissionManageGuild permission.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#get-guild-templates
func (r *requester) FetchGuildTemplates(
	ctx context.Context,
	guildID Snowflake,
) ([]GuildTemplate, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/templates",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var templates []GuildTemplate
	if err := json.NewDecoder(body).Decode(&templates); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/templates",
				"error", err.Error(),
			)
		return nil, err
	}
	return templates, nil
}

// CreateGuildTemplate creates a template for the guild.
//
// Requires the PermissionManageGuild permission.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#create-guild-template
func (r *requester) CreateGuildTemplate(
	ctx context.Context,
	guildID Snowflake,
	opts CreateGuildTemplateOptions,
) (GuildTemplate, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/" + guildID.String() + "/templates",
		Body:   reqBody,
	})
	if err != nil {
		return GuildTemplate{}, err
	}
	defer body.Close()

	var template GuildTemplate
	if err := json.NewDecoder(body).Decode(&template); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/guilds/{id}/templates",
				"error", err.Error(),
			)
		return GuildTemplate{}, err
	}
	return template, nil
}

// SyncGuildTemplate syncs the template to the guild's current state.
//
// Requires the PermissionManageGuild permission.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#sync-guild-template
func (r *requester) SyncGuildTemplate(
	ctx context.Context,
	guildID Snowflake,
	templateCode string,
) (GuildTemplate, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    "/guilds/" + guildID.String() + "/templates/" + templateCode,
	})
	if err != nil {
		return GuildTemplate{}, err
	}
	defer body.Close()

	var template GuildTemplate
	if err := json.NewDecoder(body).Decode(&template); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PUT",
				"url", "/guilds/{id}/templates/{code}",
				"error", err.Error(),
			)
		return GuildTemplate{}, err
	}
	return template, nil
}

// ModifyGuildTemplate modifies the template's metadata.
//
// Requires the PermissionManageGuild permission.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#modify-guild-template
func (r *requester) ModifyGuildTemplate(
	ctx context.Context,
	guildID Snowflake,
	templateCode string,
	opts ModifyGuildTemplateOptions,
) (GuildTemplate, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/templates/" + templateCode,
		Body:   reqBody,
	})
	if err != nil {
		return GuildTemplate{}, err
	}
	defer body.Close()

	var template GuildTemplate
	if err := json.NewDecoder(body).Decode(&template); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/templates/{code}",
				"error", err.Error(),
			)
		return GuildTemplate{}, err
	}
	return template, nil
}

// DeleteGuildTemplate deletes the template.
//
// Requires the PermissionManageGuild permission.
//
// Reference: https://discord.com/developers/docs/resources/guild-template#delete-guild-template
func (r *requester) DeleteGuildTemplate(
	ctx context.Context,
	guildID Snowflake,
	templateCode string,
) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/guilds/" + guildID.String() + "/templates/" + templateCode,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}
