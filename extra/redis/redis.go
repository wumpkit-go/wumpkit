/************************************************************************************
 *
 * wumpkit, A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD 3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package redis

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wumpkit/wumpkit"
)

const (
	defaultPrefix = "wumpkit:cache:"
)

// RedisCacheManager implements wumpkit.CacheManager using Redis for distributed caching
// across multiple bot instances / clusters.
type RedisCacheManager struct {
	client       redis.UniversalClient
	config       RedisCacheConfig
	prefix       string
	ctx          context.Context
	resourceTTLs map[wumpkit.CacheFlags]time.Duration
	slidingFlags wumpkit.CacheFlags
	mu           sync.RWMutex
}

var _ wumpkit.CacheManager = (*RedisCacheManager)(nil)

// NewRedisCacheManager creates a new RedisCacheManager backed by the provided Redis UniversalClient and optional functional options.
//
// Usage with default configuration:
//
//	cache := redis.NewRedisCacheManager(rdb)
//
// Usage with custom configuration options:
//
//	cache := redis.NewRedisCacheManager(rdb, redis.WithRedisCacheConfig(redis.RedisCacheConfig{
//		Prefix:     "mybot:cache:",
//		DefaultTTL: 1 * time.Hour,
//	}))
func NewRedisCacheManager(
	client redis.UniversalClient,
	opts ...RedisCacheOption,
) *RedisCacheManager {
	optsData := redisCacheOptions{
		config: DefaultRedisCacheConfig(),
	}
	for _, opt := range opts {
		opt(&optsData)
	}

	cfg := optsData.config
	if cfg.Prefix == "" {
		cfg.Prefix = defaultPrefix
	}
	if cfg.Flags == 0 {
		cfg.Flags = wumpkit.CacheFlagsAll
	}
	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	ttls := make(map[wumpkit.CacheFlags]time.Duration)
	if cfg.ResourceTTLs.Users > 0 {
		ttls[wumpkit.CacheFlagUsers] = cfg.ResourceTTLs.Users
	}
	if cfg.ResourceTTLs.Guilds > 0 {
		ttls[wumpkit.CacheFlagGuilds] = cfg.ResourceTTLs.Guilds
	}
	if cfg.ResourceTTLs.Members > 0 {
		ttls[wumpkit.CacheFlagMembers] = cfg.ResourceTTLs.Members
	}
	if cfg.ResourceTTLs.Channels > 0 {
		ttls[wumpkit.CacheFlagChannels] = cfg.ResourceTTLs.Channels
	}
	if cfg.ResourceTTLs.Messages > 0 {
		ttls[wumpkit.CacheFlagMessages] = cfg.ResourceTTLs.Messages
	}
	if cfg.ResourceTTLs.VoiceStates > 0 {
		ttls[wumpkit.CacheFlagVoiceStates] = cfg.ResourceTTLs.VoiceStates
	}
	if cfg.ResourceTTLs.Roles > 0 {
		ttls[wumpkit.CacheFlagRoles] = cfg.ResourceTTLs.Roles
	}
	if cfg.ResourceTTLs.Emojis > 0 {
		ttls[wumpkit.CacheFlagEmojis] = cfg.ResourceTTLs.Emojis
	}
	if cfg.ResourceTTLs.Stickers > 0 {
		ttls[wumpkit.CacheFlagStickers] = cfg.ResourceTTLs.Stickers
	}
	if cfg.ResourceTTLs.StageInstances > 0 {
		ttls[wumpkit.CacheFlagStageInstances] = cfg.ResourceTTLs.StageInstances
	}
	if cfg.ResourceTTLs.GuildScheduledEvents > 0 {
		ttls[wumpkit.CacheFlagGuildScheduledEvents] = cfg.ResourceTTLs.GuildScheduledEvents
	}
	if cfg.ResourceTTLs.AutoModerationRules > 0 {
		ttls[wumpkit.CacheFlagAutoModerationRules] = cfg.ResourceTTLs.AutoModerationRules
	}

	var sliding wumpkit.CacheFlags
	if cfg.SlidingExpiration.All {
		sliding = wumpkit.CacheFlagsAll
	} else {
		if cfg.SlidingExpiration.Users {
			sliding |= wumpkit.CacheFlagUsers
		}
		if cfg.SlidingExpiration.Guilds {
			sliding |= wumpkit.CacheFlagGuilds
		}
		if cfg.SlidingExpiration.Members {
			sliding |= wumpkit.CacheFlagMembers
		}
		if cfg.SlidingExpiration.Channels {
			sliding |= wumpkit.CacheFlagChannels
		}
		if cfg.SlidingExpiration.Messages {
			sliding |= wumpkit.CacheFlagMessages
		}
		if cfg.SlidingExpiration.VoiceStates {
			sliding |= wumpkit.CacheFlagVoiceStates
		}
		if cfg.SlidingExpiration.Roles {
			sliding |= wumpkit.CacheFlagRoles
		}
		if cfg.SlidingExpiration.Emojis {
			sliding |= wumpkit.CacheFlagEmojis
		}
		if cfg.SlidingExpiration.Stickers {
			sliding |= wumpkit.CacheFlagStickers
		}
		if cfg.SlidingExpiration.StageInstances {
			sliding |= wumpkit.CacheFlagStageInstances
		}
		if cfg.SlidingExpiration.GuildScheduledEvents {
			sliding |= wumpkit.CacheFlagGuildScheduledEvents
		}
		if cfg.SlidingExpiration.AutoModerationRules {
			sliding |= wumpkit.CacheFlagAutoModerationRules
		}
	}

	return &RedisCacheManager{
		client:       client,
		config:       cfg,
		prefix:       cfg.Prefix,
		ctx:          cfg.Context,
		resourceTTLs: ttls,
		slidingFlags: sliding,
	}
}

// Config returns the current configuration of the RedisCacheManager.
func (r *RedisCacheManager) Config() RedisCacheConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// CacheFlags returns the currently configured CacheFlags.
func (r *RedisCacheManager) CacheFlags() wumpkit.CacheFlags {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config.Flags
}

// SetCacheFlags sets the active CacheFlags.
func (r *RedisCacheManager) SetCacheFlags(flags ...wumpkit.CacheFlags) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var total wumpkit.CacheFlags
	for _, f := range flags {
		total |= f
	}
	r.config.Flags = total
}

// Close gracefully closes the underlying Redis client connection.
func (r *RedisCacheManager) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Shutdown is an alias for Close.
func (r *RedisCacheManager) Shutdown() error {
	return r.Close()
}

func (r *RedisCacheManager) getTTL(flag wumpkit.CacheFlags) time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.resourceTTLs != nil {
		if ttl, ok := r.resourceTTLs[flag]; ok && ttl > 0 {
			return ttl
		}
	}
	return r.config.DefaultTTL
}

func (r *RedisCacheManager) isSliding(flag wumpkit.CacheFlags) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.slidingFlags.Has(flag)
}

// Key helpers

func (r *RedisCacheManager) userKey(userID wumpkit.Snowflake) string {
	return r.prefix + "user:" + userID.String()
}

func (r *RedisCacheManager) guildKey(guildID wumpkit.Snowflake) string {
	return r.prefix + "guild:" + guildID.String()
}

func (r *RedisCacheManager) memberKey(guildID, userID wumpkit.Snowflake) string {
	return r.prefix + "member:" + guildID.String() + ":" + userID.String()
}

func (r *RedisCacheManager) guildMembersSetKey(guildID wumpkit.Snowflake) string {
	return r.prefix + "guild_members:" + guildID.String()
}

func (r *RedisCacheManager) channelKey(channelID wumpkit.Snowflake) string {
	return r.prefix + "channel:" + channelID.String()
}

func (r *RedisCacheManager) guildChannelsSetKey(guildID wumpkit.Snowflake) string {
	return r.prefix + "guild_channels:" + guildID.String()
}

func (r *RedisCacheManager) messageKey(messageID wumpkit.Snowflake) string {
	return r.prefix + "message:" + messageID.String()
}

func (r *RedisCacheManager) voiceStateKey(guildID, userID wumpkit.Snowflake) string {
	return r.prefix + "voicestate:" + guildID.String() + ":" + userID.String()
}

func (r *RedisCacheManager) guildVoiceStatesSetKey(guildID wumpkit.Snowflake) string {
	return r.prefix + "guild_voicestates:" + guildID.String()
}

func (r *RedisCacheManager) channelVoiceStatesSetKey(channelID wumpkit.Snowflake) string {
	return r.prefix + "channel_voicestates:" + channelID.String()
}

func (r *RedisCacheManager) roleKey(roleID wumpkit.Snowflake) string {
	return r.prefix + "role:" + roleID.String()
}

func (r *RedisCacheManager) guildRolesSetKey(guildID wumpkit.Snowflake) string {
	return r.prefix + "guild_roles:" + guildID.String()
}

func (r *RedisCacheManager) emojiKey(emojiID wumpkit.Snowflake) string {
	return r.prefix + "emoji:" + emojiID.String()
}

func (r *RedisCacheManager) guildEmojisSetKey(guildID wumpkit.Snowflake) string {
	return r.prefix + "guild_emojis:" + guildID.String()
}

func (r *RedisCacheManager) stickerKey(stickerID wumpkit.Snowflake) string {
	return r.prefix + "sticker:" + stickerID.String()
}

func (r *RedisCacheManager) guildStickersSetKey(guildID wumpkit.Snowflake) string {
	return r.prefix + "guild_stickers:" + guildID.String()
}

func (r *RedisCacheManager) stageInstanceKey(channelID wumpkit.Snowflake) string {
	return r.prefix + "stage:" + channelID.String()
}

func (r *RedisCacheManager) scheduledEventKey(eventID wumpkit.Snowflake) string {
	return r.prefix + "event:" + eventID.String()
}

func (r *RedisCacheManager) autoModRuleKey(ruleID wumpkit.Snowflake) string {
	return r.prefix + "automod:" + ruleID.String()
}

func (r *RedisCacheManager) ctxOrFallback(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

func (r *RedisCacheManager) setJSON(ctx context.Context, key string, flag wumpkit.CacheFlags, val any) error {
	ctx = r.ctxOrFallback(ctx)
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ttl := r.getTTL(flag)
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCacheManager) getJSON(ctx context.Context, key string, flag wumpkit.CacheFlags, dest any) bool {
	ctx = r.ctxOrFallback(ctx)
	ttl := r.getTTL(flag)
	var data []byte
	var err error
	if r.isSliding(flag) && ttl > 0 {
		data, err = r.client.GetEx(ctx, key, ttl).Bytes()
	} else {
		data, err = r.client.Get(ctx, key).Bytes()
	}
	if err != nil {
		return false
	}
	return json.Unmarshal(data, dest) == nil
}

func (r *RedisCacheManager) hasKey(ctx context.Context, key string) bool {
	ctx = r.ctxOrFallback(ctx)
	n, err := r.client.Exists(ctx, key).Result()
	return err == nil && n > 0
}

func (r *RedisCacheManager) delKeys(ctx context.Context, keys ...string) bool {
	ctx = r.ctxOrFallback(ctx)
	if len(keys) == 0 {
		return false
	}
	n, err := r.client.Del(ctx, keys...).Result()
	return err == nil && n > 0
}

/*****************************
 *   Single Entity Lookups
 *****************************/

