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
	"encoding/json"
	"time"
)

// ShardID represents the ID of a Discord Gateway shard, or -1 for Webhook Events.
type ShardID int

// IsWebhookEventsDispatched returns true if the event was dispatched via Webhook Events (negative shard ID).
func (s ShardID) IsWebhookEventsDispatched() bool {
	return s < 0
}

// ReadyEvent is dispatched when a shard has completed the initial handshake.
type ReadyEvent struct {
	ShardID          ShardID
	Version          int                `json:"v"`
	User             User               `json:"user"`
	Guilds           []Guild            `json:"guilds"`
	SessionID        string             `json:"session_id"`
	ResumeGatewayURL string             `json:"resume_gateway_url"`
	Shard            *[2]int            `json:"shard,omitempty"`
	Application      PartialApplication `json:"application"`
}

// ResumedEvent is dispatched when a shard has resumed a previous session.
type ResumedEvent struct {
	ShardID ShardID
}

// GuildCreateEvent is dispatched when a guild becomes available or user joins a guild.
type GuildCreateEvent struct {
	ShardID ShardID
	Guild   GatewayGuild
}

// MessageCreateEvent is dispatched when a message is created.
type MessageCreateEvent struct {
	ShardID ShardID
	Message Message
}

// MessageUpdateEvent is dispatched when a message is updated.
type MessageUpdateEvent struct {
	ShardID    ShardID
	OldMessage Message
	NewMessage Message
}

// MessageDeleteEvent is dispatched when a message is deleted.
type MessageDeleteEvent struct {
	ShardID ShardID
	Message Message
}

// InteractionCreateEvent is dispatched when an interaction is created.
type InteractionCreateEvent struct {
	ShardID     ShardID
	Interaction Interaction
}

var _ json.Unmarshaler = (*InteractionCreateEvent)(nil)

// UnmarshalJSON implements json.Unmarshaler for InteractionCreateEvent.
func (c *InteractionCreateEvent) UnmarshalJSON(buf []byte) error {
	interaction, err := UnmarshalInteraction(buf)
	if err == nil {
		c.Interaction = interaction
	}
	return err
}

// VoiceStateUpdateEvent is dispatched when a voice state is updated.
type VoiceStateUpdateEvent struct {
	ShardID  ShardID
	OldState GatewayVoiceState
	NewState GatewayVoiceState
}

// ApplicationCommandPermissionsUpdateEvent is dispatched when application command permissions are updated.
type ApplicationCommandPermissionsUpdateEvent struct {
	ShardID                                ShardID
	OldGuildApplicationCommandPermissions GuildApplicationCommandPermissions `json:"-"`
	NewGuildApplicationCommandPermissions GuildApplicationCommandPermissions `json:"-"`
}

// AutoModerationRuleCreateEvent is dispatched when an Auto Moderation rule is created.
type AutoModerationRuleCreateEvent struct {
	ShardID ShardID
	AutoModerationRule
}

// AutoModerationRuleUpdateEvent is dispatched when an Auto Moderation rule is updated.
type AutoModerationRuleUpdateEvent struct {
	ShardID               ShardID
	OldAutoModerationRule AutoModerationRule `json:"-"`
	NewAutoModerationRule AutoModerationRule `json:"-"`
}

// AutoModerationRuleDeleteEvent is dispatched when an Auto Moderation rule is deleted.
type AutoModerationRuleDeleteEvent struct {
	ShardID ShardID
	AutoModerationRule
}

// AutoModerationActionExecutionEvent is dispatched when an Auto Moderation rule executes an action.
type AutoModerationActionExecutionEvent struct {
	ShardID              ShardID
	GuildID              Snowflake                 `json:"guild_id"`
	Action               AutoModerationAction      `json:"action"`
	RuleID               Snowflake                 `json:"rule_id"`
	RuleTriggerType      AutoModerationTriggerType `json:"rule_trigger_type"`
	UserID               Snowflake                 `json:"user_id"`
	ChannelID            Snowflake                 `json:"channel_id,omitempty"`
	MessageID            Snowflake                 `json:"message_id,omitempty"`
	AlertSystemMessageID Snowflake                 `json:"alert_system_message_id,omitempty"`
	Content              string                    `json:"content,omitempty"`
	MatchedKeyword       string                    `json:"matched_keyword,omitempty"`
	MatchedContent       string                    `json:"matched_content,omitempty"`
}

