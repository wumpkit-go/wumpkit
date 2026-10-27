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
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// GatewayEvent represents the name of a Discord Gateway event.
type GatewayEvent string

const (
	GatewayEventReady                               GatewayEvent = "READY"
	GatewayEventResumed                             GatewayEvent = "RESUMED"
	GatewayEventGuildCreate                         GatewayEvent = "GUILD_CREATE"
	GatewayEventGuildUpdate                         GatewayEvent = "GUILD_UPDATE"
	GatewayEventGuildDelete                         GatewayEvent = "GUILD_DELETE"
	GatewayEventChannelCreate                       GatewayEvent = "CHANNEL_CREATE"
	GatewayEventChannelUpdate                       GatewayEvent = "CHANNEL_UPDATE"
	GatewayEventChannelDelete                       GatewayEvent = "CHANNEL_DELETE"
	GatewayEventChannelPinsUpdate                   GatewayEvent = "CHANNEL_PINS_UPDATE"
	GatewayEventThreadCreate                        GatewayEvent = "THREAD_CREATE"
	GatewayEventThreadUpdate                        GatewayEvent = "THREAD_UPDATE"
	GatewayEventThreadDelete                        GatewayEvent = "THREAD_DELETE"
	GatewayEventThreadListSync                      GatewayEvent = "THREAD_LIST_SYNC"
	GatewayEventThreadMemberUpdate                  GatewayEvent = "THREAD_MEMBER_UPDATE"
	GatewayEventThreadMembersUpdate                 GatewayEvent = "THREAD_MEMBERS_UPDATE"
	GatewayEventGuildMemberAdd                      GatewayEvent = "GUILD_MEMBER_ADD"
	GatewayEventGuildMemberUpdate                   GatewayEvent = "GUILD_MEMBER_UPDATE"
	GatewayEventGuildMemberRemove                   GatewayEvent = "GUILD_MEMBER_REMOVE"
	GatewayEventGuildMembersChunk                   GatewayEvent = "GUILD_MEMBERS_CHUNK"
	GatewayEventGuildRoleCreate                     GatewayEvent = "GUILD_ROLE_CREATE"
	GatewayEventGuildRoleUpdate                     GatewayEvent = "GUILD_ROLE_UPDATE"
	GatewayEventGuildRoleDelete                     GatewayEvent = "GUILD_ROLE_DELETE"
	GatewayEventGuildEmojisUpdate                   GatewayEvent = "GUILD_EMOJIS_UPDATE"
	GatewayEventGuildStickersUpdate                 GatewayEvent = "GUILD_STICKERS_UPDATE"
	GatewayEventGuildIntegrationsUpdate             GatewayEvent = "GUILD_INTEGRATIONS_UPDATE"
	GatewayEventGuildAuditLogEntryCreate            GatewayEvent = "GUILD_AUDIT_LOG_ENTRY_CREATE"
	GatewayEventGuildBanAdd                         GatewayEvent = "GUILD_BAN_ADD"
	GatewayEventGuildBanRemove                      GatewayEvent = "GUILD_BAN_REMOVE"
	GatewayEventGuildScheduledEventCreate           GatewayEvent = "GUILD_SCHEDULED_EVENT_CREATE"
	GatewayEventGuildScheduledEventUpdate           GatewayEvent = "GUILD_SCHEDULED_EVENT_UPDATE"
	GatewayEventGuildScheduledEventDelete           GatewayEvent = "GUILD_SCHEDULED_EVENT_DELETE"
	GatewayEventGuildScheduledEventUserAdd          GatewayEvent = "GUILD_SCHEDULED_EVENT_USER_ADD"
	GatewayEventGuildScheduledEventUserRemove       GatewayEvent = "GUILD_SCHEDULED_EVENT_USER_REMOVE"
	GatewayEventGuildSoundboardSoundCreate          GatewayEvent = "GUILD_SOUNDBOARD_SOUND_CREATE"
	GatewayEventGuildSoundboardSoundUpdate          GatewayEvent = "GUILD_SOUNDBOARD_SOUND_UPDATE"
	GatewayEventGuildSoundboardSoundDelete          GatewayEvent = "GUILD_SOUNDBOARD_SOUND_DELETE"
	GatewayEventGuildSoundboardSoundsUpdate         GatewayEvent = "GUILD_SOUNDBOARD_SOUNDS_UPDATE"
	GatewayEventSoundboardSounds                    GatewayEvent = "SOUNDBOARD_SOUNDS"
	GatewayEventIntegrationCreate                   GatewayEvent = "INTEGRATION_CREATE"
	GatewayEventIntegrationUpdate                   GatewayEvent = "INTEGRATION_UPDATE"
	GatewayEventIntegrationDelete                   GatewayEvent = "INTEGRATION_DELETE"
	GatewayEventInteractionCreate                   GatewayEvent = "INTERACTION_CREATE"
	GatewayEventInviteCreate                        GatewayEvent = "INVITE_CREATE"
	GatewayEventInviteDelete                        GatewayEvent = "INVITE_DELETE"
	GatewayEventMessageCreate                       GatewayEvent = "MESSAGE_CREATE"
	GatewayEventMessageUpdate                       GatewayEvent = "MESSAGE_UPDATE"
	GatewayEventMessageDelete                       GatewayEvent = "MESSAGE_DELETE"
	GatewayEventMessageDeleteBulk                   GatewayEvent = "MESSAGE_DELETE_BULK"
	GatewayEventMessageReactionAdd                  GatewayEvent = "MESSAGE_REACTION_ADD"
	GatewayEventMessageReactionRemove               GatewayEvent = "MESSAGE_REACTION_REMOVE"
	GatewayEventMessageReactionRemoveAll            GatewayEvent = "MESSAGE_REACTION_REMOVE_ALL"
	GatewayEventMessageReactionRemoveEmoji          GatewayEvent = "MESSAGE_REACTION_REMOVE_EMOJI"
	GatewayEventMessagePollVoteAdd                  GatewayEvent = "MESSAGE_POLL_VOTE_ADD"
	GatewayEventMessagePollVoteRemove               GatewayEvent = "MESSAGE_POLL_VOTE_REMOVE"
	GatewayEventPresenceUpdate                      GatewayEvent = "PRESENCE_UPDATE"
	GatewayEventStageInstanceCreate                 GatewayEvent = "STAGE_INSTANCE_CREATE"
	GatewayEventStageInstanceUpdate                 GatewayEvent = "STAGE_INSTANCE_UPDATE"
	GatewayEventStageInstanceDelete                 GatewayEvent = "STAGE_INSTANCE_DELETE"
	GatewayEventSubscriptionCreate                  GatewayEvent = "SUBSCRIPTION_CREATE"
	GatewayEventSubscriptionUpdate                  GatewayEvent = "SUBSCRIPTION_UPDATE"
	GatewayEventSubscriptionDelete                  GatewayEvent = "SUBSCRIPTION_DELETE"
	GatewayEventTypingStart                         GatewayEvent = "TYPING_START"
	GatewayEventUserUpdate                          GatewayEvent = "USER_UPDATE"
	GatewayEventVoiceChannelEffectSend              GatewayEvent = "VOICE_CHANNEL_EFFECT_SEND"
	GatewayEventVoiceStateUpdate                    GatewayEvent = "VOICE_STATE_UPDATE"
	GatewayEventVoiceServerUpdate                   GatewayEvent = "VOICE_SERVER_UPDATE"
	GatewayEventWebhooksUpdate                      GatewayEvent = "WEBHOOKS_UPDATE"
	GatewayEventApplicationCommandPermissionsUpdate GatewayEvent = "APPLICATION_COMMAND_PERMISSIONS_UPDATE"
	GatewayEventAutoModerationRuleCreate            GatewayEvent = "AUTO_MODERATION_RULE_CREATE"
	GatewayEventAutoModerationRuleUpdate            GatewayEvent = "AUTO_MODERATION_RULE_UPDATE"
	GatewayEventAutoModerationRuleDelete            GatewayEvent = "AUTO_MODERATION_RULE_DELETE"
	GatewayEventAutoModerationActionExecution       GatewayEvent = "AUTO_MODERATION_ACTION_EXECUTION"
	GatewayEventEntitlementCreate                   GatewayEvent = "ENTITLEMENT_CREATE"
	GatewayEventEntitlementUpdate                   GatewayEvent = "ENTITLEMENT_UPDATE"
	GatewayEventEntitlementDelete                   GatewayEvent = "ENTITLEMENT_DELETE"
	GatewayEventChannelInfo                         GatewayEvent = "CHANNEL_INFO"
	GatewayEventVoiceChannelStatusUpdate            GatewayEvent = "VOICE_CHANNEL_STATUS_UPDATE"
	GatewayEventVoiceChannelStartTimeUpdate         GatewayEvent = "VOICE_CHANNEL_START_TIME_UPDATE"
	GatewayEventApplicationAuthorized               GatewayEvent = "APPLICATION_AUTHORIZED"
	GatewayEventApplicationDeauthorized             GatewayEvent = "APPLICATION_DEAUTHORIZED"
	GatewayEventRateLimited                         GatewayEvent = "RATE_LIMITED"
	GatewayEventQuestUserEnrollment                 GatewayEvent = "QUEST_USER_ENROLLMENT"
	GatewayEventLobbyMessageCreate                  GatewayEvent = "LOBBY_MESSAGE_CREATE"
	GatewayEventLobbyMessageUpdate                  GatewayEvent = "LOBBY_MESSAGE_UPDATE"
	GatewayEventLobbyMessageDelete                  GatewayEvent = "LOBBY_MESSAGE_DELETE"
	GatewayEventGameDirectMessageCreate             GatewayEvent = "GAME_DIRECT_MESSAGE_CREATE"
	GatewayEventGameDirectMessageUpdate             GatewayEvent = "GAME_DIRECT_MESSAGE_UPDATE"
	GatewayEventGameDirectMessageDelete             GatewayEvent = "GAME_DIRECT_MESSAGE_DELETE"
)

