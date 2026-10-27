package wumpkit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequester_DoRequest_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bot test_token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Audit-Log-Reason") != "test reason" {
			http.Error(w, "bad reason", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"123","username":"wumpkit_bot"}`))
	}))
	defer ts.Close()

	cfg := DefaultRequesterConfig()
	cfg.BaseURL = ts.URL

	req := newRequester("test_token", cfg, slog.Default())
	body, err := req.DoRequest(context.Background(), Request{
		Method: http.MethodGet,
		URL:    "/users/@me",
		Reason: "test reason",
	})
	require.NoError(t, err)
	defer body.Close()

	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var u User
	err = json.Unmarshal(data, &u)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(123), u.ID)
	assert.Equal(t, "wumpkit_bot", u.Username)
}

func TestRequester_NonRetryableError(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":10003,"message":"Unknown Channel"}`))
	}))
	defer ts.Close()

	cfg := DefaultRequesterConfig()
	cfg.BaseURL = ts.URL
	cfg.MaxRetries = 3

	req := newRequester("test_token", cfg, slog.Default())
	_, err := req.DoRequest(context.Background(), Request{
		Method: http.MethodGet,
		URL:    "/channels/999",
	})

	require.Error(t, err)
	assert.Equal(t, int32(1), attempts.Load(), "non-retryable errors (404) must not be retried")
	assert.False(t, errors.Is(err, ErrMaxRetriesExceeded))
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Unknown Channel")
}

func TestRequester_RetryableStatus_ExceedsRetries(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":0,"message":"Internal Server Error"}`))
	}))
	defer ts.Close()

	cfg := DefaultRequesterConfig()
	cfg.BaseURL = ts.URL
	cfg.MaxRetries = 2

	req := newRequester("test_token", cfg, slog.Default())
	_, err := req.DoRequest(context.Background(), Request{
		Method: http.MethodGet,
		URL:    "/guilds/123",
	})

	require.Error(t, err)
	assert.Equal(t, int32(3), attempts.Load(), "initial attempt (1) + 2 retries = 3 total attempts")
	assert.True(
		t,
		errors.Is(err, ErrMaxRetriesExceeded),
		"must return sentinel ErrMaxRetriesExceeded",
	)
}

func TestRequester_RateLimit_RetriesAndSucceeds(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count == 1 {
			w.Header().Set("Retry-After", "0.05")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"You are being rate limited."}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer ts.Close()

	cfg := DefaultRequesterConfig()
	cfg.BaseURL = ts.URL
	cfg.MaxRetries = 3

	req := newRequester("test_token", cfg, slog.Default())
	body, err := req.DoRequest(context.Background(), Request{
		Method: http.MethodGet,
		URL:    "/gateway",
	})

	require.NoError(t, err)
	defer body.Close()
	assert.Equal(t, int32(2), attempts.Load())
}

func TestRequester_ContextCancelled_AbortsImmediately(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel context

	cfg := DefaultRequesterConfig()
	cfg.BaseURL = ts.URL
	cfg.MaxRetries = 3

	req := newRequester("test_token", cfg, slog.Default())
	_, err := req.DoRequest(ctx, Request{
		Method: http.MethodGet,
		URL:    "/guilds/123",
	})

	require.ErrorIs(t, err, context.Canceled)
}

func TestComputeBackoff_Jitter(t *testing.T) {
	for attempt := range 5 {
		b := computeBackoff(attempt, 50*time.Millisecond, 2*time.Second)
		assert.GreaterOrEqual(t, b, 25*time.Millisecond)
		assert.LessOrEqual(t, b, 2*time.Second)
	}
}
