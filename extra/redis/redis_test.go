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

package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/wumpkit-go/wumpkit"
	wumpkitredis "github.com/wumpkit-go/wumpkit/extra/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	redismodule "github.com/testcontainers/testcontainers-go/modules/redis"
)

// testRedisBackend encapsulates a Redis UniversalClient and provides an AdvanceTime helper
// that works seamlessly with both real Testcontainers Redis and in-process miniredis.
type testRedisBackend struct {
	client  redis.UniversalClient
	mr      *miniredis.Miniredis
	cleanup func()
}

func (b *testRedisBackend) AdvanceTime(d time.Duration) {
	if b.mr != nil {
		b.mr.FastForward(d)
	} else {
		time.Sleep(d)
	}
}

// setupRedisClient creates a Redis UniversalClient backed by a Testcontainers Redis container
// (or falls back to miniredis if Docker daemon is not available in the test environment).
func setupRedisClient(t *testing.T) *testRedisBackend {
	ctx := context.Background()

	// try spinning up a real Redis container with Testcontainers
	redisContainer, err := redismodule.Run(ctx, "redis:7-alpine")
	if err == nil && redisContainer != nil {
		connStr, err := redisContainer.ConnectionString(ctx)
		if err == nil {
			opts, err := redis.ParseURL(connStr)
			if err == nil {
				rdb := redis.NewClient(opts)
				return &testRedisBackend{
					client: rdb,
					cleanup: func() {
						_ = rdb.Close()
						_ = testcontainers.TerminateContainer(redisContainer)
					},
				}
			}
		}
		_ = testcontainers.TerminateContainer(redisContainer)
	}

	// fallback to in-process miniredis if Docker is not available in environment
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return &testRedisBackend{
		client: rdb,
		mr:     mr,
		cleanup: func() {
			_ = rdb.Close()
			mr.Close()
		},
	}
}

func setupTestRedis(t *testing.T) (wumpkit.CacheManager, func()) {
	backend := setupRedisClient(t)
	cfg := wumpkitredis.DefaultRedisCacheConfig()
	cfg.Prefix = "test:wumpkit:"
	cfg.DefaultTTL = 1 * time.Hour
	cm := wumpkitredis.NewRedisCacheManager(backend.client, wumpkitredis.WithRedisCacheConfig(cfg))
	return cm, backend.cleanup
}

