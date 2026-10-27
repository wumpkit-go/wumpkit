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

package wumpkit

import (
	"context"

	"github.com/wumpkit/wumpkit/internal/shardedmap"
)

// CacheFlags defines bit flags to configure which entities are cached.
type CacheFlags int

const (
	CacheFlagUsers CacheFlags = 1 << iota
	CacheFlagGuilds
	CacheFlagMembers
	CacheFlagChannels
	CacheFlagMessages
	CacheFlagRoles
	CacheFlagVoiceStates
	CacheFlagEmojis
	CacheFlagStickers
	CacheFlagStageInstances
	CacheFlagGuildScheduledEvents
	CacheFlagAutoModerationRules

	CacheFlagsNone CacheFlags = 0

	CacheFlagsAll = CacheFlagUsers | CacheFlagGuilds | CacheFlagMembers |
		CacheFlagChannels | CacheFlagMessages | CacheFlagRoles |
		CacheFlagVoiceStates | CacheFlagEmojis | CacheFlagStickers |
		CacheFlagStageInstances | CacheFlagGuildScheduledEvents |
		CacheFlagAutoModerationRules
)

// Has returns true if all specified flags are set.
func (f CacheFlags) Has(bits ...CacheFlags) bool {
	return BitFieldHas(f, bits...)
}

// SnowflakePairKey represents a compound key of two Snowflakes (e.g. GuildID + UserID).
type SnowflakePairKey struct {
	A Snowflake
	B Snowflake
}

// CacheConfig holds configuration for the concurrent sharded maps in InMemoryCacheManager.
type CacheConfig struct {
	UsersShardCount               int
	GuildsShardCount              int
	MembersShardCount             int
	ChannelsShardCount            int
	MessagesShardCount            int
	VoiceStatesShardCount         int
	RolesShardCount               int
	EmojisShardCount              int
	StickersShardCount            int
	StageInstancesShardCount      int
	ScheduledEventsShardCount     int
	AutoModRulesShardCount        int
	GuildToMemberShardCount       int
	GuildToChannelShardCount      int
	GuildToVoiceStateShardCount   int
	GuildToRoleShardCount         int
	GuildToEmojiShardCount        int
	GuildToStickerShardCount      int
	ChannelToVoiceStateShardCount int
	ChannelToMessageShardCount    int
	MaxMessagesPerChannel         int
}

// DefaultCacheConfig returns optimal default shard configurations for concurrent maps.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		UsersShardCount:               32,
		GuildsShardCount:              32,
		MembersShardCount:             64,
		ChannelsShardCount:            32,
		MessagesShardCount:            64,
		VoiceStatesShardCount:         32,
		RolesShardCount:               32,
		EmojisShardCount:              32,
		StickersShardCount:            32,
		StageInstancesShardCount:      16,
		ScheduledEventsShardCount:     16,
		AutoModRulesShardCount:        16,
		GuildToMemberShardCount:       32,
		GuildToChannelShardCount:      32,
		GuildToVoiceStateShardCount:   32,
		GuildToRoleShardCount:         32,
		GuildToEmojiShardCount:        32,
		GuildToStickerShardCount:      32,
		ChannelToVoiceStateShardCount: 32,
		ChannelToMessageShardCount:    32,
		MaxMessagesPerChannel:         100,
	}
}

