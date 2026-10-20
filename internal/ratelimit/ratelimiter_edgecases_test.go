package ratelimit_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wumpkit-go/wumpkit/internal/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateSnowflake creates a valid Discord Snowflake ID for a given timestamp.
func generateSnowflake(t time.Time) string {
	epoch := uint64(t.UnixMilli() - 1420070400000)
	id := epoch << 22
	return strconv.FormatUint(id, 10)
}

func TestEdgeCases_RouteNormalizationExhaustive(t *testing.T) {
	old15d := generateSnowflake(time.Now().Add(-15 * 24 * time.Hour))
	new5s := generateSnowflake(time.Now().Add(-5 * time.Second))
	mid2d := generateSnowflake(time.Now().Add(-2 * 24 * time.Hour))

	tests := []struct {
		url    string
		method string
		want   string
	}{
		// 1. guild Major
		{"/api/v10/guilds/103039963636301824", "GET", "/guilds/103039963636301824"},
		{"/api/v9/guilds/103039963636301824/channels", "GET", "/guilds/103039963636301824/channels"},
		{"/api/v10/guilds/999999999999999999/channels", "GET", "/guilds/999999999999999999/channels"},
		{"/api/v10/guilds/103039963636301824/roles", "GET", "/guilds/103039963636301824/roles"},
		{"/api/v10/guilds/103039963636301824/roles/103039963636301825", "PATCH", "/guilds/103039963636301824/roles/!"},
		{"/api/v10/guilds/103039963636301824/members/103039963636301825", "GET", "/guilds/103039963636301824/members/!"},
		{"/api/v10/guilds/103039963636301824/bans/103039963636301825", "GET", "/guilds/103039963636301824/bans/!"},
		{"/api/v10/guilds/103039963636301824/emojis/103039963636301825", "GET", "/guilds/103039963636301824/emojis/!"},
		{"/api/v10/guilds/templates/103039963636301824", "GET", "/guilds/templates/!"},

		// 2. channel Major (numParts == 2, preserves channel_id as major parameter)
		{"/api/v10/channels/203039963636301824", "GET", "/channels/203039963636301824"},
		{"/api/v10/channels/203039963636301824", "DELETE", "/channels/203039963636301824"},
		{"/api/v10/channels/203039963636301824", "PATCH", "/channels/203039963636301824"},

		// 3. channel Sub-resources
		{"/api/v10/channels/203039963636301824/pins", "GET", "/channels/203039963636301824/pins"},
		{"/api/v10/channels/203039963636301824/messages", "GET", "/channels/203039963636301824/messages"},
		{"/api/v10/channels/203039963636301824/messages", "POST", "/channels/203039963636301824/messages"},
		{"/api/v10/channels/203039963636301824/messages/bulk-delete", "POST", "/channels/203039963636301824/messages/bulk-delete"},
		{"/api/v10/channels/203039963636301824/permissions/103039963636301825", "PUT", "/channels/203039963636301824/permissions/!"},
		{"/api/v10/channels/203039963636301824/invites", "GET", "/channels/203039963636301824/invites"},
		{"/api/v10/channels/203039963636301824/threads", "POST", "/channels/203039963636301824/threads"},

		// 4. message Deletion Age Rules
		{"/api/v10/channels/203039963636301824/messages/" + old15d, "DELETE", "/channels/203039963636301824/messages/!14dmsg"},
		{"/api/v10/channels/203039963636301824/messages/" + new5s, "DELETE", "/channels/203039963636301824/messages/!10smsg"},
		{"/api/v10/channels/203039963636301824/messages/" + mid2d, "DELETE", "/channels/203039963636301824/messages/!"},
		{"/api/v10/channels/203039963636301824/messages/" + old15d, "GET", "/channels/203039963636301824/messages/!"},
		{"/api/v10/channels/203039963636301824/messages/" + old15d, "PATCH", "/channels/203039963636301824/messages/!"},

		// 5. reactions
		{"/api/v10/channels/203039963636301824/messages/" + old15d + "/reactions/%F0%9F%98%8B", "GET", "/channels/203039963636301824/messages/!/reactions/!/!"},
		{"/api/v10/channels/203039963636301824/messages/" + old15d + "/reactions/customEmoji:103039963636301825", "GET", "/channels/203039963636301824/messages/!/reactions/!/!"},
		{"/api/v10/channels/203039963636301824/messages/" + old15d + "/reactions/%F0%9F%98%8B/@me", "PUT", "/channels/203039963636301824/messages/!/reactions/!modify"},
		{"/api/v10/channels/203039963636301824/messages/" + old15d + "/reactions/%F0%9F%98%8B/@me", "DELETE", "/channels/203039963636301824/messages/!/reactions/!modify"},
		{"/api/v10/channels/203039963636301824/messages/" + old15d + "/reactions/%F0%9F%98%8B/103039963636301825", "DELETE", "/channels/203039963636301824/messages/!/reactions/!modify"},
		{"/api/v10/channels/203039963636301824/messages/" + old15d + "/reactions", "DELETE", "/channels/203039963636301824/messages/!/reactions/!modify"},

		// 6. webhooks & Followup Messages
		{"/api/v10/webhooks/203039963636301824", "GET", "/webhooks/203039963636301824"},
		{"/api/v10/webhooks/203039963636301824/VSOzAqY1OZFF5WJVtbIzFtmjGupk-84Hn0A_ZzToF_CHsPIeCk0Q9Uok_mjxR0dNtApI", "POST", "/webhooks/203039963636301824/!"},
		{"/api/v10/webhooks/203039963636301824/aW50ZXJhY3Rpb246MTEwMzA0OTQyMDkzMDU2ODMyMjpOZUllWHdNU2J4RXBFMHVYRjBpU0pHMDdEb3BhM3ZlYklBODlMUmtlUXlRbzlpZzYyTnpLU0dqdWlyVlBvZnBSUlJHbUJHYlJ0N29MbE9KQUJVTFk4bTR4UzFtZEpEeXJyY0hBUERmTEhKVE9wRkNzU1FFWUkwTnlpWFY2WHdrRg/messages/@original", "POST", "/webhooks/203039963636301824/1103049420930568322/messages/@original"},

		// 7. invites
		{"/api/v10/invites/discord-developers", "GET", "/invites/!"},
		{"/api/v10/invites/103039963636301824", "GET", "/invites/!"},

		// 8. interactions
		{"/api/v10/interactions/203039963636301824/aW50ZXJhY3Rpb246ODg3NTU5MDA01AY4NTUxNDU0OnZwS3QycDhvREk2aVF3U1BqN2prcXBkRmNqNlp4VEhGRjZvSVlXSGh4WG4yb3l6Z3B6NTBPNVc3OHphV05OULLMOHBMa2RTZmVKd3lzVDA2b2h3OTUxaFJ4QlN0dkxXallPcmhnSHNJb0tSV0M5ZzY1NkN4VGRvemFOSHY4b05c/callback", "POST", "/interactions/203039963636301824/!/callback"},

		// 9. application Commands
		{"/api/v10/applications/103039963636301824/commands", "GET", "/applications/103039963636301824/commands"},
		{"/api/v10/applications/103039963636301824/commands/103039963636301825", "DELETE", "/applications/103039963636301824/commands/!"},
		{"/api/v10/applications/103039963636301824/guilds/103039963636301825/commands", "GET", "/applications/103039963636301824/guilds/!/commands"},
		{"/api/v10/applications/103039963636301824/guilds/103039963636301825/commands/103039963636301826", "GET", "/applications/103039963636301824/guilds/!/commands/!"},

		// 10. query Parameters & Unversioned
		{"/api/v10/channels/203039963636301824/messages?limit=50&around=103039963636301825", "GET", "/channels/203039963636301824/messages"},
		{"/channels/203039963636301824/messages?limit=50", "GET", "/channels/203039963636301824/messages"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.url), func(t *testing.T) {
			got := ratelimit.NormalizeRoute(tt.url, tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEdgeCases_HeaderParsingPrecisionAndFallbacks(t *testing.T) {
	// case 1: Standard headers with millisecond decimal precision
	h1 := make(http.Header)
	h1.Set("X-RateLimit-Limit", "10")
	h1.Set("X-RateLimit-Remaining", "5")
	h1.Set("X-RateLimit-Reset-After", "0.250")
	h1.Set("X-RateLimit-Bucket", "bucketA")
	info1 := ratelimit.ParseHeaders(h1, 200)
	assert.Equal(t, int64(10), info1.Limit)
	assert.Equal(t, int64(5), info1.Remaining)
	assert.Equal(t, 250*time.Millisecond, info1.ResetAfter)
	assert.Equal(t, "bucketA", info1.BucketHash)
	assert.False(t, info1.IsGlobal)
	assert.Equal(t, "user", info1.Scope)

	// case 2: Shared rate limit preferring Retry-After
	h2 := make(http.Header)
	h2.Set("X-RateLimit-Limit", "5")
	h2.Set("X-RateLimit-Remaining", "0")
	h2.Set("X-RateLimit-Reset-After", "5.0")
	h2.Set("Retry-After", "0.500")
	h2.Set("X-RateLimit-Scope", "shared")
	info2 := ratelimit.ParseHeaders(h2, 429)
	assert.Equal(t, 500*time.Millisecond, info2.ResetAfter)
	assert.Equal(t, "shared", info2.Scope)

	// case 3: Global rate limit
	h3 := make(http.Header)
	h3.Set("X-RateLimit-Global", "true")
	h3.Set("Retry-After", "1.250")
	info3 := ratelimit.ParseHeaders(h3, 429)
	assert.True(t, info3.IsGlobal)
	assert.Equal(t, 1250*time.Millisecond, info3.ResetAfter)

	// case 4: Nil headers
	info4 := ratelimit.ParseHeaders(nil, 200)
	assert.False(t, info4.HasHeaders)
}

func TestEdgeCases_DynamicBucketHashSharing(t *testing.T) {
	limiter := ratelimit.New()
	ctx := context.Background()

	// same channel, different sub-resources that Discord groups under the same bucket hash
	// this tests that the hash correctly unifies routes within the same major parameter
	channelMessages := "/channels/111111111111111111/messages"
	channelPins := "/channels/111111111111111111/pins"

	err := limiter.Wait(ctx, channelMessages, "POST")
	require.NoError(t, err)

	h := make(http.Header)
	h.Set("X-RateLimit-Limit", "5")
	h.Set("X-RateLimit-Remaining", "0")
	h.Set("X-RateLimit-Reset-After", "0.1") // 100ms
	h.Set("X-RateLimit-Bucket", "same-bucket-hash")
	limiter.Update(channelMessages, "POST", h, 200)

	// link pins route on the same channel to the same bucket hash
	limiter.Update(channelPins, "GET", h, 200)

	// request on pins should be blocked because it shares quota with messages (same channel, same hash)
	start := time.Now()
	err = limiter.Wait(ctx, channelPins, "GET")
	require.NoError(t, err)
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 80*time.Millisecond)
}

func TestEdgeCases_GlobalLockBlocksAllRoutes(t *testing.T) {
	limiter := ratelimit.New()
	ctx := context.Background()

	guildRoute := "/guilds/103039963636301824"
	channelRoute := "/channels/203039963636301824/messages"

	// trigger 429 global rate limit for 100ms on guildRoute
	h := make(http.Header)
	h.Set("X-RateLimit-Global", "true")
	h.Set("Retry-After", "0.1")
	limiter.Update(guildRoute, "GET", h, 429)

	// attempting to send request on completely different channelRoute must be blocked by global lock!
	start := time.Now()
	err := limiter.Wait(ctx, channelRoute, "POST")
	require.NoError(t, err)
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 80*time.Millisecond)
}

func TestEdgeCases_IndependentBucketsDoNotBlockEachOther(t *testing.T) {
	limiter := ratelimit.New()
	ctx := context.Background()

	channelA := "/channels/111111111111111111/messages"
	channelB := "/channels/222222222222222222/messages"

	// exhaust quota on Channel A for 5 seconds using the SAME bucket hash as Channel B
	// discord returns identical X-RateLimit-Bucket hashes for the same endpoint template,
	// but per-route limits are independent across major parameters (channel_id)
	h := make(http.Header)
	h.Set("X-RateLimit-Limit", "5")
	h.Set("X-RateLimit-Remaining", "0")
	h.Set("X-RateLimit-Reset-After", "5.0")
	h.Set("X-RateLimit-Bucket", "same-endpoint-hash") // same hash for both channels!
	limiter.Update(channelA, "POST", h, 200)

	// channel B must proceed immediately despite sharing the same bucket hash!
	// this verifies the major parameter isolation fix
	start := time.Now()
	err := limiter.Wait(ctx, channelB, "POST")
	require.NoError(t, err)
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 50*time.Millisecond)
}

func TestEdgeCases_ConcurrentQuotaExhaustionAndSerialization(t *testing.T) {
	limiter := ratelimit.New()
	ctx := context.Background()
	route := "/channels/123456789012345678/messages"

	// set bucket capacity to 3 remaining
	h := make(http.Header)
	h.Set("X-RateLimit-Limit", "3")
	h.Set("X-RateLimit-Remaining", "3")
	h.Set("X-RateLimit-Reset-After", "0.15") // 150ms
	h.Set("X-RateLimit-Bucket", "concurrentBucket")
	limiter.Update(route, "POST", h, 200)

	var wg sync.WaitGroup
	var executedCount atomic.Int32
	totalCallers := 5

	start := time.Now()
	for range totalCallers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := limiter.Wait(ctx, route, "POST")
			assert.NoError(t, err)
			executedCount.Add(1)
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	assert.Equal(t, int32(5), executedCount.Load())
	// because 5 callers exceeded the 3 remaining quota, the last callers had to wait for the 150ms reset
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
}

func TestEdgeCases_ContextTimeoutDuringWait(t *testing.T) {
	limiter := ratelimit.New()
	route := "/channels/123456789012345678/messages"

	// set bucket remaining to 0 with 5s reset
	h := make(http.Header)
	h.Set("X-RateLimit-Limit", "5")
	h.Set("X-RateLimit-Remaining", "0")
	h.Set("X-RateLimit-Reset-After", "5.0")
	limiter.Update(route, "POST", h, 200)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx, route, "POST")
	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestEdgeCases_InteractionRoutesBypassGlobalRateLimit(t *testing.T) {
	// create a limiter with a tiny global limit (1 req/s) to easily trigger the global gate
	limiter := ratelimit.New(ratelimit.WithGlobalLimit(1))
	ctx := context.Background()

	// exhaust the global bucket with a regular route
	regularRoute := "/guilds/103039963636301824"
	err := limiter.Wait(ctx, regularRoute, "GET")
	require.NoError(t, err)

	// also set a global lock to simulate a global 429
	h := make(http.Header)
	h.Set("X-RateLimit-Global", "true")
	h.Set("Retry-After", "5.0") // 5 second global lock
	limiter.Update(regularRoute, "GET", h, 429)

	// interaction callback route must proceed immediately despite global lock + exhausted global bucket
	interactionRoute := "/interactions/203039963636301824/aW50ZXJhY3Rpb246ODg3NTU5MDA01AY4NTUxNDU0OnZwS3QycDhvREk2aVF3U1BqN2prcXBkRmNqNlp4VEhGRjZvSVlXSGh4WG4yb3l6Z3B6NTBPNVc3OHphV05OULLMOHBMa2RTZmVKd3lzVDA2b2h3OTUxaFJ4QlN0dkxXallPcmhnSHNJb0tSV0M5ZzY1NkN4VGRvemFOSHY4b05c/callback"
	start := time.Now()
	err = limiter.Wait(ctx, interactionRoute, "POST")
	require.NoError(t, err)
	elapsed := time.Since(start)

	// must complete instantly (not blocked by 5s global lock)
	assert.Less(t, elapsed, 50*time.Millisecond, "interaction route should bypass global rate limit")
}

func TestEdgeCases_InteractionWebhookFollowupBypassesGlobalRateLimit(t *testing.T) {
	limiter := ratelimit.New(ratelimit.WithGlobalLimit(1))
	ctx := context.Background()

	// exhaust global bucket
	err := limiter.Wait(ctx, "/guilds/103039963636301824", "GET")
	require.NoError(t, err)

	// set a global 429 lock
	h := make(http.Header)
	h.Set("X-RateLimit-Global", "true")
	h.Set("Retry-After", "5.0")
	limiter.Update("/guilds/103039963636301824", "GET", h, 429)

	// webhook interaction followup (with interaction token) must bypass global limit
	followupRoute := "/webhooks/203039963636301824/aW50ZXJhY3Rpb246MTEwMzA0OTQyMDkzMDU2ODMyMjpOZUllWHdNU2J4RXBFMHVYRjBpU0pHMDdEb3BhM3ZlYklBODlMUmtlUXlRbzlpZzYyTnpLU0dqdWlyVlBvZnBSUlJHbUJHYlJ0N29MbE9KQUJVTFk4bTR4UzFtZEpEeXJyY0hBUERmTEhKVE9wRkNzU1FFWUkwTnlpWFY2WHdrRg/messages/@original"
	start := time.Now()
	err = limiter.Wait(ctx, followupRoute, "POST")
	require.NoError(t, err)
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 50*time.Millisecond, "webhook interaction followup should bypass global rate limit")
}

