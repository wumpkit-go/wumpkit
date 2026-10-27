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
	"io"
	"net/url"
	"strconv"
)

// ModifyCurrentUserOptions contains options for modifying the bot's user account settings.
//
// Reference: https://discord.com/developers/docs/resources/user#modify-current-user
type ModifyCurrentUserOptions struct {
	// Username is the user's username. Zero value ("") means not provided/omitted.
	Username string `json:"username,omitempty"`

	// Avatar modifies the user's avatar image. Pointer allows nil (not modified) or image data.
	Avatar *Base64Image `json:"avatar,omitempty"`

	// Banner modifies the user's banner image. Pointer allows nil (not modified) or image data.
	Banner *Base64Image `json:"banner,omitempty"`
}

// FetchCurrentUserGuildsOptions contains query parameters for fetching the current user's guilds.
//
// Reference: https://discord.com/developers/docs/resources/user#get-current-user-guilds
type FetchCurrentUserGuildsOptions struct {
	// Before returns guilds before this guild ID.
	Before Snowflake `json:"before,omitempty"`

	// After returns guilds after this guild ID.
	After Snowflake `json:"after,omitempty"`

	// Limit is max number of guilds to return (1-200, default 200).
	Limit int `json:"limit,omitempty"`

	// WithCounts includes approximate member and presence counts in the response.
	WithCounts bool `json:"with_counts,omitempty"`

	// shard returns guilds that belong to this shard (required for bots in >10,000 guilds)
	Shard *int `json:"shard,omitempty"`
}

// CreateGroupDMOptions contains parameters for creating a new group DM channel.
//
// Reference: https://discord.com/developers/docs/resources/user#create-group-dm
type CreateGroupDMOptions struct {
	// AccessTokens are access tokens of users that have granted your app the 'gdm.join' scope.
	AccessTokens []string `json:"access_tokens"`

	// Nicks is a dictionary of user IDs to their respective nicknames.
	Nicks map[Snowflake]string `json:"nicks,omitempty"`
}

// UpdateApplicationRoleConnectionOptions contains parameters for updating a user's application role connection.
//
// Reference: https://discord.com/developers/docs/resources/user#update-current-user-application-role-connection
type UpdateApplicationRoleConnectionOptions struct {
	// PlatformName is the vanity name of the platform a bot has connected (max 50 characters).
	PlatformName *string `json:"platform_name,omitempty"`

	// PlatformUsername is the username on the platform a bot has connected (max 100 characters).
	PlatformUsername *string `json:"platform_username,omitempty"`

	// Metadata is object mapping application role connection metadata keys to their string-ified values.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// FetchCurrentUser returns the user object of the requester's account.
//
// Reference: https://discord.com/developers/docs/resources/user#get-current-user
func (r *requester) FetchCurrentUser(ctx context.Context) (User, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/users/@me",
	})
	if err != nil {
		return User{}, err
	}
	defer body.Close()

	var user User
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/users/@me",
			"error", err.Error(),
		)
		return User{}, err
	}
	return user, nil
}

// FetchUser returns a user object for a given user ID.
//
// Reference: https://discord.com/developers/docs/resources/user#get-user
func (r *requester) FetchUser(ctx context.Context, userID Snowflake) (User, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/users/" + userID.String(),
	})
	if err != nil {
		return User{}, err
	}
	defer body.Close()

	var user User
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/users/{id}",
				"error", err.Error(),
			)
		return User{}, err
	}
	return user, nil
}

// ModifyCurrentUser modifies the requester's user account settings.
//
// Reference: https://discord.com/developers/docs/resources/user#modify-current-user
func (r *requester) ModifyCurrentUser(ctx context.Context, opts ModifyCurrentUserOptions) (User, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/users/@me",
		Body:   reqBody,
	})
	if err != nil {
		return User{}, err
	}
	defer body.Close()

	var user User
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "PATCH",
			"url", "/users/@me",
			"error", err.Error(),
		)
		return User{}, err
	}
	return user, nil
}

// FetchCurrentUserGuilds returns a list of partial guild objects the current user is a member of.
//
// Reference: https://discord.com/developers/docs/resources/user#get-current-user-guilds
func (r *requester) FetchCurrentUserGuilds(ctx context.Context, opts ...FetchCurrentUserGuildsOptions) ([]PartialGuild, error) {
	var first FetchCurrentUserGuildsOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	params := url.Values{}
	if first.Before != 0 {
		params.Set("before", first.Before.String())
	}
	if first.After != 0 {
		params.Set("after", first.After.String())
	}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if first.WithCounts {
		params.Set("with_counts", "true")
	}
	if first.Shard != nil {
		params.Set("shard", strconv.Itoa(*first.Shard))
	}

	endpoint := "/users/@me/guilds"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var guilds []PartialGuild
	if err := json.NewDecoder(body).Decode(&guilds); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/users/@me/guilds",
			"error", err.Error(),
		)
		return nil, err
	}
	return guilds, nil
}

// FetchCurrentUserGuildMember returns a guild member object for the current user in a guild.
//
// Reference: https://discord.com/developers/docs/resources/user#get-current-user-guild-member
func (r *requester) FetchCurrentUserGuildMember(ctx context.Context, guildID Snowflake) (FullMember, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/users/@me/guilds/" + guildID.String() + "/member",
	})
	if err != nil {
		return FullMember{}, err
	}
	defer body.Close()

	var member FullMember
	if err := json.NewDecoder(body).Decode(&member); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/users/@me/guilds/{id}/member",
				"error", err.Error(),
			)
		return FullMember{}, err
	}
	member.GuildID = guildID
	if member.ID == 0 && member.User.ID != 0 {
		member.ID = member.User.ID
	}
	return member, nil
}

