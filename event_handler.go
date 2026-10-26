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
	"encoding/json"
)

// dispatchHandlers executes a slice of event handlers either concurrently (in new goroutines) or sequentially.
func dispatchHandlers[E any](ctx context.Context, handlers []func(context.Context, E), evt E, concurrent bool) {
	if len(handlers) == 0 {
		return
	}
	if concurrent {
		for _, handler := range handlers {
			go handler(ctx, evt)
		}
	} else {
		for _, handler := range handlers {
			handler(ctx, evt)
		}
	}
}

/*****************************
 *   READY Handler
 *****************************/

type readyHandlers struct {
	logger   Logger
	handlers []func(context.Context, ReadyEvent)
}

func (h *readyHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ReadyEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "READY", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagGuilds) {
		for i := range evt.Guilds {
			// update unavailable status for existing guilds without overwriting full data
			if existing, ok := cache.GetGuild(ctx, evt.Guilds[i].ID); ok {
				existing.Unavailable = evt.Guilds[i].Unavailable
				cache.AddGuild(ctx, existing)
			}
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *readyHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ReadyEvent)))
}

/*****************************
 *   RESUMED Handler
 *****************************/

type resumedHandlers struct {
	logger   Logger
	handlers []func(context.Context, ResumedEvent)
}

func (h *resumedHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ResumedEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "RESUMED", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *resumedHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ResumedEvent)))
}

/*****************************
 *   GUILD_CREATE Handler
 *****************************/

type guildCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildCreateEvent)
}

func (h *guildCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildCreateEvent{ShardID: shardID}

	if err := json.Unmarshal(data, &evt.Guild); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_CREATE", "error", err)
		return
	}

	if cache != nil {
		flags := cache.CacheFlags()
		if flags.Has(CacheFlagGuilds) {
			cache.AddGuild(ctx, evt.Guild.Guild)
		}
		if flags.Has(CacheFlagChannels) {
			for i := range len(evt.Guild.Channels) {
				cache.AddChannel(ctx, evt.Guild.Channels[i])
			}
		}
		if flags.Has(CacheFlagRoles) {
			for i := range len(evt.Guild.Roles) {
				cache.AddRole(ctx, evt.Guild.Roles[i])
			}
		}
		if flags.Has(CacheFlagVoiceStates) {
			for i := range len(evt.Guild.VoiceStates) {
				cache.AddVoiceState(ctx, evt.Guild.VoiceStates[i])
			}
		}
		if flags.Has(CacheFlagMembers) {
			for i := range len(evt.Guild.Members) {
				cache.AddMember(ctx, evt.Guild.Members[i].Member)
			}
		}
		if flags.Has(CacheFlagUsers) {
			for i := range len(evt.Guild.Members) {
				cache.AddUser(ctx, evt.Guild.Members[i].User)
			}
		}
		if flags.Has(CacheFlagChannels) {
			for i := range len(evt.Guild.Threads) {
				cache.AddChannel(ctx, &evt.Guild.Threads[i])
			}
		}
		if flags.Has(CacheFlagStageInstances) {
			for i := range len(evt.Guild.StageInstances) {
				cache.AddStageInstance(ctx, evt.Guild.StageInstances[i])
			}
		}
		if flags.Has(CacheFlagEmojis) {
			for i := range len(evt.Guild.Emojis) {
				evt.Guild.Emojis[i].GuildID = evt.Guild.ID
				cache.AddEmoji(ctx, evt.Guild.Emojis[i])
			}
		}
		if flags.Has(CacheFlagStickers) {
			for i := range len(evt.Guild.Stickers) {
				evt.Guild.Stickers[i].GuildID = evt.Guild.ID
				cache.AddSticker(ctx, evt.Guild.Stickers[i])
			}
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildCreateEvent)))
}

/*****************************
 *   GUILD_UPDATE Handler
 *****************************/

type guildUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildUpdateEvent)
}

func (h *guildUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt.NewGuild); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if oldGuild, ok := cache.GetGuild(ctx, evt.NewGuild.ID); ok {
			evt.OldGuild = oldGuild
		}
		if cache.CacheFlags().Has(CacheFlagGuilds) {
			cache.AddGuild(ctx, evt.NewGuild)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildUpdateEvent)))
}

/*****************************
 *   GUILD_DELETE Handler
 *****************************/

type guildDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildDeleteEvent)
}

func (h *guildDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_DELETE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagGuilds) {
		cache.DelGuild(ctx, evt.ID)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildDeleteEvent)))
}

/*****************************
 *   CHANNEL_CREATE Handler
 *****************************/

type channelCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, ChannelCreateEvent)
}

func (h *channelCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ChannelCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "CHANNEL_CREATE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagChannels) {
		cache.AddChannel(ctx, evt.Channel)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *channelCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ChannelCreateEvent)))
}

/*****************************
 *   CHANNEL_UPDATE Handler
 *****************************/

type channelUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, ChannelUpdateEvent)
}

func (h *channelUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ChannelUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "CHANNEL_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if ch, ok := evt.NewChannel.(GuildChannel); ok {
			if oldCh, ok := cache.GetChannel(ctx, ch.GetID()); ok {
				evt.OldChannel = oldCh
			}
		}
		if cache.CacheFlags().Has(CacheFlagChannels) {
			cache.AddChannel(ctx, evt.NewChannel)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *channelUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ChannelUpdateEvent)))
}

/*****************************
 *   CHANNEL_DELETE Handler
 *****************************/

type channelDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, ChannelDeleteEvent)
}

func (h *channelDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ChannelDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "CHANNEL_DELETE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagChannels) {
		cache.DelChannel(ctx, evt.Channel.GetID())
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *channelDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ChannelDeleteEvent)))
}

/*****************************
 *   CHANNEL_PINS_UPDATE Handler
 *****************************/

type channelPinsUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, ChannelPinsUpdateEvent)
}

func (h *channelPinsUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ChannelPinsUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "CHANNEL_PINS_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *channelPinsUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ChannelPinsUpdateEvent)))
}

/*****************************
 *   THREAD_CREATE Handler
 *****************************/

type threadCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, ThreadCreateEvent)
}

func (h *threadCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ThreadCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "THREAD_CREATE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagChannels) && evt.Thread != nil {
		cache.AddChannel(ctx, evt.Thread)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *threadCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ThreadCreateEvent)))
}

/*****************************
 *   THREAD_UPDATE Handler
 *****************************/

type threadUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, ThreadUpdateEvent)
}

func (h *threadUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ThreadUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "THREAD_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if evt.NewThread != nil {
			if oldCh, ok := cache.GetChannel(ctx, evt.NewThread.ID); ok {
				if tc, ok := oldCh.(*ThreadChannel); ok {
					evt.OldThread = tc
				}
			}
		}
		if cache.CacheFlags().Has(CacheFlagChannels) && evt.NewThread != nil {
			cache.AddChannel(ctx, evt.NewThread)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *threadUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ThreadUpdateEvent)))
}

/*****************************
 *   THREAD_DELETE Handler
 *****************************/

type threadDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, ThreadDeleteEvent)
}

func (h *threadDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ThreadDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "THREAD_DELETE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagChannels) {
		cache.DelChannel(ctx, evt.ID)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *threadDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ThreadDeleteEvent)))
}

/*****************************
 *   THREAD_LIST_SYNC Handler
 *****************************/

type threadListSyncHandlers struct {
	logger   Logger
	handlers []func(context.Context, ThreadListSyncEvent)
}

func (h *threadListSyncHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ThreadListSyncEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "THREAD_LIST_SYNC", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *threadListSyncHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ThreadListSyncEvent)))
}

/*****************************
 *   THREAD_MEMBER_UPDATE Handler
 *****************************/

type threadMemberUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, ThreadMemberUpdateEvent)
}

func (h *threadMemberUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ThreadMemberUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "THREAD_MEMBER_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *threadMemberUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ThreadMemberUpdateEvent)))
}

/*****************************
 *   THREAD_MEMBERS_UPDATE Handler
 *****************************/

type threadMembersUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, ThreadMembersUpdateEvent)
}

func (h *threadMembersUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ThreadMembersUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "THREAD_MEMBERS_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *threadMembersUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ThreadMembersUpdateEvent)))
}

/*****************************
 *   GUILD_MEMBER_ADD Handler
 *****************************/

type guildMemberAddHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildMemberAddEvent)
}

func (h *guildMemberAddHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildMemberAddEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_MEMBER_ADD", "error", err)
		return
	}

	if cache != nil {
		if cache.CacheFlags().Has(CacheFlagMembers) {
			cache.AddMember(ctx, evt.Member.Member)
		}
		if cache.CacheFlags().Has(CacheFlagUsers) {
			cache.AddUser(ctx, evt.Member.User)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildMemberAddHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildMemberAddEvent)))
}

/*****************************
 *   GUILD_MEMBER_UPDATE Handler
 *****************************/

type guildMemberUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildMemberUpdateEvent)
}

func (h *guildMemberUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildMemberUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_MEMBER_UPDATE", "error", err)
		return
	}

	evt.NewMember = Member{
		ID:                         evt.User.ID,
		GuildID:                    evt.GuildID,
		Nickname:                   evt.Nick,
		Avatar:                     evt.Avatar,
		RoleIDs:                    evt.Roles,
		JoinedAt:                   evt.JoinedAt,
		PremiumSince:               evt.PremiumSince,
		Deaf:                       evt.Deaf,
		Mute:                       evt.Mute,
		Pending:                    evt.Pending,
		CommunicationDisabledUntil: evt.CommunicationDisabledUntil,
		Flags:                      evt.Flags,
		AvatarDecorationData:       evt.AvatarDecorationData,
	}

	if cache != nil {
		if oldMember, ok := cache.GetMember(ctx, evt.GuildID, evt.User.ID); ok {
			evt.OldMember = oldMember
		}
		if cache.CacheFlags().Has(CacheFlagMembers) {
			cache.AddMember(ctx, evt.NewMember)
		}
		if cache.CacheFlags().Has(CacheFlagUsers) {
			cache.AddUser(ctx, evt.User)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildMemberUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildMemberUpdateEvent)))
}