// CacheManager defines the swappable interface for caching Discord entities.
type CacheManager interface {
	// CacheFlags returns the bitmask of enabled entity caching flags.
	CacheFlags() CacheFlags
	// SetCacheFlags updates the entity caching flags.
	SetCacheFlags(flags ...CacheFlags)

	// GetUser returns the cached User for the given userID, or false if not in cache.
	GetUser(ctx context.Context, userID Snowflake) (User, bool)
	// GetGuild returns the cached Guild for the given guildID, or false if not in cache.
	GetGuild(ctx context.Context, guildID Snowflake) (Guild, bool)
	// GetMember returns the cached Member for the given guildID and userID, or false if not in cache.
	GetMember(ctx context.Context, guildID, userID Snowflake) (Member, bool)
	// GetChannel returns the cached Channel for the given channelID, or false if not in cache.
	GetChannel(ctx context.Context, channelID Snowflake) (Channel, bool)
	// GetMessage returns the cached Message for the given messageID, or false if not in cache.
	GetMessage(ctx context.Context, messageID Snowflake) (Message, bool)
	// GetVoiceState returns the cached VoiceState for the given guildID and userID, or false if not in cache.
	GetVoiceState(ctx context.Context, guildID, userID Snowflake) (VoiceState, bool)
	// GetRole returns the cached Role for the given roleID, or false if not in cache.
	GetRole(ctx context.Context, roleID Snowflake) (Role, bool)
	// GetRoles returns a map of cached Roles for the provided roleIDs.
	GetRoles(ctx context.Context, rolesIDs ...Snowflake) map[Snowflake]Role
	// GetEmoji returns the cached Emoji for the given emojiID, or false if not in cache.
	GetEmoji(ctx context.Context, emojiID Snowflake) (Emoji, bool)
	// GetSticker returns the cached Sticker for the given stickerID, or false if not in cache.
	GetSticker(ctx context.Context, stickerID Snowflake) (Sticker, bool)
	// GetStageInstance returns the cached StageInstance for the given channelID, or false if not in cache.
	GetStageInstance(ctx context.Context, channelID Snowflake) (StageInstance, bool)
	// GetGuildScheduledEvent returns the cached GuildScheduledEvent for the given eventID, or false if not in cache.
	GetGuildScheduledEvent(ctx context.Context, eventID Snowflake) (GuildScheduledEvent, bool)
	// GetAutoModerationRule returns the cached AutoModerationRule for the given ruleID, or false if not in cache.
	GetAutoModerationRule(ctx context.Context, ruleID Snowflake) (AutoModerationRule, bool)

	// GetGuildChannels returns all cached channels belonging to the given guildID as a map keyed by ChannelID.
	GetGuildChannels(ctx context.Context, guildID Snowflake) (map[Snowflake]GuildChannel, bool)
	// GetGuildMembers returns all cached members belonging to the given guildID as a map keyed by UserID.
	GetGuildMembers(ctx context.Context, guildID Snowflake) (map[Snowflake]Member, bool)
	// GetGuildVoiceStates returns all cached voice states for the given guildID as a map keyed by UserID.
	GetGuildVoiceStates(ctx context.Context, guildID Snowflake) (map[Snowflake]VoiceState, bool)
	// GetGuildRoles returns all cached roles for the given guildID as a map keyed by RoleID.
	GetGuildRoles(ctx context.Context, guildID Snowflake) (map[Snowflake]Role, bool)
	// GetGuildEmojis returns all cached emojis for the given guildID as a map keyed by EmojiID.
	GetGuildEmojis(ctx context.Context, guildID Snowflake) (map[Snowflake]Emoji, bool)
	// GetGuildStickers returns all cached stickers for the given guildID as a map keyed by StickerID.
	GetGuildStickers(ctx context.Context, guildID Snowflake) (map[Snowflake]Sticker, bool)
	// GetVoiceChannelMembers returns all cached members currently connected to the given voice channelID.
	GetVoiceChannelMembers(ctx context.Context, channelID Snowflake) (map[Snowflake]Member, bool)

	// HasUser checks if the user is in cache.
	HasUser(ctx context.Context, userID Snowflake) bool
	// HasGuild checks if the guild is in cache.
	HasGuild(ctx context.Context, guildID Snowflake) bool
	// HasMember checks if the member is in cache.
	HasMember(ctx context.Context, guildID, userID Snowflake) bool
	// HasChannel checks if the channel is in cache.
	HasChannel(ctx context.Context, channelID Snowflake) bool
	// HasMessage checks if the message is in cache.
	HasMessage(ctx context.Context, messageID Snowflake) bool
	// HasVoiceState checks if the voice state is in cache.
	HasVoiceState(ctx context.Context, guildID, userID Snowflake) bool
	// HasRole checks if the role is in cache.
	HasRole(ctx context.Context, roleID Snowflake) bool
	// HasRoles checks if all specified roles are in cache.
	HasRoles(ctx context.Context, rolesIDs ...Snowflake) bool
	// HasEmoji checks if the emoji is in cache.
	HasEmoji(ctx context.Context, emojiID Snowflake) bool
	// HasSticker checks if the sticker is in cache.
	HasSticker(ctx context.Context, stickerID Snowflake) bool
	// HasStageInstance checks if the stage instance is in cache.
	HasStageInstance(ctx context.Context, channelID Snowflake) bool
	// HasGuildScheduledEvent checks if the scheduled event is in cache.
	HasGuildScheduledEvent(ctx context.Context, eventID Snowflake) bool
	// HasAutoModerationRule checks if the auto moderation rule is in cache.
	HasAutoModerationRule(ctx context.Context, ruleID Snowflake) bool

	// CountUsers returns the total number of cached users.
	CountUsers(ctx context.Context) int
	// CountGuilds returns the total number of cached guilds.
	CountGuilds(ctx context.Context) int
	// CountMembers returns the total number of cached members across all guilds.
	CountMembers(ctx context.Context) int
	// CountChannels returns the total number of cached channels across all guilds.
	CountChannels(ctx context.Context) int
	// CountMessages returns the total number of cached messages.
	CountMessages(ctx context.Context) int
	// CountVoiceStates returns the total number of cached voice states.
	CountVoiceStates(ctx context.Context) int
	// CountRoles returns the total number of cached roles.
	CountRoles(ctx context.Context) int
	// CountEmojis returns the total number of cached emojis.
	CountEmojis(ctx context.Context) int
	// CountStickers returns the total number of cached stickers.
	CountStickers(ctx context.Context) int
	// CountStageInstances returns the total number of cached stage instances.
	CountStageInstances(ctx context.Context) int
	// CountGuildScheduledEvents returns the total number of cached scheduled events.
	CountGuildScheduledEvents(ctx context.Context) int
	// CountAutoModerationRules returns the total number of cached auto moderation rules.
	CountAutoModerationRules(ctx context.Context) int
	// CountGuildChannels returns the number of cached channels in the given guild.
	CountGuildChannels(ctx context.Context, guildID Snowflake) int
	// CountGuildMembers returns the number of cached members in the given guild.
	CountGuildMembers(ctx context.Context, guildID Snowflake) int
	// CountGuildRoles returns the number of cached roles in the given guild.
	CountGuildRoles(ctx context.Context, guildID Snowflake) int
	// CountGuildEmojis returns the number of cached emojis in the given guild.
	CountGuildEmojis(ctx context.Context, guildID Snowflake) int
	// CountGuildStickers returns the number of cached stickers in the given guild.
	CountGuildStickers(ctx context.Context, guildID Snowflake) int

	// AddUser inserts or updates a User in the cache.
	AddUser(ctx context.Context, user User)
	// AddGuild inserts or updates a Guild in the cache.
	AddGuild(ctx context.Context, guild Guild)
	// AddMember inserts or updates a Member in the cache.
	AddMember(ctx context.Context, member Member)
	// AddChannel inserts or updates a Channel in the cache.
	AddChannel(ctx context.Context, channel Channel)
	// AddMessage inserts or updates a Message in the cache.
	AddMessage(ctx context.Context, message Message)
	// AddVoiceState inserts or updates a VoiceState in the cache.
	AddVoiceState(ctx context.Context, voiceState VoiceState)
	// AddRole inserts or updates a Role in the cache.
	AddRole(ctx context.Context, role Role)
	// AddRoles inserts or updates multiple Roles in the cache.
	AddRoles(ctx context.Context, roles ...Role)
	// AddEmoji inserts or updates an Emoji in the cache.
	AddEmoji(ctx context.Context, emoji Emoji)
	// AddEmojis inserts or updates multiple Emojis in the cache.
	AddEmojis(ctx context.Context, emojis ...Emoji)
	// AddSticker inserts or updates a Sticker in the cache.
	AddSticker(ctx context.Context, sticker Sticker)
	// AddStickers inserts or updates multiple Stickers in the cache.
	AddStickers(ctx context.Context, stickers ...Sticker)
	// AddStageInstance inserts or updates a StageInstance in the cache.
	AddStageInstance(ctx context.Context, stage StageInstance)
	// AddGuildScheduledEvent inserts or updates a GuildScheduledEvent in the cache.
	AddGuildScheduledEvent(ctx context.Context, event GuildScheduledEvent)
	// AddAutoModerationRule inserts or updates an AutoModerationRule in the cache.
	AddAutoModerationRule(ctx context.Context, rule AutoModerationRule)

	// DelUser removes a User from cache.
	DelUser(ctx context.Context, userID Snowflake) bool
	// DelGuild removes a Guild and all its associated child entities (channels, members, roles, emojis, stickers, voice states) from cache.
	DelGuild(ctx context.Context, guildID Snowflake) bool
	// DelMember removes a Member from cache.
	DelMember(ctx context.Context, guildID, userID Snowflake) bool
	// DelChannel removes a Channel from cache.
	DelChannel(ctx context.Context, channelID Snowflake) bool
	// DelMessage removes a Message from cache.
	DelMessage(ctx context.Context, messageID Snowflake) bool
	// DelVoiceState removes a VoiceState from cache.
	DelVoiceState(ctx context.Context, guildID, userID Snowflake) bool
	// DelRole removes a Role from cache.
	DelRole(ctx context.Context, guildID, roleID Snowflake) bool
	// DelRoles removes multiple Roles from cache.
	DelRoles(ctx context.Context, roleIDs ...Snowflake) bool
	// DelEmoji removes an Emoji from cache.
	DelEmoji(ctx context.Context, guildID, emojiID Snowflake) bool
	// DelSticker removes a Sticker from cache.
	DelSticker(ctx context.Context, guildID, stickerID Snowflake) bool
	// DelStageInstance removes a StageInstance from cache.
	DelStageInstance(ctx context.Context, channelID Snowflake) bool
	// DelGuildScheduledEvent removes a GuildScheduledEvent from cache.
	DelGuildScheduledEvent(ctx context.Context, eventID Snowflake) bool
	// DelAutoModerationRule removes an AutoModerationRule from cache.
	DelAutoModerationRule(ctx context.Context, guildID, ruleID Snowflake) bool
	// DelGuildChannels removes all cached channels belonging to a guild.
	DelGuildChannels(ctx context.Context, guildID Snowflake) bool
	// DelGuildMembers removes all cached members belonging to a guild.
	DelGuildMembers(ctx context.Context, guildID Snowflake) bool
	// DelGuildRoles removes all cached roles belonging to a guild.
	DelGuildRoles(ctx context.Context, guildID Snowflake) bool
	// DelGuildVoiceStates removes all cached voice states belonging to a guild.
	DelGuildVoiceStates(ctx context.Context, guildID Snowflake) bool
	// DelGuildEmojis removes all cached emojis belonging to a guild.
	DelGuildEmojis(ctx context.Context, guildID Snowflake) bool
	// DelGuildStickers removes all cached stickers belonging to a guild.
	DelGuildStickers(ctx context.Context, guildID Snowflake) bool
}