func (r *RedisCacheManager) GetUser(ctx context.Context, userID wumpkit.Snowflake) (wumpkit.User, bool) {
	var user wumpkit.User
	ok := r.getJSON(ctx, r.userKey(userID), wumpkit.CacheFlagUsers, &user)
	return user, ok
}

func (r *RedisCacheManager) GetGuild(ctx context.Context, guildID wumpkit.Snowflake) (wumpkit.Guild, bool) {
	var guild wumpkit.Guild
	ok := r.getJSON(ctx, r.guildKey(guildID), wumpkit.CacheFlagGuilds, &guild)
	return guild, ok
}

func (r *RedisCacheManager) GetMember(ctx context.Context, guildID, userID wumpkit.Snowflake) (wumpkit.Member, bool) {
	var member wumpkit.Member
	ok := r.getJSON(ctx, r.memberKey(guildID, userID), wumpkit.CacheFlagMembers, &member)
	return member, ok
}

func (r *RedisCacheManager) GetChannel(ctx context.Context, channelID wumpkit.Snowflake) (wumpkit.Channel, bool) {
	ttl := r.getTTL(wumpkit.CacheFlagChannels)
	var data []byte
	var err error
	if r.isSliding(wumpkit.CacheFlagChannels) && ttl > 0 {
		data, err = r.client.GetEx(ctx, r.channelKey(channelID), ttl).Bytes()
	} else {
		data, err = r.client.Get(ctx, r.channelKey(channelID)).Bytes()
	}
	if err != nil {
		return nil, false
	}
	ch, err := wumpkit.UnmarshalChannel(data)
	if err != nil {
		return nil, false
	}
	return ch, true
}

