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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FetchApplicationCommandsOptions contains query options for fetching commands.
type FetchApplicationCommandsOptions struct {
	// WithLocalizations whether to include full localization dictionaries.
	WithLocalizations bool `json:"with_localizations,omitempty"`
}

// EditApplicationCommandPermissionsOptions contains permissions to set.
type EditApplicationCommandPermissionsOptions struct {
	// Permissions is the permissions for the command in the guild (up to 100).
	Permissions []ApplicationCommandPermission `json:"permissions"`
}

// FetchGlobalApplicationCommands fetches all global commands for your application.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#get-global-application-commands
func (r *requester) FetchGlobalApplicationCommands(
	ctx context.Context,
	appID Snowflake,
	opts ...FetchApplicationCommandsOptions,
) ([]ApplicationCommand, error) {
	endpoint := fmt.Sprintf("/applications/%s/commands", appID)
	if len(opts) > 0 && opts[0].WithLocalizations {
		endpoint += "?with_localizations=true"
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodGet,
		URL:    endpoint,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommands(bodyBytes)
}

// CreateGlobalApplicationCommand creates a new global command.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#create-global-application-command
func (r *requester) CreateGlobalApplicationCommand(
	ctx context.Context,
	appID Snowflake,
	cmd ApplicationCommand,
) (ApplicationCommand, error) {
	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ApplicationCommand: %w", err)
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/applications/%s/commands", appID),
		Body:   payload,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommand(bodyBytes)
}

// FetchGlobalApplicationCommand fetches a global command by ID.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#get-global-application-command
func (r *requester) FetchGlobalApplicationCommand(
	ctx context.Context,
	appID, cmdID Snowflake,
) (ApplicationCommand, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/applications/%s/commands/%s", appID, cmdID),
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommand(bodyBytes)
}

// EditGlobalApplicationCommand edits a global command.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#edit-global-application-command
func (r *requester) EditGlobalApplicationCommand(
	ctx context.Context,
	appID, cmdID Snowflake,
	cmd ApplicationCommand,
) (ApplicationCommand, error) {
	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ApplicationCommand: %w", err)
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/applications/%s/commands/%s", appID, cmdID),
		Body:   payload,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommand(bodyBytes)
}

// DeleteGlobalApplicationCommand deletes a global command.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#delete-global-application-command
func (r *requester) DeleteGlobalApplicationCommand(
	ctx context.Context,
	appID, cmdID Snowflake,
) error {
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/applications/%s/commands/%s", appID, cmdID),
	})
}

// BulkOverwriteGlobalApplicationCommands takes a list of application commands, overwriting all existing commands.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#bulk-overwrite-global-application-commands
func (r *requester) BulkOverwriteGlobalApplicationCommands(
	ctx context.Context,
	appID Snowflake,
	cmds []ApplicationCommand,
) ([]ApplicationCommand, error) {
	payload, err := json.Marshal(cmds)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal commands: %w", err)
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/applications/%s/commands", appID),
		Body:   payload,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommands(bodyBytes)
}