// ChannelCreateEvent is dispatched when a channel is created.
type ChannelCreateEvent struct {
	ShardID ShardID
	Channel Channel
}

func (e *ChannelCreateEvent) UnmarshalJSON(buf []byte) error {
	c, err := UnmarshalChannel(buf)
	if err != nil {
		return err
	}
	e.Channel = c
	return nil
}

// ChannelUpdateEvent is dispatched when a channel is updated.
type ChannelUpdateEvent struct {
	ShardID    ShardID
	OldChannel Channel `json:"-"`
	NewChannel Channel `json:"-"`
}

func (e *ChannelUpdateEvent) UnmarshalJSON(buf []byte) error {
	c, err := UnmarshalChannel(buf)
	if err != nil {
		return err
	}
	e.NewChannel = c
	return nil
}

// ChannelDeleteEvent is dispatched when a channel is deleted.
type ChannelDeleteEvent struct {
	ShardID ShardID
	Channel Channel
}

func (e *ChannelDeleteEvent) UnmarshalJSON(buf []byte) error {
	c, err := UnmarshalChannel(buf)
	if err != nil {
		return err
	}
	e.Channel = c
	return nil
}

// ChannelPinsUpdateEvent is dispatched when a message is pinned or unpinned in a channel.
type ChannelPinsUpdateEvent struct {
	ShardID          ShardID
	GuildID          Snowflake  `json:"guild_id,omitempty"`
	ChannelID        Snowflake  `json:"channel_id"`
	LastPinTimestamp *time.Time `json:"last_pin_timestamp,omitempty"`
}

// ThreadCreateEvent is dispatched when a thread is created.
type ThreadCreateEvent struct {
	ShardID      ShardID
	Thread       *ThreadChannel `json:"thread"`
	NewlyCreated bool           `json:"newly_created,omitempty"`
}

func (e *ThreadCreateEvent) UnmarshalJSON(buf []byte) error {
	type threadAlias ThreadChannel
	var raw struct {
		threadAlias
		NewlyCreated bool `json:"newly_created,omitempty"`
	}
	if err := json.Unmarshal(buf, &raw); err != nil {
		return err
	}
	tc := ThreadChannel(raw.threadAlias)
	e.Thread = &tc
	e.NewlyCreated = raw.NewlyCreated
	return nil
}

// ThreadUpdateEvent is dispatched when a thread is updated.
type ThreadUpdateEvent struct {
	ShardID   ShardID
	OldThread *ThreadChannel `json:"-"`
	NewThread *ThreadChannel `json:"-"`
}

func (e *ThreadUpdateEvent) UnmarshalJSON(buf []byte) error {
	var tc ThreadChannel
	if err := json.Unmarshal(buf, &tc); err != nil {
		return err
	}
	e.NewThread = &tc
	return nil
}

// ThreadDeleteEvent is dispatched when a thread is deleted.
type ThreadDeleteEvent struct {
	ShardID  ShardID
	ID       Snowflake   `json:"id"`
	GuildID  Snowflake   `json:"guild_id"`
	ParentID Snowflake   `json:"parent_id"`
	Type     ChannelType `json:"type"`
}

// ThreadListSyncEvent is dispatched when gaining access to a channel with active threads.
type ThreadListSyncEvent struct {
	ShardID    ShardID
	GuildID    Snowflake       `json:"guild_id"`
	ChannelIDs []Snowflake     `json:"channel_ids,omitempty"`
	Threads    []ThreadChannel `json:"threads"`
	Members    []ThreadMember  `json:"members"`
}

// ThreadMemberUpdateEvent is dispatched when a thread member for the current user is updated.
type ThreadMemberUpdateEvent struct {
	ShardID   ShardID
	GuildID   Snowflake    `json:"guild_id"`
	OldMember ThreadMember `json:"-"`
	NewMember ThreadMember `json:"-"`
}

// ThreadMembersUpdateEvent is dispatched when users are added to or removed from a thread.
type ThreadMembersUpdateEvent struct {
	ShardID          ShardID
	ID               Snowflake      `json:"id"`
	GuildID          Snowflake      `json:"guild_id"`
	MemberCount      int            `json:"member_count"`
	AddedMembers     []ThreadMember `json:"added_members,omitempty"`
	RemovedMemberIDs []Snowflake    `json:"removed_member_ids,omitempty"`
}