func (r *RedisCacheManager) GetMessage(ctx context.Context, messageID wumpkit.Snowflake) (wumpkit.Message, bool) {
	var msg wumpkit.Message
	ok := r.getJSON(ctx, r.messageKey(messageID), wumpkit.CacheFlagMessages, &msg)
	return msg, ok
}

func (r *RedisCacheManager) GetVoiceState(ctx context.Context, guildID, userID wumpkit.Snowflake) (wumpkit.VoiceState, bool) {
	var vs wumpkit.VoiceState
	ok := r.getJSON(ctx, r.voiceStateKey(guildID, userID), wumpkit.CacheFlagVoiceStates, &vs)
	return vs, ok
}

func (r *RedisCacheManager) GetRole(ctx context.Context, roleID wumpkit.Snowflake) (wumpkit.Role, bool) {
	var role wumpkit.Role
	ok := r.getJSON(ctx, r.roleKey(roleID), wumpkit.CacheFlagRoles, &role)
	return role, ok
}

func (r *RedisCacheManager) GetRoles(ctx context.Context, rolesIDs ...wumpkit.Snowflake) map[wumpkit.Snowflake]wumpkit.Role {
	ctx = r.ctxOrFallback(ctx)
	out := make(map[wumpkit.Snowflake]wumpkit.Role, len(rolesIDs))
	if len(rolesIDs) == 0 {
		return out
	}
	keys := make([]string, len(rolesIDs))
	for i, id := range rolesIDs {
		keys[i] = r.roleKey(id)
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return out
	}
	if r.isSliding(wumpkit.CacheFlagRoles) && r.getTTL(wumpkit.CacheFlagRoles) > 0 {
		ttl := r.getTTL(wumpkit.CacheFlagRoles)
		pipe := r.client.Pipeline()
		for _, k := range keys {
			pipe.Expire(ctx, k, ttl)
		}
		_, _ = pipe.Exec(ctx)
	}
	for i, val := range vals {
		if val == nil {
			continue
		}
		var role wumpkit.Role
		switch v := val.(type) {
		case string:
			if json.Unmarshal([]byte(v), &role) == nil {
				out[rolesIDs[i]] = role
			}
		case []byte:
			if json.Unmarshal(v, &role) == nil {
				out[rolesIDs[i]] = role
			}
		}
	}
	return out
}

func (r *RedisCacheManager) GetEmoji(ctx context.Context, emojiID wumpkit.Snowflake) (wumpkit.Emoji, bool) {
	var emoji wumpkit.Emoji
	ok := r.getJSON(ctx, r.emojiKey(emojiID), wumpkit.CacheFlagEmojis, &emoji)
	return emoji, ok
}

func (r *RedisCacheManager) GetSticker(ctx context.Context, stickerID wumpkit.Snowflake) (wumpkit.Sticker, bool) {
	var sticker wumpkit.Sticker
	ok := r.getJSON(ctx, r.stickerKey(stickerID), wumpkit.CacheFlagStickers, &sticker)
	return sticker, ok
}

func (r *RedisCacheManager) GetStageInstance(ctx context.Context, channelID wumpkit.Snowflake) (wumpkit.StageInstance, bool) {
	var si wumpkit.StageInstance
	ok := r.getJSON(ctx, r.stageInstanceKey(channelID), wumpkit.CacheFlagStageInstances, &si)
	return si, ok
}