type InMemoryCacheManager struct {
	flags  CacheFlags
	config CacheConfig

	users           shardedmap.Map[Snowflake, User]
	guilds          shardedmap.Map[Snowflake, Guild]
	members         shardedmap.Map[SnowflakePairKey, Member]
	channels        shardedmap.Map[Snowflake, Channel]
	messages        shardedmap.Map[Snowflake, Message]
	voiceStates     shardedmap.Map[SnowflakePairKey, VoiceState]
	roles           shardedmap.Map[Snowflake, Role]
	emojis          shardedmap.Map[Snowflake, Emoji]
	stickers        shardedmap.Map[Snowflake, Sticker]
	stageInstances  shardedmap.Map[Snowflake, StageInstance]
	scheduledEvents shardedmap.Map[Snowflake, GuildScheduledEvent]
	autoModRules    shardedmap.Map[Snowflake, AutoModerationRule]

	// Secondary index maps: GuildID -> Set[ID]
	guildToMemberIDs           shardedmap.Map[Snowflake, map[Snowflake]struct{}]
	guildToChannelIDs          shardedmap.Map[Snowflake, map[Snowflake]struct{}]
	guildToVoiceStateUserIDs   shardedmap.Map[Snowflake, map[Snowflake]struct{}]
	guildToRoleIDs             shardedmap.Map[Snowflake, map[Snowflake]struct{}]
	guildToEmojiIDs            shardedmap.Map[Snowflake, map[Snowflake]struct{}]
	guildToStickerIDs          shardedmap.Map[Snowflake, map[Snowflake]struct{}]
	guildToAutoModRuleIDs      shardedmap.Map[Snowflake, map[Snowflake]struct{}]
	channelToVoiceStateUserIDs shardedmap.Map[Snowflake, map[Snowflake]struct{}]
	channelToMessageIDs        shardedmap.Map[Snowflake, []Snowflake]
}