func TestRedisCacheManager_CRUD(t *testing.T) {
	ctx := context.Background()
	cm, cleanup := setupTestRedis(t)
	defer cleanup()

	guildID := wumpkit.Snowflake(1001)
	userID := wumpkit.Snowflake(2001)
	channelID := wumpkit.Snowflake(3001)
	roleID := wumpkit.Snowflake(4001)
	emojiID := wumpkit.Snowflake(5001)
	stickerID := wumpkit.Snowflake(6001)
	messageID := wumpkit.Snowflake(7001)
	stageID := wumpkit.Snowflake(8001)
	eventID := wumpkit.Snowflake(9001)
	ruleID := wumpkit.Snowflake(9002)

	// 1. user
	user := wumpkit.User{ID: userID, Username: "redis_user"}
	cm.AddUser(ctx, user)
	assert.True(t, cm.HasUser(ctx, userID))
	u, ok := cm.GetUser(ctx, userID)
	require.True(t, ok)
	assert.Equal(t, "redis_user", u.Username)
	assert.Equal(t, 1, cm.CountUsers(ctx))

	// 2. guild
	guild := wumpkit.Guild{ID: guildID, Name: "Redis Guild"}
	cm.AddGuild(ctx, guild)
	assert.True(t, cm.HasGuild(ctx, guildID))
	g, ok := cm.GetGuild(ctx, guildID)
	require.True(t, ok)
	assert.Equal(t, "Redis Guild", g.Name)
	assert.Equal(t, 1, cm.CountGuilds(ctx))

	// 3. member
	member := wumpkit.Member{ID: userID, GuildID: guildID, Nickname: "Redis Member"}
	cm.AddMember(ctx, member)
	assert.True(t, cm.HasMember(ctx, guildID, userID))
	m, ok := cm.GetMember(ctx, guildID, userID)
	require.True(t, ok)
	assert.Equal(t, "Redis Member", m.Nickname)
	members, ok := cm.GetGuildMembers(ctx, guildID)
	require.True(t, ok)
	assert.Len(t, members, 1)
	assert.Equal(t, 1, cm.CountGuildMembers(ctx, guildID))

	// 4. channel
	ch := &wumpkit.TextChannel{
		GuildChannelFields: wumpkit.GuildChannelFields{
			ChannelFields: wumpkit.ChannelFields{
				ID:   channelID,
				Type: wumpkit.ChannelTypeGuildText,
			},
			GuildID: guildID,
			Name:    "redis-general",
		},
	}
	cm.AddChannel(ctx, ch)
	assert.True(t, cm.HasChannel(ctx, channelID))
	c, ok := cm.GetChannel(ctx, channelID)
	require.True(t, ok)
	assert.Equal(t, channelID, c.GetID())
	nc, ok := c.(wumpkit.NamedChannel)
	require.True(t, ok)
	assert.Equal(t, "redis-general", nc.GetName())
	channels, ok := cm.GetGuildChannels(ctx, guildID)
	require.True(t, ok)
	assert.Len(t, channels, 1)
	assert.Equal(t, 1, cm.CountGuildChannels(ctx, guildID))

	// 5. message
	msg := wumpkit.Message{ID: messageID, ChannelID: channelID, Content: "hello redis"}
	cm.AddMessage(ctx, msg)
	assert.True(t, cm.HasMessage(ctx, messageID))
	msgGot, ok := cm.GetMessage(ctx, messageID)
	require.True(t, ok)
	assert.Equal(t, "hello redis", msgGot.Content)
	assert.Equal(t, 1, cm.CountMessages(ctx))

	// 6. voiceState
	vs := wumpkit.VoiceState{GuildID: guildID, ChannelID: channelID, UserID: userID}
	cm.AddVoiceState(ctx, vs)
	assert.True(t, cm.HasVoiceState(ctx, guildID, userID))
	vsGot, ok := cm.GetVoiceState(ctx, guildID, userID)
	require.True(t, ok)
	assert.Equal(t, channelID, vsGot.ChannelID)
	assert.Equal(t, 1, cm.CountVoiceStates(ctx))
	vss, ok := cm.GetGuildVoiceStates(ctx, guildID)
	require.True(t, ok)
	assert.Len(t, vss, 1)

	// 7. role
	role := wumpkit.Role{ID: roleID, GuildID: guildID, Name: "Admin"}
	cm.AddRole(ctx, role)
	assert.True(t, cm.HasRole(ctx, roleID))
	rGot, ok := cm.GetRole(ctx, roleID)
	require.True(t, ok)
	assert.Equal(t, "Admin", rGot.Name)
	roles, ok := cm.GetGuildRoles(ctx, guildID)
	require.True(t, ok)
	assert.Len(t, roles, 1)
	assert.Equal(t, 1, cm.CountGuildRoles(ctx, guildID))

	// 8. emoji
	emoji := wumpkit.Emoji{ID: emojiID, GuildID: guildID, Name: "custom_emoji"}
	cm.AddEmoji(ctx, emoji)
	assert.True(t, cm.HasEmoji(ctx, emojiID))
	eGot, ok := cm.GetEmoji(ctx, emojiID)
	require.True(t, ok)
	assert.Equal(t, "custom_emoji", eGot.Name)
	emojis, ok := cm.GetGuildEmojis(ctx, guildID)
	require.True(t, ok)
	assert.Len(t, emojis, 1)
	assert.Equal(t, 1, cm.CountGuildEmojis(ctx, guildID))

	// 9. sticker
	sticker := wumpkit.Sticker{ID: stickerID, GuildID: guildID, Name: "custom_sticker"}
	cm.AddSticker(ctx, sticker)
	assert.True(t, cm.HasSticker(ctx, stickerID))
	sGot, ok := cm.GetSticker(ctx, stickerID)
	require.True(t, ok)
	assert.Equal(t, "custom_sticker", sGot.Name)
	stickers, ok := cm.GetGuildStickers(ctx, guildID)
	require.True(t, ok)
	assert.Len(t, stickers, 1)
	assert.Equal(t, 1, cm.CountGuildStickers(ctx, guildID))

	// 10. stageInstance
	stage := wumpkit.StageInstance{ID: stageID, ChannelID: channelID, Topic: "Dev Talk"}
	cm.AddStageInstance(ctx, stage)
	assert.True(t, cm.HasStageInstance(ctx, channelID))
	stGot, ok := cm.GetStageInstance(ctx, channelID)
	require.True(t, ok)
	assert.Equal(t, "Dev Talk", stGot.Topic)
	assert.Equal(t, 1, cm.CountStageInstances(ctx))

	// 11. guildScheduledEvent
	event := wumpkit.GuildScheduledEvent{ID: eventID, GuildID: guildID, Name: "Weekly Meet"}
	cm.AddGuildScheduledEvent(ctx, event)
	assert.True(t, cm.HasGuildScheduledEvent(ctx, eventID))
	evGot, ok := cm.GetGuildScheduledEvent(ctx, eventID)
	require.True(t, ok)
	assert.Equal(t, "Weekly Meet", evGot.Name)
	assert.Equal(t, 1, cm.CountGuildScheduledEvents(ctx))

	// 12. autoModerationRule
	rule := wumpkit.AutoModerationRule{ID: ruleID, GuildID: guildID, Name: "Anti-Spam"}
	cm.AddAutoModerationRule(ctx, rule)
	assert.True(t, cm.HasAutoModerationRule(ctx, ruleID))
	ruGot, ok := cm.GetAutoModerationRule(ctx, ruleID)
	require.True(t, ok)
	assert.Equal(t, "Anti-Spam", ruGot.Name)
	assert.Equal(t, 1, cm.CountAutoModerationRules(ctx))

	// deletions
	cm.DelUser(ctx, userID)
	assert.False(t, cm.HasUser(ctx, userID))

	cm.DelMessage(ctx, messageID)
	assert.False(t, cm.HasMessage(ctx, messageID))

	cm.DelVoiceState(ctx, guildID, userID)
	assert.False(t, cm.HasVoiceState(ctx, guildID, userID))

	cm.DelRole(ctx, guildID, roleID)
	assert.False(t, cm.HasRole(ctx, roleID))

	cm.DelEmoji(ctx, guildID, emojiID)
	assert.False(t, cm.HasEmoji(ctx, emojiID))

	cm.DelSticker(ctx, guildID, stickerID)
	assert.False(t, cm.HasSticker(ctx, stickerID))

	cm.DelStageInstance(ctx, channelID)
	assert.False(t, cm.HasStageInstance(ctx, channelID))

	cm.DelGuildScheduledEvent(ctx, eventID)
	assert.False(t, cm.HasGuildScheduledEvent(ctx, eventID))

	cm.DelAutoModerationRule(ctx, guildID, ruleID)
	assert.False(t, cm.HasAutoModerationRule(ctx, ruleID))

	cm.DelMember(ctx, guildID, userID)
	assert.False(t, cm.HasMember(ctx, guildID, userID))

	cm.DelChannel(ctx, channelID)
	assert.False(t, cm.HasChannel(ctx, channelID))

	cm.DelGuild(ctx, guildID)
	assert.False(t, cm.HasGuild(ctx, guildID))
}