func (r *RedisCacheManager) GetGuildScheduledEvent(ctx context.Context, eventID wumpkit.Snowflake) (wumpkit.GuildScheduledEvent, bool) {
	var event wumpkit.GuildScheduledEvent
	ok := r.getJSON(ctx, r.scheduledEventKey(eventID), wumpkit.CacheFlagGuildScheduledEvents, &event)
	return event, ok
}

func (r *RedisCacheManager) GetAutoModerationRule(ctx context.Context, ruleID wumpkit.Snowflake) (wumpkit.AutoModerationRule, bool) {
	var rule wumpkit.AutoModerationRule
	ok := r.getJSON(ctx, r.autoModRuleKey(ruleID), wumpkit.CacheFlagAutoModerationRules, &rule)
	return rule, ok
}

/*****************************
 *   Group / Relational Lookups
 *****************************/

func (r *RedisCacheManager) GetGuildChannels(ctx context.Context, guildID wumpkit.Snowflake) (map[wumpkit.Snowflake]wumpkit.GuildChannel, bool) {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildChannelsSetKey(guildID)).Result()
	if err != nil || len(ids) == 0 {
		return nil, false
	}
	keys := make([]string, len(ids))
	for i, idStr := range ids {
		keys[i] = r.prefix + "channel:" + idStr
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, false
	}
	out := make(map[wumpkit.Snowflake]wumpkit.GuildChannel, len(vals))
	for _, val := range vals {
		if val == nil {
			continue
		}
		var raw []byte
		switch v := val.(type) {
		case string:
			raw = []byte(v)
		case []byte:
			raw = v
		}
		ch, err := wumpkit.UnmarshalChannel(raw)
		if err == nil {
			if gc, ok := ch.(wumpkit.GuildChannel); ok {
				out[gc.GetID()] = gc
			}
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func (r *RedisCacheManager) GetGuildMembers(ctx context.Context, guildID wumpkit.Snowflake) (map[wumpkit.Snowflake]wumpkit.Member, bool) {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildMembersSetKey(guildID)).Result()
	if err != nil || len(ids) == 0 {
		return nil, false
	}
	keys := make([]string, len(ids))
	for i, idStr := range ids {
		keys[i] = r.prefix + "member:" + guildID.String() + ":" + idStr
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, false
	}
	out := make(map[wumpkit.Snowflake]wumpkit.Member, len(vals))
	for _, val := range vals {
		if val == nil {
			continue
		}
		var m wumpkit.Member
		var raw []byte
		switch v := val.(type) {
		case string:
			raw = []byte(v)
		case []byte:
			raw = v
		}
		if json.Unmarshal(raw, &m) == nil {
			out[m.ID] = m
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func (r *RedisCacheManager) GetGuildVoiceStates(ctx context.Context, guildID wumpkit.Snowflake) (map[wumpkit.Snowflake]wumpkit.VoiceState, bool) {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildVoiceStatesSetKey(guildID)).Result()
	if err != nil || len(ids) == 0 {
		return nil, false
	}
	keys := make([]string, len(ids))
	for i, idStr := range ids {
		keys[i] = r.prefix + "voicestate:" + guildID.String() + ":" + idStr
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, false
	}
	out := make(map[wumpkit.Snowflake]wumpkit.VoiceState, len(vals))
	for _, val := range vals {
		if val == nil {
			continue
		}
		var vs wumpkit.VoiceState
		var raw []byte
		switch v := val.(type) {
		case string:
			raw = []byte(v)
		case []byte:
			raw = v
		}
		if json.Unmarshal(raw, &vs) == nil {
			out[vs.UserID] = vs
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func (r *RedisCacheManager) GetGuildRoles(ctx context.Context, guildID wumpkit.Snowflake) (map[wumpkit.Snowflake]wumpkit.Role, bool) {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildRolesSetKey(guildID)).Result()
	if err != nil || len(ids) == 0 {
		return nil, false
	}
	keys := make([]string, len(ids))
	for i, idStr := range ids {
		keys[i] = r.prefix + "role:" + idStr
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, false
	}
	out := make(map[wumpkit.Snowflake]wumpkit.Role, len(vals))
	for _, val := range vals {
		if val == nil {
			continue
		}
		var role wumpkit.Role
		var raw []byte
		switch v := val.(type) {
		case string:
			raw = []byte(v)
		case []byte:
			raw = v
		}
		if json.Unmarshal(raw, &role) == nil {
			out[role.ID] = role
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func (r *RedisCacheManager) GetGuildEmojis(ctx context.Context, guildID wumpkit.Snowflake) (map[wumpkit.Snowflake]wumpkit.Emoji, bool) {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildEmojisSetKey(guildID)).Result()
	if err != nil || len(ids) == 0 {
		return nil, false
	}
	keys := make([]string, len(ids))
	for i, idStr := range ids {
		keys[i] = r.prefix + "emoji:" + idStr
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, false
	}
	out := make(map[wumpkit.Snowflake]wumpkit.Emoji, len(vals))
	for _, val := range vals {
		if val == nil {
			continue
		}
		var emoji wumpkit.Emoji
		var raw []byte
		switch v := val.(type) {
		case string:
			raw = []byte(v)
		case []byte:
			raw = v
		}
		if json.Unmarshal(raw, &emoji) == nil {
			out[emoji.ID] = emoji
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func (r *RedisCacheManager) GetGuildStickers(ctx context.Context, guildID wumpkit.Snowflake) (map[wumpkit.Snowflake]wumpkit.Sticker, bool) {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildStickersSetKey(guildID)).Result()
	if err != nil || len(ids) == 0 {
		return nil, false
	}
	keys := make([]string, len(ids))
	for i, idStr := range ids {
		keys[i] = r.prefix + "sticker:" + idStr
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, false
	}
	out := make(map[wumpkit.Snowflake]wumpkit.Sticker, len(vals))
	for _, val := range vals {
		if val == nil {
			continue
		}
		var sticker wumpkit.Sticker
		var raw []byte
		switch v := val.(type) {
		case string:
			raw = []byte(v)
		case []byte:
			raw = v
		}
		if json.Unmarshal(raw, &sticker) == nil {
			out[sticker.ID] = sticker
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func (r *RedisCacheManager) GetVoiceChannelMembers(ctx context.Context, channelID wumpkit.Snowflake) (map[wumpkit.Snowflake]wumpkit.Member, bool) {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.channelVoiceStatesSetKey(channelID)).Result()
	if err != nil || len(ids) == 0 {
		return nil, false
	}
	out := make(map[wumpkit.Snowflake]wumpkit.Member, len(ids))
	for _, idStr := range ids {
		uid, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		// search member across guilds or get voicestate first to find guild
		var vs wumpkit.VoiceState
		// find voicestate matching channel
		pattern := r.prefix + "voicestate:*:" + idStr
		keys, err := r.client.Keys(ctx, pattern).Result()
		if err == nil && len(keys) > 0 {
			if r.getJSON(ctx, keys[0], wumpkit.CacheFlagVoiceStates, &vs) && vs.ChannelID == channelID {
				if m, ok := r.GetMember(ctx, vs.GuildID, wumpkit.Snowflake(uid)); ok {
					out[m.ID] = m
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

/*****************************
 *   Presence Checks
 *****************************/

func (r *RedisCacheManager) HasUser(ctx context.Context, userID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.userKey(userID))
}

func (r *RedisCacheManager) HasGuild(ctx context.Context, guildID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.guildKey(guildID))
}

func (r *RedisCacheManager) HasMember(ctx context.Context, guildID, userID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.memberKey(guildID, userID))
}

func (r *RedisCacheManager) HasChannel(ctx context.Context, channelID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.channelKey(channelID))
}

func (r *RedisCacheManager) HasMessage(ctx context.Context, messageID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.messageKey(messageID))
}

func (r *RedisCacheManager) HasVoiceState(ctx context.Context, guildID, userID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.voiceStateKey(guildID, userID))
}

func (r *RedisCacheManager) HasRole(ctx context.Context, roleID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.roleKey(roleID))
}

func (r *RedisCacheManager) HasRoles(ctx context.Context, rolesIDs ...wumpkit.Snowflake) bool {
	for _, id := range rolesIDs {
		if !r.HasRole(ctx, id) {
			return false
		}
	}
	return true
}

func (r *RedisCacheManager) HasEmoji(ctx context.Context, emojiID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.emojiKey(emojiID))
}

func (r *RedisCacheManager) HasSticker(ctx context.Context, stickerID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.stickerKey(stickerID))
}

func (r *RedisCacheManager) HasStageInstance(ctx context.Context, channelID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.stageInstanceKey(channelID))
}

func (r *RedisCacheManager) HasGuildScheduledEvent(ctx context.Context, eventID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.scheduledEventKey(eventID))
}

func (r *RedisCacheManager) HasAutoModerationRule(ctx context.Context, ruleID wumpkit.Snowflake) bool {
	return r.hasKey(ctx, r.autoModRuleKey(ruleID))
}

/*****************************
 *   Counts
 *****************************/

func (r *RedisCacheManager) countKeysByPattern(ctx context.Context, pattern string) int {
	ctx = r.ctxOrFallback(ctx)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0
	}
	return len(keys)
}

func (r *RedisCacheManager) countSetMembers(ctx context.Context, key string) int {
	ctx = r.ctxOrFallback(ctx)
	n, err := r.client.SCard(ctx, key).Result()
	if err != nil {
		return 0
	}
	return int(n)
}

func (r *RedisCacheManager) CountUsers(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"user:*")
}

func (r *RedisCacheManager) CountGuilds(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"guild:*")
}

func (r *RedisCacheManager) CountMembers(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"member:*")
}

func (r *RedisCacheManager) CountChannels(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"channel:*")
}