// Is returns true if this GatewayEvent matches the provided one.
func (e GatewayEvent) Is(other GatewayEvent) bool {
	return e == other
}

// eventhandlersManager defines the interface for managing event handlers of a specific event type.
type eventhandlersManager interface {
	handleEvent(
		ctx context.Context,
		cache CacheManager,
		concurrent bool,
		shardID ShardID,
		buf []byte,
	)
	addHandler(handler any)
}

// HandlersExecutionMode defines how event handlers are executed.
type HandlersExecutionMode int

const (
	HandlersExecutionSequential HandlersExecutionMode = iota
	HandlersExecutionConcurrent
)

// Is returns true if this HandlersExecutionMode matches the provided one.
func (v HandlersExecutionMode) Is(other HandlersExecutionMode) bool {
	return v == other
}

// DispatcherConfig holds configuration parameters for the event dispatcher.
type DispatcherConfig struct {
	// ExecutionMode defines whether event handlers are executed sequentially or concurrently.
	ExecutionMode HandlersExecutionMode
}

// DefaultDispatcherConfig returns standard dispatcher configuration defaults.
func DefaultDispatcherConfig() DispatcherConfig {
	return DispatcherConfig{
		ExecutionMode: HandlersExecutionSequential,
	}
}

// eventTask represents a queued gateway dispatch payload
type eventTask struct {
	shardID   ShardID
	eventName string
	data      []byte
}

// dispatcher manages registration of event handlers and dispatching of events.
type dispatcher struct {
	ctx                   context.Context
	cache                 CacheManager
	logger                Logger
	handlersManagers      map[GatewayEvent]eventhandlersManager
	handlersExecutionMode HandlersExecutionMode
	wg                    sync.WaitGroup
	closed                atomic.Bool
	stopCh                chan struct{}
	queuesMu              sync.Mutex
	queues                map[ShardID]chan eventTask
}