func TestRedisCacheManager_Flags(t *testing.T) {
	ctx := context.Background()
	cm, cleanup := setupTestRedis(t)
	defer cleanup()

	cm.SetCacheFlags(wumpkit.CacheFlagsNone)
	assert.Equal(t, wumpkit.CacheFlagsNone, cm.CacheFlags())

	cm.AddUser(ctx, wumpkit.User{ID: 9999, Username: "ignored"})
	assert.False(t, cm.HasUser(ctx, 9999))

	cm.SetCacheFlags(wumpkit.CacheFlagUsers)
	assert.True(t, cm.CacheFlags().Has(wumpkit.CacheFlagUsers))

	cm.AddUser(ctx, wumpkit.User{ID: 9999, Username: "accepted"})
	assert.True(t, cm.HasUser(ctx, 9999))
}

func TestRedisCacheManager_PerResourceTTL(t *testing.T) {
	ctx := context.Background()
	backend := setupRedisClient(t)
	defer backend.cleanup()

	cfg := wumpkitredis.DefaultRedisCacheConfig()
	cfg.DefaultTTL = 1 * time.Hour
	cfg.ResourceTTLs.Messages = 500 * time.Millisecond
	cfg.ResourceTTLs.Users = 2000 * time.Millisecond

	cm := wumpkitredis.NewRedisCacheManager(backend.client, wumpkitredis.WithRedisCacheConfig(cfg))

	// add message (500ms TTL) and user (2000ms TTL)
	cm.AddMessage(ctx, wumpkit.Message{ID: 7777, Content: "short lived"})
	cm.AddUser(ctx, wumpkit.User{ID: 8888, Username: "longer lived"})

	assert.True(t, cm.HasMessage(ctx, 7777))
	assert.True(t, cm.HasUser(ctx, 8888))

	// advance time 600ms
	backend.AdvanceTime(600 * time.Millisecond)

	// message should have expired, user should still be alive
	assert.False(t, cm.HasMessage(ctx, 7777))
	assert.True(t, cm.HasUser(ctx, 8888))

	// advance another 1500ms
	backend.AdvanceTime(1500 * time.Millisecond)
	assert.False(t, cm.HasUser(ctx, 8888))
}

