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
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// WebhookEventType represents the type of outgoing webhook payload from Discord.
//
// Reference: https://discord.com/developers/docs/events/webhook-events#webhook-types
type WebhookEventType int

const (
	// WebhookEventTypePing is sent by Discord to verify that your Webhook Event URL is active.
	WebhookEventTypePing WebhookEventType = 0

	// WebhookEventTypeEvent contains event data in the Event object.
	WebhookEventTypeEvent WebhookEventType = 1
)

// WebhookPayload represents the outer payload for a Discord Webhook Event.
//
// Reference: https://discord.com/developers/docs/events/webhook-events#payload-structure
type WebhookPayload struct {
	Version       int               `json:"version"`
	ApplicationID Snowflake         `json:"application_id"`
	Type          WebhookEventType  `json:"type"`
	Event         *WebhookEventBody `json:"event,omitempty"`
}

// WebhookEventBody represents the inner event payload of a Discord Webhook Event.
//
// Reference: https://discord.com/developers/docs/events/webhook-events#event-body-object
type WebhookEventBody struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// WebhookOnlyEvent represents Discord events that are exclusively sent over Outgoing HTTP Webhooks
// and are never received over the Gateway WebSocket connection.
//
// Reference: https://discord.com/developers/docs/events/webhook-events#event-types
type WebhookOnlyEvent string

const (
	// WebhookOnlyEventApplicationAuthorized is sent when an app was authorized by a user to a server or account.
	WebhookOnlyEventApplicationAuthorized WebhookOnlyEvent = "APPLICATION_AUTHORIZED"

	// WebhookOnlyEventApplicationDeauthorized is sent when an app was deauthorized by a user.
	WebhookOnlyEventApplicationDeauthorized WebhookOnlyEvent = "APPLICATION_DEAUTHORIZED"

	// WebhookOnlyEventQuestUserEnrollment is sent when a user was added to a Quest on Discord.
	WebhookOnlyEventQuestUserEnrollment WebhookOnlyEvent = "QUEST_USER_ENROLLMENT"

	// WebhookOnlyEventLobbyMessageCreate is sent when a message is created in a Discord Social SDK lobby.
	WebhookOnlyEventLobbyMessageCreate WebhookOnlyEvent = "LOBBY_MESSAGE_CREATE"

	// WebhookOnlyEventLobbyMessageUpdate is sent when a message is updated in a Discord Social SDK lobby.
	WebhookOnlyEventLobbyMessageUpdate WebhookOnlyEvent = "LOBBY_MESSAGE_UPDATE"

	// WebhookOnlyEventLobbyMessageDelete is sent when a message is deleted from a Discord Social SDK lobby.
	WebhookOnlyEventLobbyMessageDelete WebhookOnlyEvent = "LOBBY_MESSAGE_DELETE"

	// WebhookOnlyEventGameDirectMessageCreate is sent when an in-game direct message is created during a Social SDK session.
	WebhookOnlyEventGameDirectMessageCreate WebhookOnlyEvent = "GAME_DIRECT_MESSAGE_CREATE"

	// WebhookOnlyEventGameDirectMessageUpdate is sent when an in-game direct message is updated during a Social SDK session.
	WebhookOnlyEventGameDirectMessageUpdate WebhookOnlyEvent = "GAME_DIRECT_MESSAGE_UPDATE"

	// WebhookOnlyEventGameDirectMessageDelete is sent when an in-game direct message is deleted during a Social SDK session.
	WebhookOnlyEventGameDirectMessageDelete WebhookOnlyEvent = "GAME_DIRECT_MESSAGE_DELETE"
)

// Is returns true if this WebhookOnlyEvent matches the provided one.
func (w WebhookOnlyEvent) Is(other WebhookOnlyEvent) bool {
	return w == other
}

// WebhookEvent represents all events that can be received over Discord Webhook Events (HTTP),
// including webhook-exclusive events, monetization entitlements, and interactions.
type WebhookEvent string