/*****************************
 *   GUILD_MEMBER_REMOVE Handler
 *****************************/

type guildMemberRemoveHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildMemberRemoveEvent)
}

func (h *guildMemberRemoveHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildMemberRemoveEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_MEMBER_REMOVE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagMembers) {
		cache.DelMember(ctx, evt.GuildID, evt.User.ID)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildMemberRemoveHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildMemberRemoveEvent)))
}

/*****************************
 *   GUILD_MEMBERS_CHUNK Handler
 *****************************/

type guildMembersChunkHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildMembersChunkEvent)
}

func (h *guildMembersChunkHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildMembersChunkEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_MEMBERS_CHUNK", "error", err)
		return
	}

	if cache != nil {
		if cache.CacheFlags().Has(CacheFlagMembers) {
			for i := range len(evt.Members) {
				cache.AddMember(ctx, evt.Members[i].Member)
			}
		}
		if cache.CacheFlags().Has(CacheFlagUsers) {
			for i := range len(evt.Members) {
				cache.AddUser(ctx, evt.Members[i].User)
			}
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildMembersChunkHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildMembersChunkEvent)))
}

/*****************************
 *   GUILD_ROLE_CREATE Handler
 *****************************/

type guildRoleCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildRoleCreateEvent)
}

func (h *guildRoleCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildRoleCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_ROLE_CREATE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagRoles) {
		cache.AddRole(ctx, evt.Role)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildRoleCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildRoleCreateEvent)))
}

/*****************************
 *   GUILD_ROLE_UPDATE Handler
 *****************************/

type guildRoleUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildRoleUpdateEvent)
}

func (h *guildRoleUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildRoleUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_ROLE_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if oldRole, ok := cache.GetRole(ctx, evt.NewRole.ID); ok {
			evt.OldRole = oldRole
		}
		if cache.CacheFlags().Has(CacheFlagRoles) {
			cache.AddRole(ctx, evt.NewRole)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildRoleUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildRoleUpdateEvent)))
}

/*****************************
 *   GUILD_ROLE_DELETE Handler
 *****************************/

type guildRoleDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildRoleDeleteEvent)
}

func (h *guildRoleDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildRoleDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_ROLE_DELETE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagRoles) {
		cache.DelRole(ctx, evt.GuildID, evt.RoleID)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildRoleDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildRoleDeleteEvent)))
}

/*****************************
 *   GUILD_EMOJIS_UPDATE Handler
 *****************************/

type guildEmojisUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildEmojisUpdateEvent)
}

func (h *guildEmojisUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildEmojisUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_EMOJIS_UPDATE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagEmojis) {
		for i := range len(evt.NewEmojis) {
			evt.NewEmojis[i].GuildID = evt.GuildID
			cache.AddEmoji(ctx, evt.NewEmojis[i])
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildEmojisUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildEmojisUpdateEvent)))
}

/*****************************
 *   GUILD_STICKERS_UPDATE Handler
 *****************************/

type guildStickersUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildStickersUpdateEvent)
}

func (h *guildStickersUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildStickersUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_STICKERS_UPDATE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagStickers) {
		for i := range len(evt.NewStickers) {
			evt.NewStickers[i].GuildID = evt.GuildID
			cache.AddSticker(ctx, evt.NewStickers[i])
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildStickersUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildStickersUpdateEvent)))
}

/*****************************
 *   GUILD_INTEGRATIONS_UPDATE Handler
 *****************************/

type guildIntegrationsUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildIntegrationsUpdateEvent)
}

func (h *guildIntegrationsUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildIntegrationsUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_INTEGRATIONS_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildIntegrationsUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildIntegrationsUpdateEvent)))
}

/*****************************
 *   GUILD_AUDIT_LOG_ENTRY_CREATE Handler
 *****************************/

type guildAuditLogEntryCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildAuditLogEntryCreateEvent)
}

func (h *guildAuditLogEntryCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildAuditLogEntryCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_AUDIT_LOG_ENTRY_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildAuditLogEntryCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildAuditLogEntryCreateEvent)))
}

/*****************************
 *   GUILD_BAN_ADD Handler
 *****************************/

type guildBanAddHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildBanAddEvent)
}

func (h *guildBanAddHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildBanAddEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_BAN_ADD", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildBanAddHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildBanAddEvent)))
}