func TestRedisCacheManager_SlidingExpiration(t *testing.T) {
	ctx := context.Background()
	backend := setupRedisClient(t)
	defer backend.cleanup()

	cfg := wumpkitredis.DefaultRedisCacheConfig()
	cfg.ResourceTTLs.Users = 800 * time.Millisecond
	cfg.ResourceTTLs.Messages = 800 * time.Millisecond
	// enable sliding expiration ONLY for Users, NOT Messages
	cfg.SlidingExpiration.Users = true

	cm := wumpkitredis.NewRedisCacheManager(backend.client, wumpkitredis.WithRedisCacheConfig(cfg))

	cm.AddUser(ctx, wumpkit.User{ID: 5555, Username: "sliding_user"})
	cm.AddMessage(ctx, wumpkit.Message{ID: 6666, Content: "fixed_ttl_message"})

	// advance 500ms
	backend.AdvanceTime(500 * time.Millisecond)

	// read user (triggers sliding expiration -> resets user back to 800ms)
	u, ok := cm.GetUser(ctx, 5555)
	require.True(t, ok)
	assert.Equal(t, "sliding_user", u.Username)

	// read message (does NOT trigger sliding expiration because Messages sliding is false)
	m, ok := cm.GetMessage(ctx, 6666)
	require.True(t, ok)
	assert.Equal(t, "fixed_ttl_message", m.Content)

	// advance 500ms more
	// total elapsed for message: 1000ms (> 800ms TTL) -> should expire
	// total elapsed for user since read: 500ms (< 800ms TTL) -> should still exist
	backend.AdvanceTime(500 * time.Millisecond)
	assert.True(t, cm.HasUser(ctx, 5555))
	assert.False(t, cm.HasMessage(ctx, 6666))

	// wait 900ms without reading user -> user should expire too
	backend.AdvanceTime(900 * time.Millisecond)
	assert.False(t, cm.HasUser(ctx, 5555))
}