// NewInMemoryCacheManager creates an InMemoryCacheManager with the specified flags and sharding config.
func NewInMemoryCacheManager(flags CacheFlags, config CacheConfig) *InMemoryCacheManager {
	channelToMsgShards := config.ChannelToMessageShardCount
	if channelToMsgShards <= 0 {
		channelToMsgShards = 32
	}
	if config.MaxMessagesPerChannel <= 0 {
		config.MaxMessagesPerChannel = 100
	}
	return &InMemoryCacheManager{
		flags:                      flags,
		config:                     config,
		users:                      shardedmap.New[Snowflake, User](config.UsersShardCount),
		guilds:                     shardedmap.New[Snowflake, Guild](config.GuildsShardCount),
		members:                    shardedmap.New[SnowflakePairKey, Member](config.MembersShardCount),
		channels:                   shardedmap.New[Snowflake, Channel](config.ChannelsShardCount),
		messages:                   shardedmap.New[Snowflake, Message](config.MessagesShardCount),
		voiceStates:                shardedmap.New[SnowflakePairKey, VoiceState](config.VoiceStatesShardCount),
		roles:                      shardedmap.New[Snowflake, Role](config.RolesShardCount),
		emojis:                     shardedmap.New[Snowflake, Emoji](config.EmojisShardCount),
		stickers:                   shardedmap.New[Snowflake, Sticker](config.StickersShardCount),
		stageInstances:             shardedmap.New[Snowflake, StageInstance](config.StageInstancesShardCount),
		scheduledEvents:            shardedmap.New[Snowflake, GuildScheduledEvent](config.ScheduledEventsShardCount),
		autoModRules:               shardedmap.New[Snowflake, AutoModerationRule](config.AutoModRulesShardCount),
		guildToMemberIDs:           shardedmap.New[Snowflake, map[Snowflake]struct{}](config.GuildToMemberShardCount),
		guildToChannelIDs:          shardedmap.New[Snowflake, map[Snowflake]struct{}](config.GuildToChannelShardCount),
		guildToVoiceStateUserIDs:   shardedmap.New[Snowflake, map[Snowflake]struct{}](config.GuildToVoiceStateShardCount),
		guildToRoleIDs:             shardedmap.New[Snowflake, map[Snowflake]struct{}](config.GuildToRoleShardCount),
		guildToEmojiIDs:            shardedmap.New[Snowflake, map[Snowflake]struct{}](config.GuildToEmojiShardCount),
		guildToStickerIDs:          shardedmap.New[Snowflake, map[Snowflake]struct{}](config.GuildToStickerShardCount),
		guildToAutoModRuleIDs:      shardedmap.New[Snowflake, map[Snowflake]struct{}](config.GuildToRoleShardCount),
		channelToVoiceStateUserIDs: shardedmap.New[Snowflake, map[Snowflake]struct{}](config.ChannelToVoiceStateShardCount),
		channelToMessageIDs:        shardedmap.New[Snowflake, []Snowflake](channelToMsgShards),
	}
}

func (c *InMemoryCacheManager) CacheFlags() CacheFlags {
	return c.flags
}

func (c *InMemoryCacheManager) SetCacheFlags(flags ...CacheFlags) {
	c.flags = CacheFlagsNone
	for _, f := range flags {
		c.flags |= f
	}
}

func (c *InMemoryCacheManager) GetUser(ctx context.Context, userID Snowflake) (User, bool) {
	return c.users.Get(userID)
}

func (c *InMemoryCacheManager) GetGuild(ctx context.Context, guildID Snowflake) (Guild, bool) {
	return c.guilds.Get(guildID)
}

func (c *InMemoryCacheManager) GetMember(ctx context.Context, guildID, userID Snowflake) (Member, bool) {
	return c.members.Get(SnowflakePairKey{A: guildID, B: userID})
}

func (c *InMemoryCacheManager) GetChannel(ctx context.Context, channelID Snowflake) (Channel, bool) {
	return c.channels.Get(channelID)
}

func (c *InMemoryCacheManager) GetMessage(ctx context.Context, messageID Snowflake) (Message, bool) {
	return c.messages.Get(messageID)
}

func (c *InMemoryCacheManager) GetVoiceState(ctx context.Context, guildID, userID Snowflake) (VoiceState, bool) {
	return c.voiceStates.Get(SnowflakePairKey{A: guildID, B: userID})
}

func (c *InMemoryCacheManager) GetRole(ctx context.Context, roleID Snowflake) (Role, bool) {
	return c.roles.Get(roleID)
}

func (c *InMemoryCacheManager) GetRoles(ctx context.Context, rolesIDs ...Snowflake) map[Snowflake]Role {
	res := make(map[Snowflake]Role, len(rolesIDs))
	for _, id := range rolesIDs {
		if role, ok := c.roles.Get(id); ok {
			res[id] = role
		}
	}
	return res
}

func (c *InMemoryCacheManager) GetEmoji(ctx context.Context, emojiID Snowflake) (Emoji, bool) {
	return c.emojis.Get(emojiID)
}

func (c *InMemoryCacheManager) GetSticker(ctx context.Context, stickerID Snowflake) (Sticker, bool) {
	return c.stickers.Get(stickerID)
}

func (c *InMemoryCacheManager) GetStageInstance(ctx context.Context, channelID Snowflake) (StageInstance, bool) {
	return c.stageInstances.Get(channelID)
}

func (c *InMemoryCacheManager) GetGuildScheduledEvent(ctx context.Context, eventID Snowflake) (GuildScheduledEvent, bool) {
	return c.scheduledEvents.Get(eventID)
}

func (c *InMemoryCacheManager) GetAutoModerationRule(ctx context.Context, ruleID Snowflake) (AutoModerationRule, bool) {
	return c.autoModRules.Get(ruleID)
}

func (c *InMemoryCacheManager) GetGuildChannels(ctx context.Context, guildID Snowflake) (map[Snowflake]GuildChannel, bool) {
	set, ok := c.guildToChannelIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]GuildChannel, len(set))
	for channelID := range set {
		if channel, ok := c.channels.Get(channelID); ok {
			if gc, ok := channel.(GuildChannel); ok {
				res[channelID] = gc
			}
		}
	}
	return res, true
}

