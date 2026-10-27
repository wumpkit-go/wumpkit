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

package redis_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wumpkitredis "github.com/wumpkit/wumpkit/extra/redis"
)

func TestRedisShardsIdentifyRateLimiter_Basic(t *testing.T) {
	backend := setupRedisClient(t)
	defer backend.cleanup()

	limiter := wumpkitredis.NewRedisShardsIdentifyRateLimiter(
		backend.client,
		wumpkitredis.WithIdentifyMaxConcurrency(1),
		wumpkitredis.WithIdentifyInterval(150*time.Millisecond),
		wumpkitredis.WithIdentifyPrefix("test:identify:basic:"),
	)

	ctx := context.Background()

	// first identify succeeds immediately
	start := time.Now()
	err1 := limiter.Wait(ctx)
	require.NoError(t, err1)
	assert.Less(t, time.Since(start), 50*time.Millisecond)

	// second identify in the same bucket waits for the interval window
	err2 := limiter.Wait(ctx)
	require.NoError(t, err2)
	assert.GreaterOrEqual(t, time.Since(start), 150*time.Millisecond)
}

func TestRedisShardsIdentifyRateLimiter_ConcurrentBuckets(t *testing.T) {
	backend := setupRedisClient(t)
	defer backend.cleanup()

	// 4 concurrent buckets (max_concurrency = 4)
	limiter := wumpkitredis.NewRedisShardsIdentifyRateLimiter(
		backend.client,
		wumpkitredis.WithIdentifyMaxConcurrency(4),
		wumpkitredis.WithIdentifyInterval(300*time.Millisecond),
		wumpkitredis.WithIdentifyPrefix("test:identify:buckets:"),
	)

	ctx := context.Background()

	// shards 0, 1, 2, 3 belong to 4 different buckets (0, 1, 2, 3)
	// they should all identify concurrently at the same time without waiting on each other!
	start := time.Now()
	var wg sync.WaitGroup
	for shardID := 0; shardID < 4; shardID++ {
		wg.Add(1)
		go func(sID int) {
			defer wg.Done()
			err := limiter.WaitShard(ctx, sID)
			assert.NoError(t, err)
		}(shardID)
	}
	wg.Wait()

	// all 4 different buckets executed in parallel in under 100ms
	assert.Less(t, time.Since(start), 100*time.Millisecond)

	// now Shards 4, 5, 6, 7 map to buckets 0, 1, 2, 3 (4 % 4 = 0, etc.)
	// they must wait for the 300ms window
	start2 := time.Now()
	for shardID := 4; shardID < 8; shardID++ {
		wg.Add(1)
		go func(sID int) {
			defer wg.Done()
			err := limiter.WaitShard(ctx, sID)
			assert.NoError(t, err)
		}(shardID)
	}
	wg.Wait()

	assert.GreaterOrEqual(t, time.Since(start2), 200*time.Millisecond)
}

func TestRedisShardsIdentifyRateLimiter_ForShardAndBucket(t *testing.T) {
	backend := setupRedisClient(t)
	defer backend.cleanup()

	limiter := wumpkitredis.NewRedisShardsIdentifyRateLimiter(
		backend.client,
		wumpkitredis.WithIdentifyMaxConcurrency(2),
		wumpkitredis.WithIdentifyInterval(100*time.Millisecond),
		wumpkitredis.WithIdentifyPrefix("test:identify:forshard:"),
	)

	shard0Limiter := limiter.ForShard(0)
	shard1Limiter := limiter.ForShard(1)

	ctx := context.Background()

	start := time.Now()
	err0 := shard0Limiter.Wait(ctx)
	require.NoError(t, err0)

	err1 := shard1Limiter.Wait(ctx)
	require.NoError(t, err1)

	// shard 0 (bucket 0) and Shard 1 (bucket 1) are concurrent
	assert.Less(t, time.Since(start), 50*time.Millisecond)
}

func TestRedisShardsIdentifyRateLimiter_ContextCancellation(t *testing.T) {
	backend := setupRedisClient(t)
	defer backend.cleanup()

	limiter := wumpkitredis.NewRedisShardsIdentifyRateLimiter(
		backend.client,
		wumpkitredis.WithIdentifyMaxConcurrency(1),
		wumpkitredis.WithIdentifyInterval(5*time.Second),
		wumpkitredis.WithIdentifyPrefix("test:identify:cancel:"),
	)

	ctx := context.Background()
	// consume single token
	err := limiter.Wait(ctx)
	require.NoError(t, err)

	// second token with short timeout should cancel cleanly
	cancelCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = limiter.Wait(cancelCtx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