const (
	WebhookEventApplicationAuthorized   WebhookEvent = "APPLICATION_AUTHORIZED"
	WebhookEventApplicationDeauthorized WebhookEvent = "APPLICATION_DEAUTHORIZED"
	WebhookEventEntitlementCreate       WebhookEvent = "ENTITLEMENT_CREATE"
	WebhookEventEntitlementUpdate       WebhookEvent = "ENTITLEMENT_UPDATE"
	WebhookEventEntitlementDelete       WebhookEvent = "ENTITLEMENT_DELETE"
	WebhookEventQuestUserEnrollment     WebhookEvent = "QUEST_USER_ENROLLMENT"
	WebhookEventLobbyMessageCreate      WebhookEvent = "LOBBY_MESSAGE_CREATE"
	WebhookEventLobbyMessageUpdate      WebhookEvent = "LOBBY_MESSAGE_UPDATE"
	WebhookEventLobbyMessageDelete      WebhookEvent = "LOBBY_MESSAGE_DELETE"
	WebhookEventGameDirectMessageCreate WebhookEvent = "GAME_DIRECT_MESSAGE_CREATE"
	WebhookEventGameDirectMessageUpdate WebhookEvent = "GAME_DIRECT_MESSAGE_UPDATE"
	WebhookEventGameDirectMessageDelete WebhookEvent = "GAME_DIRECT_MESSAGE_DELETE"
	WebhookEventInteractionCreate       WebhookEvent = "INTERACTION_CREATE"
)

// Is returns true if this WebhookEvent matches the provided one.
func (w WebhookEvent) Is(other WebhookEvent) bool {
	return w == other
}

// IsExclusive returns true if this event is exclusive to Webhook Events (never sent over Gateway).
func (w WebhookEvent) IsExclusive() bool {
	switch w {
	case WebhookEventApplicationAuthorized,
		WebhookEventApplicationDeauthorized,
		WebhookEventQuestUserEnrollment,
		WebhookEventLobbyMessageCreate,
		WebhookEventLobbyMessageUpdate,
		WebhookEventLobbyMessageDelete,
		WebhookEventGameDirectMessageCreate,
		WebhookEventGameDirectMessageUpdate,
		WebhookEventGameDirectMessageDelete:
		return true
	default:
		return false
	}
}

// WithPublicKey sets the Discord Application Public Key used to verify Webhook Events and HTTP interactions.
// You can find your Public Key in the Discord Developer Portal under General Information.
//
// Usage:
//
//	bot := wumpkit.New(ctx, wumpkit.WithToken(token), wumpkit.WithPublicKey("your_public_key_hex"))
func WithPublicKey(publicKeyHex string) clientOption {
	publicKeyHex = strings.TrimSpace(publicKeyHex)
	return func(cfg *clientConfig) {
		cfg.publicKey = publicKeyHex
		cfg.webhookEventsConfig.PublicKey = publicKeyHex
	}
}

// WithWebhookEventsPublicKey is an explicit alias for WithPublicKey.
func WithWebhookEventsPublicKey(publicKeyHex string) clientOption {
	return WithPublicKey(publicKeyHex)
}

// PublicKey returns the application public key configured on the client.
func (c *Client) PublicKey() string {
	return c.publicKey
}

// VerifyWebhookEvent verifies an incoming Discord Webhook Event request using ED25519 cryptography.
//
// Parameters:
//   - publicKeyHex: The 32-byte hex-encoded public key from the Discord Developer Portal.
//   - timestamp: The value of the X-Signature-Timestamp header.
//   - body: The raw HTTP request body bytes.
//   - signatureHex: The 64-byte hex-encoded signature from the X-Signature-Ed25519 header.
func VerifyWebhookEvent(publicKeyHex, timestamp string, body []byte, signatureHex string) bool {
	pubKeyBytes, err := hex.DecodeString(strings.TrimSpace(publicKeyHex))
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return false
	}
	return VerifyWebhookEventWithKey(ed25519.PublicKey(pubKeyBytes), timestamp, body, signatureHex)
}

// VerifyWebhookEventWithKey verifies an incoming Discord Webhook Event request using a pre-parsed ed25519.PublicKey.
func VerifyWebhookEventWithKey(pubKey ed25519.PublicKey, timestamp string, body []byte, signatureHex string) bool {
	if len(pubKey) != ed25519.PublicKeySize || timestamp == "" || len(body) == 0 || signatureHex == "" {
		return false
	}
	sigBytes, err := hex.DecodeString(strings.TrimSpace(signatureHex))
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false
	}

	msg := make([]byte, 0, len(timestamp)+len(body))
	msg = append(msg, timestamp...)
	msg = append(msg, body...)

	return ed25519.Verify(pubKey, msg, sigBytes)
}