func (c *InMemoryCacheManager) GetGuildMembers(ctx context.Context, guildID Snowflake) (map[Snowflake]Member, bool) {
	set, ok := c.guildToMemberIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]Member, len(set))
	for userID := range set {
		key := SnowflakePairKey{A: guildID, B: userID}
		if member, ok := c.members.Get(key); ok {
			res[userID] = member
		}
	}
	return res, true
}

func (c *InMemoryCacheManager) GetGuildVoiceStates(ctx context.Context, guildID Snowflake) (map[Snowflake]VoiceState, bool) {
	set, ok := c.guildToVoiceStateUserIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]VoiceState, len(set))
	for userID := range set {
		key := SnowflakePairKey{A: guildID, B: userID}
		if voiceState, ok := c.voiceStates.Get(key); ok {
			res[userID] = voiceState
		}
	}
	return res, true
}

func (c *InMemoryCacheManager) GetGuildRoles(ctx context.Context, guildID Snowflake) (map[Snowflake]Role, bool) {
	set, ok := c.guildToRoleIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]Role, len(set))
	for roleID := range set {
		if role, ok := c.roles.Get(roleID); ok {
			res[roleID] = role
		}
	}
	return res, true
}

func (c *InMemoryCacheManager) GetGuildEmojis(ctx context.Context, guildID Snowflake) (map[Snowflake]Emoji, bool) {
	set, ok := c.guildToEmojiIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]Emoji, len(set))
	for emojiID := range set {
		if emoji, ok := c.emojis.Get(emojiID); ok {
			res[emojiID] = emoji
		}
	}
	return res, true
}

func (c *InMemoryCacheManager) GetGuildStickers(ctx context.Context, guildID Snowflake) (map[Snowflake]Sticker, bool) {
	set, ok := c.guildToStickerIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]Sticker, len(set))
	for stickerID := range set {
		if sticker, ok := c.stickers.Get(stickerID); ok {
			res[stickerID] = sticker
		}
	}
	return res, true
}

func (c *InMemoryCacheManager) GetVoiceChannelMembers(ctx context.Context, channelID Snowflake) (map[Snowflake]Member, bool) {
	ch, ok := c.GetChannel(ctx, channelID)
	if !ok {
		return nil, false
	}
	guildChannel, ok := ch.(GuildChannel)
	if !ok {
		return nil, false
	}
	guildID := guildChannel.GetGuildID()

	set, ok := c.channelToVoiceStateUserIDs.Get(channelID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]Member, len(set))
	for userID := range set {
		if member, found := c.GetMember(ctx, guildID, userID); found {
			res[userID] = member
		}
	}
	return res, true
}

func (c *InMemoryCacheManager) HasUser(ctx context.Context, userID Snowflake) bool {
	return c.flags.Has(CacheFlagUsers) && c.users.Has(userID)
}

func (c *InMemoryCacheManager) HasGuild(ctx context.Context, guildID Snowflake) bool {
	return c.flags.Has(CacheFlagGuilds) && c.guilds.Has(guildID)
}

func (c *InMemoryCacheManager) HasMember(ctx context.Context, guildID, userID Snowflake) bool {
	return c.flags.Has(CacheFlagMembers) && c.members.Has(SnowflakePairKey{A: guildID, B: userID})
}

func (c *InMemoryCacheManager) HasChannel(ctx context.Context, channelID Snowflake) bool {
	return c.flags.Has(CacheFlagChannels) && c.channels.Has(channelID)
}

func (c *InMemoryCacheManager) HasMessage(ctx context.Context, messageID Snowflake) bool {
	return c.flags.Has(CacheFlagMessages) && c.messages.Has(messageID)
}

func (c *InMemoryCacheManager) HasVoiceState(ctx context.Context, guildID, userID Snowflake) bool {
	return c.flags.Has(CacheFlagVoiceStates) && c.voiceStates.Has(SnowflakePairKey{A: guildID, B: userID})
}

func (c *InMemoryCacheManager) HasRole(ctx context.Context, roleID Snowflake) bool {
	return c.flags.Has(CacheFlagRoles) && c.roles.Has(roleID)
}

func (c *InMemoryCacheManager) HasRoles(ctx context.Context, rolesIDs ...Snowflake) bool {
	if !c.flags.Has(CacheFlagRoles) {
		return false
	}
	for _, id := range rolesIDs {
		if !c.roles.Has(id) {
			return false
		}
	}
	return true
}

func (c *InMemoryCacheManager) HasEmoji(ctx context.Context, emojiID Snowflake) bool {
	return c.flags.Has(CacheFlagEmojis) && c.emojis.Has(emojiID)
}

func (c *InMemoryCacheManager) HasSticker(ctx context.Context, stickerID Snowflake) bool {
	return c.flags.Has(CacheFlagStickers) && c.stickers.Has(stickerID)
}

func (c *InMemoryCacheManager) HasStageInstance(ctx context.Context, channelID Snowflake) bool {
	return c.flags.Has(CacheFlagStageInstances) && c.stageInstances.Has(channelID)
}

func (c *InMemoryCacheManager) HasGuildScheduledEvent(ctx context.Context, eventID Snowflake) bool {
	return c.flags.Has(CacheFlagGuildScheduledEvents) && c.scheduledEvents.Has(eventID)
}

func (c *InMemoryCacheManager) HasAutoModerationRule(ctx context.Context, ruleID Snowflake) bool {
	return c.flags.Has(CacheFlagAutoModerationRules) && c.autoModRules.Has(ruleID)
}