// LeaveGuild leaves a guild.
//
// Reference: https://discord.com/developers/docs/resources/user#leave-guild
func (r *requester) LeaveGuild(ctx context.Context, guildID Snowflake) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/users/@me/guilds/" + guildID.String(),
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// CreateDM creates a new direct message channel with a user.
//
// Reference: https://discord.com/developers/docs/resources/user#create-dm
func (r *requester) CreateDM(ctx context.Context, recipientID Snowflake) (*DMChannel, error) {
	reqBody, _ := json.Marshal(map[string]Snowflake{"recipient_id": recipientID})
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/users/@me/channels",
		Body:   reqBody,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	channel, err := UnmarshalChannel(bodyBytes)
	if err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "POST",
			"url", "/users/@me/channels",
			"error", err.Error(),
		)
		return nil, err
	}
	if dm, ok := channel.(*DMChannel); ok {
		return dm, nil
	}
	return nil, fmt.Errorf("CreateDM: created channel is not a DMChannel")
}

// CreateGroupDM creates a new group direct message channel with multiple users.
//
// Reference: https://discord.com/developers/docs/resources/user#create-group-dm
func (r *requester) CreateGroupDM(ctx context.Context, opts CreateGroupDMOptions) (*GroupDMChannel, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/users/@me/channels",
		Body:   reqBody,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	channel, err := UnmarshalChannel(bodyBytes)
	if err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "POST",
			"url", "/users/@me/channels",
			"error", err.Error(),
		)
		return nil, err
	}
	if gdm, ok := channel.(*GroupDMChannel); ok {
		return gdm, nil
	}
	return nil, fmt.Errorf("CreateGroupDM: created channel is not a GroupDMChannel")
}

// FetchCurrentUserConnections returns a list of connection objects for the current user.
//
// Reference: https://discord.com/developers/docs/resources/user#get-current-user-connections
func (r *requester) FetchCurrentUserConnections(ctx context.Context) ([]Connection, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/users/@me/connections",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var connections []Connection
	if err := json.NewDecoder(body).Decode(&connections); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/users/@me/connections",
			"error", err.Error(),
		)
		return nil, err
	}
	return connections, nil
}

// FetchCurrentUserApplicationRoleConnection returns the application role connection for the user.
//
// Reference: https://discord.com/developers/docs/resources/user#get-current-user-application-role-connection
func (r *requester) FetchCurrentUserApplicationRoleConnection(ctx context.Context, applicationID Snowflake) (ApplicationRoleConnection, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/users/@me/applications/" + applicationID.String() + "/role-connection",
	})
	if err != nil {
		return ApplicationRoleConnection{}, err
	}
	defer body.Close()

	var roleConn ApplicationRoleConnection
	if err := json.NewDecoder(body).Decode(&roleConn); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/users/@me/applications/{id}/role-connection",
				"error", err.Error(),
			)
		return ApplicationRoleConnection{}, err
	}
	return roleConn, nil
}

// UpdateCurrentUserApplicationRoleConnection updates and returns the application role connection for the user.
//
// Reference: https://discord.com/developers/docs/resources/user#update-current-user-application-role-connection
func (r *requester) UpdateCurrentUserApplicationRoleConnection(ctx context.Context, applicationID Snowflake, opts UpdateApplicationRoleConnectionOptions) (ApplicationRoleConnection, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    "/users/@me/applications/" + applicationID.String() + "/role-connection",
		Body:   reqBody,
	})
	if err != nil {
		return ApplicationRoleConnection{}, err
	}
	defer body.Close()

	var roleConn ApplicationRoleConnection
	if err := json.NewDecoder(body).Decode(&roleConn); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PUT",
				"url", "/users/@me/applications/{id}/role-connection",
				"error", err.Error(),
			)
		return ApplicationRoleConnection{}, err
	}
	return roleConn, nil
}

// FetchCurrentUserWithToken retrieves the user object of the requester using a user Bearer access token.
//
// Reference: https://discord.com/developers/docs/resources/user#get-current-user
func (r *requester) FetchCurrentUserWithToken(ctx context.Context, accessToken string) (User, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/users/@me",
		Headers: map[string]string{
			"Authorization": "Bearer " + accessToken,
		},
		NoAuth: true,
	})
	if err != nil {
		return User{}, err
	}
	defer body.Close()

	var user User
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/users/@me",
			"error", err.Error(),
		)
		return User{}, err
	}
	return user, nil
}

// FetchCurrentUserGuildsWithToken returns a list of partial guild objects the current user is a member of using a user Bearer token.
//
// Reference: https://discord.com/developers/docs/resources/user#get-current-user-guilds
func (r *requester) FetchCurrentUserGuildsWithToken(ctx context.Context, accessToken string, opts ...FetchCurrentUserGuildsOptions) ([]PartialGuild, error) {
	endpoint := "/users/@me/guilds"
	if len(opts) > 0 {
		query := url.Values{}
		if opts[0].Before != 0 {
			query.Set("before", opts[0].Before.String())
		}
		if opts[0].After != 0 {
			query.Set("after", opts[0].After.String())
		}
		if opts[0].Limit > 0 {
			query.Set("limit", strconv.Itoa(opts[0].Limit))
		}
		if opts[0].WithCounts {
			query.Set("with_counts", "true")
		}
		if opts[0].Shard != nil {
			query.Set("shard", strconv.Itoa(*opts[0].Shard))
		}
		if q := query.Encode(); q != "" {
			endpoint += "?" + q
		}
	}

	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
		Headers: map[string]string{
			"Authorization": "Bearer " + accessToken,
		},
		NoAuth: true,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var guilds []PartialGuild
	if err := json.NewDecoder(body).Decode(&guilds); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/users/@me/guilds",
			"error", err.Error(),
		)
		return nil, err
	}
	return guilds, nil
}
