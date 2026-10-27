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

package wumpkit_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wumpkit/wumpkit"
)

func TestShardID_IsWebhookEventsDispatched(t *testing.T) {
	assert.True(t, wumpkit.ShardID(-1).IsWebhookEventsDispatched())
	assert.False(t, wumpkit.ShardID(0).IsWebhookEventsDispatched())
}

func TestVerifyWebhookEvent(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	pubKeyHex := hex.EncodeToString(pubKey)
	timestamp := "1700000000"
	body := []byte(`{"version":1,"application_id":"123","type":0}`)

	msg := append([]byte(timestamp), body...)
	sig := ed25519.Sign(privKey, msg)
	sigHex := hex.EncodeToString(sig)

	// valid verification
	assert.True(t, wumpkit.VerifyWebhookEvent(pubKeyHex, timestamp, body, sigHex))
	assert.True(t, wumpkit.VerifyWebhookEventWithKey(pubKey, timestamp, body, sigHex))
	assert.True(t, wumpkit.VerifyInteraction(pubKeyHex, timestamp, body, sigHex))

	// tampered body
	assert.False(t, wumpkit.VerifyWebhookEvent(pubKeyHex, timestamp, []byte(`{"version":1,"type":1}`), sigHex))

	// tampered timestamp
	assert.False(t, wumpkit.VerifyWebhookEvent(pubKeyHex, "1700000001", body, sigHex))

	// invalid signature hex
	assert.False(t, wumpkit.VerifyWebhookEvent(pubKeyHex, timestamp, body, "invalid_sig_hex"))

	// invalid public key hex
	assert.False(t, wumpkit.VerifyWebhookEvent("invalid_pub_key", timestamp, body, sigHex))
}

func TestWebhookEventsHandler_PingAndEvents(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	pubKeyHex := hex.EncodeToString(pubKey)

	ctx := context.Background()
	client := wumpkit.New(
		ctx,
		wumpkit.WithToken("dummy_test_token"),
		wumpkit.WithPublicKey(pubKeyHex),
	)

	var (
		authorizedReceived   atomic.Bool
		deauthorizedReceived atomic.Bool
		interactionReceived  atomic.Bool
	)

	client.OnApplicationAuthorized(func(ctx context.Context, e wumpkit.ApplicationAuthorizedEvent) {
		if e.User.ID == 9988776655 {
			authorizedReceived.Store(true)
		}
	})

	client.OnApplicationDeauthorized(func(ctx context.Context, e wumpkit.ApplicationDeauthorizedEvent) {
		if e.User.ID == 9988776655 {
			deauthorizedReceived.Store(true)
		}
	})

	client.OnInteractionCreate(func(ctx context.Context, e wumpkit.InteractionCreateEvent) {
		if e.Interaction.GetID() == 123456789012345678 {
			interactionReceived.Store(true)
		}
	})

	handler := client.WebhookEventsHandler()

	signRequest := func(body []byte, ts string) (string, string) {
		msg := append([]byte(ts), body...)
		sig := ed25519.Sign(privKey, msg)
		return ts, hex.EncodeToString(sig)
	}

	// 1. method not allowed (GET)
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)

	// 2. missing signature headers
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer([]byte(`{"type":0}`)))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// 3. invalid signature
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer([]byte(`{"type":0}`)))
	req.Header.Set("X-Signature-Ed25519", "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	req.Header.Set("X-Signature-Timestamp", "1700000000")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// 4. valid Webhook Events PING (type: 0) -> responds with HTTP 204 No Content
	pingWebhookBody := []byte(`{"version":1,"application_id":"123","type":0}`)
	ts, sigHex := signRequest(pingWebhookBody, "1700000000")
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(pingWebhookBody))
	req.Header.Set("X-Signature-Ed25519", sigHex)
	req.Header.Set("X-Signature-Timestamp", ts)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// 5. valid Webhook Event: APPLICATION_AUTHORIZED -> responds with HTTP 204 No Content & dispatches
	authEventBody := []byte(`{"version":1,"application_id":"123","type":1,"event":{"type":"APPLICATION_AUTHORIZED","timestamp":"2024-10-18T14:42:53Z","data":{"integration_type":1,"scopes":["applications.commands"],"user":{"id":"9988776655","username":"tester"}}}}`)
	ts, sigHex = signRequest(authEventBody, "1700000001")
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(authEventBody))
	req.Header.Set("X-Signature-Ed25519", sigHex)
	req.Header.Set("X-Signature-Timestamp", ts)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Eventually(t, func() bool {
		return authorizedReceived.Load()
	}, 1*time.Second, 10*time.Millisecond)

	// 6. valid Webhook Event: APPLICATION_DEAUTHORIZED -> responds with HTTP 204 No Content & dispatches
	deauthEventBody := []byte(`{"version":1,"application_id":"123","type":1,"event":{"type":"APPLICATION_DEAUTHORIZED","timestamp":"2024-10-18T14:42:53Z","data":{"user":{"id":"9988776655","username":"tester"}}}}`)
	ts, sigHex = signRequest(deauthEventBody, "1700000002")
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(deauthEventBody))
	req.Header.Set("X-Signature-Ed25519", sigHex)
	req.Header.Set("X-Signature-Timestamp", ts)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Eventually(t, func() bool {
		return deauthorizedReceived.Load()
	}, 1*time.Second, 10*time.Millisecond)

	// 7. valid HTTP Interaction PING (type 1) -> responds with HTTP 200 {"type":1}
	pingInteractionBody := []byte(`{"type":1}`)
	ts, sigHex = signRequest(pingInteractionBody, "1700000003")
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(pingInteractionBody))
	req.Header.Set("X-Signature-Ed25519", sigHex)
	req.Header.Set("X-Signature-Timestamp", ts)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"type":1}`, rec.Body.String())

	// 8. valid HTTP Component Interaction (type 3) -> responds with HTTP 200 OK & dispatches
	componentBody := []byte(`{"id":"123456789012345678","type":3,"data":{"custom_id":"btn_click","component_type":2},"token":"xyz","version":1}`)
	ts, sigHex = signRequest(componentBody, "1700000004")
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(componentBody))
	req.Header.Set("X-Signature-Ed25519", sigHex)
	req.Header.Set("X-Signature-Timestamp", ts)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"type":6}`, rec.Body.String())
	assert.Eventually(t, func() bool {
		return interactionReceived.Load()
	}, 1*time.Second, 10*time.Millisecond)
}