func newDispatcher(
	ctx context.Context,
	cache CacheManager,
	logger Logger,
	mode HandlersExecutionMode,
) *dispatcher {
	d := &dispatcher{
		ctx:                   ctx,
		cache:                 cache,
		logger:                logger,
		handlersExecutionMode: mode,
		handlersManagers:      make(map[GatewayEvent]eventhandlersManager, 80),
		queues:                make(map[ShardID]chan eventTask),
		stopCh:                make(chan struct{}),
	}

	// register all handler managers
	d.handlersManagers[GatewayEventReady] = &readyHandlers{logger: logger}
	d.handlersManagers[GatewayEventResumed] = &resumedHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildCreate] = &guildCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildUpdate] = &guildUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildDelete] = &guildDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventChannelCreate] = &channelCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventChannelUpdate] = &channelUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventChannelDelete] = &channelDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventChannelPinsUpdate] = &channelPinsUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventThreadCreate] = &threadCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventThreadUpdate] = &threadUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventThreadDelete] = &threadDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventThreadListSync] = &threadListSyncHandlers{logger: logger}
	d.handlersManagers[GatewayEventThreadMemberUpdate] = &threadMemberUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventThreadMembersUpdate] = &threadMembersUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildMemberAdd] = &guildMemberAddHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildMemberUpdate] = &guildMemberUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildMemberRemove] = &guildMemberRemoveHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildMembersChunk] = &guildMembersChunkHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildRoleCreate] = &guildRoleCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildRoleUpdate] = &guildRoleUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildRoleDelete] = &guildRoleDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildEmojisUpdate] = &guildEmojisUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildStickersUpdate] = &guildStickersUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildIntegrationsUpdate] = &guildIntegrationsUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildAuditLogEntryCreate] = &guildAuditLogEntryCreateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildBanAdd] = &guildBanAddHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildBanRemove] = &guildBanRemoveHandlers{logger: logger}
	d.handlersManagers[GatewayEventGuildScheduledEventCreate] = &guildScheduledEventCreateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildScheduledEventUpdate] = &guildScheduledEventUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildScheduledEventDelete] = &guildScheduledEventDeleteHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildScheduledEventUserAdd] = &guildScheduledEventUserAddHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildScheduledEventUserRemove] = &guildScheduledEventUserRemoveHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildSoundboardSoundCreate] = &guildSoundboardSoundCreateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildSoundboardSoundUpdate] = &guildSoundboardSoundUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildSoundboardSoundDelete] = &guildSoundboardSoundDeleteHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGuildSoundboardSoundsUpdate] = &guildSoundboardSoundsUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventSoundboardSounds] = &soundboardSoundsHandlers{logger: logger}
	d.handlersManagers[GatewayEventIntegrationCreate] = &integrationCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventIntegrationUpdate] = &integrationUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventIntegrationDelete] = &integrationDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventInteractionCreate] = &interactionCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventInviteCreate] = &inviteCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventInviteDelete] = &inviteDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventMessageCreate] = &messageCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventMessageUpdate] = &messageUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventMessageDelete] = &messageDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventMessageDeleteBulk] = &messageDeleteBulkHandlers{logger: logger}
	d.handlersManagers[GatewayEventMessageReactionAdd] = &messageReactionAddHandlers{logger: logger}
	d.handlersManagers[GatewayEventMessageReactionRemove] = &messageReactionRemoveHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventMessageReactionRemoveAll] = &messageReactionRemoveAllHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventMessageReactionRemoveEmoji] = &messageReactionRemoveEmojiHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventMessagePollVoteAdd] = &messagePollVoteAddHandlers{logger: logger}
	d.handlersManagers[GatewayEventMessagePollVoteRemove] = &messagePollVoteRemoveHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventPresenceUpdate] = &presenceUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventStageInstanceCreate] = &stageInstanceCreateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventStageInstanceUpdate] = &stageInstanceUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventStageInstanceDelete] = &stageInstanceDeleteHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventSubscriptionCreate] = &subscriptionCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventSubscriptionUpdate] = &subscriptionUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventSubscriptionDelete] = &subscriptionDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventTypingStart] = &typingStartHandlers{logger: logger}
	d.handlersManagers[GatewayEventUserUpdate] = &userUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventVoiceChannelEffectSend] = &voiceChannelEffectSendHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventVoiceStateUpdate] = &voiceStateUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventVoiceServerUpdate] = &voiceServerUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventWebhooksUpdate] = &webhooksUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventApplicationCommandPermissionsUpdate] = &applicationCommandPermissionsUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventAutoModerationRuleCreate] = &autoModerationRuleCreateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventAutoModerationRuleUpdate] = &autoModerationRuleUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventAutoModerationRuleDelete] = &autoModerationRuleDeleteHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventAutoModerationActionExecution] = &autoModerationActionExecutionHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventEntitlementCreate] = &entitlementCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventEntitlementUpdate] = &entitlementUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventEntitlementDelete] = &entitlementDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventChannelInfo] = &channelInfoHandlers{logger: logger}
	d.handlersManagers[GatewayEventVoiceChannelStatusUpdate] = &voiceChannelStatusUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventVoiceChannelStartTimeUpdate] = &voiceChannelStartTimeUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventApplicationAuthorized] = &applicationAuthorizedHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventApplicationDeauthorized] = &applicationDeauthorizedHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventRateLimited] = &rateLimitedHandlers{logger: logger}
	d.handlersManagers[GatewayEventQuestUserEnrollment] = &questUserEnrollmentHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventLobbyMessageCreate] = &lobbyMessageCreateHandlers{logger: logger}
	d.handlersManagers[GatewayEventLobbyMessageUpdate] = &lobbyMessageUpdateHandlers{logger: logger}
	d.handlersManagers[GatewayEventLobbyMessageDelete] = &lobbyMessageDeleteHandlers{logger: logger}
	d.handlersManagers[GatewayEventGameDirectMessageCreate] = &gameDirectMessageCreateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGameDirectMessageUpdate] = &gameDirectMessageUpdateHandlers{
		logger: logger,
	}
	d.handlersManagers[GatewayEventGameDirectMessageDelete] = &gameDirectMessageDeleteHandlers{
		logger: logger,
	}

	return d
}

func (d *dispatcher) Dispatch(shardID ShardID, eventName string, data []byte) {
	if d.closed.Load() {
		return
	}

	d.queuesMu.Lock()
	q, ok := d.queues[shardID]
	if !ok {
		if d.closed.Load() {
			d.queuesMu.Unlock()
			return
		}
		q = make(chan eventTask, 2048)
		d.queues[shardID] = q
		d.wg.Add(1)
		go d.runWorker(q)
	}
	d.queuesMu.Unlock()

	d.logger.Debug("event dispatched", "shard_id", shardID, "event", eventName)

	select {
	case q <- eventTask{shardID: shardID, eventName: eventName, data: data}:
	default:
		// queue overflow fallback to avoid blocking websocket read loop
		go func() {
			defer func() {
				_ = recover()
			}()
			if d.closed.Load() {
				return
			}
			select {
			case q <- eventTask{shardID: shardID, eventName: eventName, data: data}:
			case <-d.stopCh:
			case <-d.ctx.Done():
			}
		}()
	}
}