func (c *InMemoryCacheManager) CountUsers(ctx context.Context) int       { return c.users.Len() }
func (c *InMemoryCacheManager) CountGuilds(ctx context.Context) int      { return c.guilds.Len() }
func (c *InMemoryCacheManager) CountMembers(ctx context.Context) int     { return c.members.Len() }
func (c *InMemoryCacheManager) CountChannels(ctx context.Context) int    { return c.channels.Len() }
func (c *InMemoryCacheManager) CountMessages(ctx context.Context) int    { return c.messages.Len() }
func (c *InMemoryCacheManager) CountVoiceStates(ctx context.Context) int { return c.voiceStates.Len() }
func (c *InMemoryCacheManager) CountRoles(ctx context.Context) int       { return c.roles.Len() }
func (c *InMemoryCacheManager) CountEmojis(ctx context.Context) int      { return c.emojis.Len() }
func (c *InMemoryCacheManager) CountStickers(ctx context.Context) int    { return c.stickers.Len() }
func (c *InMemoryCacheManager) CountStageInstances(ctx context.Context) int {
	return c.stageInstances.Len()
}

func (c *InMemoryCacheManager) CountGuildScheduledEvents(ctx context.Context) int {
	return c.scheduledEvents.Len()
}

func (c *InMemoryCacheManager) CountAutoModerationRules(ctx context.Context) int {
	return c.autoModRules.Len()
}

func (c *InMemoryCacheManager) CountGuildChannels(ctx context.Context, guildID Snowflake) int {
	set, ok := c.guildToChannelIDs.Get(guildID)
	if !ok {
		return 0
	}
	return len(set)
}

func (c *InMemoryCacheManager) CountGuildMembers(ctx context.Context, guildID Snowflake) int {
	set, ok := c.guildToMemberIDs.Get(guildID)
	if !ok {
		return 0
	}
	return len(set)
}

func (c *InMemoryCacheManager) CountGuildRoles(ctx context.Context, guildID Snowflake) int {
	set, ok := c.guildToRoleIDs.Get(guildID)
	if !ok {
		return 0
	}
	return len(set)
}

func (c *InMemoryCacheManager) CountGuildEmojis(ctx context.Context, guildID Snowflake) int {
	set, ok := c.guildToEmojiIDs.Get(guildID)
	if !ok {
		return 0
	}
	return len(set)
}

func (c *InMemoryCacheManager) CountGuildStickers(ctx context.Context, guildID Snowflake) int {
	set, ok := c.guildToStickerIDs.Get(guildID)
	if !ok {
		return 0
	}
	return len(set)
}

func (c *InMemoryCacheManager) AddUser(ctx context.Context, user User) {
	if !c.flags.Has(CacheFlagUsers) {
		return
	}
	c.users.Set(user.ID, user)
}

func (c *InMemoryCacheManager) AddGuild(ctx context.Context, guild Guild) {
	if !c.flags.Has(CacheFlagGuilds) {
		return
	}
	c.guilds.Set(guild.ID, guild)
}

func (c *InMemoryCacheManager) AddMember(ctx context.Context, member Member) {
	if !c.flags.Has(CacheFlagMembers) {
		return
	}
	userID := member.ID
	guildID := member.GuildID
	key := SnowflakePairKey{A: guildID, B: userID}
	c.members.Set(key, member)
	c.guildToMemberIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
		if !exists || set == nil {
			set = make(map[Snowflake]struct{})
		}
		set[userID] = struct{}{}
		return set
	})
}

func (c *InMemoryCacheManager) AddChannel(ctx context.Context, channel Channel) {
	if !c.flags.Has(CacheFlagChannels) || channel == nil {
		return
	}
	channelID := channel.GetID()
	c.channels.Set(channelID, channel)
	if guildChannel, ok := channel.(GuildChannel); ok {
		guildID := guildChannel.GetGuildID()
		c.guildToChannelIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if !exists || set == nil {
				set = make(map[Snowflake]struct{})
			}
			set[channelID] = struct{}{}
			return set
		})
	}
}

func (c *InMemoryCacheManager) AddMessage(ctx context.Context, message Message) {
	if !c.flags.Has(CacheFlagMessages) {
		return
	}
	c.messages.Set(message.ID, message)

	maxPerChannel := c.config.MaxMessagesPerChannel
	if maxPerChannel <= 0 {
		maxPerChannel = 100
	}

	channelID := message.ChannelID
	if channelID != 0 {
		var evictedID Snowflake
		c.channelToMessageIDs.Compute(channelID, func(list []Snowflake, exists bool) []Snowflake {
			if !exists || list == nil {
				list = make([]Snowflake, 0, min(maxPerChannel, 16))
			}
			list = append(list, message.ID)
			if len(list) > maxPerChannel {
				evictedID = list[0]
				list = list[1:]
			}
			return list
		})
		if evictedID != 0 {
			c.messages.Delete(evictedID)
		}
	}
}

func (c *InMemoryCacheManager) AddVoiceState(ctx context.Context, voiceState VoiceState) {
	if !c.flags.Has(CacheFlagVoiceStates) {
		return
	}
	guildID := voiceState.GuildID
	userID := voiceState.UserID
	key := SnowflakePairKey{A: guildID, B: userID}

	var oldChannelID Snowflake
	if oldVS, ok := c.voiceStates.Get(key); ok {
		oldChannelID = oldVS.ChannelID
	}

	if voiceState.ChannelID == 0 {
		c.voiceStates.Delete(key)
		res := c.guildToVoiceStateUserIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if exists && set != nil {
				delete(set, userID)
				return set
			}
			return nil
		})
		if len(res) == 0 {
			c.guildToVoiceStateUserIDs.Delete(guildID)
		}
	} else {
		c.voiceStates.Set(key, voiceState)
		c.guildToVoiceStateUserIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if !exists || set == nil {
				set = make(map[Snowflake]struct{})
			}
			set[userID] = struct{}{}
			return set
		})
	}

	if oldChannelID != 0 && oldChannelID != voiceState.ChannelID {
		res := c.channelToVoiceStateUserIDs.Compute(oldChannelID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if exists && set != nil {
				delete(set, userID)
				return set
			}
			return nil
		})
		if len(res) == 0 {
			c.channelToVoiceStateUserIDs.Delete(oldChannelID)
		}
	}

	if voiceState.ChannelID != 0 {
		c.channelToVoiceStateUserIDs.Compute(voiceState.ChannelID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if !exists || set == nil {
				set = make(map[Snowflake]struct{})
			}
			set[userID] = struct{}{}
			return set
		})
	}
}