func (r *RedisCacheManager) CountMessages(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"message:*")
}

func (r *RedisCacheManager) CountVoiceStates(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"voicestate:*")
}

func (r *RedisCacheManager) CountRoles(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"role:*")
}

func (r *RedisCacheManager) CountEmojis(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"emoji:*")
}

func (r *RedisCacheManager) CountStickers(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"sticker:*")
}

func (r *RedisCacheManager) CountStageInstances(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"stage:*")
}

func (r *RedisCacheManager) CountGuildScheduledEvents(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"event:*")
}

func (r *RedisCacheManager) CountAutoModerationRules(ctx context.Context) int {
	return r.countKeysByPattern(ctx, r.prefix+"automod:*")
}

func (r *RedisCacheManager) CountGuildChannels(ctx context.Context, guildID wumpkit.Snowflake) int {
	return r.countSetMembers(ctx, r.guildChannelsSetKey(guildID))
}

func (r *RedisCacheManager) CountGuildMembers(ctx context.Context, guildID wumpkit.Snowflake) int {
	return r.countSetMembers(ctx, r.guildMembersSetKey(guildID))
}

func (r *RedisCacheManager) CountGuildRoles(ctx context.Context, guildID wumpkit.Snowflake) int {
	return r.countSetMembers(ctx, r.guildRolesSetKey(guildID))
}