func (d *dispatcher) runWorker(ch chan eventTask) {
	defer d.wg.Done()
	for task := range ch {
		d.processEvent(task.shardID, task.eventName, task.data)
	}
}

func (d *dispatcher) processEvent(shardID ShardID, eventName string, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			if tmErr, ok := r.(*TypeMismatchError); ok {
				panic(tmErr)
			}
			d.logger.Error(
				"Recovered from panic while handling event",
				"event", eventName,
				"shard_id", shardID,
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()

	if hm, ok := d.handlersManagers[GatewayEvent(eventName)]; ok {
		hm.handleEvent(
			d.ctx,
			d.cache,
			d.handlersExecutionMode == HandlersExecutionConcurrent,
			shardID,
			data,
		)
	}
}

// Shutdown stops dispatching new events and gracefully waits for in-flight event handlers to complete.
func (d *dispatcher) Shutdown(timeout time.Duration) {
	if !d.closed.CompareAndSwap(false, true) {
		return
	}
	close(d.stopCh)

	d.queuesMu.Lock()
	for _, q := range d.queues {
		close(q)
	}
	d.queues = make(map[ShardID]chan eventTask)
	d.queuesMu.Unlock()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		d.logger.Debug("all in-flight event handlers drained successfully")
	case <-time.After(timeout):
		d.logger.Warn(
			"timed out waiting for in-flight event handlers to drain",
			"timeout",
			timeout.String(),
		)
	}
}

func (d *dispatcher) register(event GatewayEvent, h any) {
	d.logger.Debug("handler registered", "event", string(event))
	if hm, ok := d.handlersManagers[event]; ok {
		hm.addHandler(h)
	}
}

// OnReady registers a handler for 'READY' (GatewayEventReady) events.
func (d *dispatcher) OnReady(h func(context.Context, ReadyEvent)) {
	d.register(GatewayEventReady, h)
}

// OnResumed registers a handler for 'RESUMED' (GatewayEventResumed) events.
func (d *dispatcher) OnResumed(h func(context.Context, ResumedEvent)) {
	d.register(GatewayEventResumed, h)
}

// OnGuildCreate registers a handler for 'GUILD_CREATE' (GatewayEventGuildCreate) events.
func (d *dispatcher) OnGuildCreate(h func(context.Context, GuildCreateEvent)) {
	d.register(GatewayEventGuildCreate, h)
}

// OnGuildUpdate registers a handler for 'GUILD_UPDATE' (GatewayEventGuildUpdate) events.
func (d *dispatcher) OnGuildUpdate(h func(context.Context, GuildUpdateEvent)) {
	d.register(GatewayEventGuildUpdate, h)
}

// OnGuildDelete registers a handler for 'GUILD_DELETE' (GatewayEventGuildDelete) events.
func (d *dispatcher) OnGuildDelete(h func(context.Context, GuildDeleteEvent)) {
	d.register(GatewayEventGuildDelete, h)
}

// OnChannelCreate registers a handler for 'CHANNEL_CREATE' (GatewayEventChannelCreate) events.
func (d *dispatcher) OnChannelCreate(h func(context.Context, ChannelCreateEvent)) {
	d.register(GatewayEventChannelCreate, h)
}

// OnChannelUpdate registers a handler for 'CHANNEL_UPDATE' (GatewayEventChannelUpdate) events.
func (d *dispatcher) OnChannelUpdate(h func(context.Context, ChannelUpdateEvent)) {
	d.register(GatewayEventChannelUpdate, h)
}

// OnChannelDelete registers a handler for 'CHANNEL_DELETE' (GatewayEventChannelDelete) events.
func (d *dispatcher) OnChannelDelete(h func(context.Context, ChannelDeleteEvent)) {
	d.register(GatewayEventChannelDelete, h)
}

// OnChannelPinsUpdate registers a handler for 'CHANNEL_PINS_UPDATE' (GatewayEventChannelPinsUpdate) events.
func (d *dispatcher) OnChannelPinsUpdate(h func(context.Context, ChannelPinsUpdateEvent)) {
	d.register(GatewayEventChannelPinsUpdate, h)
}

// OnThreadCreate registers a handler for 'THREAD_CREATE' (GatewayEventThreadCreate) events.
func (d *dispatcher) OnThreadCreate(h func(context.Context, ThreadCreateEvent)) {
	d.register(GatewayEventThreadCreate, h)
}

// OnThreadUpdate registers a handler for 'THREAD_UPDATE' (GatewayEventThreadUpdate) events.
func (d *dispatcher) OnThreadUpdate(h func(context.Context, ThreadUpdateEvent)) {
	d.register(GatewayEventThreadUpdate, h)
}

// OnThreadDelete registers a handler for 'THREAD_DELETE' (GatewayEventThreadDelete) events.
func (d *dispatcher) OnThreadDelete(h func(context.Context, ThreadDeleteEvent)) {
	d.register(GatewayEventThreadDelete, h)
}

// OnThreadListSync registers a handler for 'THREAD_LIST_SYNC' (GatewayEventThreadListSync) events.
func (d *dispatcher) OnThreadListSync(h func(context.Context, ThreadListSyncEvent)) {
	d.register(GatewayEventThreadListSync, h)
}

// OnThreadMemberUpdate registers a handler for 'THREAD_MEMBER_UPDATE' (GatewayEventThreadMemberUpdate) events.
func (d *dispatcher) OnThreadMemberUpdate(h func(context.Context, ThreadMemberUpdateEvent)) {
	d.register(GatewayEventThreadMemberUpdate, h)
}

// OnThreadMembersUpdate registers a handler for 'THREAD_MEMBERS_UPDATE' (GatewayEventThreadMembersUpdate) events.
func (d *dispatcher) OnThreadMembersUpdate(h func(context.Context, ThreadMembersUpdateEvent)) {
	d.register(GatewayEventThreadMembersUpdate, h)
}

