package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/wumpkit/wumpkit"
)

func main() {
	clientIDStr := os.Getenv("DISCORD_CLIENT_ID")
	clientSecret := os.Getenv("DISCORD_CLIENT_SECRET")
	redirectURI := os.Getenv("DISCORD_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/callback"
	}

	if clientIDStr == "" || clientSecret == "" {
		log.Fatal("DISCORD_CLIENT_ID and DISCORD_CLIENT_SECRET environment variables are required")
	}

	clientIDNum, err := strconv.ParseUint(clientIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid DISCORD_CLIENT_ID: %v", err)
	}
	clientID := wumpkit.Snowflake(clientIDNum)

	// create client
	client := wumpkit.New(context.Background())

	// route 1: Home page with login and invite links
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loginURL := wumpkit.GenerateOAuth2AuthorizeURL(wumpkit.OAuth2AuthorizeURLOptions{
			ClientID:    clientID,
			RedirectURI: redirectURI,
			Scopes:      []wumpkit.OAuth2Scope{wumpkit.OAuth2ScopeIdentify, wumpkit.OAuth2ScopeGuilds},
			State:       "random_security_state",
			Prompt:      "consent",
		})

		inviteURL := wumpkit.GenerateBotInviteURL(
			clientID,
			wumpkit.PermissionAdministrator,
			wumpkit.OAuth2ScopeBot,
			wumpkit.OAuth2ScopeApplicationsCommands,
		)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
			<h1>Discord OAuth2 Example</h1>
			<p><a href="%s">Login with Discord</a></p>
			<p><a href="%s">Invite Bot to Server</a></p>
		`, loginURL, inviteURL)
	})

	// route 2: OAuth2 Callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code query param", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// exchange authorization code for OAuth2 tokens
		tokens, err := client.ExchangeOAuth2Code(ctx, clientID, clientSecret, code, redirectURI)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed exchanging code: %v", err), http.StatusInternalServerError)
			return
		}

		// fetch user authorization information using the access token
		authInfo, err := client.FetchCurrentAuthorization(ctx, tokens.AccessToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed fetching authorization: %v", err), http.StatusInternalServerError)
			return
		}

		var username string = "Unknown"
		var userID string
		if authInfo.User != nil {
			username = authInfo.User.Username
			userID = authInfo.User.ID.String()
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
			<h2>Authentication Successful</h2>
			<p><strong>User:</strong> %s (ID: %s)</p>
			<p><strong>Scopes Authorized:</strong> %v</p>
			<p><strong>Access Token Expiry:</strong> %d seconds</p>
			<p><a href="/">Back to Home</a></p>
		`, username, userID, authInfo.Scopes, tokens.ExpiresIn)
	})

	addr := ":8080"
	log.Printf("OAuth2 Web Server running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
