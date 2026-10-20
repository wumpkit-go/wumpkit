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

package ratelimit

import (
	"sync/atomic"
	"time"

	"github.com/wumpkit-go/wumpkit/internal/shardedmap"
)

// BucketState represents the quota state for a Discord rate limit bucket.
type BucketState struct {
	Limit      int64     // Maximum requests allowed in the window.
	Remaining  int64     // Remaining requests in the current window.
	ResetAt    time.Time // Time at which the bucket resets.
	BucketHash string    // Opaque bucket identifier from Discord.
}

// Store abstracts the storage of rate limit bucket states and global locks.
// Implementations must be safe for concurrent use.
//
// The default implementation is MemoryStore (single-process).
// For multi-cluster deployments, use an external store (e.g. Redis).
type Store interface {
	// Reserve atomically checks and consumes 1 quota token from the bucket.
	// Returns 0 if quota was available, or waitDuration > 0 if caller must sleep until reset.
	Reserve(bucketKey string) (waitDuration time.Duration)

	// SetBucket stores the bucket state for the given key.
	SetBucket(bucketKey string, state BucketState)

	// GetBucket retrieves the bucket state for the given key.
	GetBucket(bucketKey string) (BucketState, bool)

	// GetGlobalLock returns the time until which all requests should be paused.
	// Returns zero time if no global lock is active.
	GetGlobalLock() time.Time

	// SetGlobalLock sets a global pause until the given time.
	// Implementations should only advance the lock (never move it backwards).
	SetGlobalLock(until time.Time)

	// GetBucketHash returns the Discord bucket hash mapped to the given optimistic route key.
	GetBucketHash(optimisticKey string) (string, bool)

	// SetBucketHash maps an optimistic route key to a Discord bucket hash.
	SetBucketHash(optimisticKey, bucketHash string)
}

// MemoryStore is a high-performance, single-process Store implementation using sharded
// concurrent maps and atomic operations for minimal lock contention.
type MemoryStore struct {
	buckets      shardedmap.Map[string, BucketState]
	bucketHashes shardedmap.Map[string, string]
	globalLock   atomic.Int64 // UnixNano timestamp of global lock expiry
}

// NewMemoryStore creates a new in-memory rate limit store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		buckets:      shardedmap.New[string, BucketState](64),
		bucketHashes: shardedmap.New[string, string](32),
	}
}

// Reserve atomically checks and consumes 1 quota token from the bucket.
// Returns 0 if quota was available, or a positive duration if the caller must wait.
//
// When the bucket has expired (now > ResetAt), remaining is restored to the bucket limit
// and ResetAt is cleared (set to zero time). The next Update() from Discord's response
// will re-establish the correct window timing.
func (s *MemoryStore) Reserve(bucketKey string) time.Duration {
	var wait time.Duration
	s.buckets.Compute(bucketKey, func(state BucketState, exists bool) BucketState {
		if !exists {
			wait = 0
			return state
		}

		now := time.Now()

		// bucket has expired — reset remaining and clear the window
		// the next Update() call will set a fresh ResetAt from Discord's headers
		if !state.ResetAt.IsZero() && now.After(state.ResetAt) {
			state.Remaining = state.Limit
			state.ResetAt = time.Time{}
		} else if state.Remaining <= 0 && !state.ResetAt.IsZero() {
			wait = max(0, state.ResetAt.Sub(now))
			return state
		}

		if state.Remaining > 0 {
			state.Remaining--
		}
		wait = 0
		return state
	})
	return wait
}

// GetBucket retrieves the bucket state for the given key.
func (s *MemoryStore) GetBucket(bucketKey string) (BucketState, bool) {
	return s.buckets.Get(bucketKey)
}

// SetBucket stores the bucket state for the given key.
func (s *MemoryStore) SetBucket(bucketKey string, state BucketState) {
	s.buckets.Set(bucketKey, state)
}

// GetGlobalLock returns the time until which all requests should be paused.
// Returns zero time if no global lock is active.
func (s *MemoryStore) GetGlobalLock() time.Time {
	nano := s.globalLock.Load()
	if nano == 0 {
		return time.Time{}
	}
	return time.Unix(0, nano)
}

// SetGlobalLock sets a global pause until the given time.
// Uses CAS to ensure the lock only advances forward (never moves backwards).
func (s *MemoryStore) SetGlobalLock(until time.Time) {
	untilNano := until.UnixNano()
	for {
		cur := s.globalLock.Load()
		if untilNano <= cur {
			return
		}
		if s.globalLock.CompareAndSwap(cur, untilNano) {
			return
		}
	}
}

// GetBucketHash returns the Discord bucket hash mapped to the given optimistic route key.
func (s *MemoryStore) GetBucketHash(optimisticKey string) (string, bool) {
	return s.bucketHashes.Get(optimisticKey)
}

// SetBucketHash maps an optimistic route key to a Discord bucket hash.
func (s *MemoryStore) SetBucketHash(optimisticKey, bucketHash string) {
	s.bucketHashes.Set(optimisticKey, bucketHash)
}

// Cleanup removes expired bucket entries to prevent unbounded memory growth.
// This method is only available on MemoryStore (not on the Store interface)
// to avoid breaking external Store implementations.
func (s *MemoryStore) Cleanup() {
	now := time.Now()

	// collect expired keys first (Range holds a read lock per shard, so we
	// cannot call Delete inside Range without causing a deadlock)
	var expiredKeys []string
	s.buckets.Range(func(key string, state BucketState) bool {
		// remove buckets that have been expired for more than 5 minutes
		// this grace period prevents removing buckets that are about to be refreshed
		// by an in-flight response
		if !state.ResetAt.IsZero() && now.After(state.ResetAt.Add(5*time.Minute)) {
			expiredKeys = append(expiredKeys, key)
		}
		return true
	})

	for _, key := range expiredKeys {
		s.buckets.Delete(key)
	}
}