/*****************************
 *   GUILD_BAN_REMOVE Handler
 *****************************/

type guildBanRemoveHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildBanRemoveEvent)
}

func (h *guildBanRemoveHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildBanRemoveEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_BAN_REMOVE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildBanRemoveHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildBanRemoveEvent)))
}

/*****************************
 *   GUILD_SCHEDULED_EVENT_CREATE Handler
 *****************************/

type guildScheduledEventCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildScheduledEventCreateEvent)
}

func (h *guildScheduledEventCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildScheduledEventCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SCHEDULED_EVENT_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildScheduledEventCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildScheduledEventCreateEvent)))
}

/*****************************
 *   GUILD_SCHEDULED_EVENT_UPDATE Handler
 *****************************/

type guildScheduledEventUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildScheduledEventUpdateEvent)
}

func (h *guildScheduledEventUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildScheduledEventUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SCHEDULED_EVENT_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildScheduledEventUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildScheduledEventUpdateEvent)))
}

/*****************************
 *   GUILD_SCHEDULED_EVENT_DELETE Handler
 *****************************/

type guildScheduledEventDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildScheduledEventDeleteEvent)
}

func (h *guildScheduledEventDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildScheduledEventDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SCHEDULED_EVENT_DELETE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildScheduledEventDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildScheduledEventDeleteEvent)))
}

/*****************************
 *   GUILD_SCHEDULED_EVENT_USER_ADD Handler
 *****************************/

type guildScheduledEventUserAddHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildScheduledEventUserAddEvent)
}

func (h *guildScheduledEventUserAddHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildScheduledEventUserAddEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SCHEDULED_EVENT_USER_ADD", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildScheduledEventUserAddHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildScheduledEventUserAddEvent)))
}

/*****************************
 *   GUILD_SCHEDULED_EVENT_USER_REMOVE Handler
 *****************************/

type guildScheduledEventUserRemoveHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildScheduledEventUserRemoveEvent)
}

func (h *guildScheduledEventUserRemoveHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildScheduledEventUserRemoveEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SCHEDULED_EVENT_USER_REMOVE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildScheduledEventUserRemoveHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildScheduledEventUserRemoveEvent)))
}

/*****************************
 *   GUILD_SOUNDBOARD_SOUND_CREATE Handler
 *****************************/

type guildSoundboardSoundCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildSoundboardSoundCreateEvent)
}

func (h *guildSoundboardSoundCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildSoundboardSoundCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SOUNDBOARD_SOUND_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildSoundboardSoundCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildSoundboardSoundCreateEvent)))
}

/*****************************
 *   GUILD_SOUNDBOARD_SOUND_UPDATE Handler
 *****************************/

type guildSoundboardSoundUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildSoundboardSoundUpdateEvent)
}

func (h *guildSoundboardSoundUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildSoundboardSoundUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SOUNDBOARD_SOUND_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildSoundboardSoundUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildSoundboardSoundUpdateEvent)))
}

/*****************************
 *   GUILD_SOUNDBOARD_SOUND_DELETE Handler
 *****************************/

type guildSoundboardSoundDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildSoundboardSoundDeleteEvent)
}

func (h *guildSoundboardSoundDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildSoundboardSoundDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SOUNDBOARD_SOUND_DELETE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildSoundboardSoundDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildSoundboardSoundDeleteEvent)))
}

/*****************************
 *   GUILD_SOUNDBOARD_SOUNDS_UPDATE Handler
 *****************************/

type guildSoundboardSoundsUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GuildSoundboardSoundsUpdateEvent)
}

func (h *guildSoundboardSoundsUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GuildSoundboardSoundsUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GUILD_SOUNDBOARD_SOUNDS_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *guildSoundboardSoundsUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GuildSoundboardSoundsUpdateEvent)))
}

/*****************************
 *   SOUNDBOARD_SOUNDS Handler
 *****************************/

type soundboardSoundsHandlers struct {
	logger   Logger
	handlers []func(context.Context, SoundboardSoundsEvent)
}

func (h *soundboardSoundsHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := SoundboardSoundsEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "SOUNDBOARD_SOUNDS", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *soundboardSoundsHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, SoundboardSoundsEvent)))
}

/*****************************
 *   INTEGRATION_CREATE Handler
 *****************************/

type integrationCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, IntegrationCreateEvent)
}

func (h *integrationCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := IntegrationCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "INTEGRATION_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *integrationCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, IntegrationCreateEvent)))
}

/*****************************
 *   INTEGRATION_UPDATE Handler
 *****************************/

type integrationUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, IntegrationUpdateEvent)
}

func (h *integrationUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := IntegrationUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "INTEGRATION_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *integrationUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, IntegrationUpdateEvent)))
}

/*****************************
 *   INTEGRATION_DELETE Handler
 *****************************/

type integrationDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, IntegrationDeleteEvent)
}

func (h *integrationDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := IntegrationDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "INTEGRATION_DELETE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *integrationDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, IntegrationDeleteEvent)))
}

/*****************************
 *   INTERACTION_CREATE Handler
 *****************************/

type interactionCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, InteractionCreateEvent)
}

func (h *interactionCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := InteractionCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "INTERACTION_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *interactionCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, InteractionCreateEvent)))
}

/*****************************
 *   INVITE_CREATE Handler
 *****************************/

type inviteCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, InviteCreateEvent)
}

func (h *inviteCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := InviteCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "INVITE_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *inviteCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, InviteCreateEvent)))
}

/*****************************
 *   INVITE_DELETE Handler
 *****************************/

type inviteDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, InviteDeleteEvent)
}

func (h *inviteDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := InviteDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "INVITE_DELETE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *inviteDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, InviteDeleteEvent)))
}

/*****************************
 *   MESSAGE_CREATE Handler
 *****************************/

type messageCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessageCreateEvent)
}

func (h *messageCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessageCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt.Message); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_CREATE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagMessages) {
		cache.AddMessage(ctx, evt.Message)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messageCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessageCreateEvent)))
}

/*****************************
 *   MESSAGE_UPDATE Handler
 *****************************/

type messageUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessageUpdateEvent)
}

func (h *messageUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessageUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt.NewMessage); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if oldMsg, ok := cache.GetMessage(ctx, evt.NewMessage.ID); ok {
			evt.OldMessage = oldMsg
		}
		if cache.CacheFlags().Has(CacheFlagMessages) {
			cache.AddMessage(ctx, evt.NewMessage)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messageUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessageUpdateEvent)))
}

/*****************************
 *   MESSAGE_DELETE Handler
 *****************************/

type messageDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessageDeleteEvent)
}

func (h *messageDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessageDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt.Message); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_DELETE", "error", err)
		return
	}

	if cache != nil {
		if msg, ok := cache.GetMessage(ctx, evt.Message.ID); ok {
			evt.Message = msg
		}
		cache.DelMessage(ctx, evt.Message.ID)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messageDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessageDeleteEvent)))
}

/*****************************
 *   MESSAGE_DELETE_BULK Handler
 *****************************/

type messageDeleteBulkHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessageDeleteBulkEvent)
}

func (h *messageDeleteBulkHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessageDeleteBulkEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_DELETE_BULK", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagMessages) {
		for _, id := range evt.IDs {
			cache.DelMessage(ctx, id)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messageDeleteBulkHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessageDeleteBulkEvent)))
}

/*****************************
 *   MESSAGE_REACTION_ADD Handler
 *****************************/

type messageReactionAddHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessageReactionAddEvent)
}

func (h *messageReactionAddHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessageReactionAddEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_REACTION_ADD", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messageReactionAddHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessageReactionAddEvent)))
}

/*****************************
 *   MESSAGE_REACTION_REMOVE Handler
 *****************************/

type messageReactionRemoveHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessageReactionRemoveEvent)
}

func (h *messageReactionRemoveHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessageReactionRemoveEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_REACTION_REMOVE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messageReactionRemoveHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessageReactionRemoveEvent)))
}

/*****************************
 *   MESSAGE_REACTION_REMOVE_ALL Handler
 *****************************/

type messageReactionRemoveAllHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessageReactionRemoveAllEvent)
}

func (h *messageReactionRemoveAllHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessageReactionRemoveAllEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_REACTION_REMOVE_ALL", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messageReactionRemoveAllHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessageReactionRemoveAllEvent)))
}

/*****************************
 *   MESSAGE_REACTION_REMOVE_EMOJI Handler
 *****************************/

type messageReactionRemoveEmojiHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessageReactionRemoveEmojiEvent)
}

func (h *messageReactionRemoveEmojiHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessageReactionRemoveEmojiEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_REACTION_REMOVE_EMOJI", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messageReactionRemoveEmojiHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessageReactionRemoveEmojiEvent)))
}

/*****************************
 *   MESSAGE_POLL_VOTE_ADD Handler
 *****************************/

type messagePollVoteAddHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessagePollVoteAddEvent)
}

func (h *messagePollVoteAddHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessagePollVoteAddEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_POLL_VOTE_ADD", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messagePollVoteAddHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessagePollVoteAddEvent)))
}

/*****************************
 *   MESSAGE_POLL_VOTE_REMOVE Handler
 *****************************/

type messagePollVoteRemoveHandlers struct {
	logger   Logger
	handlers []func(context.Context, MessagePollVoteRemoveEvent)
}

func (h *messagePollVoteRemoveHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := MessagePollVoteRemoveEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "MESSAGE_POLL_VOTE_REMOVE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *messagePollVoteRemoveHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, MessagePollVoteRemoveEvent)))
}