// EntitlementCreateEvent is dispatched when an entitlement is created.
type EntitlementCreateEvent struct {
	ShardID ShardID
	Entitlement
}

// EntitlementUpdateEvent is dispatched when an entitlement is updated.
type EntitlementUpdateEvent struct {
	ShardID        ShardID
	OldEntitlement Entitlement `json:"-"`
	NewEntitlement Entitlement `json:"-"`
}

// EntitlementDeleteEvent is dispatched when an entitlement is deleted.
type EntitlementDeleteEvent struct {
	ShardID ShardID
	Entitlement
}

// GuildUpdateEvent is dispatched when a guild is updated.
type GuildUpdateEvent struct {
	ShardID  ShardID
	OldGuild Guild `json:"-"`
	NewGuild Guild `json:"-"`
}

// GuildDeleteEvent is dispatched when a guild is deleted or becomes unavailable.
type GuildDeleteEvent struct {
	ShardID     ShardID
	ID          Snowflake `json:"id"`
	Unavailable bool      `json:"unavailable,omitempty"`
}

// GuildAuditLogEntryCreateEvent is dispatched when an audit log entry is created.
type GuildAuditLogEntryCreateEvent struct {
	ShardID ShardID
	AuditLogEntry
	GuildID Snowflake `json:"guild_id"`
}

// GuildBanAddEvent is dispatched when a user is banned from a guild.
type GuildBanAddEvent struct {
	ShardID ShardID
	GuildID Snowflake `json:"guild_id"`
	User    User      `json:"user"`
}

// GuildBanRemoveEvent is dispatched when a user is unbanned from a guild.
type GuildBanRemoveEvent struct {
	ShardID ShardID
	GuildID Snowflake `json:"guild_id"`
	User    User      `json:"user"`
}

// GuildEmojisUpdateEvent is dispatched when guild emojis are updated.
type GuildEmojisUpdateEvent struct {
	ShardID   ShardID
	GuildID   Snowflake `json:"guild_id"`
	OldEmojis []Emoji   `json:"-"`
	NewEmojis []Emoji   `json:"emojis"`
}

// GuildStickersUpdateEvent is dispatched when guild stickers are updated.
type GuildStickersUpdateEvent struct {
	ShardID     ShardID
	GuildID     Snowflake `json:"guild_id"`
	OldStickers []Sticker `json:"-"`
	NewStickers []Sticker `json:"stickers"`
}

// GuildIntegrationsUpdateEvent is dispatched when guild integrations are updated.
type GuildIntegrationsUpdateEvent struct {
	ShardID ShardID
	GuildID Snowflake `json:"guild_id"`
}

// GuildMemberAddEvent is dispatched when a new member joins a guild.
type GuildMemberAddEvent struct {
	ShardID ShardID
	Member  FullMember `json:"member"`
	GuildID Snowflake  `json:"guild_id"`
}

func (e *GuildMemberAddEvent) UnmarshalJSON(buf []byte) error {
	type localEvent struct {
		FullMember
		GuildID Snowflake `json:"guild_id"`
	}
	var v localEvent
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	e.Member = v.FullMember
	e.GuildID = v.GuildID
	e.Member.GuildID = v.GuildID
	if e.Member.User.ID != 0 {
		e.Member.ID = e.Member.User.ID
	}
	return nil
}

// GuildMemberRemoveEvent is dispatched when a user is removed from a guild.
type GuildMemberRemoveEvent struct {
	ShardID ShardID
	GuildID Snowflake `json:"guild_id"`
	User    User      `json:"user"`
}

