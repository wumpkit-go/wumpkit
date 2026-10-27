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
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/wumpkit/wumpkit"
)

// RedisRateLimitStore implements wumpkit.RateLimitStore using Redis for multi-cluster synchronization.
type RedisRateLimitStore struct {
	client goredis.UniversalClient
	prefix string
}

// NewRedisRateLimitStore creates a new RedisRateLimitStore with an optional key prefix (defaults to "wumpkit:ratelimit:").
func NewRedisRateLimitStore(client goredis.UniversalClient, prefix ...string) *RedisRateLimitStore {
	p := "wumpkit:ratelimit:"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
		if !strings.HasSuffix(p, ":") {
			p += ":"
		}
	}
	return &RedisRateLimitStore{
		client: client,
		prefix: p,
	}
}

func (s *RedisRateLimitStore) bucketKey(key string) string {
	return s.prefix + "bucket:" + key
}

func (s *RedisRateLimitStore) hashKey(key string) string {
	return s.prefix + "hash:" + key
}

func (s *RedisRateLimitStore) globalKey() string {
	return s.prefix + "global_lock"
}

// reserve atomically checks remaining quota and consumes 1 token in redis using optimistic locking
func (s *RedisRateLimitStore) Reserve(bucketKey string) time.Duration {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := s.bucketKey(bucketKey)

	for attempt := 0; attempt < 3; attempt++ {
		var waitDuration time.Duration

		err := s.client.Watch(ctx, func(tx *goredis.Tx) error {
			val, err := tx.Get(ctx, key).Result()
			if errors.Is(err, goredis.Nil) {
				waitDuration = 0
				return nil
			}
			if err != nil {
				return err
			}

			var state wumpkit.RateLimitBucketState
			if err := json.Unmarshal([]byte(val), &state); err != nil {
				waitDuration = 0
				return nil
			}

			now := time.Now()
			if now.After(state.ResetAt) {
				state.Remaining = state.Limit
			} else if state.Remaining <= 0 {
				waitDuration = max(0, state.ResetAt.Sub(now))
				return nil
			}

			if state.Remaining > 0 {
				state.Remaining--
				data, err := json.Marshal(state)
				if err != nil {
					return err
				}
				ttl := max(time.Minute, time.Until(state.ResetAt)+time.Minute)
				_, err = tx.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
					pipe.Set(ctx, key, data, ttl)
					return nil
				})
				return err
			}

			waitDuration = 0
			return nil
		}, key)

		if errors.Is(err, goredis.TxFailedErr) {
			continue
		}
		if err != nil {
			return 0
		}
		return waitDuration
	}

	return 0
}

func (s *RedisRateLimitStore) GetBucket(bucketKey string) (wumpkit.RateLimitBucketState, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var zero wumpkit.RateLimitBucketState
	val, err := s.client.Get(ctx, s.bucketKey(bucketKey)).Result()
	if err != nil {
		return zero, false
	}

	var state wumpkit.RateLimitBucketState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return zero, false
	}
	return state, true
}

func (s *RedisRateLimitStore) SetBucket(bucketKey string, state wumpkit.RateLimitBucketState) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	ttl := time.Until(state.ResetAt) + 1*time.Minute
	if ttl <= 0 {
		ttl = 1 * time.Minute
	}

	_ = s.client.Set(ctx, s.bucketKey(bucketKey), data, ttl).Err()
}

func (s *RedisRateLimitStore) GetGlobalLock() time.Time {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	val, err := s.client.Get(ctx, s.globalKey()).Result()
	if err != nil {
		return time.Time{}
	}

	unixNano, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(0, unixNano)
}

func (s *RedisRateLimitStore) SetGlobalLock(until time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ttl := time.Until(until)
	if ttl <= 0 {
		return
	}

	_ = s.client.Set(ctx, s.globalKey(), strconv.FormatInt(until.UnixNano(), 10), ttl).Err()
}

func (s *RedisRateLimitStore) GetBucketHash(optimisticKey string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	val, err := s.client.Get(ctx, s.hashKey(optimisticKey)).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

func (s *RedisRateLimitStore) SetBucketHash(optimisticKey, bucketHash string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = s.client.Set(ctx, s.hashKey(optimisticKey), bucketHash, 24*time.Hour).Err()
}