func TestWebhookEvents_Enums(t *testing.T) {
	assert.True(t, wumpkit.WebhookEventApplicationAuthorized.IsExclusive())
	assert.True(t, wumpkit.WebhookEventApplicationDeauthorized.IsExclusive())
	assert.True(t, wumpkit.WebhookEventQuestUserEnrollment.IsExclusive())
	assert.True(t, wumpkit.WebhookEventLobbyMessageCreate.IsExclusive())
	assert.True(t, wumpkit.WebhookEventGameDirectMessageCreate.IsExclusive())
	assert.False(t, wumpkit.WebhookEventEntitlementCreate.IsExclusive())
	assert.False(t, wumpkit.WebhookEventInteractionCreate.IsExclusive())

	assert.True(t, wumpkit.WebhookOnlyEventApplicationAuthorized.Is(wumpkit.WebhookOnlyEventApplicationAuthorized))
	assert.True(t, wumpkit.WebhookEventApplicationAuthorized.Is(wumpkit.WebhookEventApplicationAuthorized))
}

func TestStartWebhookEvents_LifecycleAndConcurrency(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := wumpkit.New(
		ctx,
		wumpkit.WithToken("dummy_test_token"),
		wumpkit.WithPublicKey(hex.EncodeToString(pubKey)),
	)

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.StartWebhookEvents("127.0.0.1:0", wumpkit.WithWebhookEventsPath("/events"))
	}()

	time.Sleep(50 * time.Millisecond)

	// attempting to start duplicate Webhook Events server concurrently on the same client must fail
	err = client.StartWebhookEvents("127.0.0.1:0")
	assert.ErrorContains(t, err, "cannot start webhook events server: webhook events server is already running on this client")

	cancel()

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for webhook events server to shut down")
	}
}