// GuildMemberUpdateEvent is dispatched when a guild member is updated.
type GuildMemberUpdateEvent struct {
	ShardID                    ShardID
	GuildID                    Snowflake             `json:"guild_id"`
	Roles                      []Snowflake           `json:"roles"`
	User                       User                  `json:"user"`
	Nick                       string                `json:"nick,omitempty"`
	Avatar                     string                `json:"avatar,omitempty"`
	JoinedAt                   *time.Time            `json:"joined_at,omitempty"`
	PremiumSince               *time.Time            `json:"premium_since,omitempty"`
	Deaf                       bool                  `json:"deaf,omitempty"`
	Mute                       bool                  `json:"mute,omitempty"`
	Pending                    bool                  `json:"pending,omitempty"`
	CommunicationDisabledUntil *time.Time            `json:"communication_disabled_until,omitempty"`
	Flags                      MemberFlags           `json:"flags,omitempty"`
	AvatarDecorationData       *AvatarDecorationData `json:"avatar_decoration_data,omitempty"`
	OldMember                  Member                `json:"-"`
	NewMember                  Member                `json:"-"`
}

// GuildMembersChunkEvent is dispatched in response to RequestGuildMembers.
type GuildMembersChunkEvent struct {
	ShardID    ShardID
	GuildID    Snowflake         `json:"guild_id"`
	Members    []FullMember      `json:"members"`
	ChunkIndex int               `json:"chunk_index"`
	ChunkCount int               `json:"chunk_count"`
	NotFound   []Snowflake       `json:"not_found,omitempty"`
	Presences  []GatewayPresence `json:"presences,omitempty"`
	Nonce      string            `json:"nonce,omitempty"`
}

func (e *GuildMembersChunkEvent) UnmarshalJSON(buf []byte) error {
	type localEvent GuildMembersChunkEvent
	var v localEvent
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	*e = GuildMembersChunkEvent(v)
	if e.GuildID != 0 {
		for i := range e.Members {
			e.Members[i].GuildID = e.GuildID
			if e.Members[i].User.ID != 0 {
				e.Members[i].ID = e.Members[i].User.ID
			}
		}
	}
	return nil
}

// GuildRoleCreateEvent is dispatched when a guild role is created.
type GuildRoleCreateEvent struct {
	ShardID ShardID
	GuildID Snowflake `json:"guild_id"`
	Role    Role      `json:"role"`
}

// GuildRoleUpdateEvent is dispatched when a guild role is updated.
type GuildRoleUpdateEvent struct {
	ShardID ShardID
	GuildID Snowflake `json:"guild_id"`
	OldRole Role      `json:"-"`
	NewRole Role      `json:"role"`
}

// GuildRoleDeleteEvent is dispatched when a guild role is deleted.
type GuildRoleDeleteEvent struct {
	ShardID ShardID
	GuildID Snowflake `json:"guild_id"`
	RoleID  Snowflake `json:"role_id"`
}

// GuildScheduledEventCreateEvent is dispatched when a scheduled event is created.
type GuildScheduledEventCreateEvent struct {
	ShardID ShardID
	GuildScheduledEvent
}

// GuildScheduledEventUpdateEvent is dispatched when a scheduled event is updated.
type GuildScheduledEventUpdateEvent struct {
	ShardID                ShardID
	OldGuildScheduledEvent GuildScheduledEvent `json:"-"`
	NewGuildScheduledEvent GuildScheduledEvent `json:"-"`
}

// GuildScheduledEventDeleteEvent is dispatched when a scheduled event is deleted.
type GuildScheduledEventDeleteEvent struct {
	ShardID ShardID
	GuildScheduledEvent
}

// GuildScheduledEventUserAddEvent is dispatched when a user subscribes to a scheduled event.
type GuildScheduledEventUserAddEvent struct {
	ShardID               ShardID
	GuildScheduledEventID Snowflake `json:"guild_scheduled_event_id"`
	UserID                Snowflake `json:"user_id"`
	GuildID               Snowflake `json:"guild_id"`
}

// GuildScheduledEventUserRemoveEvent is dispatched when a user unsubscribes from a scheduled event.
type GuildScheduledEventUserRemoveEvent struct {
	ShardID               ShardID
	GuildScheduledEventID Snowflake `json:"guild_scheduled_event_id"`
	UserID                Snowflake `json:"user_id"`
	GuildID               Snowflake `json:"guild_id"`
}

// GuildSoundboardSoundCreateEvent is dispatched when a soundboard sound is created.
type GuildSoundboardSoundCreateEvent struct {
	ShardID ShardID
	SoundboardSound
}