func (c *InMemoryCacheManager) AddRole(ctx context.Context, role Role) {
	if !c.flags.Has(CacheFlagRoles) {
		return
	}
	guildID := role.GuildID
	roleID := role.ID
	c.roles.Set(roleID, role)
	c.guildToRoleIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
		if !exists || set == nil {
			set = make(map[Snowflake]struct{})
		}
		set[roleID] = struct{}{}
		return set
	})
}

func (c *InMemoryCacheManager) AddRoles(ctx context.Context, roles ...Role) {
	if !c.flags.Has(CacheFlagRoles) || len(roles) == 0 {
		return
	}
	guildToRoles := make(map[Snowflake][]Role)
	for _, role := range roles {
		guildToRoles[role.GuildID] = append(guildToRoles[role.GuildID], role)
		c.roles.Set(role.ID, role)
	}
	for guildID, guildRoles := range guildToRoles {
		c.guildToRoleIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if !exists || set == nil {
				set = make(map[Snowflake]struct{})
			}
			for _, role := range guildRoles {
				set[role.ID] = struct{}{}
			}
			return set
		})
	}
}

func (c *InMemoryCacheManager) AddEmoji(ctx context.Context, emoji Emoji) {
	if !c.flags.Has(CacheFlagEmojis) {
		return
	}
	c.emojis.Set(emoji.ID, emoji)
	if emoji.GuildID != 0 {
		c.guildToEmojiIDs.Compute(emoji.GuildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if !exists || set == nil {
				set = make(map[Snowflake]struct{})
			}
			set[emoji.ID] = struct{}{}
			return set
		})
	}
}

func (c *InMemoryCacheManager) AddEmojis(ctx context.Context, emojis ...Emoji) {
	if !c.flags.Has(CacheFlagEmojis) || len(emojis) == 0 {
		return
	}
	for _, emoji := range emojis {
		c.AddEmoji(ctx, emoji)
	}
}

func (c *InMemoryCacheManager) AddSticker(ctx context.Context, sticker Sticker) {
	if !c.flags.Has(CacheFlagStickers) {
		return
	}
	c.stickers.Set(sticker.ID, sticker)
	if sticker.GuildID != 0 {
		c.guildToStickerIDs.Compute(sticker.GuildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if !exists || set == nil {
				set = make(map[Snowflake]struct{})
			}
			set[sticker.ID] = struct{}{}
			return set
		})
	}
}

func (c *InMemoryCacheManager) AddStickers(ctx context.Context, stickers ...Sticker) {
	if !c.flags.Has(CacheFlagStickers) || len(stickers) == 0 {
		return
	}
	for _, sticker := range stickers {
		c.AddSticker(ctx, sticker)
	}
}

func (c *InMemoryCacheManager) AddStageInstance(ctx context.Context, stage StageInstance) {
	if !c.flags.Has(CacheFlagStageInstances) {
		return
	}
	c.stageInstances.Set(stage.ChannelID, stage)
}

func (c *InMemoryCacheManager) AddGuildScheduledEvent(ctx context.Context, event GuildScheduledEvent) {
	if !c.flags.Has(CacheFlagGuildScheduledEvents) {
		return
	}
	c.scheduledEvents.Set(event.ID, event)
}

func (c *InMemoryCacheManager) AddAutoModerationRule(ctx context.Context, rule AutoModerationRule) {
	if !c.flags.Has(CacheFlagAutoModerationRules) {
		return
	}
	c.autoModRules.Set(rule.ID, rule)
	if rule.GuildID != 0 {
		c.guildToAutoModRuleIDs.Compute(rule.GuildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if !exists || set == nil {
				set = make(map[Snowflake]struct{})
			}
			set[rule.ID] = struct{}{}
			return set
		})
	}
}

func (c *InMemoryCacheManager) DelUser(ctx context.Context, userID Snowflake) bool {
	_, ok := c.users.Delete(userID)
	return ok
}

func (c *InMemoryCacheManager) DelGuild(ctx context.Context, guildID Snowflake) bool {
	_, ok := c.guilds.Delete(guildID)
	if ok {
		c.DelGuildMembers(ctx, guildID)
		c.DelGuildChannels(ctx, guildID)
		c.DelGuildVoiceStates(ctx, guildID)
		c.DelGuildRoles(ctx, guildID)
		c.DelGuildEmojis(ctx, guildID)
		c.DelGuildStickers(ctx, guildID)
	}
	return ok
}

func (c *InMemoryCacheManager) DelMember(ctx context.Context, guildID, userID Snowflake) bool {
	key := SnowflakePairKey{A: guildID, B: userID}
	_, ok := c.members.Delete(key)
	if ok {
		res := c.guildToMemberIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if exists && set != nil {
				delete(set, userID)
				return set
			}
			return nil
		})
		if len(res) == 0 {
			c.guildToMemberIDs.Delete(guildID)
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelChannel(ctx context.Context, channelID Snowflake) bool {
	channel, ok := c.channels.Delete(channelID)
	if ok {
		if list, hasMessages := c.channelToMessageIDs.Delete(channelID); hasMessages {
			for _, mid := range list {
				c.messages.Delete(mid)
			}
		}
		if guildChannel, ok := channel.(GuildChannel); ok {
			res := c.guildToChannelIDs.Compute(guildChannel.GetGuildID(), func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
				if exists && set != nil {
					delete(set, channelID)
					return set
				}
				return nil
			})
			if len(res) == 0 {
				c.guildToChannelIDs.Delete(guildChannel.GetGuildID())
			}
		}
		c.channelToVoiceStateUserIDs.Delete(channelID)
	}
	return ok
}

