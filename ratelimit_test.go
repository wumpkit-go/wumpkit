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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wumpkit/wumpkit"
)

func TestClient_WithRateLimitStore(t *testing.T) {
	memStore := wumpkit.NewMemoryRateLimitStore()
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken("dummy_token"),
		wumpkit.WithRESTConfig(wumpkit.RequesterConfig{
			RateLimitStore:  memStore,
			GlobalRateLimit: 30,
		}),
	)

	assert.Equal(t, memStore, bot.RateLimitStore())
}

func TestRequester_RateLimiterIntegration(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			w.Header().Set("X-RateLimit-Limit", "2")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset-After", "0.1") // 100ms
			w.Header().Set("X-RateLimit-Bucket", "chanBucket")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "123"}`))
			return
		}

		w.Header().Set("X-RateLimit-Limit", "2")
		w.Header().Set("X-RateLimit-Remaining", "1")
		w.Header().Set("X-RateLimit-Reset-After", "1.0")
		w.Header().Set("X-RateLimit-Bucket", "chanBucket")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	memStore := wumpkit.NewMemoryRateLimitStore()
	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken("test_token"),
		wumpkit.WithRESTConfig(wumpkit.RequesterConfig{
			BaseURL:        server.URL,
			RateLimitStore: memStore,
		}),
	)

	ctx := context.Background()

	// 1. first request succeeds and sets remaining to 0
	_, err := bot.FetchChannel(ctx, 123)
	require.NoError(t, err)

	// verify state in store — bucket key is now "majorParam:hash" for major param isolation
	state, ok := memStore.GetBucket("123:chanBucket")
	require.True(t, ok)
	assert.Equal(t, int64(2), state.Limit)
	assert.Equal(t, int64(0), state.Remaining)

	// 2. second request should automatically wait for bucket reset (100ms) before sending
	start := time.Now()
	_, err = bot.FetchChannel(ctx, 123)
	require.NoError(t, err)
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 80*time.Millisecond)
	assert.Equal(t, 2, requestCount)
}