// GuildSoundboardSoundUpdateEvent is dispatched when a soundboard sound is updated.
type GuildSoundboardSoundUpdateEvent struct {
	ShardID  ShardID
	OldSound SoundboardSound `json:"-"`
	NewSound SoundboardSound `json:"sound"`
}

// GuildSoundboardSoundDeleteEvent is dispatched when a soundboard sound is deleted.
type GuildSoundboardSoundDeleteEvent struct {
	ShardID ShardID
	SoundID Snowflake `json:"sound_id"`
	GuildID Snowflake `json:"guild_id"`
}

// GuildSoundboardSoundsUpdateEvent is dispatched when all soundboard sounds for a guild are updated.
type GuildSoundboardSoundsUpdateEvent struct {
	ShardID             ShardID
	GuildID             Snowflake         `json:"guild_id"`
	OldSoundboardSounds []SoundboardSound `json:"-"`
	NewSoundboardSounds []SoundboardSound `json:"soundboard_sounds"`
}

// SoundboardSoundsEvent is dispatched in response to RequestSoundboardSounds.
type SoundboardSoundsEvent struct {
	ShardID          ShardID
	SoundboardSounds []SoundboardSound `json:"soundboard_sounds"`
	GuildID          Snowflake         `json:"guild_id"`
}

// IntegrationCreateEvent is dispatched when an integration is created.
type IntegrationCreateEvent struct {
	ShardID ShardID
	Integration
	GuildID Snowflake `json:"guild_id"`
}

// IntegrationUpdateEvent is dispatched when an integration is updated.
type IntegrationUpdateEvent struct {
	ShardID        ShardID
	GuildID        Snowflake   `json:"guild_id"`
	OldIntegration Integration `json:"-"`
	NewIntegration Integration `json:"integration"`
}

// IntegrationDeleteEvent is dispatched when an integration is deleted.
type IntegrationDeleteEvent struct {
	ShardID       ShardID
	ID            Snowflake `json:"id"`
	GuildID       Snowflake `json:"guild_id"`
	ApplicationID Snowflake `json:"application_id,omitempty"`
}

// InviteCreateEvent is dispatched when an invite is created.
type InviteCreateEvent struct {
	ShardID           ShardID
	ChannelID         Snowflake           `json:"channel_id"`
	Code              string              `json:"code"`
	CreatedAt         time.Time           `json:"created_at"`
	GuildID           Snowflake           `json:"guild_id,omitempty"`
	Inviter           *User               `json:"inviter,omitempty"`
	MaxAge            int                 `json:"max_age"`
	MaxUses           int                 `json:"max_uses"`
	TargetType        *InviteTargetType   `json:"target_type,omitempty"`
	TargetUser        *User               `json:"target_user,omitempty"`
	TargetApplication *PartialApplication `json:"target_application,omitempty"`
	Temporary         bool                `json:"temporary"`
	Uses              int                 `json:"uses"`
}

// InviteDeleteEvent is dispatched when an invite is deleted.
type InviteDeleteEvent struct {
	ShardID   ShardID
	ChannelID Snowflake `json:"channel_id"`
	GuildID   Snowflake `json:"guild_id,omitempty"`
	Code      string    `json:"code"`
}

// MessageDeleteBulkEvent is dispatched when multiple messages are deleted in bulk.
type MessageDeleteBulkEvent struct {
	ShardID   ShardID
	IDs       []Snowflake `json:"ids"`
	ChannelID Snowflake   `json:"channel_id"`
	GuildID   Snowflake   `json:"guild_id,omitempty"`
}

// MessageReactionAddEvent is dispatched when a reaction is added to a message.
type MessageReactionAddEvent struct {
	ShardID         ShardID
	UserID          Snowflake    `json:"user_id"`
	ChannelID       Snowflake    `json:"channel_id"`
	MessageID       Snowflake    `json:"message_id"`
	GuildID         Snowflake    `json:"guild_id,omitempty"`
	Member          *Member      `json:"member,omitempty"`
	Emoji           Emoji        `json:"emoji"`
	MessageAuthorID Snowflake    `json:"message_author_id,omitempty"`
	Burst           bool         `json:"burst,omitempty"`
	BurstColors     []string     `json:"burst_colors,omitempty"`
	Type            ReactionType `json:"type"`
}

