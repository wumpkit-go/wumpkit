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

package redis

import (
	"context"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/wumpkit/wumpkit"

	_ "embed"
)

//go:embed scripts/identify.lua
var identifyLuaScript string

var identifyScript = goredis.NewScript(identifyLuaScript)

// RedisShardsIdentifyRateLimiter coordinates Discord Gateway Identify rate limits across multiple clusters / processes.
// It supports per-bucket rate limiting for Discord's MaxConcurrency large-bot sharding protocol.
type RedisShardsIdentifyRateLimiter struct {
	client         goredis.UniversalClient
	maxConcurrency int
	interval       time.Duration
	prefix         string
}

var _ wumpkit.ShardsIdentifyRateLimiter = (*RedisShardsIdentifyRateLimiter)(nil)

// RedisIdentifyLimiterOption defines functional options for RedisShardsIdentifyRateLimiter.
type RedisIdentifyLimiterOption func(*RedisShardsIdentifyRateLimiter)

// WithIdentifyMaxConcurrency sets the max concurrency for Identify payloads (default: 1).
// Shards are automatically partitioned into buckets: shardID % maxConcurrency.
func WithIdentifyMaxConcurrency(maxConcurrency int) RedisIdentifyLimiterOption {
	return func(r *RedisShardsIdentifyRateLimiter) {
		if maxConcurrency > 0 {
			r.maxConcurrency = maxConcurrency
		}
	}
}

// WithIdentifyInterval sets the reset interval per bucket (default: 5 * time.Second).
func WithIdentifyInterval(interval time.Duration) RedisIdentifyLimiterOption {
	return func(r *RedisShardsIdentifyRateLimiter) {
		if interval > 0 {
			r.interval = interval
		}
	}
}

// WithIdentifyPrefix sets the Redis key prefix for identify rate limiting (default: "wumpkit:identify_ratelimit:").
func WithIdentifyPrefix(prefix string) RedisIdentifyLimiterOption {
	return func(r *RedisShardsIdentifyRateLimiter) {
		if prefix != "" {
			if !strings.HasSuffix(prefix, ":") {
				prefix += ":"
			}
			r.prefix = prefix
		}
	}
}

// WithIdentifyKey sets the Redis key prefix for identify rate limiting.
func WithIdentifyKey(key string) RedisIdentifyLimiterOption {
	return WithIdentifyPrefix(key)
}

// NewRedisShardsIdentifyRateLimiter creates a new RedisShardsIdentifyRateLimiter for multi-process cluster synchronization.
func NewRedisShardsIdentifyRateLimiter(
	client goredis.UniversalClient,
	options ...RedisIdentifyLimiterOption,
) *RedisShardsIdentifyRateLimiter {
	r := &RedisShardsIdentifyRateLimiter{
		client:         client,
		maxConcurrency: 1,
		interval:       5 * time.Second,
		prefix:         "wumpkit:identify_ratelimit:",
	}

	for _, opt := range options {
		opt(r)
	}

	return r
}

func (r *RedisShardsIdentifyRateLimiter) bucketKey(bucket int) string {
	return r.prefix + "bucket:" + strconv.Itoa(bucket)
}

// WaitShard blocks until the specified shard is allowed to send an Identify payload or the context is cancelled.
// It calculates the bucket index as (shardID % maxConcurrency) for concurrent bucket throughput.
func (r *RedisShardsIdentifyRateLimiter) WaitShard(ctx context.Context, shardID int) error {
	if ctx == nil {
		ctx = context.Background()
	}

	bucket := 0
	if r.maxConcurrency > 1 {
		bucket = shardID % r.maxConcurrency
		if bucket < 0 {
			bucket = -bucket
		}
	}

	bucketKey := r.bucketKey(bucket)

	// execute atomic slot reservation in Redis (evaluated with server-side time)
	res, err := identifyScript.Run(ctx, r.client, []string{bucketKey}, int(r.interval.Milliseconds())).
		Int64()
	if err != nil {
		return err
	}

	if res <= 0 {
		return nil
	}

	timer := time.NewTimer(time.Duration(res) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Wait blocks until a slot in bucket 0 is available to send an Identify payload or the context is cancelled.
func (r *RedisShardsIdentifyRateLimiter) Wait(ctx context.Context) error {
	return r.WaitShard(ctx, 0)
}

// ForShard returns a wumpkit.ShardsIdentifyRateLimiter scoped to a specific shard ID.
func (r *RedisShardsIdentifyRateLimiter) ForShard(shardID int) wumpkit.ShardsIdentifyRateLimiter {
	return &shardBoundIdentifyLimiter{
		limiter: r,
		shardID: shardID,
	}
}

// Bucket returns a wumpkit.ShardsIdentifyRateLimiter scoped to a specific bucket index.
func (r *RedisShardsIdentifyRateLimiter) Bucket(bucketID int) wumpkit.ShardsIdentifyRateLimiter {
	return &shardBoundIdentifyLimiter{
		limiter: r,
		shardID: bucketID,
	}
}

type shardBoundIdentifyLimiter struct {
	limiter *RedisShardsIdentifyRateLimiter
	shardID int
}

func (s *shardBoundIdentifyLimiter) Wait(ctx context.Context) error {
	return s.limiter.WaitShard(ctx, s.shardID)
}