// OnGuildMemberAdd registers a handler for 'GUILD_MEMBER_ADD' (GatewayEventGuildMemberAdd) events.
func (d *dispatcher) OnGuildMemberAdd(h func(context.Context, GuildMemberAddEvent)) {
	d.register(GatewayEventGuildMemberAdd, h)
}

// OnGuildMemberUpdate registers a handler for 'GUILD_MEMBER_UPDATE' (GatewayEventGuildMemberUpdate) events.
func (d *dispatcher) OnGuildMemberUpdate(h func(context.Context, GuildMemberUpdateEvent)) {
	d.register(GatewayEventGuildMemberUpdate, h)
}

// OnGuildMemberRemove registers a handler for 'GUILD_MEMBER_REMOVE' (GatewayEventGuildMemberRemove) events.
func (d *dispatcher) OnGuildMemberRemove(h func(context.Context, GuildMemberRemoveEvent)) {
	d.register(GatewayEventGuildMemberRemove, h)
}

// OnGuildMembersChunk registers a handler for 'GUILD_MEMBERS_CHUNK' (GatewayEventGuildMembersChunk) events.
func (d *dispatcher) OnGuildMembersChunk(h func(context.Context, GuildMembersChunkEvent)) {
	d.register(GatewayEventGuildMembersChunk, h)
}

// OnGuildRoleCreate registers a handler for 'GUILD_ROLE_CREATE' (GatewayEventGuildRoleCreate) events.
func (d *dispatcher) OnGuildRoleCreate(h func(context.Context, GuildRoleCreateEvent)) {
	d.register(GatewayEventGuildRoleCreate, h)
}

// OnGuildRoleUpdate registers a handler for 'GUILD_ROLE_UPDATE' (GatewayEventGuildRoleUpdate) events.
func (d *dispatcher) OnGuildRoleUpdate(h func(context.Context, GuildRoleUpdateEvent)) {
	d.register(GatewayEventGuildRoleUpdate, h)
}

// OnGuildRoleDelete registers a handler for 'GUILD_ROLE_DELETE' (GatewayEventGuildRoleDelete) events.
func (d *dispatcher) OnGuildRoleDelete(h func(context.Context, GuildRoleDeleteEvent)) {
	d.register(GatewayEventGuildRoleDelete, h)
}

// OnGuildEmojisUpdate registers a handler for 'GUILD_EMOJIS_UPDATE' (GatewayEventGuildEmojisUpdate) events.
func (d *dispatcher) OnGuildEmojisUpdate(h func(context.Context, GuildEmojisUpdateEvent)) {
	d.register(GatewayEventGuildEmojisUpdate, h)
}

// OnGuildStickersUpdate registers a handler for 'GUILD_STICKERS_UPDATE' (GatewayEventGuildStickersUpdate) events.
func (d *dispatcher) OnGuildStickersUpdate(h func(context.Context, GuildStickersUpdateEvent)) {
	d.register(GatewayEventGuildStickersUpdate, h)
}

// OnGuildIntegrationsUpdate registers a handler for 'GUILD_INTEGRATIONS_UPDATE' (GatewayEventGuildIntegrationsUpdate) events.
func (d *dispatcher) OnGuildIntegrationsUpdate(
	h func(context.Context, GuildIntegrationsUpdateEvent),
) {
	d.register(GatewayEventGuildIntegrationsUpdate, h)
}

// OnGuildAuditLogEntryCreate registers a handler for 'GUILD_AUDIT_LOG_ENTRY_CREATE' (GatewayEventGuildAuditLogEntryCreate) events.
func (d *dispatcher) OnGuildAuditLogEntryCreate(
	h func(context.Context, GuildAuditLogEntryCreateEvent),
) {
	d.register(GatewayEventGuildAuditLogEntryCreate, h)
}

// OnGuildBanAdd registers a handler for 'GUILD_BAN_ADD' (GatewayEventGuildBanAdd) events.
func (d *dispatcher) OnGuildBanAdd(h func(context.Context, GuildBanAddEvent)) {
	d.register(GatewayEventGuildBanAdd, h)
}

// OnGuildBanRemove registers a handler for 'GUILD_BAN_REMOVE' (GatewayEventGuildBanRemove) events.
func (d *dispatcher) OnGuildBanRemove(h func(context.Context, GuildBanRemoveEvent)) {
	d.register(GatewayEventGuildBanRemove, h)
}

// OnGuildScheduledEventCreate registers a handler for 'GUILD_SCHEDULED_EVENT_CREATE' (GatewayEventGuildScheduledEventCreate) events.
func (d *dispatcher) OnGuildScheduledEventCreate(
	h func(context.Context, GuildScheduledEventCreateEvent),
) {
	d.register(GatewayEventGuildScheduledEventCreate, h)
}

// OnGuildScheduledEventUpdate registers a handler for 'GUILD_SCHEDULED_EVENT_UPDATE' (GatewayEventGuildScheduledEventUpdate) events.
func (d *dispatcher) OnGuildScheduledEventUpdate(
	h func(context.Context, GuildScheduledEventUpdateEvent),
) {
	d.register(GatewayEventGuildScheduledEventUpdate, h)
}

// OnGuildScheduledEventDelete registers a handler for 'GUILD_SCHEDULED_EVENT_DELETE' (GatewayEventGuildScheduledEventDelete) events.
func (d *dispatcher) OnGuildScheduledEventDelete(
	h func(context.Context, GuildScheduledEventDeleteEvent),
) {
	d.register(GatewayEventGuildScheduledEventDelete, h)
}

// OnGuildScheduledEventUserAdd registers a handler for 'GUILD_SCHEDULED_EVENT_USER_ADD' (GatewayEventGuildScheduledEventUserAdd) events.
func (d *dispatcher) OnGuildScheduledEventUserAdd(
	h func(context.Context, GuildScheduledEventUserAddEvent),
) {
	d.register(GatewayEventGuildScheduledEventUserAdd, h)
}

// OnGuildScheduledEventUserRemove registers a handler for 'GUILD_SCHEDULED_EVENT_USER_REMOVE' (GatewayEventGuildScheduledEventUserRemove) events.
func (d *dispatcher) OnGuildScheduledEventUserRemove(
	h func(context.Context, GuildScheduledEventUserRemoveEvent),
) {
	d.register(GatewayEventGuildScheduledEventUserRemove, h)
}