func (e *MessageReactionAddEvent) UnmarshalJSON(buf []byte) error {
	type localEvent MessageReactionAddEvent
	var v localEvent
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	*e = MessageReactionAddEvent(v)
	if e.Member != nil {
		if e.GuildID != 0 {
			e.Member.GuildID = e.GuildID
		}
		if e.UserID != 0 {
			e.Member.ID = e.UserID
		}
	}
	return nil
}

// MessageReactionRemoveEvent is dispatched when a reaction is removed from a message.
type MessageReactionRemoveEvent struct {
	ShardID   ShardID
	UserID    Snowflake    `json:"user_id"`
	ChannelID Snowflake    `json:"channel_id"`
	MessageID Snowflake    `json:"message_id"`
	GuildID   Snowflake    `json:"guild_id,omitempty"`
	Emoji     Emoji        `json:"emoji"`
	Burst     bool         `json:"burst,omitempty"`
	Type      ReactionType `json:"type"`
}

// MessageReactionRemoveAllEvent is dispatched when all reactions are removed from a message.
type MessageReactionRemoveAllEvent struct {
	ShardID   ShardID
	ChannelID Snowflake `json:"channel_id"`
	MessageID Snowflake `json:"message_id"`
	GuildID   Snowflake `json:"guild_id,omitempty"`
}

// MessageReactionRemoveEmojiEvent is dispatched when all reactions of a specific emoji are removed.
type MessageReactionRemoveEmojiEvent struct {
	ShardID   ShardID
	ChannelID Snowflake `json:"channel_id"`
	GuildID   Snowflake `json:"guild_id,omitempty"`
	MessageID Snowflake `json:"message_id"`
	Emoji     Emoji     `json:"emoji"`
}

// PresenceUpdateEvent is dispatched when a user's presence or status changes.
type PresenceUpdateEvent struct {
	ShardID     ShardID
	OldPresence *GatewayPresence `json:"-"`
	NewPresence GatewayPresence  `json:"-"`
	GatewayPresence
}

// StageInstanceCreateEvent is dispatched when a stage instance is created.
type StageInstanceCreateEvent struct {
	ShardID ShardID
	StageInstance
}

// StageInstanceUpdateEvent is dispatched when a stage instance is updated.
type StageInstanceUpdateEvent struct {
	ShardID          ShardID
	OldStageInstance StageInstance `json:"-"`
	NewStageInstance StageInstance `json:"-"`
}

// StageInstanceDeleteEvent is dispatched when a stage instance is deleted or closed.
type StageInstanceDeleteEvent struct {
	ShardID ShardID
	StageInstance
}

// SubscriptionCreateEvent is dispatched when an application subscription is created.
type SubscriptionCreateEvent struct {
	ShardID ShardID
	Subscription
}

// SubscriptionUpdateEvent is dispatched when an application subscription is updated.
type SubscriptionUpdateEvent struct {
	ShardID         ShardID
	OldSubscription Subscription `json:"-"`
	NewSubscription Subscription `json:"-"`
}

// SubscriptionDeleteEvent is dispatched when an application subscription is deleted.
type SubscriptionDeleteEvent struct {
	ShardID ShardID
	Subscription
}

// TypingStartEvent is dispatched when a user starts typing.
type TypingStartEvent struct {
	ShardID   ShardID
	ChannelID Snowflake `json:"channel_id"`
	GuildID   Snowflake `json:"guild_id,omitempty"`
	UserID    Snowflake `json:"user_id"`
	Timestamp int64     `json:"timestamp"`
	Member    *Member   `json:"member,omitempty"`
}

func (e *TypingStartEvent) UnmarshalJSON(buf []byte) error {
	type localEvent TypingStartEvent
	var v localEvent
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	*e = TypingStartEvent(v)
	if e.Member != nil {
		if e.GuildID != 0 {
			e.Member.GuildID = e.GuildID
		}
		if e.UserID != 0 {
			e.Member.ID = e.UserID
		}
	}
	return nil
}

// UserUpdateEvent is dispatched when the current user's profile is updated.
type UserUpdateEvent struct {
	ShardID ShardID
	OldUser User `json:"-"`
	NewUser User `json:"-"`
}

// VoiceChannelEffectAnimationType represents voice channel effect animation types.
type VoiceChannelEffectAnimationType int