func (r *RedisCacheManager) CountGuildEmojis(ctx context.Context, guildID wumpkit.Snowflake) int {
	return r.countSetMembers(ctx, r.guildEmojisSetKey(guildID))
}

func (r *RedisCacheManager) CountGuildStickers(ctx context.Context, guildID wumpkit.Snowflake) int {
	return r.countSetMembers(ctx, r.guildStickersSetKey(guildID))
}

/*****************************
 *   Mutators (Add)
 *****************************/

func (r *RedisCacheManager) AddUser(ctx context.Context, user wumpkit.User) {
	if !r.CacheFlags().Has(wumpkit.CacheFlagUsers) {
		return
	}
	_ = r.setJSON(ctx, r.userKey(user.ID), wumpkit.CacheFlagUsers, user)
}

func (r *RedisCacheManager) AddGuild(ctx context.Context, guild wumpkit.Guild) {
	if !r.CacheFlags().Has(wumpkit.CacheFlagGuilds) {
		return
	}
	_ = r.setJSON(ctx, r.guildKey(guild.ID), wumpkit.CacheFlagGuilds, guild)
}

func (r *RedisCacheManager) AddMember(ctx context.Context, member wumpkit.Member) {
	ctx = r.ctxOrFallback(ctx)
	if !r.CacheFlags().Has(wumpkit.CacheFlagMembers) {
		return
	}
	pipe := r.client.Pipeline()
	data, _ := json.Marshal(member)
	ttl := r.getTTL(wumpkit.CacheFlagMembers)
	pipe.Set(ctx, r.memberKey(member.GuildID, member.ID), data, ttl)
	pipe.SAdd(ctx, r.guildMembersSetKey(member.GuildID), member.ID.String())
	_, _ = pipe.Exec(ctx)
}

func (r *RedisCacheManager) AddChannel(ctx context.Context, channel wumpkit.Channel) {
	ctx = r.ctxOrFallback(ctx)
	if !r.CacheFlags().Has(wumpkit.CacheFlagChannels) || channel == nil {
		return
	}
	pipe := r.client.Pipeline()
	data, _ := json.Marshal(channel)
	ttl := r.getTTL(wumpkit.CacheFlagChannels)
	pipe.Set(ctx, r.channelKey(channel.GetID()), data, ttl)
	if gc, ok := channel.(wumpkit.GuildChannel); ok && gc.GetGuildID() != 0 {
		pipe.SAdd(ctx, r.guildChannelsSetKey(gc.GetGuildID()), channel.GetID().String())
	}
	_, _ = pipe.Exec(ctx)
}

func (r *RedisCacheManager) AddMessage(ctx context.Context, message wumpkit.Message) {
	if !r.CacheFlags().Has(wumpkit.CacheFlagMessages) {
		return
	}
	_ = r.setJSON(ctx, r.messageKey(message.ID), wumpkit.CacheFlagMessages, message)
}

func (r *RedisCacheManager) AddVoiceState(ctx context.Context, voiceState wumpkit.VoiceState) {
	ctx = r.ctxOrFallback(ctx)
	if !r.CacheFlags().Has(wumpkit.CacheFlagVoiceStates) {
		return
	}
	pipe := r.client.Pipeline()
	data, _ := json.Marshal(voiceState)
	ttl := r.getTTL(wumpkit.CacheFlagVoiceStates)
	pipe.Set(ctx, r.voiceStateKey(voiceState.GuildID, voiceState.UserID), data, ttl)
	pipe.SAdd(ctx, r.guildVoiceStatesSetKey(voiceState.GuildID), voiceState.UserID.String())
	if voiceState.ChannelID != 0 {
		pipe.SAdd(ctx, r.channelVoiceStatesSetKey(voiceState.ChannelID), voiceState.UserID.String())
	}
	_, _ = pipe.Exec(ctx)
}

func (r *RedisCacheManager) AddRole(ctx context.Context, role wumpkit.Role) {
	ctx = r.ctxOrFallback(ctx)
	if !r.CacheFlags().Has(wumpkit.CacheFlagRoles) {
		return
	}
	pipe := r.client.Pipeline()
	data, _ := json.Marshal(role)
	ttl := r.getTTL(wumpkit.CacheFlagRoles)
	pipe.Set(ctx, r.roleKey(role.ID), data, ttl)
	if role.GuildID != 0 {
		pipe.SAdd(ctx, r.guildRolesSetKey(role.GuildID), role.ID.String())
	}
	_, _ = pipe.Exec(ctx)
}

func (r *RedisCacheManager) AddRoles(ctx context.Context, roles ...wumpkit.Role) {
	for _, role := range roles {
		r.AddRole(ctx, role)
	}
}

