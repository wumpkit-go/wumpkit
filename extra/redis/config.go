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
	"time"

	"github.com/wumpkit/wumpkit"
)

// RedisCacheResourceTTLs specifies Time-To-Live durations for individual Discord entity types.
// A duration of 0 means the entity will use RedisCacheConfig.DefaultTTL (or never expire if DefaultTTL is 0).
type RedisCacheResourceTTLs struct {
	Users                time.Duration
	Guilds               time.Duration
	Members              time.Duration
	Channels             time.Duration
	Messages             time.Duration
	VoiceStates          time.Duration
	Roles                time.Duration
	Emojis               time.Duration
	Stickers             time.Duration
	StageInstances       time.Duration
	GuildScheduledEvents time.Duration
	AutoModerationRules  time.Duration
}

// RedisCacheSlidingExpiration configures sliding expiration behavior (resetting TTL on read for hot data).
// When enabled for a resource, reading that cached entity resets its expiration timer.
type RedisCacheSlidingExpiration struct {
	// All turns on sliding expiration for all resources that have a TTL configured.
	All bool

	// Granular per-resource sliding expiration toggles:
	Users                bool
	Guilds               bool
	Members              bool
	Channels             bool
	Messages             bool
	VoiceStates          bool
	Roles                bool
	Emojis               bool
	Stickers             bool
	StageInstances       bool
	GuildScheduledEvents bool
	AutoModerationRules  bool
}

// RedisCacheConfig configures the RedisCacheManager.
type RedisCacheConfig struct {
	// Prefix is the Redis key prefix used for namespacing (default: "wumpkit:cache:").
	Prefix string

	// Flags defines which entities to cache (default: wumpkit.CacheFlagsAll).
	Flags wumpkit.CacheFlags

	// Context is the base context for Redis commands (default: context.Background()).
	Context context.Context

	// DefaultTTL is the fallback expiration duration for cached items (0 = no expiration).
	DefaultTTL time.Duration

	// ResourceTTLs specifies fine-grained TTLs for each individual entity type.
	ResourceTTLs RedisCacheResourceTTLs

	// SlidingExpiration configures whether reading cached data resets its TTL.
	SlidingExpiration RedisCacheSlidingExpiration
}

// DefaultRedisCacheConfig returns standard default configuration for RedisCacheManager.
func DefaultRedisCacheConfig() RedisCacheConfig {
	return RedisCacheConfig{
		Prefix:  "wumpkit:cache:",
		Flags:   wumpkit.CacheFlagsAll,
		Context: context.Background(),
	}
}

// RedisCacheOption is a functional option for configuring RedisCacheManager.
type RedisCacheOption func(*redisCacheOptions)

type redisCacheOptions struct {
	config RedisCacheConfig
}

// WithRedisCacheConfig sets the complete RedisCacheConfig for RedisCacheManager.
func WithRedisCacheConfig(cfg RedisCacheConfig) RedisCacheOption {
	return func(o *redisCacheOptions) {
		o.config = cfg
	}
}

// WithRedisCachePrefix sets the key namespace prefix for Redis keys.
func WithRedisCachePrefix(prefix string) RedisCacheOption {
	return func(o *redisCacheOptions) {
		o.config.Prefix = prefix
	}
}

// WithRedisCacheFlags sets which Discord entities should be cached.
func WithRedisCacheFlags(flags wumpkit.CacheFlags) RedisCacheOption {
	return func(o *redisCacheOptions) {
		o.config.Flags = flags
	}
}

// WithRedisCacheDefaultTTL sets the fallback expiration duration for cached items.
func WithRedisCacheDefaultTTL(ttl time.Duration) RedisCacheOption {
	return func(o *redisCacheOptions) {
		o.config.DefaultTTL = ttl
	}
}
