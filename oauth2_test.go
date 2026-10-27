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

package wumpkit_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wumpkit/wumpkit"
)

func TestOAuth2_GenerateAuthorizeURL(t *testing.T) {
	integrationType := wumpkit.ApplicationIntegrationTypeGuildInstall
	authURL := wumpkit.GenerateOAuth2AuthorizeURL(wumpkit.OAuth2AuthorizeURLOptions{
		ClientID:           123456789012345678,
		RedirectURI:        "https://example.com/callback",
		Scopes:             []wumpkit.OAuth2Scope{wumpkit.OAuth2ScopeBot, wumpkit.OAuth2ScopeApplicationsCommands, wumpkit.OAuth2ScopeIdentify},
		State:              "secret_csrf_state",
		Permissions:        wumpkit.PermissionSendMessages | wumpkit.PermissionViewChannel,
		Prompt:             "consent",
		GuildID:            987654321098765432,
		DisableGuildSelect: true,
		IntegrationType:    &integrationType,
	})

	parsed, err := url.Parse(authURL)
	require.NoError(t, err)

	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "discord.com", parsed.Host)
	assert.Equal(t, "/oauth2/authorize", parsed.Path)

	q := parsed.Query()
	assert.Equal(t, "123456789012345678", q.Get("client_id"))
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "https://example.com/callback", q.Get("redirect_uri"))
	assert.Equal(t, "bot applications.commands identify", q.Get("scope"))
	assert.Equal(t, "secret_csrf_state", q.Get("state"))
	assert.Equal(t, "3072", q.Get("permissions")) // sendMessages (2048) + ViewChannel (1024) = 3072
	assert.Equal(t, "consent", q.Get("prompt"))
	assert.Equal(t, "987654321098765432", q.Get("guild_id"))
	assert.Equal(t, "true", q.Get("disable_guild_select"))
	assert.Equal(t, "0", q.Get("integration_type"))
}

func TestOAuth2_GenerateBotInviteURL(t *testing.T) {
	inviteURL := wumpkit.GenerateBotInviteURL(
		123456789012345678,
		wumpkit.PermissionAdministrator,
	)

	parsed, err := url.Parse(inviteURL)
	require.NoError(t, err)

	q := parsed.Query()
	assert.Equal(t, "123456789012345678", q.Get("client_id"))
	assert.Equal(t, "bot applications.commands", q.Get("scope"))
	assert.Equal(t, "8", q.Get("permissions")) // permissionAdministrator = 8
}

func TestOAuth2_ExchangeOAuth2Code(t *testing.T) {
	var capturedForm url.Values
	var capturedPath, capturedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		bodyBytes, _ := io.ReadAll(r.Body)
		capturedForm, _ = url.ParseQuery(string(bodyBytes))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wumpkit.OAuth2TokenResponse{
			AccessToken:  "mock_access_token_123",
			TokenType:    "Bearer",
			ExpiresIn:    604800,
			RefreshToken: "mock_refresh_token_456",
			Scope:        "identify guilds",
		})
	}))
	defer server.Close()

	client := wumpkit.New(context.Background(), wumpkit.WithRESTConfig(wumpkit.RequesterConfig{BaseURL: server.URL}))

	resp, err := client.ExchangeOAuth2Code(context.Background(), 111222333, "client_secret_xyz", "auth_code_789", "https://app.local/callback")
	require.NoError(t, err)

	assert.Equal(t, "POST", capturedMethod)
	assert.Equal(t, "/oauth2/token", capturedPath)
	assert.Equal(t, "authorization_code", capturedForm.Get("grant_type"))
	assert.Equal(t, "auth_code_789", capturedForm.Get("code"))
	assert.Equal(t, "https://app.local/callback", capturedForm.Get("redirect_uri"))
	assert.Equal(t, "111222333", capturedForm.Get("client_id"))
	assert.Equal(t, "client_secret_xyz", capturedForm.Get("client_secret"))

	assert.Equal(t, "mock_access_token_123", resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, 604800, resp.ExpiresIn)
	assert.Equal(t, "mock_refresh_token_456", resp.RefreshToken)
	assert.Equal(t, []wumpkit.OAuth2Scope{wumpkit.OAuth2ScopeIdentify, wumpkit.OAuth2ScopeGuilds}, resp.ScopesList())
}