/*****************************
 *   PRESENCE_UPDATE Handler
 *****************************/

type presenceUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, PresenceUpdateEvent)
}

func (h *presenceUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := PresenceUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "PRESENCE_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *presenceUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, PresenceUpdateEvent)))
}

/*****************************
 *   STAGE_INSTANCE_CREATE Handler
 *****************************/

type stageInstanceCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, StageInstanceCreateEvent)
}

func (h *stageInstanceCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := StageInstanceCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "STAGE_INSTANCE_CREATE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagStageInstances) {
		cache.AddStageInstance(ctx, evt.StageInstance)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *stageInstanceCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, StageInstanceCreateEvent)))
}

/*****************************
 *   STAGE_INSTANCE_UPDATE Handler
 *****************************/

type stageInstanceUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, StageInstanceUpdateEvent)
}

func (h *stageInstanceUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := StageInstanceUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "STAGE_INSTANCE_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if oldStage, ok := cache.GetStageInstance(ctx, evt.NewStageInstance.ChannelID); ok {
			evt.OldStageInstance = oldStage
		}
		if cache.CacheFlags().Has(CacheFlagStageInstances) {
			cache.AddStageInstance(ctx, evt.NewStageInstance)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *stageInstanceUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, StageInstanceUpdateEvent)))
}

/*****************************
 *   STAGE_INSTANCE_DELETE Handler
 *****************************/

type stageInstanceDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, StageInstanceDeleteEvent)
}

func (h *stageInstanceDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := StageInstanceDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "STAGE_INSTANCE_DELETE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagStageInstances) {
		cache.DelStageInstance(ctx, evt.ChannelID)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *stageInstanceDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, StageInstanceDeleteEvent)))
}

/*****************************
 *   SUBSCRIPTION_CREATE Handler
 *****************************/

type subscriptionCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, SubscriptionCreateEvent)
}

func (h *subscriptionCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := SubscriptionCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "SUBSCRIPTION_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *subscriptionCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, SubscriptionCreateEvent)))
}

/*****************************
 *   SUBSCRIPTION_UPDATE Handler
 *****************************/

type subscriptionUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, SubscriptionUpdateEvent)
}

func (h *subscriptionUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := SubscriptionUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "SUBSCRIPTION_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *subscriptionUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, SubscriptionUpdateEvent)))
}

/*****************************
 *   SUBSCRIPTION_DELETE Handler
 *****************************/

type subscriptionDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, SubscriptionDeleteEvent)
}

func (h *subscriptionDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := SubscriptionDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "SUBSCRIPTION_DELETE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *subscriptionDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, SubscriptionDeleteEvent)))
}

/*****************************
 *   TYPING_START Handler
 *****************************/

type typingStartHandlers struct {
	logger   Logger
	handlers []func(context.Context, TypingStartEvent)
}

func (h *typingStartHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := TypingStartEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "TYPING_START", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *typingStartHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, TypingStartEvent)))
}

/*****************************
 *   USER_UPDATE Handler
 *****************************/

type userUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, UserUpdateEvent)
}

func (h *userUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := UserUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "USER_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if oldUser, ok := cache.GetUser(ctx, evt.NewUser.ID); ok {
			evt.OldUser = oldUser
		}
		if cache.CacheFlags().Has(CacheFlagUsers) {
			cache.AddUser(ctx, evt.NewUser)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *userUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, UserUpdateEvent)))
}

/*****************************
 *   VOICE_CHANNEL_EFFECT_SEND Handler
 *****************************/

type voiceChannelEffectSendHandlers struct {
	logger   Logger
	handlers []func(context.Context, VoiceChannelEffectSendEvent)
}

func (h *voiceChannelEffectSendHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := VoiceChannelEffectSendEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "VOICE_CHANNEL_EFFECT_SEND", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *voiceChannelEffectSendHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, VoiceChannelEffectSendEvent)))
}

/*****************************
 *   VOICE_STATE_UPDATE Handler
 *****************************/

type voiceStateUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, VoiceStateUpdateEvent)
}

func (h *voiceStateUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := VoiceStateUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "VOICE_STATE_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if oldState, ok := cache.GetVoiceState(ctx, evt.NewState.GuildID, evt.NewState.UserID); ok {
			evt.OldState = GatewayVoiceState{VoiceState: oldState}
		}
		if cache.CacheFlags().Has(CacheFlagVoiceStates) {
			if evt.NewState.ChannelID == 0 {
				cache.DelVoiceState(ctx, evt.NewState.GuildID, evt.NewState.UserID)
			} else {
				cache.AddVoiceState(ctx, evt.NewState.VoiceState)
			}
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *voiceStateUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, VoiceStateUpdateEvent)))
}

/*****************************
 *   VOICE_SERVER_UPDATE Handler
 *****************************/

type voiceServerUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, VoiceServerUpdateEvent)
}

