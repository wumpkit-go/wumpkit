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

// Package ratelimit implements Discord REST API rate limiting with a two-tier design:
//
//  1. A global token bucket enforcing Discord's 50 requests/second limit (with exemptions
//     for interaction endpoints per Discord documentation).
//  2. Per-route bucket tracking using optimistic path normalization and Discord's
//     X-RateLimit-Bucket header for dynamic bucket discovery.
//
// The package uses lock-free atomics for the global bucket and sharded concurrent maps
// for per-route state, achieving high throughput with minimal contention under concurrent load.
package ratelimit

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/wumpkit-go/wumpkit/internal/shardedmap"
)

type routeKey struct {
	method string
	route  string
}

// Limiter orchestrates pre-request rate limit checks combining a global token bucket
// with per-route bucket state managed through a pluggable Store.
type Limiter struct {
	store        Store
	globalBucket Bucket
	pathCache    shardedmap.Map[routeKey, string]
}

// Option configures a Limiter during construction.
type Option func(*Limiter)

// WithStore sets a custom Store implementation for rate limit state persistence.
// Pass nil to keep the default in-memory store.
func WithStore(store Store) Option {
	return func(l *Limiter) {
		if store != nil {
			l.store = store
		}
	}
}

// WithGlobalLimit sets the global requests-per-second limit.
// Defaults to 50 per Discord documentation.
func WithGlobalLimit(limit uint) Option {
	return func(l *Limiter) {
		if limit > 0 {
			l.globalBucket = NewTokenBucket(limit, 1*time.Second)
		}
	}
}

// New creates a new Limiter with the given options.
// Default configuration: in-memory store, 50 req/s global limit.
func New(opts ...Option) *Limiter {
	l := &Limiter{
		store:        NewMemoryStore(),
		globalBucket: NewTokenBucket(50, 1*time.Second),
		pathCache:    shardedmap.New[routeKey, string](32),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// getOptimisticPath returns the normalized bucket path for a route, using a cache
// to avoid repeated path parsing on hot paths.
func (l *Limiter) bucketPath(route, method string) string {
	key := routeKey{method: method, route: route}
	if path, ok := l.pathCache.Get(key); ok {
		return path
	}

	path := NormalizeRoute(route, method)
	l.pathCache.Set(key, path)
	return path
}

// resolveBucketKey maps a normalized route path to a composite bucket key combining
// the Discord bucket hash with the route's major parameter.
func (l *Limiter) resolveBucketKey(bPath string) string {
	hash, ok := l.store.GetBucketHash(bPath)
	if !ok || hash == "" {
		return bPath
	}

	major := majorParam(bPath)
	if major != "" && major != "!" {
		return major + ":" + hash
	}
	return hash
}

// majorParam extracts the major parameter (second path segment) from a normalized path.
func majorParam(bPath string) string {
	if len(bPath) == 0 || bPath[0] != '/' {
		return ""
	}
	rest := bPath[1:]

	slashIdx := strings.IndexByte(rest, '/')
	if slashIdx == -1 {
		return ""
	}
	rest = rest[slashIdx+1:]

	if nextSlash := strings.IndexByte(rest, '/'); nextSlash != -1 {
		return rest[:nextSlash]
	}
	return rest
}

// Wait blocks until the caller is permitted to send a request to the given route.
// It enforces both the global rate limit (50 req/s, skipped for interaction endpoints)
// and per-route bucket limits. Returns ctx.Err() if the context is cancelled while waiting.
func (l *Limiter) Wait(ctx context.Context, route, method string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	bPath := l.bucketPath(route, method)
	isInteraction := IsInteractionRoute(bPath)

	// 1. global pause check (skip for interaction endpoints per Discord docs)
	if !isInteraction {
		if globalUntil := l.store.GetGlobalLock(); !globalUntil.IsZero() {
			now := time.Now()
			if now.Before(globalUntil) {
				select {
				case <-time.After(time.Until(globalUntil)):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		// 2. global 50 req/s token bucket (skip for interaction endpoints per Discord docs)
		for {
			if err := l.globalBucket.Add(1); err == nil {
				break
			}
			reset := l.globalBucket.Reset()
			wait := time.Until(reset)
			if wait <= 0 {
				continue
			}
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// 3. per-route bucket check (applies to all routes including interactions)
	bucketKey := l.resolveBucketKey(bPath)

	for {
		waitDuration := l.store.Reserve(bucketKey)
		if waitDuration <= 0 {
			break
		}

		select {
		case <-time.After(waitDuration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// Update processes Discord rate limit response headers and updates internal state.
// Must be called after every HTTP response to keep bucket state synchronized.
func (l *Limiter) Update(route, method string, headers http.Header, statusCode int) {
	if headers == nil {
		return
	}

	info := ParseHeaders(headers, statusCode)
	bPath := l.bucketPath(route, method)

	if info.BucketHash != "" {
		l.store.SetBucketHash(bPath, info.BucketHash)
	}

	bucketKey := l.resolveBucketKey(bPath)

	// handle global rate limit: pause all non-interaction requests
	if info.IsGlobal || (statusCode == http.StatusTooManyRequests && info.Scope == "global") {
		until := time.Now().Add(info.ResetAfter)
		l.store.SetGlobalLock(until)
	}

	if info.HasHeaders || statusCode == http.StatusTooManyRequests {
		resetAt := time.Now().Add(info.ResetAfter)
		remaining := info.Remaining

		if statusCode == http.StatusTooManyRequests {
			// for shared scope 429s, the per-route bucket may still have capacity —
			// only the underlying resource is throttled. Don't zero out remaining;
			// instead, use the reported remaining (which may be > 0) and set the
			// reset time from the Retry-After to block this specific bucket
			if info.Scope != "shared" {
				remaining = 0
			}
		}

		l.store.SetBucket(bucketKey, BucketState{
			Limit:      info.Limit,
			Remaining:  remaining,
			ResetAt:    resetAt,
			BucketHash: info.BucketHash,
		})
	}
}