// VerifyInteraction is an alias for VerifyWebhookEvent.
func VerifyInteraction(publicKeyHex, timestamp string, body []byte, signatureHex string) bool {
	return VerifyWebhookEvent(publicKeyHex, timestamp, body, signatureHex)
}

// VerifyInteractionWithKey is an alias for VerifyWebhookEventWithKey.
func VerifyInteractionWithKey(pubKey ed25519.PublicKey, timestamp string, body []byte, signatureHex string) bool {
	return VerifyWebhookEventWithKey(pubKey, timestamp, body, signatureHex)
}

// WebhookEventsHandler returns an http.Handler that processes incoming Discord Webhook Events
// and HTTP interactions, validates ED25519 signatures, automatically responds to PING requests,
// and dispatches events to your client event handlers via dispatcher.
func (c *Client) WebhookEventsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// read signature headers
		signature := r.Header.Get("X-Signature-Ed25519")
		timestamp := r.Header.Get("X-Signature-Timestamp")
		if signature == "" || timestamp == "" {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		// read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed reading body", http.StatusBadRequest)
			return
		}

		// verify signature against client's public key
		if len(c.publicKeyBytes) != ed25519.PublicKeySize && c.publicKey != "" {
			if b, err := hex.DecodeString(c.publicKey); err == nil && len(b) == ed25519.PublicKeySize {
				c.publicKeyBytes = ed25519.PublicKey(b)
			}
		}

		if len(c.publicKeyBytes) != ed25519.PublicKeySize {
			if c.logger != nil {
				c.logger.Error("webhook event received but client has no valid PublicKey configured")
			}
			http.Error(w, "server configuration error: public key missing", http.StatusInternalServerError)
			return
		}

		if !VerifyWebhookEventWithKey(c.publicKeyBytes, timestamp, body, signature) {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		// check for Discord Webhook Event payload structure (outer wrapper)
		var webhookPayload struct {
			Version       int               `json:"version"`
			ApplicationID Snowflake         `json:"application_id"`
			Type          int               `json:"type"`
			Event         *WebhookEventBody `json:"event,omitempty"`
		}
		if err := json.Unmarshal(body, &webhookPayload); err == nil && webhookPayload.Version == 1 && webhookPayload.ApplicationID != 0 {
			// discord webhook event
			if webhookPayload.Type == int(WebhookEventTypePing) {
				// ping for webhook events responds with 204 No Content
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if webhookPayload.Type == int(WebhookEventTypeEvent) && webhookPayload.Event != nil {
				// dispatch specific event type to dispatcher
				if c.dispatcher != nil {
					payloadData := webhookPayload.Event.Data
					if len(payloadData) == 0 {
						payloadData = body
					}
					c.dispatcher.Dispatch(-1, webhookPayload.Event.Type, payloadData)
				}
				// webhook events acknowledge with 204 No Content
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		// check for Discord HTTP Interaction payload structure
		var rawInteraction struct {
			Type InteractionType `json:"type"`
		}
		if err := json.Unmarshal(body, &rawInteraction); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		// handle type 1 PING for HTTP Interactions
		if rawInteraction.Type == InteractionTypePing {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"type":1}`))
			return
		}

		// dispatch interaction to event handlers
		if c.dispatcher != nil {
			c.dispatcher.Dispatch(-1, "INTERACTION_CREATE", body)
		}

		// acknowledge http interaction with deferred response to satisfy discord protocol
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		switch rawInteraction.Type {
		case InteractionTypeComponent:
			_, _ = w.Write([]byte(`{"type":6}`))
		case InteractionTypeAutocomplete:
			_, _ = w.Write([]byte(`{"type":8,"data":{"choices":[]}}`))
		default:
			_, _ = w.Write([]byte(`{"type":5}`))
		}
	})
}

// HTTPHandler is an alias for WebhookEventsHandler.
func (c *Client) HTTPHandler() http.Handler {
	return c.WebhookEventsHandler()
}

// WebhookEventsStartOption configures settings when starting a Webhook Events server.
type WebhookEventsStartOption func(*WebhookEventsConfig)

// HTTPStartOption is an alias for WebhookEventsStartOption.
type HTTPStartOption = WebhookEventsStartOption

// WebhookEventsConfig holds configuration parameters for receiving Discord Webhook Events over HTTP.
type WebhookEventsConfig struct {
	// Path is the URL route path to listen for webhook events (default: "/events").
	Path string
	// PublicKey is the hex-encoded ED25519 public key used to verify Discord webhook signatures.
	PublicKey string
	// ReadTimeout is the maximum duration for reading the entire request (default: 10s).
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response (default: 10s).
	WriteTimeout time.Duration
	// IdleTimeout is the maximum amount of time to wait for the next request (default: 60s).
	IdleTimeout time.Duration
	// Server is an optional pre-configured *http.Server instance.
	Server *http.Server
}

// DefaultWebhookEventsConfig returns standard Webhook Events configuration defaults.
func DefaultWebhookEventsConfig() WebhookEventsConfig {
	return WebhookEventsConfig{
		Path:         "/events",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// WithWebhookEventsPath sets the URL path for the webhook events endpoint (default: "/events").
func WithWebhookEventsPath(path string) WebhookEventsStartOption {
	return func(cfg *WebhookEventsConfig) {
		cfg.Path = path
	}
}

// WithWebhookEventsReadTimeout sets the maximum duration for reading the entire request.
func WithWebhookEventsReadTimeout(d time.Duration) WebhookEventsStartOption {
	return func(cfg *WebhookEventsConfig) {
		cfg.ReadTimeout = d
	}
}

// WithWebhookEventsWriteTimeout sets the maximum duration before timing out writes of the response.
func WithWebhookEventsWriteTimeout(d time.Duration) WebhookEventsStartOption {
	return func(cfg *WebhookEventsConfig) {
		cfg.WriteTimeout = d
	}
}

// WithWebhookEventsIdleTimeout sets the maximum amount of time to wait for the next request.
func WithWebhookEventsIdleTimeout(d time.Duration) WebhookEventsStartOption {
	return func(cfg *WebhookEventsConfig) {
		cfg.IdleTimeout = d
	}
}

// WithWebhookEventsCustomServer specifies a custom http.Server to use.
func WithWebhookEventsCustomServer(srv *http.Server) WebhookEventsStartOption {
	return func(cfg *WebhookEventsConfig) {
		cfg.Server = srv
	}
}

// StartWebhookEvents starts a dedicated HTTP server to receive Discord Webhook Events on the specified address.
// Blocks until client context is cancelled or Shutdown() is called.
//
// Usage:
//
//	err := bot.StartWebhookEvents(":8080", wumpkit.WithWebhookEventsPath("/events"))
func (c *Client) StartWebhookEvents(addr string, options ...WebhookEventsStartOption) error {
	if !c.isWebhookEventsRunning.CompareAndSwap(false, true) {
		return errors.New("cannot start webhook events server: webhook events server is already running on this client")
	}
	defer c.isWebhookEventsRunning.Store(false)

	cfg := DefaultWebhookEventsConfig()
	if c.publicKey != "" {
		cfg.PublicKey = c.publicKey
	}
	for _, opt := range options {
		opt(&cfg)
	}

	if cfg.PublicKey != "" {
		c.publicKey = cfg.PublicKey
		if b, err := hex.DecodeString(cfg.PublicKey); err == nil && len(b) == ed25519.PublicKeySize {
			c.publicKeyBytes = ed25519.PublicKey(b)
		}
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, c.WebhookEventsHandler())

	server := cfg.Server
	if server == nil {
		server = &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		}
	} else {
		if server.Addr == "" {
			server.Addr = addr
		}
		if server.Handler == nil {
			server.Handler = mux
		}
	}
	c.httpServer = server

	errChan := make(chan error, 1)
	go func() {
		if c.logger != nil {
			c.logger.Info("starting webhook events server", "addr", addr, "path", cfg.Path)
		}
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-c.ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		return nil
	}
}

// StartWebhookEventsWithSignals starts the webhook events server and blocks until an interrupt or termination signal
// (SIGINT, SIGTERM) is received, then gracefully shuts down the server.
func (c *Client) StartWebhookEventsWithSignals(addr string, options ...WebhookEventsStartOption) error {
	sigCtx, stop := signal.NotifyContext(c.ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 1)
	go func() {
		errChan <- c.StartWebhookEvents(addr, options...)
	}()

	select {
	case err := <-errChan:
		return err
	case <-sigCtx.Done():
		if c.logger != nil {
			c.logger.Info("OS signal received, shutting down webhook events server")
		}
		c.Shutdown()
		return <-errChan
	}
}