func (h *voiceServerUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := VoiceServerUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "VOICE_SERVER_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *voiceServerUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, VoiceServerUpdateEvent)))
}

/*****************************
 *   WEBHOOKS_UPDATE Handler
 *****************************/

type webhooksUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, WebhooksUpdateEvent)
}

func (h *webhooksUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := WebhooksUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "WEBHOOKS_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *webhooksUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, WebhooksUpdateEvent)))
}

/*****************************
 *   APPLICATION_COMMAND_PERMISSIONS_UPDATE Handler
 *****************************/

type applicationCommandPermissionsUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, ApplicationCommandPermissionsUpdateEvent)
}

func (h *applicationCommandPermissionsUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ApplicationCommandPermissionsUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "APPLICATION_COMMAND_PERMISSIONS_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *applicationCommandPermissionsUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ApplicationCommandPermissionsUpdateEvent)))
}

/*****************************
 *   AUTO_MODERATION_RULE_CREATE Handler
 *****************************/

type autoModerationRuleCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, AutoModerationRuleCreateEvent)
}

func (h *autoModerationRuleCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := AutoModerationRuleCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "AUTO_MODERATION_RULE_CREATE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagAutoModerationRules) {
		cache.AddAutoModerationRule(ctx, evt.AutoModerationRule)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *autoModerationRuleCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, AutoModerationRuleCreateEvent)))
}

/*****************************
 *   AUTO_MODERATION_RULE_UPDATE Handler
 *****************************/

type autoModerationRuleUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, AutoModerationRuleUpdateEvent)
}

func (h *autoModerationRuleUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := AutoModerationRuleUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "AUTO_MODERATION_RULE_UPDATE", "error", err)
		return
	}

	if cache != nil {
		if oldRule, ok := cache.GetAutoModerationRule(ctx, evt.NewAutoModerationRule.ID); ok {
			evt.OldAutoModerationRule = oldRule
		}
		if cache.CacheFlags().Has(CacheFlagAutoModerationRules) {
			cache.AddAutoModerationRule(ctx, evt.NewAutoModerationRule)
		}
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *autoModerationRuleUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, AutoModerationRuleUpdateEvent)))
}

/*****************************
 *   AUTO_MODERATION_RULE_DELETE Handler
 *****************************/

type autoModerationRuleDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, AutoModerationRuleDeleteEvent)
}

func (h *autoModerationRuleDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := AutoModerationRuleDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "AUTO_MODERATION_RULE_DELETE", "error", err)
		return
	}

	if cache != nil && cache.CacheFlags().Has(CacheFlagAutoModerationRules) {
		cache.DelAutoModerationRule(ctx, evt.GuildID, evt.ID)
	}

	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *autoModerationRuleDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, AutoModerationRuleDeleteEvent)))
}

/*****************************
 *   AUTO_MODERATION_ACTION_EXECUTION Handler
 *****************************/

type autoModerationActionExecutionHandlers struct {
	logger   Logger
	handlers []func(context.Context, AutoModerationActionExecutionEvent)
}

func (h *autoModerationActionExecutionHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := AutoModerationActionExecutionEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "AUTO_MODERATION_ACTION_EXECUTION", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *autoModerationActionExecutionHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, AutoModerationActionExecutionEvent)))
}

/*****************************
 *   ENTITLEMENT_CREATE Handler
 *****************************/

type entitlementCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, EntitlementCreateEvent)
}

func (h *entitlementCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := EntitlementCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "ENTITLEMENT_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *entitlementCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, EntitlementCreateEvent)))
}

/*****************************
 *   ENTITLEMENT_UPDATE Handler
 *****************************/

type entitlementUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, EntitlementUpdateEvent)
}

func (h *entitlementUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := EntitlementUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "ENTITLEMENT_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *entitlementUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, EntitlementUpdateEvent)))
}

/*****************************
 *   ENTITLEMENT_DELETE Handler
 *****************************/

type entitlementDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, EntitlementDeleteEvent)
}

func (h *entitlementDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := EntitlementDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "ENTITLEMENT_DELETE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *entitlementDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, EntitlementDeleteEvent)))
}

/*****************************
 *   CHANNEL_INFO Handler
 *****************************/

type channelInfoHandlers struct {
	logger   Logger
	handlers []func(context.Context, ChannelInfoEvent)
}

func (h *channelInfoHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ChannelInfoEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "CHANNEL_INFO", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *channelInfoHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ChannelInfoEvent)))
}

/*****************************
 *   VOICE_CHANNEL_STATUS_UPDATE Handler
 *****************************/

type voiceChannelStatusUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, VoiceChannelStatusUpdateEvent)
}

func (h *voiceChannelStatusUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := VoiceChannelStatusUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "VOICE_CHANNEL_STATUS_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *voiceChannelStatusUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, VoiceChannelStatusUpdateEvent)))
}

/*****************************
 *   VOICE_CHANNEL_START_TIME_UPDATE Handler
 *****************************/

type voiceChannelStartTimeUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, VoiceChannelStartTimeUpdateEvent)
}

func (h *voiceChannelStartTimeUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := VoiceChannelStartTimeUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "VOICE_CHANNEL_START_TIME_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *voiceChannelStartTimeUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, VoiceChannelStartTimeUpdateEvent)))
}

/*****************************
 *   APPLICATION_AUTHORIZED Handler
 *****************************/

type applicationAuthorizedHandlers struct {
	logger   Logger
	handlers []func(context.Context, ApplicationAuthorizedEvent)
}

func (h *applicationAuthorizedHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ApplicationAuthorizedEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "APPLICATION_AUTHORIZED", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *applicationAuthorizedHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ApplicationAuthorizedEvent)))
}

/*****************************
 *   APPLICATION_DEAUTHORIZED Handler
 *****************************/

type applicationDeauthorizedHandlers struct {
	logger   Logger
	handlers []func(context.Context, ApplicationDeauthorizedEvent)
}

func (h *applicationDeauthorizedHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := ApplicationDeauthorizedEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "APPLICATION_DEAUTHORIZED", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *applicationDeauthorizedHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, ApplicationDeauthorizedEvent)))
}

/*****************************
 *   RATE_LIMITED Handler
 *****************************/

type rateLimitedHandlers struct {
	logger   Logger
	handlers []func(context.Context, RateLimitedEvent)
}

func (h *rateLimitedHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := RateLimitedEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "RATE_LIMITED", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *rateLimitedHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, RateLimitedEvent)))
}

/*****************************
 *   QUEST_USER_ENROLLMENT Handler
 *****************************/

type questUserEnrollmentHandlers struct {
	logger   Logger
	handlers []func(context.Context, QuestUserEnrollmentEvent)
}

func (h *questUserEnrollmentHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := QuestUserEnrollmentEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "QUEST_USER_ENROLLMENT", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *questUserEnrollmentHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, QuestUserEnrollmentEvent)))
}

/*****************************
 *   LOBBY_MESSAGE_CREATE Handler
 *****************************/

type lobbyMessageCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, LobbyMessageCreateEvent)
}

func (h *lobbyMessageCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := LobbyMessageCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "LOBBY_MESSAGE_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *lobbyMessageCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, LobbyMessageCreateEvent)))
}

/*****************************
 *   LOBBY_MESSAGE_UPDATE Handler
 *****************************/

type lobbyMessageUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, LobbyMessageUpdateEvent)
}

func (h *lobbyMessageUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := LobbyMessageUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "LOBBY_MESSAGE_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *lobbyMessageUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, LobbyMessageUpdateEvent)))
}

/*****************************
 *   LOBBY_MESSAGE_DELETE Handler
 *****************************/

type lobbyMessageDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, LobbyMessageDeleteEvent)
}

func (h *lobbyMessageDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := LobbyMessageDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "LOBBY_MESSAGE_DELETE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *lobbyMessageDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, LobbyMessageDeleteEvent)))
}

/*****************************
 *   GAME_DIRECT_MESSAGE_CREATE Handler
 *****************************/

type gameDirectMessageCreateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GameDirectMessageCreateEvent)
}

func (h *gameDirectMessageCreateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GameDirectMessageCreateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GAME_DIRECT_MESSAGE_CREATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *gameDirectMessageCreateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GameDirectMessageCreateEvent)))
}

/*****************************
 *   GAME_DIRECT_MESSAGE_UPDATE Handler
 *****************************/

type gameDirectMessageUpdateHandlers struct {
	logger   Logger
	handlers []func(context.Context, GameDirectMessageUpdateEvent)
}

func (h *gameDirectMessageUpdateHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GameDirectMessageUpdateEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GAME_DIRECT_MESSAGE_UPDATE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *gameDirectMessageUpdateHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GameDirectMessageUpdateEvent)))
}

/*****************************
 *   GAME_DIRECT_MESSAGE_DELETE Handler
 *****************************/

type gameDirectMessageDeleteHandlers struct {
	logger   Logger
	handlers []func(context.Context, GameDirectMessageDeleteEvent)
}

func (h *gameDirectMessageDeleteHandlers) handleEvent(ctx context.Context, cache CacheManager, concurrent bool, shardID ShardID, data []byte) {
	evt := GameDirectMessageDeleteEvent{ShardID: shardID}
	if err := json.Unmarshal(data, &evt); err != nil {
		h.logger.Error("failed parsing event data", "event", "GAME_DIRECT_MESSAGE_DELETE", "error", err)
		return
	}
	dispatchHandlers(ctx, h.handlers, evt, concurrent)
}

func (h *gameDirectMessageDeleteHandlers) addHandler(handler any) {
	h.handlers = append(h.handlers, handler.(func(context.Context, GameDirectMessageDeleteEvent)))
}