// OnGuildSoundboardSoundCreate registers a handler for 'GUILD_SOUNDBOARD_SOUND_CREATE' (GatewayEventGuildSoundboardSoundCreate) events.
func (d *dispatcher) OnGuildSoundboardSoundCreate(
	h func(context.Context, GuildSoundboardSoundCreateEvent),
) {
	d.register(GatewayEventGuildSoundboardSoundCreate, h)
}

// OnGuildSoundboardSoundUpdate registers a handler for 'GUILD_SOUNDBOARD_SOUND_UPDATE' (GatewayEventGuildSoundboardSoundUpdate) events.
func (d *dispatcher) OnGuildSoundboardSoundUpdate(
	h func(context.Context, GuildSoundboardSoundUpdateEvent),
) {
	d.register(GatewayEventGuildSoundboardSoundUpdate, h)
}

// OnGuildSoundboardSoundDelete registers a handler for 'GUILD_SOUNDBOARD_SOUND_DELETE' (GatewayEventGuildSoundboardSoundDelete) events.
func (d *dispatcher) OnGuildSoundboardSoundDelete(
	h func(context.Context, GuildSoundboardSoundDeleteEvent),
) {
	d.register(GatewayEventGuildSoundboardSoundDelete, h)
}

// OnGuildSoundboardSoundsUpdate registers a handler for 'GUILD_SOUNDBOARD_SOUNDS_UPDATE' (GatewayEventGuildSoundboardSoundsUpdate) events.
func (d *dispatcher) OnGuildSoundboardSoundsUpdate(
	h func(context.Context, GuildSoundboardSoundsUpdateEvent),
) {
	d.register(GatewayEventGuildSoundboardSoundsUpdate, h)
}

// OnSoundboardSounds registers a handler for 'SOUNDBOARD_SOUNDS' (GatewayEventSoundboardSounds) events.
func (d *dispatcher) OnSoundboardSounds(h func(context.Context, SoundboardSoundsEvent)) {
	d.register(GatewayEventSoundboardSounds, h)
}

// OnIntegrationCreate registers a handler for 'INTEGRATION_CREATE' (GatewayEventIntegrationCreate) events.
func (d *dispatcher) OnIntegrationCreate(h func(context.Context, IntegrationCreateEvent)) {
	d.register(GatewayEventIntegrationCreate, h)
}

// OnIntegrationUpdate registers a handler for 'INTEGRATION_UPDATE' (GatewayEventIntegrationUpdate) events.
func (d *dispatcher) OnIntegrationUpdate(h func(context.Context, IntegrationUpdateEvent)) {
	d.register(GatewayEventIntegrationUpdate, h)
}

// OnIntegrationDelete registers a handler for 'INTEGRATION_DELETE' (GatewayEventIntegrationDelete) events.
func (d *dispatcher) OnIntegrationDelete(h func(context.Context, IntegrationDeleteEvent)) {
	d.register(GatewayEventIntegrationDelete, h)
}

// OnInteractionCreate registers a handler for 'INTERACTION_CREATE' (GatewayEventInteractionCreate) events.
func (d *dispatcher) OnInteractionCreate(h func(context.Context, InteractionCreateEvent)) {
	d.register(GatewayEventInteractionCreate, h)
}

// OnInviteCreate registers a handler for 'INVITE_CREATE' (GatewayEventInviteCreate) events.
func (d *dispatcher) OnInviteCreate(h func(context.Context, InviteCreateEvent)) {
	d.register(GatewayEventInviteCreate, h)
}

// OnInviteDelete registers a handler for 'INVITE_DELETE' (GatewayEventInviteDelete) events.
func (d *dispatcher) OnInviteDelete(h func(context.Context, InviteDeleteEvent)) {
	d.register(GatewayEventInviteDelete, h)
}

// OnMessageCreate registers a handler for 'MESSAGE_CREATE' (GatewayEventMessageCreate) events.
func (d *dispatcher) OnMessageCreate(h func(context.Context, MessageCreateEvent)) {
	d.register(GatewayEventMessageCreate, h)
}

// OnMessageUpdate registers a handler for 'MESSAGE_UPDATE' (GatewayEventMessageUpdate) events.
func (d *dispatcher) OnMessageUpdate(h func(context.Context, MessageUpdateEvent)) {
	d.register(GatewayEventMessageUpdate, h)
}

// OnMessageDelete registers a handler for 'MESSAGE_DELETE' (GatewayEventMessageDelete) events.
func (d *dispatcher) OnMessageDelete(h func(context.Context, MessageDeleteEvent)) {
	d.register(GatewayEventMessageDelete, h)
}

// OnMessageDeleteBulk registers a handler for 'MESSAGE_DELETE_BULK' (GatewayEventMessageDeleteBulk) events.
func (d *dispatcher) OnMessageDeleteBulk(h func(context.Context, MessageDeleteBulkEvent)) {
	d.register(GatewayEventMessageDeleteBulk, h)
}

// OnMessageReactionAdd registers a handler for 'MESSAGE_REACTION_ADD' (GatewayEventMessageReactionAdd) events.
func (d *dispatcher) OnMessageReactionAdd(h func(context.Context, MessageReactionAddEvent)) {
	d.register(GatewayEventMessageReactionAdd, h)
}

// OnMessageReactionRemove registers a handler for 'MESSAGE_REACTION_REMOVE' (GatewayEventMessageReactionRemove) events.
func (d *dispatcher) OnMessageReactionRemove(h func(context.Context, MessageReactionRemoveEvent)) {
	d.register(GatewayEventMessageReactionRemove, h)
}

// OnMessageReactionRemoveAll registers a handler for 'MESSAGE_REACTION_REMOVE_ALL' (GatewayEventMessageReactionRemoveAll) events.
func (d *dispatcher) OnMessageReactionRemoveAll(
	h func(context.Context, MessageReactionRemoveAllEvent),
) {
	d.register(GatewayEventMessageReactionRemoveAll, h)
}