const (
	VoiceChannelEffectAnimationTypePremium VoiceChannelEffectAnimationType = 0
	VoiceChannelEffectAnimationTypeBasic   VoiceChannelEffectAnimationType = 1
)

// Is returns true if this VoiceChannelEffectAnimationType matches the provided one.
func (v VoiceChannelEffectAnimationType) Is(other VoiceChannelEffectAnimationType) bool {
	return v == other
}

// VoiceChannelEffectSendEvent is dispatched when an effect is sent in a voice channel.
type VoiceChannelEffectSendEvent struct {
	ShardID       ShardID
	ChannelID     Snowflake                       `json:"channel_id"`
	GuildID       Snowflake                       `json:"guild_id"`
	UserID        Snowflake                       `json:"user_id"`
	Emoji         *Emoji                          `json:"emoji,omitempty"`
	AnimationType VoiceChannelEffectAnimationType `json:"animation_type,omitempty"`
	AnimationID   int                             `json:"animation_id,omitempty"`
	SoundID       Snowflake                       `json:"sound_id,omitempty"`
	SoundVolume   *float64                        `json:"sound_volume,omitempty"`
}

// VoiceServerUpdateEvent is dispatched when a guild's voice server changes.
type VoiceServerUpdateEvent struct {
	ShardID  ShardID
	Token    string    `json:"token"`
	GuildID  Snowflake `json:"guild_id"`
	Endpoint *string   `json:"endpoint"`
}

// WebhooksUpdateEvent is dispatched when a channel's webhooks are modified.
type WebhooksUpdateEvent struct {
	ShardID   ShardID
	GuildID   Snowflake `json:"guild_id"`
	ChannelID Snowflake `json:"channel_id"`
}

// MessagePollVoteAddEvent is dispatched when a user votes on a poll.
type MessagePollVoteAddEvent struct {
	ShardID   ShardID
	UserID    Snowflake `json:"user_id"`
	ChannelID Snowflake `json:"channel_id"`
	MessageID Snowflake `json:"message_id"`
	GuildID   Snowflake `json:"guild_id,omitempty"`
	AnswerID  int       `json:"answer_id"`
}

// MessagePollVoteRemoveEvent is dispatched when a user removes their vote on a poll.
type MessagePollVoteRemoveEvent struct {
	ShardID   ShardID
	UserID    Snowflake `json:"user_id"`
	ChannelID Snowflake `json:"channel_id"`
	MessageID Snowflake `json:"message_id"`
	GuildID   Snowflake `json:"guild_id,omitempty"`
	AnswerID  int       `json:"answer_id"`
}

// ChannelInfoChannel represents ephemeral data for a single channel in ChannelInfoEvent.
type ChannelInfoChannel struct {
	ID             Snowflake `json:"id"`
	Status         *string   `json:"status,omitempty"`
	VoiceStartTime *int64    `json:"voice_start_time,omitempty"`
}

// ChannelInfoEvent is dispatched in response to RequestChannelInfo.
type ChannelInfoEvent struct {
	ShardID  ShardID
	GuildID  Snowflake            `json:"guild_id"`
	Channels []ChannelInfoChannel `json:"channels"`
}

// VoiceChannelStatusUpdateEvent is dispatched when a voice channel's status changes.
type VoiceChannelStatusUpdateEvent struct {
	ShardID   ShardID
	ID        Snowflake `json:"id"`
	GuildID   Snowflake `json:"guild_id"`
	OldStatus *string   `json:"-"`
	NewStatus *string   `json:"status,omitempty"`
}

// VoiceChannelStartTimeUpdateEvent is dispatched when a voice channel's start time changes.
type VoiceChannelStartTimeUpdateEvent struct {
	ShardID           ShardID
	ID                Snowflake  `json:"id"`
	GuildID           Snowflake  `json:"guild_id"`
	OldVoiceStartTime *int64     `json:"-"`
	NewVoiceStartTime *int64     `json:"voice_start_time,omitempty"`
}

// ApplicationAuthorizedEvent is dispatched when an app was authorized by a user to a server or their account.
type ApplicationAuthorizedEvent struct {
	ShardID         ShardID                         `json:"-"`
	IntegrationType *ApplicationIntegrationType `json:"integration_type,omitempty"`
	User            User                        `json:"user"`
	Scopes          []OAuth2Scope               `json:"scopes"`
	Guild           *Guild                      `json:"guild,omitempty"`
}