func TestEdgeCases_SharedScope429DoesNotZeroRemaining(t *testing.T) {
	limiter := ratelimit.New()
	ctx := context.Background()
	route := "/channels/123456789012345678/messages"

	// simulate a shared scope 429 where the per-route bucket still has capacity
	// this happens when the underlying Discord resource is throttled across multiple bots
	h := make(http.Header)
	h.Set("X-RateLimit-Limit", "10")
	h.Set("X-RateLimit-Remaining", "9") // per-route bucket still has 9 remaining!
	h.Set("X-RateLimit-Reset-After", "64.57")
	h.Set("Retry-After", "0.1") // shared resource is throttled for 100ms
	h.Set("X-RateLimit-Bucket", "sharedBucket")
	h.Set("X-RateLimit-Scope", "shared")
	limiter.Update(route, "POST", h, 429)

	// the per-route bucket should still have remaining capacity (9, not 0)
	// the request should NOT be blocked for 64.57s — it should use the Retry-After
	// duration from the shared scope as the reset time
	start := time.Now()
	err := limiter.Wait(ctx, route, "POST")
	require.NoError(t, err)
	elapsed := time.Since(start)

	// should wait only ~100ms (Retry-After) not the full 64.57s
	// since remaining was set to 9, the request should proceed after the bucket resets
	assert.Less(t, elapsed, 200*time.Millisecond, "shared scope 429 should not block for the per-route reset duration")
}