func (r *RedisCacheManager) AddEmoji(ctx context.Context, emoji wumpkit.Emoji) {
	ctx = r.ctxOrFallback(ctx)
	if !r.CacheFlags().Has(wumpkit.CacheFlagEmojis) {
		return
	}
	pipe := r.client.Pipeline()
	data, _ := json.Marshal(emoji)
	ttl := r.getTTL(wumpkit.CacheFlagEmojis)
	pipe.Set(ctx, r.emojiKey(emoji.ID), data, ttl)
	if emoji.GuildID != 0 {
		pipe.SAdd(ctx, r.guildEmojisSetKey(emoji.GuildID), emoji.ID.String())
	}
	_, _ = pipe.Exec(ctx)
}

func (r *RedisCacheManager) AddEmojis(ctx context.Context, emojis ...wumpkit.Emoji) {
	for _, emoji := range emojis {
		r.AddEmoji(ctx, emoji)
	}
}

func (r *RedisCacheManager) AddSticker(ctx context.Context, sticker wumpkit.Sticker) {
	ctx = r.ctxOrFallback(ctx)
	if !r.CacheFlags().Has(wumpkit.CacheFlagStickers) {
		return
	}
	pipe := r.client.Pipeline()
	data, _ := json.Marshal(sticker)
	ttl := r.getTTL(wumpkit.CacheFlagStickers)
	pipe.Set(ctx, r.stickerKey(sticker.ID), data, ttl)
	if sticker.GuildID != 0 {
		pipe.SAdd(ctx, r.guildStickersSetKey(sticker.GuildID), sticker.ID.String())
	}
	_, _ = pipe.Exec(ctx)
}

func (r *RedisCacheManager) AddStickers(ctx context.Context, stickers ...wumpkit.Sticker) {
	for _, sticker := range stickers {
		r.AddSticker(ctx, sticker)
	}
}

func (r *RedisCacheManager) AddStageInstance(ctx context.Context, stage wumpkit.StageInstance) {
	if !r.CacheFlags().Has(wumpkit.CacheFlagStageInstances) {
		return
	}
	_ = r.setJSON(ctx, r.stageInstanceKey(stage.ChannelID), wumpkit.CacheFlagStageInstances, stage)
}

func (r *RedisCacheManager) AddGuildScheduledEvent(ctx context.Context, event wumpkit.GuildScheduledEvent) {
	if !r.CacheFlags().Has(wumpkit.CacheFlagGuildScheduledEvents) {
		return
	}
	_ = r.setJSON(ctx, r.scheduledEventKey(event.ID), wumpkit.CacheFlagGuildScheduledEvents, event)
}

func (r *RedisCacheManager) AddAutoModerationRule(ctx context.Context, rule wumpkit.AutoModerationRule) {
	if !r.CacheFlags().Has(wumpkit.CacheFlagAutoModerationRules) {
		return
	}
	_ = r.setJSON(ctx, r.autoModRuleKey(rule.ID), wumpkit.CacheFlagAutoModerationRules, rule)
}

/*****************************
 *   Deleters (Del)
 *****************************/

func (r *RedisCacheManager) DelUser(ctx context.Context, userID wumpkit.Snowflake) bool {
	return r.delKeys(ctx, r.userKey(userID))
}

func (r *RedisCacheManager) DelGuild(ctx context.Context, guildID wumpkit.Snowflake) bool {
	_ = r.DelGuildMembers(ctx, guildID)
	_ = r.DelGuildChannels(ctx, guildID)
	_ = r.DelGuildRoles(ctx, guildID)
	_ = r.DelGuildVoiceStates(ctx, guildID)
	_ = r.DelGuildEmojis(ctx, guildID)
	_ = r.DelGuildStickers(ctx, guildID)
	return r.delKeys(ctx, r.guildKey(guildID))
}

func (r *RedisCacheManager) DelMember(ctx context.Context, guildID, userID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	pipe := r.client.Pipeline()
	pipe.Del(ctx, r.memberKey(guildID, userID))
	pipe.SRem(ctx, r.guildMembersSetKey(guildID), userID.String())
	cmds, err := pipe.Exec(ctx)
	if err != nil || len(cmds) == 0 {
		return false
	}
	delCmd, ok := cmds[0].(*redis.IntCmd)
	return ok && delCmd.Val() > 0
}

func (r *RedisCacheManager) DelChannel(ctx context.Context, channelID wumpkit.Snowflake) bool {
	if ch, ok := r.GetChannel(ctx, channelID); ok {
		if gc, ok := ch.(wumpkit.GuildChannel); ok && gc.GetGuildID() != 0 {
			r.client.SRem(ctx, r.guildChannelsSetKey(gc.GetGuildID()), channelID.String())
		}
	}
	return r.delKeys(ctx, r.channelKey(channelID))
}

func (r *RedisCacheManager) DelMessage(ctx context.Context, messageID wumpkit.Snowflake) bool {
	return r.delKeys(ctx, r.messageKey(messageID))
}

func (r *RedisCacheManager) DelVoiceState(ctx context.Context, guildID, userID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	if vs, ok := r.GetVoiceState(ctx, guildID, userID); ok && vs.ChannelID != 0 {
		r.client.SRem(ctx, r.channelVoiceStatesSetKey(vs.ChannelID), userID.String())
	}
	pipe := r.client.Pipeline()
	pipe.Del(ctx, r.voiceStateKey(guildID, userID))
	pipe.SRem(ctx, r.guildVoiceStatesSetKey(guildID), userID.String())
	cmds, err := pipe.Exec(ctx)
	if err != nil || len(cmds) == 0 {
		return false
	}
	delCmd, ok := cmds[0].(*redis.IntCmd)
	return ok && delCmd.Val() > 0
}

