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
	"github.com/wumpkit/wumpkit/internal/ratelimit"
)

// RateLimitBucketState represents the quota state for a rate-limit bucket.
type RateLimitBucketState = ratelimit.BucketState

// RateLimitStore abstracts the storage of rate-limit bucket states and global locks.
// Users can provide custom store implementations (e.g. Redis for multi-cluster setups).
type RateLimitStore = ratelimit.Store

// NewMemoryRateLimitStore returns a new in-memory rate-limit store.
func NewMemoryRateLimitStore() RateLimitStore {
	return ratelimit.NewMemoryStore()
}
