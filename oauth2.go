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
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// OAuth2AuthorizeURLOptions configures the generation of a Discord OAuth2 authorization URL.
//
// Reference: https://discord.com/developers/docs/topics/oauth2#authorization-code-grant-authorization-url-example
type OAuth2AuthorizeURLOptions struct {
	// ClientID is the application's client ID.
	ClientID Snowflake

	// RedirectURI is the URL the user will be redirected to after authorization.
	RedirectURI string

	// Scopes is the list of OAuth2 scopes requested.
	Scopes []OAuth2Scope

	// State is an optional state string to prevent CSRF attacks.
	State string

	// Permissions is the bitwise permissions value to request (used when scope includes 'bot').
	Permissions Permissions

	// Prompt controls whether the user is prompted for consent ("consent" or "none").
	Prompt string

	// GuildID pre-selects a guild for the user to add the bot to.
	GuildID Snowflake

	// DisableGuildSelect prevents the user from changing the selected guild.
	DisableGuildSelect bool

	// ResponseType is the response type ("code" for authorization code, "token" for implicit grant).
	// Default is "code".
	ResponseType string

	// IntegrationType specifies 0 for Guild Install, 1 for User Install.
	IntegrationType *ApplicationIntegrationType
}

// GenerateOAuth2AuthorizeURL builds an authorization URL with query parameters according to the Discord OAuth2 spec.
func GenerateOAuth2AuthorizeURL(opts OAuth2AuthorizeURLOptions) string {
	baseURL := "https://discord.com/oauth2/authorize"
	query := url.Values{}

	if opts.ClientID != 0 {
		query.Set("client_id", opts.ClientID.String())
	}

	responseType := "code"
	if opts.ResponseType != "" {
		responseType = opts.ResponseType
	}
	query.Set("response_type", responseType)

	if opts.RedirectURI != "" {
		query.Set("redirect_uri", opts.RedirectURI)
	}

	if len(opts.Scopes) > 0 {
		scopeStrs := make([]string, len(opts.Scopes))
		for i, s := range opts.Scopes {
			scopeStrs[i] = string(s)
		}
		query.Set("scope", strings.Join(scopeStrs, " "))
	}

	if opts.State != "" {
		query.Set("state", opts.State)
	}

	if opts.Permissions != 0 {
		query.Set("permissions", strconv.FormatInt(int64(opts.Permissions), 10))
	}

	if opts.Prompt != "" {
		query.Set("prompt", opts.Prompt)
	}

	if opts.GuildID != 0 {
		query.Set("guild_id", opts.GuildID.String())
	}

	if opts.DisableGuildSelect {
		query.Set("disable_guild_select", "true")
	}

	if opts.IntegrationType != nil {
		query.Set("integration_type", strconv.Itoa(int(*opts.IntegrationType)))
	}

	return fmt.Sprintf("%s?%s", baseURL, query.Encode())
}

// GenerateBotInviteURL generates a standard bot invite URL with permissions and scopes.
// By default, it requests the 'bot' and 'applications.commands' scopes if none are provided.
func GenerateBotInviteURL(
	clientID Snowflake,
	permissions Permissions,
	scopes ...OAuth2Scope,
) string {
	if len(scopes) == 0 {
		scopes = []OAuth2Scope{OAuth2ScopeBot, OAuth2ScopeApplicationsCommands}
	}
	return GenerateOAuth2AuthorizeURL(OAuth2AuthorizeURLOptions{
		ClientID:    clientID,
		Permissions: permissions,
		Scopes:      scopes,
	})
}

// OAuth2TokenResponse represents the response received from Discord after exchanging or refreshing an OAuth2 token.
//
// Reference: https://discord.com/developers/docs/topics/oauth2#authorization-code-grant-access-token-response
type OAuth2TokenResponse struct {
	// AccessToken is the token used to authorize API requests on behalf of the user.
	AccessToken string `json:"access_token"`

	// TokenType is the type of token (typically "Bearer").
	TokenType string `json:"token_type"`

	// ExpiresIn is the number of seconds until the access token expires.
	ExpiresIn int `json:"expires_in"`

	// RefreshToken is the token used to refresh the access token once expired.
	RefreshToken string `json:"refresh_token"`

	// Scope is the space-delimited list of scopes granted.
	Scope string `json:"scope"`

	// Webhook is returned when the 'webhook.incoming' scope is authorized.
	Webhook *Webhook `json:"webhook,omitempty"`

	// Guild is returned when the 'bot' scope is authorized and a guild is selected.
	Guild *OAuth2Guild `json:"guild,omitempty"`
}

// ScopesList returns the granted scopes as a slice of OAuth2Scope.
func (t OAuth2TokenResponse) ScopesList() []OAuth2Scope {
	parts := strings.Fields(t.Scope)
	scopes := make([]OAuth2Scope, len(parts))
	for i, p := range parts {
		scopes[i] = OAuth2Scope(p)
	}
	return scopes
}

// OAuth2Guild represents the partial guild object returned in an OAuth2 bot authorization response.
type OAuth2Guild struct {
	ID          Snowflake   `json:"id"`
	Name        string      `json:"name"`
	Icon        string      `json:"icon"`
	Owner       bool        `json:"owner"`
	Permissions Permissions `json:"permissions"`
	Features    []string    `json:"features"`
}

// CurrentAuthorizationInformation represents the metadata returned by GET /oauth2/@me.
//
// Reference: https://discord.com/developers/docs/topics/oauth2#get-current-authorization-information
type CurrentAuthorizationInformation struct {
	// Application contains the current application details.
	Application Application `json:"application"`

	// Scopes contains the scopes granted to the access token.
	Scopes []OAuth2Scope `json:"scopes"`

	// Expires is when the access token expires.
	Expires time.Time `json:"expires"`

	// User contains user details if the 'identify' scope was authorized.
	User *User `json:"user,omitempty"`
}