func (r *RedisCacheManager) DelRole(ctx context.Context, guildID, roleID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	pipe := r.client.Pipeline()
	pipe.Del(ctx, r.roleKey(roleID))
	pipe.SRem(ctx, r.guildRolesSetKey(guildID), roleID.String())
	cmds, err := pipe.Exec(ctx)
	if err != nil || len(cmds) == 0 {
		return false
	}
	delCmd, ok := cmds[0].(*redis.IntCmd)
	return ok && delCmd.Val() > 0
}

func (r *RedisCacheManager) DelRoles(ctx context.Context, roleIDs ...wumpkit.Snowflake) bool {
	var anyDel bool
	for _, id := range roleIDs {
		if role, ok := r.GetRole(ctx, id); ok {
			if r.DelRole(ctx, role.GuildID, id) {
				anyDel = true
			}
		} else {
			if r.delKeys(ctx, r.roleKey(id)) {
				anyDel = true
			}
		}
	}
	return anyDel
}

func (r *RedisCacheManager) DelEmoji(ctx context.Context, guildID, emojiID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	pipe := r.client.Pipeline()
	pipe.Del(ctx, r.emojiKey(emojiID))
	pipe.SRem(ctx, r.guildEmojisSetKey(guildID), emojiID.String())
	cmds, err := pipe.Exec(ctx)
	if err != nil || len(cmds) == 0 {
		return false
	}
	delCmd, ok := cmds[0].(*redis.IntCmd)
	return ok && delCmd.Val() > 0
}

func (r *RedisCacheManager) DelSticker(ctx context.Context, guildID, stickerID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	pipe := r.client.Pipeline()
	pipe.Del(ctx, r.stickerKey(stickerID))
	pipe.SRem(ctx, r.guildStickersSetKey(guildID), stickerID.String())
	cmds, err := pipe.Exec(ctx)
	if err != nil || len(cmds) == 0 {
		return false
	}
	delCmd, ok := cmds[0].(*redis.IntCmd)
	return ok && delCmd.Val() > 0
}

func (r *RedisCacheManager) DelStageInstance(ctx context.Context, channelID wumpkit.Snowflake) bool {
	return r.delKeys(ctx, r.stageInstanceKey(channelID))
}

func (r *RedisCacheManager) DelGuildScheduledEvent(ctx context.Context, eventID wumpkit.Snowflake) bool {
	return r.delKeys(ctx, r.scheduledEventKey(eventID))
}

func (r *RedisCacheManager) DelAutoModerationRule(ctx context.Context, guildID, ruleID wumpkit.Snowflake) bool {
	return r.delKeys(ctx, r.autoModRuleKey(ruleID))
}

func (r *RedisCacheManager) DelGuildChannels(ctx context.Context, guildID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildChannelsSetKey(guildID)).Result()
	if err != nil {
		return false
	}
	keys := make([]string, len(ids)+1)
	for i, id := range ids {
		keys[i] = r.prefix + "channel:" + id
	}
	keys[len(ids)] = r.guildChannelsSetKey(guildID)
	return r.delKeys(ctx, keys...)
}

func (r *RedisCacheManager) DelGuildMembers(ctx context.Context, guildID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildMembersSetKey(guildID)).Result()
	if err != nil {
		return false
	}
	keys := make([]string, len(ids)+1)
	for i, id := range ids {
		keys[i] = r.prefix + "member:" + guildID.String() + ":" + id
	}
	keys[len(ids)] = r.guildMembersSetKey(guildID)
	return r.delKeys(ctx, keys...)
}

func (r *RedisCacheManager) DelGuildRoles(ctx context.Context, guildID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildRolesSetKey(guildID)).Result()
	if err != nil {
		return false
	}
	keys := make([]string, len(ids)+1)
	for i, id := range ids {
		keys[i] = r.prefix + "role:" + id
	}
	keys[len(ids)] = r.guildRolesSetKey(guildID)
	return r.delKeys(ctx, keys...)
}

func (r *RedisCacheManager) DelGuildVoiceStates(ctx context.Context, guildID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildVoiceStatesSetKey(guildID)).Result()
	if err != nil {
		return false
	}
	keys := make([]string, len(ids)+1)
	for i, id := range ids {
		keys[i] = r.prefix + "voicestate:" + guildID.String() + ":" + id
	}
	keys[len(ids)] = r.guildVoiceStatesSetKey(guildID)
	return r.delKeys(ctx, keys...)
}

func (r *RedisCacheManager) DelGuildEmojis(ctx context.Context, guildID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildEmojisSetKey(guildID)).Result()
	if err != nil {
		return false
	}
	keys := make([]string, len(ids)+1)
	for i, id := range ids {
		keys[i] = r.prefix + "emoji:" + id
	}
	keys[len(ids)] = r.guildEmojisSetKey(guildID)
	return r.delKeys(ctx, keys...)
}

func (r *RedisCacheManager) DelGuildStickers(ctx context.Context, guildID wumpkit.Snowflake) bool {
	ctx = r.ctxOrFallback(ctx)
	ids, err := r.client.SMembers(ctx, r.guildStickersSetKey(guildID)).Result()
	if err != nil {
		return false
	}
	keys := make([]string, len(ids)+1)
	for i, id := range ids {
		keys[i] = r.prefix + "sticker:" + id
	}
	keys[len(ids)] = r.guildStickersSetKey(guildID)
	return r.delKeys(ctx, keys...)
}