func TestOAuth2_RefreshOAuth2Token(t *testing.T) {
	var capturedForm url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		capturedForm, _ = url.ParseQuery(string(bodyBytes))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wumpkit.OAuth2TokenResponse{
			AccessToken:  "new_access_token",
			TokenType:    "Bearer",
			ExpiresIn:    604800,
			RefreshToken: "new_refresh_token",
			Scope:        "identify",
		})
	}))
	defer server.Close()

	client := wumpkit.New(context.Background(), wumpkit.WithRESTConfig(wumpkit.RequesterConfig{BaseURL: server.URL}))

	resp, err := client.RefreshOAuth2Token(context.Background(), 111222333, "client_secret_xyz", "old_refresh_token")
	require.NoError(t, err)

	assert.Equal(t, "refresh_token", capturedForm.Get("grant_type"))
	assert.Equal(t, "old_refresh_token", capturedForm.Get("refresh_token"))
	assert.Equal(t, "new_access_token", resp.AccessToken)
}

func TestOAuth2_ClientCredentialsGrant(t *testing.T) {
	var capturedForm url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		capturedForm, _ = url.ParseQuery(string(bodyBytes))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wumpkit.OAuth2TokenResponse{
			AccessToken: "client_creds_token",
			TokenType:   "Bearer",
			ExpiresIn:   604800,
			Scope:       "applications.commands.update",
		})
	}))
	defer server.Close()

	client := wumpkit.New(context.Background(), wumpkit.WithRESTConfig(wumpkit.RequesterConfig{BaseURL: server.URL}))

	resp, err := client.ClientCredentialsGrant(context.Background(), 111222333, "client_secret_xyz", wumpkit.OAuth2ScopeApplicationsCommandsUpdate)
	require.NoError(t, err)

	assert.Equal(t, "client_credentials", capturedForm.Get("grant_type"))
	assert.Equal(t, "applications.commands.update", capturedForm.Get("scope"))
	assert.Equal(t, "client_creds_token", resp.AccessToken)
}

func TestOAuth2_RevokeOAuth2Token(t *testing.T) {
	var capturedForm url.Values
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		bodyBytes, _ := io.ReadAll(r.Body)
		capturedForm, _ = url.ParseQuery(string(bodyBytes))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := wumpkit.New(context.Background(), wumpkit.WithRESTConfig(wumpkit.RequesterConfig{BaseURL: server.URL}))

	err := client.RevokeOAuth2Token(context.Background(), 111222333, "client_secret_xyz", "token_to_revoke", "access_token")
	require.NoError(t, err)

	assert.Equal(t, "/oauth2/token/revoke", capturedPath)
	assert.Equal(t, "token_to_revoke", capturedForm.Get("token"))
	assert.Equal(t, "access_token", capturedForm.Get("token_type_hint"))
}

func TestOAuth2_FetchCurrentAuthorization(t *testing.T) {
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wumpkit.CurrentAuthorizationInformation{
			Application: wumpkit.Application{
				ID:   123456,
				Name: "Test App",
			},
			Scopes:  []wumpkit.OAuth2Scope{wumpkit.OAuth2ScopeIdentify, wumpkit.OAuth2ScopeGuilds},
			Expires: time.Now().Add(24 * time.Hour),
			User: &wumpkit.User{
				ID:       7891011,
				Username: "authed_user",
			},
		})
	}))
	defer server.Close()

	client := wumpkit.New(context.Background(), wumpkit.WithRESTConfig(wumpkit.RequesterConfig{BaseURL: server.URL}))

	info, err := client.FetchCurrentAuthorization(context.Background(), "mock_user_bearer_token")
	require.NoError(t, err)

	assert.Equal(t, "Bearer mock_user_bearer_token", capturedAuth)
	assert.Equal(t, wumpkit.Snowflake(123456), info.Application.ID)
	assert.Equal(t, "Test App", info.Application.Name)
	assert.NotNil(t, info.User)
	assert.Equal(t, "authed_user", info.User.Username)
	assert.Contains(t, info.Scopes, wumpkit.OAuth2ScopeIdentify)
	assert.Contains(t, info.Scopes, wumpkit.OAuth2ScopeGuilds)
}
