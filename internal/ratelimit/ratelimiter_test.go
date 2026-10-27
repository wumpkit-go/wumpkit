package ratelimit

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimiter_BasicWaitAndUpdate(t *testing.T) {
	limiter := New()

	ctx := context.Background()
	route := "/channels/123456789012345678/messages"
	method := "POST"

	// 1. initial wait should proceed immediately
	start := time.Now()
	err := limiter.Wait(ctx, route, method)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), start, 50*time.Millisecond)

	// 2. simulate Discord response headers
	headers := make(http.Header)
	headers.Set("X-RateLimit-Limit", "5")
	headers.Set("X-RateLimit-Remaining", "4")
	headers.Set("X-RateLimit-Reset-After", "1.5")
	headers.Set("X-RateLimit-Bucket", "hash123")

	limiter.Update(route, method, headers, 200)

	// verify bucket state
	hash, ok := limiter.store.GetBucketHash(NormalizeRoute(route, method))
	assert.True(t, ok)
	assert.Equal(t, "hash123", hash)

	state, ok := limiter.store.GetBucket("123456789012345678:hash123")
	assert.True(t, ok)
	assert.Equal(t, int64(5), state.Limit)
	assert.Equal(t, int64(4), state.Remaining)
	assert.Equal(t, "hash123", state.BucketHash)
}

func TestLimiter_BlocksWhenRemainingZero(t *testing.T) {
	limiter := New()
	ctx := context.Background()
	route := "/channels/123456789012345678/messages"
	method := "POST"

	// set remaining to 0 with reset in 100ms
	headers := make(http.Header)
	headers.Set("X-RateLimit-Limit", "5")
	headers.Set("X-RateLimit-Remaining", "0")
	headers.Set("X-RateLimit-Reset-After", "0.1") // 100ms
	headers.Set("X-RateLimit-Bucket", "hashZero")
	limiter.Update(route, method, headers, 200)

	start := time.Now()
	err := limiter.Wait(ctx, route, method)
	require.NoError(t, err)
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 80*time.Millisecond)
}

func TestLimiter_GlobalRateLimit(t *testing.T) {
	limiter := New()
	ctx := context.Background()
	route := "/guilds/123456789012345678"
	method := "GET"

	// set global rate limit pause for 100ms
	headers := make(http.Header)
	headers.Set("X-RateLimit-Global", "true")
	headers.Set("Retry-After", "0.1") // 100ms
	limiter.Update(route, method, headers, 429)

	start := time.Now()
	err := limiter.Wait(ctx, route, method)
	require.NoError(t, err)
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 80*time.Millisecond)
}

func TestLimiter_ContextCancellation(t *testing.T) {
	limiter := New()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	route := "/channels/123456789012345678/messages"
	method := "POST"

	// set remaining to 0 with long reset (5 seconds)
	headers := make(http.Header)
	headers.Set("X-RateLimit-Limit", "5")
	headers.Set("X-RateLimit-Remaining", "0")
	headers.Set("X-RateLimit-Reset-After", "5.0")
	limiter.Update(route, method, headers, 200)

	err := limiter.Wait(ctx, route, method)
	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestLimiter_ConcurrentRequests(t *testing.T) {
	limiter := New()
	ctx := context.Background()
	route := "/channels/123456789012345678/messages"
	method := "POST"

	var wg sync.WaitGroup
	workers := 10
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()
			err := limiter.Wait(ctx, route, method)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestLimiter_InteractionRouteBypassesGlobalBucket(t *testing.T) {
	// create a limiter with global limit of 1 to easily exhaust it
	limiter := New(WithGlobalLimit(1))
	ctx := context.Background()

	// exhaust the single global token with a regular route
	err := limiter.Wait(ctx, "/guilds/103039963636301824", "GET")
	require.NoError(t, err)

	// interaction route must proceed immediately despite exhausted global bucket
	start := time.Now()
	err = limiter.Wait(ctx, "/interactions/203039963636301824/token/callback", "POST")
	require.NoError(t, err)
	elapsed := time.Since(start)

	assert.Less(
		t,
		elapsed,
		50*time.Millisecond,
		"interaction route should not wait for global bucket",
	)
}