func TestRedisCacheManager_DefaultAndFunctionalOptions(t *testing.T) {
	ctx := context.Background()
	backend := setupRedisClient(t)
	defer backend.cleanup()

	// 1. default creation without options
	cmDefault := wumpkitredis.NewRedisCacheManager(backend.client)
	require.NotNil(t, cmDefault)
	cmDefault.AddUser(ctx, wumpkit.User{ID: 101, Username: "default_user"})
	u, ok := cmDefault.GetUser(ctx, 101)
	require.True(t, ok)
	assert.Equal(t, "default_user", u.Username)

	// 2. creation with individual functional options
	cmCustom := wumpkitredis.NewRedisCacheManager(
		backend.client,
		wumpkitredis.WithRedisCachePrefix("custom:prefix:"),
		wumpkitredis.WithRedisCacheFlags(wumpkit.CacheFlagUsers|wumpkit.CacheFlagGuilds),
		wumpkitredis.WithRedisCacheDefaultTTL(10*time.Minute),
	)
	require.NotNil(t, cmCustom)
	cmCustom.AddUser(ctx, wumpkit.User{ID: 202, Username: "custom_user"})
	u2, ok2 := cmCustom.GetUser(ctx, 202)
	require.True(t, ok2)
	assert.Equal(t, "custom_user", u2.Username)
}

func TestRedisCacheManager_TestcontainersDirect(t *testing.T) {
	ctx := context.Background()
	container, err := redismodule.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Skipf("Skipping direct Testcontainers test (Docker not running): %v", err)
		return
	}
	defer func() {
		_ = testcontainers.TerminateContainer(container)
	}()

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	opts, err := redis.ParseURL(connStr)
	require.NoError(t, err)

	rdb := redis.NewClient(opts)
	defer rdb.Close()

	cfg := wumpkitredis.DefaultRedisCacheConfig()
	cm := wumpkitredis.NewRedisCacheManager(rdb, wumpkitredis.WithRedisCacheConfig(cfg))

	cm.AddUser(ctx, wumpkit.User{ID: 12345, Username: "container_user"})
	u, ok := cm.GetUser(ctx, 12345)
	require.True(t, ok)
	assert.Equal(t, "container_user", u.Username)
}

func TestRedisRateLimitStore(t *testing.T) {
	backend := setupRedisClient(t)
	defer backend.cleanup()

	store := wumpkitredis.NewRedisRateLimitStore(backend.client, "test:ratelimit:")
	require.NotNil(t, store)

	// 1. bucket Hash mapping
	store.SetBucketHash("/channels/123/messages", "chanHash123")
	hash, ok := store.GetBucketHash("/channels/123/messages")
	require.True(t, ok)
	assert.Equal(t, "chanHash123", hash)

	// 2. bucket State
	resetTime := time.Now().Add(5 * time.Second).Truncate(time.Millisecond)
	store.SetBucket("chanHash123", wumpkit.RateLimitBucketState{
		Limit:      5,
		Remaining:  3,
		ResetAt:    resetTime,
		BucketHash: "chanHash123",
	})

	state, ok := store.GetBucket("chanHash123")
	require.True(t, ok)
	assert.Equal(t, int64(5), state.Limit)
	assert.Equal(t, int64(3), state.Remaining)
	assert.Equal(t, "chanHash123", state.BucketHash)
	assert.WithinDuration(t, resetTime, state.ResetAt, 50*time.Millisecond)

	// 3. global Lock
	globalUntil := time.Now().Add(2 * time.Second).Truncate(time.Millisecond)
	store.SetGlobalLock(globalUntil)
	gotGlobal := store.GetGlobalLock()
	assert.WithinDuration(t, globalUntil, gotGlobal, 50*time.Millisecond)
}