// ApplicationDeauthorizedEvent is dispatched when an app was deauthorized by a user.
type ApplicationDeauthorizedEvent struct {
	ShardID ShardID  `json:"-"`
	User    User `json:"user"`
}

// RateLimitMeta contains metadata for a rate limited gateway opcode.
type RateLimitMeta struct {
	GuildID Snowflake `json:"guild_id,omitempty"`
	Nonce   *string   `json:"nonce,omitempty"`
}

// RateLimitedEvent is dispatched when an app encounters a gateway rate limit for an opcode.
type RateLimitedEvent struct {
	ShardID    ShardID       `json:"-"`
	Opcode     gatewayOpcode `json:"opcode"`
	RetryAfter float64       `json:"retry_after"`
	Meta       RateLimitMeta `json:"meta"`
}

// QuestUserEnrollmentEvent is dispatched when a user is added to a Quest on Discord.
type QuestUserEnrollmentEvent struct {
	ShardID ShardID   `json:"-"`
	UserID  Snowflake `json:"user_id"`
	QuestID Snowflake `json:"quest_id"`
}

// LobbyMessage represents a message sent in a lobby or linked channel.
type LobbyMessage struct {
	ID            Snowflake         `json:"id"`
	Type          MessageType       `json:"type"`
	Content       string            `json:"content"`
	LobbyID       Snowflake         `json:"lobby_id"`
	ChannelID     Snowflake         `json:"channel_id"`
	Author        User              `json:"author"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Flags         MessageFlags      `json:"flags,omitempty"`
	ApplicationID *Snowflake        `json:"application_id,omitempty"`
}

// LobbyMessageCreateEvent is dispatched when a message is created in a lobby.
type LobbyMessageCreateEvent struct {
	ShardID ShardID `json:"-"`
	LobbyMessage
}

// LobbyMessageUpdateEvent is dispatched when a message is updated in a lobby.
type LobbyMessageUpdateEvent struct {
	ShardID         ShardID    `json:"-"`
	LobbyMessage
	EditedTimestamp *time.Time `json:"edited_timestamp,omitempty"`
	Timestamp       *time.Time `json:"timestamp,omitempty"`
}

// LobbyMessageDeleteEvent is dispatched when a message is deleted from a lobby.
type LobbyMessageDeleteEvent struct {
	ShardID ShardID   `json:"-"`
	ID      Snowflake `json:"id"`
	LobbyID Snowflake `json:"lobby_id"`
}

// GameDirectMessageCreateEvent is dispatched when a direct message is created during an active Social SDK session.
type GameDirectMessageCreateEvent struct {
	ShardID       ShardID      `json:"-"`
	ID            Snowflake    `json:"id"`
	Type          MessageType  `json:"type"`
	Content       string       `json:"content"`
	ChannelID     Snowflake    `json:"channel_id"`
	Author        User         `json:"author"`
	Timestamp     *time.Time   `json:"timestamp,omitempty"`
	ApplicationID *Snowflake   `json:"application_id,omitempty"`
	Attachments   []Attachment `json:"attachments,omitempty"`
}

// GameDirectMessageUpdateEvent is dispatched when a direct message is updated during an active Social SDK session.
type GameDirectMessageUpdateEvent struct {
	ShardID         ShardID    `json:"-"`
	ID              Snowflake  `json:"id"`
	ChannelID       Snowflake  `json:"channel_id"`
	Content         string     `json:"content,omitempty"`
	Author          User       `json:"author"`
	RecipientID     Snowflake  `json:"recipient_id,omitempty"`
	EditedTimestamp *time.Time `json:"edited_timestamp,omitempty"`
}

// GameDirectMessageDeleteEvent is dispatched when a direct message is deleted during an active Social SDK session.
type GameDirectMessageDeleteEvent struct {
	ShardID     ShardID      `json:"-"`
	ID          Snowflake    `json:"id"`
	ChannelID   Snowflake    `json:"channel_id"`
	Author      User         `json:"author"`
	Timestamp   *time.Time   `json:"timestamp,omitempty"`
	Flags       MessageFlags `json:"flags,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Components  []Component  `json:"components,omitempty"`
}
