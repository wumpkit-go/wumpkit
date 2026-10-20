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
	"cmp"
	"net/http"
	"strconv"
	"time"
)

// HeaderInfo contains parsed Discord rate limit information extracted from HTTP response headers.
type HeaderInfo struct {
	Limit      int64         // Maximum requests allowed in the current window (-1 if absent).
	Remaining  int64         // Remaining requests in the current window (-1 if absent).
	ResetAfter time.Duration // Duration until the rate limit bucket resets (0 if absent).
	IsGlobal   bool          // True if this is a global rate limit (X-RateLimit-Global: true).
	Scope      string        // Rate limit scope: "user", "shared", or "global".
	BucketHash string        // Opaque bucket identifier from X-RateLimit-Bucket.
	HasHeaders bool          // True if any rate limit headers were present in the response.
}

// getHeader looks up a header value using both the canonical key (as Go normalizes it)
// and an alternative casing. This defensive dual-lookup handles edge cases where
// responses may pass through proxies or CDNs that don't canonicalize header names.
func getHeader(headers http.Header, canonicalKey, altKey string) string {
	if v, ok := headers[canonicalKey]; ok && len(v) > 0 {
		return v[0]
	}
	if v, ok := headers[altKey]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

// ParseHeaders extracts Discord rate limit information from HTTP response headers.
// It handles all three rate limit scopes (user, global, shared) and correctly
// selects between X-RateLimit-Reset-After and Retry-After based on the scope.
//
// For shared and global scopes, Retry-After is preferred over Reset-After since
// it reflects the actual wait time for the specific limit that was hit.
func ParseHeaders(headers http.Header, statusCode int) HeaderInfo {
	if headers == nil {
		return HeaderInfo{}
	}

	limitStr := getHeader(headers, "X-Ratelimit-Limit", "X-RateLimit-Limit")
	remainingStr := getHeader(headers, "X-Ratelimit-Remaining", "X-RateLimit-Remaining")
	resetAfterStr := getHeader(headers, "X-Ratelimit-Reset-After", "X-RateLimit-Reset-After")
	retryAfterStr := getHeader(headers, "Retry-After", "retry-after")
	scope := cmp.Or(getHeader(headers, "X-Ratelimit-Scope", "X-RateLimit-Scope"), "user")
	globalStr := getHeader(headers, "X-Ratelimit-Global", "X-RateLimit-Global")
	isGlobal := globalStr == "true"
	bucketHash := getHeader(headers, "X-Ratelimit-Bucket", "X-RateLimit-Bucket")

	// use Retry-After for globals, shared limits, or when reset-after is omitted
	// for shared/global 429s, Retry-After reflects the actual resource/global wait time
	// which may differ from the per-route Reset-After
	if resetAfterStr == "" || (scope != "user" && retryAfterStr != "") || (isGlobal && retryAfterStr != "") {
		if retryAfterStr != "" {
			resetAfterStr = retryAfterStr
		}
	}

	var resetDuration time.Duration
	if resetAfterStr != "" {
		if resetSeconds, err := strconv.ParseFloat(resetAfterStr, 64); err == nil && resetSeconds > 0 {
			resetDuration = time.Duration(resetSeconds * float64(time.Second))
		}
	}

	var limit int64 = -1
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = parsedLimit
		}
	}

	var remaining int64 = -1
	if remainingStr != "" {
		if parsedRemaining, err := strconv.ParseInt(remainingStr, 10, 64); err == nil {
			remaining = parsedRemaining
		}
	}

	hasHeaders := limit != -1 || remaining != -1 || resetDuration > 0 || isGlobal || bucketHash != ""

	return HeaderInfo{
		Limit:      limit,
		Remaining:  remaining,
		ResetAfter: resetDuration,
		IsGlobal:   isGlobal,
		Scope:      scope,
		BucketHash: bucketHash,
		HasHeaders: hasHeaders,
	}
}