// OnMessageReactionRemoveEmoji registers a handler for 'MESSAGE_REACTION_REMOVE_EMOJI' (GatewayEventMessageReactionRemoveEmoji) events.
func (d *dispatcher) OnMessageReactionRemoveEmoji(
	h func(context.Context, MessageReactionRemoveEmojiEvent),
) {
	d.register(GatewayEventMessageReactionRemoveEmoji, h)
}

// OnMessagePollVoteAdd registers a handler for 'MESSAGE_POLL_VOTE_ADD' (GatewayEventMessagePollVoteAdd) events.
func (d *dispatcher) OnMessagePollVoteAdd(h func(context.Context, MessagePollVoteAddEvent)) {
	d.register(GatewayEventMessagePollVoteAdd, h)
}

// OnMessagePollVoteRemove registers a handler for 'MESSAGE_POLL_VOTE_REMOVE' (GatewayEventMessagePollVoteRemove) events.
func (d *dispatcher) OnMessagePollVoteRemove(h func(context.Context, MessagePollVoteRemoveEvent)) {
	d.register(GatewayEventMessagePollVoteRemove, h)
}

// OnPresenceUpdate registers a handler for 'PRESENCE_UPDATE' (GatewayEventPresenceUpdate) events.
func (d *dispatcher) OnPresenceUpdate(h func(context.Context, PresenceUpdateEvent)) {
	d.register(GatewayEventPresenceUpdate, h)
}

// OnStageInstanceCreate registers a handler for 'STAGE_INSTANCE_CREATE' (GatewayEventStageInstanceCreate) events.
func (d *dispatcher) OnStageInstanceCreate(h func(context.Context, StageInstanceCreateEvent)) {
	d.register(GatewayEventStageInstanceCreate, h)
}

// OnStageInstanceUpdate registers a handler for 'STAGE_INSTANCE_UPDATE' (GatewayEventStageInstanceUpdate) events.
func (d *dispatcher) OnStageInstanceUpdate(h func(context.Context, StageInstanceUpdateEvent)) {
	d.register(GatewayEventStageInstanceUpdate, h)
}

// OnStageInstanceDelete registers a handler for 'STAGE_INSTANCE_DELETE' (GatewayEventStageInstanceDelete) events.
func (d *dispatcher) OnStageInstanceDelete(h func(context.Context, StageInstanceDeleteEvent)) {
	d.register(GatewayEventStageInstanceDelete, h)
}

// OnSubscriptionCreate registers a handler for 'SUBSCRIPTION_CREATE' (GatewayEventSubscriptionCreate) events.
func (d *dispatcher) OnSubscriptionCreate(h func(context.Context, SubscriptionCreateEvent)) {
	d.register(GatewayEventSubscriptionCreate, h)
}

// OnSubscriptionUpdate registers a handler for 'SUBSCRIPTION_UPDATE' (GatewayEventSubscriptionUpdate) events.
func (d *dispatcher) OnSubscriptionUpdate(h func(context.Context, SubscriptionUpdateEvent)) {
	d.register(GatewayEventSubscriptionUpdate, h)
}

// OnSubscriptionDelete registers a handler for 'SUBSCRIPTION_DELETE' (GatewayEventSubscriptionDelete) events.
func (d *dispatcher) OnSubscriptionDelete(h func(context.Context, SubscriptionDeleteEvent)) {
	d.register(GatewayEventSubscriptionDelete, h)
}

// OnTypingStart registers a handler for 'TYPING_START' (GatewayEventTypingStart) events.
func (d *dispatcher) OnTypingStart(h func(context.Context, TypingStartEvent)) {
	d.register(GatewayEventTypingStart, h)
}

// OnUserUpdate registers a handler for 'USER_UPDATE' (GatewayEventUserUpdate) events.
func (d *dispatcher) OnUserUpdate(h func(context.Context, UserUpdateEvent)) {
	d.register(GatewayEventUserUpdate, h)
}

// OnVoiceChannelEffectSend registers a handler for 'VOICE_CHANNEL_EFFECT_SEND' (GatewayEventVoiceChannelEffectSend) events.
func (d *dispatcher) OnVoiceChannelEffectSend(
	h func(context.Context, VoiceChannelEffectSendEvent),
) {
	d.register(GatewayEventVoiceChannelEffectSend, h)
}

// OnVoiceStateUpdate registers a handler for 'VOICE_STATE_UPDATE' (GatewayEventVoiceStateUpdate) events.
func (d *dispatcher) OnVoiceStateUpdate(h func(context.Context, VoiceStateUpdateEvent)) {
	d.register(GatewayEventVoiceStateUpdate, h)
}

// OnVoiceServerUpdate registers a handler for 'VOICE_SERVER_UPDATE' (GatewayEventVoiceServerUpdate) events.
func (d *dispatcher) OnVoiceServerUpdate(h func(context.Context, VoiceServerUpdateEvent)) {
	d.register(GatewayEventVoiceServerUpdate, h)
}

// OnWebhooksUpdate registers a handler for 'WEBHOOKS_UPDATE' (GatewayEventWebhooksUpdate) events.
func (d *dispatcher) OnWebhooksUpdate(h func(context.Context, WebhooksUpdateEvent)) {
	d.register(GatewayEventWebhooksUpdate, h)
}

// OnApplicationCommandPermissionsUpdate registers a handler for 'APPLICATION_COMMAND_PERMISSIONS_UPDATE' (GatewayEventApplicationCommandPermissionsUpdate) events.
func (d *dispatcher) OnApplicationCommandPermissionsUpdate(
	h func(context.Context, ApplicationCommandPermissionsUpdateEvent),
) {
	d.register(GatewayEventApplicationCommandPermissionsUpdate, h)
}

// OnAutoModerationRuleCreate registers a handler for 'AUTO_MODERATION_RULE_CREATE' (GatewayEventAutoModerationRuleCreate) events.
func (d *dispatcher) OnAutoModerationRuleCreate(
	h func(context.Context, AutoModerationRuleCreateEvent),
) {
	d.register(GatewayEventAutoModerationRuleCreate, h)
}

// OnAutoModerationRuleUpdate registers a handler for 'AUTO_MODERATION_RULE_UPDATE' (GatewayEventAutoModerationRuleUpdate) events.
func (d *dispatcher) OnAutoModerationRuleUpdate(
	h func(context.Context, AutoModerationRuleUpdateEvent),
) {
	d.register(GatewayEventAutoModerationRuleUpdate, h)
}