// FetchGuildApplicationCommands fetches all commands for a specific guild.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#get-guild-application-commands
func (r *requester) FetchGuildApplicationCommands(
	ctx context.Context,
	appID, guildID Snowflake,
	opts ...FetchApplicationCommandsOptions,
) ([]ApplicationCommand, error) {
	endpoint := fmt.Sprintf("/applications/%s/guilds/%s/commands", appID, guildID)
	if len(opts) > 0 && opts[0].WithLocalizations {
		endpoint += "?with_localizations=true"
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodGet,
		URL:    endpoint,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommands(bodyBytes)
}

// CreateGuildApplicationCommand creates a new command in a specific guild.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#create-guild-application-command
func (r *requester) CreateGuildApplicationCommand(
	ctx context.Context,
	appID, guildID Snowflake,
	cmd ApplicationCommand,
) (ApplicationCommand, error) {
	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ApplicationCommand: %w", err)
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/applications/%s/guilds/%s/commands", appID, guildID),
		Body:   payload,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommand(bodyBytes)
}

// FetchGuildApplicationCommand fetches a guild command by ID.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#get-guild-application-command
func (r *requester) FetchGuildApplicationCommand(
	ctx context.Context,
	appID, guildID, cmdID Snowflake,
) (ApplicationCommand, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/applications/%s/guilds/%s/commands/%s", appID, guildID, cmdID),
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommand(bodyBytes)
}

// EditGuildApplicationCommand edits a guild command.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#edit-guild-application-command
func (r *requester) EditGuildApplicationCommand(
	ctx context.Context,
	appID, guildID, cmdID Snowflake,
	cmd ApplicationCommand,
) (ApplicationCommand, error) {
	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ApplicationCommand: %w", err)
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/applications/%s/guilds/%s/commands/%s", appID, guildID, cmdID),
		Body:   payload,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommand(bodyBytes)
}

// DeleteGuildApplicationCommand deletes a guild command.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#delete-guild-application-command
func (r *requester) DeleteGuildApplicationCommand(
	ctx context.Context,
	appID, guildID, cmdID Snowflake,
) error {
	return requestVoid(ctx, r, Request{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/applications/%s/guilds/%s/commands/%s", appID, guildID, cmdID),
	})
}

// BulkOverwriteGuildApplicationCommands overwrites all commands in a guild.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#bulk-overwrite-guild-application-commands
func (r *requester) BulkOverwriteGuildApplicationCommands(
	ctx context.Context,
	appID, guildID Snowflake,
	cmds []ApplicationCommand,
) ([]ApplicationCommand, error) {
	payload, err := json.Marshal(cmds)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal commands: %w", err)
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/applications/%s/guilds/%s/commands", appID, guildID),
		Body:   payload,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return UnmarshalApplicationCommands(bodyBytes)
}

// FetchGuildApplicationCommandPermissions fetches command permissions for all commands for your application in a guild.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#get-guild-application-command-permissions
func (r *requester) FetchGuildApplicationCommandPermissions(
	ctx context.Context,
	appID, guildID Snowflake,
) ([]GuildApplicationCommandPermissions, error) {
	return requestJSON[[]GuildApplicationCommandPermissions](ctx, r, Request{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/applications/%s/guilds/%s/commands/permissions", appID, guildID),
	})
}

// FetchApplicationCommandPermissions fetches command permissions for a specific command for your application in a guild.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#get-application-command-permissions
func (r *requester) FetchApplicationCommandPermissions(
	ctx context.Context,
	appID, guildID, cmdID Snowflake,
) (GuildApplicationCommandPermissions, error) {
	return requestJSON[GuildApplicationCommandPermissions](ctx, r, Request{
		Method: http.MethodGet,
		URL: fmt.Sprintf(
			"/applications/%s/guilds/%s/commands/%s/permissions",
			appID,
			guildID,
			cmdID,
		),
	})
}

// EditApplicationCommandPermissions edits permissions for a specific command for your application in a guild.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#edit-application-command-permissions
func (r *requester) EditApplicationCommandPermissions(
	ctx context.Context,
	appID, guildID, cmdID Snowflake,
	opts EditApplicationCommandPermissionsOptions,
) (GuildApplicationCommandPermissions, error) {
	payload, err := json.Marshal(opts)
	if err != nil {
		return GuildApplicationCommandPermissions{}, fmt.Errorf(
			"failed to marshal EditApplicationCommandPermissionsOptions: %w",
			err,
		)
	}

	return requestJSON[GuildApplicationCommandPermissions](ctx, r, Request{
		Method: http.MethodPut,
		URL: fmt.Sprintf(
			"/applications/%s/guilds/%s/commands/%s/permissions",
			appID,
			guildID,
			cmdID,
		),
		Body: payload,
	})
}