func TestEdgeCases_MemoryStoreCleanup(t *testing.T) {
	store := ratelimit.NewMemoryStore()

	// create an expired bucket (reset was 10 minutes ago)
	store.SetBucket("expired-bucket", ratelimit.BucketState{
		Limit:      5,
		Remaining:  0,
		ResetAt:    time.Now().Add(-10 * time.Minute),
		BucketHash: "expired-bucket",
	})

	// create an active bucket (reset in 1 minute)
	store.SetBucket("active-bucket", ratelimit.BucketState{
		Limit:      5,
		Remaining:  3,
		ResetAt:    time.Now().Add(1 * time.Minute),
		BucketHash: "active-bucket",
	})

	// verify both exist
	_, ok := store.GetBucket("expired-bucket")
	assert.True(t, ok, "expired bucket should exist before cleanup")
	_, ok = store.GetBucket("active-bucket")
	assert.True(t, ok, "active bucket should exist before cleanup")

	// run cleanup
	store.Cleanup()

	// expired bucket (10 min past reset + 5 min grace = well past) should be removed
	_, ok = store.GetBucket("expired-bucket")
	assert.False(t, ok, "expired bucket should be removed after cleanup")

	// active bucket should remain
	state, ok := store.GetBucket("active-bucket")
	assert.True(t, ok, "active bucket should remain after cleanup")
	assert.Equal(t, int64(3), state.Remaining)
}

func TestEdgeCases_ReserveResetRaceCondition(t *testing.T) {
	store := ratelimit.NewMemoryStore()

	// create a bucket that has expired (ResetAt in the past)
	store.SetBucket("race-bucket", ratelimit.BucketState{
		Limit:      5,
		Remaining:  0,
		ResetAt:    time.Now().Add(-100 * time.Millisecond), // expired 100ms ago
		BucketHash: "race-bucket",
	})

	// first Reserve should reset remaining to Limit (5) and consume 1, returning 0 wait
	wait := store.Reserve("race-bucket")
	assert.Equal(t, time.Duration(0), wait, "first reserve after expiry should succeed immediately")

	// subsequent reserves should also succeed (remaining was reset to 5, minus 1 = 4 left)
	for i := range 4 {
		wait = store.Reserve("race-bucket")
		assert.Equal(t, time.Duration(0), wait, fmt.Sprintf("reserve %d should succeed", i+2))
	}

	// the 6th reserve should also succeed because ResetAt was cleared after expiry reset,
	// so remaining stays at 0 but there's no ResetAt to wait for
	wait = store.Reserve("race-bucket")
	assert.Equal(t, time.Duration(0), wait, "reserve after full consumption with cleared ResetAt should succeed")
}