// OnAutoModerationRuleDelete registers a handler for 'AUTO_MODERATION_RULE_DELETE' (GatewayEventAutoModerationRuleDelete) events.
func (d *dispatcher) OnAutoModerationRuleDelete(
	h func(context.Context, AutoModerationRuleDeleteEvent),
) {
	d.register(GatewayEventAutoModerationRuleDelete, h)
}

// OnAutoModerationActionExecution registers a handler for 'AUTO_MODERATION_ACTION_EXECUTION' (GatewayEventAutoModerationActionExecution) events.
func (d *dispatcher) OnAutoModerationActionExecution(
	h func(context.Context, AutoModerationActionExecutionEvent),
) {
	d.register(GatewayEventAutoModerationActionExecution, h)
}

// OnEntitlementCreate registers a handler for 'ENTITLEMENT_CREATE' (GatewayEventEntitlementCreate) events.
func (d *dispatcher) OnEntitlementCreate(h func(context.Context, EntitlementCreateEvent)) {
	d.register(GatewayEventEntitlementCreate, h)
}

// OnEntitlementUpdate registers a handler for 'ENTITLEMENT_UPDATE' (GatewayEventEntitlementUpdate) events.
func (d *dispatcher) OnEntitlementUpdate(h func(context.Context, EntitlementUpdateEvent)) {
	d.register(GatewayEventEntitlementUpdate, h)
}

// OnEntitlementDelete registers a handler for 'ENTITLEMENT_DELETE' (GatewayEventEntitlementDelete) events.
func (d *dispatcher) OnEntitlementDelete(h func(context.Context, EntitlementDeleteEvent)) {
	d.register(GatewayEventEntitlementDelete, h)
}

// OnChannelInfo registers a handler for 'CHANNEL_INFO' (GatewayEventChannelInfo) events.
func (d *dispatcher) OnChannelInfo(h func(context.Context, ChannelInfoEvent)) {
	d.register(GatewayEventChannelInfo, h)
}

// OnVoiceChannelStatusUpdate registers a handler for 'VOICE_CHANNEL_STATUS_UPDATE' (GatewayEventVoiceChannelStatusUpdate) events.
func (d *dispatcher) OnVoiceChannelStatusUpdate(
	h func(context.Context, VoiceChannelStatusUpdateEvent),
) {
	d.register(GatewayEventVoiceChannelStatusUpdate, h)
}

// OnVoiceChannelStartTimeUpdate registers a handler for 'VOICE_CHANNEL_START_TIME_UPDATE' (GatewayEventVoiceChannelStartTimeUpdate) events.
func (d *dispatcher) OnVoiceChannelStartTimeUpdate(
	h func(context.Context, VoiceChannelStartTimeUpdateEvent),
) {
	d.register(GatewayEventVoiceChannelStartTimeUpdate, h)
}

// OnApplicationAuthorized registers a handler for 'APPLICATION_AUTHORIZED' (GatewayEventApplicationAuthorized) webhook events.
func (d *dispatcher) OnApplicationAuthorized(h func(context.Context, ApplicationAuthorizedEvent)) {
	d.register(GatewayEventApplicationAuthorized, h)
}

// OnApplicationDeauthorized registers a handler for 'APPLICATION_DEAUTHORIZED' (GatewayEventApplicationDeauthorized) webhook events.
func (d *dispatcher) OnApplicationDeauthorized(
	h func(context.Context, ApplicationDeauthorizedEvent),
) {
	d.register(GatewayEventApplicationDeauthorized, h)
}

// OnRateLimited registers a handler for 'RATE_LIMITED' (GatewayEventRateLimited) gateway events.
func (d *dispatcher) OnRateLimited(h func(context.Context, RateLimitedEvent)) {
	d.register(GatewayEventRateLimited, h)
}

// OnQuestUserEnrollment registers a handler for 'QUEST_USER_ENROLLMENT' (GatewayEventQuestUserEnrollment) webhook events.
func (d *dispatcher) OnQuestUserEnrollment(h func(context.Context, QuestUserEnrollmentEvent)) {
	d.register(GatewayEventQuestUserEnrollment, h)
}

// OnLobbyMessageCreate registers a handler for 'LOBBY_MESSAGE_CREATE' (GatewayEventLobbyMessageCreate) webhook events.
func (d *dispatcher) OnLobbyMessageCreate(h func(context.Context, LobbyMessageCreateEvent)) {
	d.register(GatewayEventLobbyMessageCreate, h)
}

// OnLobbyMessageUpdate registers a handler for 'LOBBY_MESSAGE_UPDATE' (GatewayEventLobbyMessageUpdate) webhook events.
func (d *dispatcher) OnLobbyMessageUpdate(h func(context.Context, LobbyMessageUpdateEvent)) {
	d.register(GatewayEventLobbyMessageUpdate, h)
}

// OnLobbyMessageDelete registers a handler for 'LOBBY_MESSAGE_DELETE' (GatewayEventLobbyMessageDelete) webhook events.
func (d *dispatcher) OnLobbyMessageDelete(h func(context.Context, LobbyMessageDeleteEvent)) {
	d.register(GatewayEventLobbyMessageDelete, h)
}

// OnGameDirectMessageCreate registers a handler for 'GAME_DIRECT_MESSAGE_CREATE' (GatewayEventGameDirectMessageCreate) webhook events.
func (d *dispatcher) OnGameDirectMessageCreate(
	h func(context.Context, GameDirectMessageCreateEvent),
) {
	d.register(GatewayEventGameDirectMessageCreate, h)
}

// OnGameDirectMessageUpdate registers a handler for 'GAME_DIRECT_MESSAGE_UPDATE' (GatewayEventGameDirectMessageUpdate) webhook events.
func (d *dispatcher) OnGameDirectMessageUpdate(
	h func(context.Context, GameDirectMessageUpdateEvent),
) {
	d.register(GatewayEventGameDirectMessageUpdate, h)
}

// OnGameDirectMessageDelete registers a handler for 'GAME_DIRECT_MESSAGE_DELETE' (GatewayEventGameDirectMessageDelete) webhook events.
func (d *dispatcher) OnGameDirectMessageDelete(
	h func(context.Context, GameDirectMessageDeleteEvent),
) {
	d.register(GatewayEventGameDirectMessageDelete, h)
}