func (c *InMemoryCacheManager) DelMessage(ctx context.Context, messageID Snowflake) bool {
	msg, ok := c.messages.Delete(messageID)
	if ok && msg.ChannelID != 0 {
		c.channelToMessageIDs.Compute(msg.ChannelID, func(list []Snowflake, exists bool) []Snowflake {
			if !exists || len(list) == 0 {
				return nil
			}
			for i, id := range list {
				if id == messageID {
					return append(list[:i], list[i+1:]...)
				}
			}
			return list
		})
	}
	return ok
}

func (c *InMemoryCacheManager) DelVoiceState(ctx context.Context, guildID, userID Snowflake) bool {
	key := SnowflakePairKey{A: guildID, B: userID}
	vs, ok := c.voiceStates.Delete(key)
	if ok {
		res := c.guildToVoiceStateUserIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if exists && set != nil {
				delete(set, userID)
				return set
			}
			return nil
		})
		if len(res) == 0 {
			c.guildToVoiceStateUserIDs.Delete(guildID)
		}

		if vs.ChannelID != 0 {
			res := c.channelToVoiceStateUserIDs.Compute(vs.ChannelID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
				if exists && set != nil {
					delete(set, userID)
					return set
				}
				return nil
			})
			if len(res) == 0 {
				c.channelToVoiceStateUserIDs.Delete(vs.ChannelID)
			}
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelGuildVoiceStates(ctx context.Context, guildID Snowflake) bool {
	set, ok := c.guildToVoiceStateUserIDs.Delete(guildID)
	if ok {
		for userID := range set {
			key := SnowflakePairKey{A: guildID, B: userID}
			if vs, ok := c.voiceStates.Delete(key); ok {
				if vs.ChannelID != 0 {
					res := c.channelToVoiceStateUserIDs.Compute(vs.ChannelID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
						if exists && set != nil {
							delete(set, userID)
							return set
						}
						return nil
					})
					if len(res) == 0 {
						c.channelToVoiceStateUserIDs.Delete(vs.ChannelID)
					}
				}
			}
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelRole(ctx context.Context, guildID, roleID Snowflake) bool {
	_, ok := c.roles.Delete(roleID)
	if ok {
		res := c.guildToRoleIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if exists && set != nil {
				delete(set, roleID)
				return set
			}
			return nil
		})
		if len(res) == 0 {
			c.guildToRoleIDs.Delete(guildID)
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelRoles(ctx context.Context, roleIDs ...Snowflake) bool {
	anyDeleted := false
	for _, id := range roleIDs {
		if _, ok := c.roles.Delete(id); ok {
			anyDeleted = true
		}
	}
	return anyDeleted
}

func (c *InMemoryCacheManager) DelEmoji(ctx context.Context, guildID, emojiID Snowflake) bool {
	_, ok := c.emojis.Delete(emojiID)
	if ok && guildID != 0 {
		res := c.guildToEmojiIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if exists && set != nil {
				delete(set, emojiID)
				return set
			}
			return nil
		})
		if len(res) == 0 {
			c.guildToEmojiIDs.Delete(guildID)
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelSticker(ctx context.Context, guildID, stickerID Snowflake) bool {
	_, ok := c.stickers.Delete(stickerID)
	if ok && guildID != 0 {
		res := c.guildToStickerIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if exists && set != nil {
				delete(set, stickerID)
				return set
			}
			return nil
		})
		if len(res) == 0 {
			c.guildToStickerIDs.Delete(guildID)
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelStageInstance(ctx context.Context, channelID Snowflake) bool {
	_, ok := c.stageInstances.Delete(channelID)
	return ok
}

func (c *InMemoryCacheManager) DelGuildScheduledEvent(ctx context.Context, eventID Snowflake) bool {
	_, ok := c.scheduledEvents.Delete(eventID)
	return ok
}

func (c *InMemoryCacheManager) DelAutoModerationRule(ctx context.Context, guildID, ruleID Snowflake) bool {
	_, ok := c.autoModRules.Delete(ruleID)
	if ok && guildID != 0 {
		res := c.guildToAutoModRuleIDs.Compute(guildID, func(set map[Snowflake]struct{}, exists bool) map[Snowflake]struct{} {
			if exists && set != nil {
				delete(set, ruleID)
				return set
			}
			return nil
		})
		if len(res) == 0 {
			c.guildToAutoModRuleIDs.Delete(guildID)
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelGuildRoles(ctx context.Context, guildID Snowflake) bool {
	set, ok := c.guildToRoleIDs.Delete(guildID)
	if ok {
		for roleID := range set {
			c.roles.Delete(roleID)
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelGuildChannels(ctx context.Context, guildID Snowflake) bool {
	set, ok := c.guildToChannelIDs.Delete(guildID)
	if ok {
		for channelID := range set {
			c.channels.Delete(channelID)
			c.channelToVoiceStateUserIDs.Delete(channelID)
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelGuildMembers(ctx context.Context, guildID Snowflake) bool {
	set, ok := c.guildToMemberIDs.Delete(guildID)
	if ok {
		for userID := range set {
			c.members.Delete(SnowflakePairKey{A: guildID, B: userID})
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelGuildEmojis(ctx context.Context, guildID Snowflake) bool {
	set, ok := c.guildToEmojiIDs.Delete(guildID)
	if ok {
		for emojiID := range set {
			c.emojis.Delete(emojiID)
		}
	}
	return ok
}

func (c *InMemoryCacheManager) DelGuildStickers(ctx context.Context, guildID Snowflake) bool {
	set, ok := c.guildToStickerIDs.Delete(guildID)
	if ok {
		for stickerID := range set {
			c.stickers.Delete(stickerID)
		}
	}
	return ok
}
