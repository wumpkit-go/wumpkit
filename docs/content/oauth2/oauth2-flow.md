---
title: "OAuth2 Flow"
weight: 1
---

# Implementing OAuth2

Discord OAuth2 lets your website or web service authenticate Discord users and access their servers and profile information.

## 1. Generating Login and Invite URLs

Use `GenerateOAuth2AuthorizeURL` to create standard login URLs for web users, and `GenerateBotInviteURL` to create clean bot invitation links:

```go
package main

import (
	"fmt"

	"github.com/wumpkit/wumpkit"
)

func buildLinks(appID wumpkit.Snowflake) {
	// url for users logging into your web dashboard
	loginURL := wumpkit.GenerateOAuth2AuthorizeURL(wumpkit.OAuth2AuthorizeURLOptions{
		ClientID:    appID,
		RedirectURI: "https://example.com/api/callback",
		Scopes: []wumpkit.OAuth2Scope{
			wumpkit.OAuth2ScopeIdentify,
			wumpkit.OAuth2ScopeGuilds,
		},
		State:  "secure_random_state_string",
		Prompt: "consent",
	})

	// url for adding the bot to a server with specific permissions
	inviteURL := wumpkit.GenerateBotInviteURL(
		appID,
		wumpkit.PermissionSendMessages|wumpkit.PermissionEmbedLinks,
		wumpkit.OAuth2ScopeBot,
		wumpkit.OAuth2ScopeApplicationsCommands,
	)

	fmt.Printf("Login URL: %s\n", loginURL)
	fmt.Printf("Bot Invite URL: %s\n", inviteURL)
}
```

## 2. Handling the OAuth2 Callback

When a user approves access, Discord redirects back to your `redirect_uri` with a `?code=` query parameter. Exchange this code for access and refresh tokens:

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wumpkit/wumpkit"
)

func handleCallback(w http.ResponseWriter, r *http.Request, client *wumpkit.Client, clientID wumpkit.Snowflake, clientSecret, redirectURI string) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// exchange authorization code for access and refresh tokens
	tokens, err := client.ExchangeOAuth2Code(ctx, clientID, clientSecret, code, redirectURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	// fetch authorized user identity using the access token
	authInfo, err := client.FetchCurrentAuthorization(ctx, tokens.AccessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed fetching user info: %v", err), http.StatusInternalServerError)
		return
	}

	if authInfo.User != nil {
		fmt.Fprintf(w, "Welcome, %s (ID: %s)!\n", authInfo.User.Username, authInfo.User.ID)
		fmt.Fprintf(w, "Token expires in %d seconds.\n", tokens.ExpiresIn)
	}
}
```

## 3. Refreshing Expired Tokens

Access tokens typically expire after 7 days. Use the `RefreshToken` to request a fresh token without requiring the user to log in again:

```go
newTokens, err := client.RefreshOAuth2Token(ctx, clientID, clientSecret, storedRefreshToken)
if err != nil {
	// refresh token expired or was revoked by user
	return err
}

_ = newTokens.AccessToken
_ = newTokens.RefreshToken
```
