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
)

// GatewayIntent represents Discord Gateway Intents.
//
// Intents are bit flags that specify which events your bot receives over the WebSocket connection.
// Combine multiple intents using bitwise OR (|).
//
// Example:
//
//	intents := GatewayIntent_Guilds | GatewayIntent_GuildMessages
type GatewayIntent uint32

const (
	// Guilds includes:
	//   GuildCreate, GuildUpdate, GuildDelete
	//   GuildRoleCreate, GuildRoleUpdate, GuildRoleDelete
	//   ChannelCreate, ChannelUpdate, ChannelDelete, ChannelPinsUpdate
	//   ThreadCreate, ThreadUpdate, ThreadDelete, ThreadListSync
	//   ThreadMemberUpdate, ThreadMembersUpdate
	//   StageInstanceCreate, StageInstanceUpdate, StageInstanceDelete
	GatewayIntentGuilds GatewayIntent = 1 << 0

	// GuildMembers includes:
	//   GuildMemberAdd, GuildMemberUpdate, GuildMemberRemove
	//   ThreadMembersUpdate
	GatewayIntentGuildMembers GatewayIntent = 1 << 1

	// GuildModeration includes:
	//   GuildAuditLogEntryCreate, GuildBanAdd, GuildBanRemove
	GatewayIntentGuildModeration GatewayIntent = 1 << 2

	// GuildExpressions includes:
	//   GuildEmojisUpdate, GuildStickersUpdate
	//   GuildSoundboardSoundCreate, GuildSoundboardSoundUpdate, GuildSoundboardSoundDelete
	//   GuildSoundboardSoundsUpdate
	GatewayIntentGuildExpressions GatewayIntent = 1 << 3

	// GuildIntegrations includes:
	//   GuildIntegrationsUpdate, IntegrationCreate, IntegrationUpdate, IntegrationDelete
	GatewayIntentGuildIntegrations GatewayIntent = 1 << 4

	// GuildWebhooks includes:
	//   WebhooksUpdate
	GatewayIntentGuildWebhooks GatewayIntent = 1 << 5

	// GuildInvites includes:
	//   InviteCreate, InviteDelete
	GatewayIntentGuildInvites GatewayIntent = 1 << 6

	// GuildVoiceStates includes:
	//   VoiceChannelEffectSend, VoiceStateUpdate
	GatewayIntentGuildVoiceStates GatewayIntent = 1 << 7

	// GuildPresences includes:
	//   PresenceUpdate
	GatewayIntentGuildPresences GatewayIntent = 1 << 8

	// GuildMessages includes:
	//   MessageCreate, MessageUpdate, MessageDelete, MessageDeleteBulk
	GatewayIntentGuildMessages GatewayIntent = 1 << 9

	// GuildMessageReactions includes:
	//   MessageReactionAdd, MessageReactionRemove, MessageReactionRemoveAll, MessageReactionRemoveEmoji
	GatewayIntentGuildMessageReactions GatewayIntent = 1 << 10

	// GuildMessageTyping includes:
	//   TypingStart
	GatewayIntentGuildMessageTyping GatewayIntent = 1 << 11

	// DirectMessages includes:
	//   MessageCreate, MessageUpdate, MessageDelete, ChannelPinsUpdate
	GatewayIntentDirectMessages GatewayIntent = 1 << 12

	// DirectMessageReactions includes:
	//   MessageReactionAdd, MessageReactionRemove, MessageReactionRemoveAll, MessageReactionRemoveEmoji
	GatewayIntentDirectMessageReactions GatewayIntent = 1 << 13

	// DirectMessageTyping includes:
	//   TypingStart
	GatewayIntentDirectMessageTyping GatewayIntent = 1 << 14

	// MessageContent enables access to message content in events.
	GatewayIntentMessageContent GatewayIntent = 1 << 15

	// GuildScheduledEvents includes:
	//   GuildScheduledEventCreate, GuildScheduledEventUpdate, GuildScheduledEventDelete
	//   GuildScheduledEventUserAdd, GuildScheduledEventUserRemove
	GatewayIntentGuildScheduledEvents GatewayIntent = 1 << 16

	// AutoModerationConfiguration includes:
	//   AutoModerationRuleCreate, AutoModerationRuleUpdate, AutoModerationRuleDelete
	GatewayIntentAutoModerationConfiguration GatewayIntent = 1 << 20

	// AutoModerationExecution includes:
	//   AutoModerationActionExecution
	GatewayIntentAutoModerationExecution GatewayIntent = 1 << 21

	// GuildMessagePolls includes:
	//   MessagePollVoteAdd, MessagePollVoteRemove
	GatewayIntentGuildMessagePolls GatewayIntent = 1 << 24

	// DirectMessagePolls includes:
	//   MessagePollVoteAdd, MessagePollVoteRemove
	GatewayIntentDirectMessagePolls GatewayIntent = 1 << 25
)

// GatewayCompression defines the transport compression mode for Gateway WebSocket connections.
type GatewayCompression string

const (
	// GatewayCompressionNone disables transport compression (uses plaintext JSON frames).
	GatewayCompressionNone GatewayCompression = "none"

	// GatewayCompressionZlibStream enables shared zlib stream compression (default, recommended).
	GatewayCompressionZlibStream GatewayCompression = "zlib-stream"

	// GatewayCompressionZstdStream enables zstd stream compression.
	GatewayCompressionZstdStream GatewayCompression = "zstd-stream"
)

// Is returns true if this GatewayCompression matches the provided one.
func (c GatewayCompression) Is(other GatewayCompression) bool {
	return c == other
}

// gatewayOpcode represents the operation codes used in Discord Gateway WebSocket frames.
//
// Each opcode defines a specific action or message type in the client-server communication.
type gatewayOpcode int

const (
	// Dispatch
	//
	//	- Client Action: Receive.
	// 	- Description: An event was dispatched by the gateway.
	gatewayOpcodeDispatch gatewayOpcode = 0

	// Heartbeat
	//
	//	- Client Action: Send/Receive.
	// 	- Description: Fired periodically by the client to keep the connection alive.
	gatewayOpcodeHeartbeat gatewayOpcode = 1

	// Identify
	//
	//  - Client Action: Send.
	//  - Description: Starts a new session during the initial handshake.
	gatewayOpcodeIdentify gatewayOpcode = 2

	// PresenceUpdate
	//
	//	- Client Action: Send.
	// 	- Description: Update the client's presence.
	gatewayOpcodePresenceUpdate gatewayOpcode = 3

	// VoiceStateUpdate
	//
	//	- Client Action: Send.
	// 	- Description: Used to join, leave, or move between voice channels.
	gatewayOpcodeVoiceStateUpdate gatewayOpcode = 4

	// Resume
	//
	//	- Client Action: Send.
	// 	- Description: Resume a previous session that was disconnected.
	gatewayOpcodeResume gatewayOpcode = 6

	// Reconnect
	//
	//	- Client Action: Receive.
	// 	- Description: Server requests the client to reconnect and resume immediately.
	gatewayOpcodeReconnect gatewayOpcode = 7

	// RequestGuildMembers
	//
	//	- Client Action: Send.
	// 	- Description: Request information about offline guild members in a large guild.
	gatewayOpcodeRequestGuildMembers gatewayOpcode = 8

	// InvalidSession
	//
	//	- Client Action: Receive.
	// 	- Description: The session has been invalidated. Client must reconnect and identify or resume accordingly.
	gatewayOpcodeInvalidSession gatewayOpcode = 9

	// Hello
	//
	//	- Client Action: Receive.
	// 	- Description: Sent immediately after connecting. Contains the heartbeat_interval to use.
	gatewayOpcodeHello gatewayOpcode = 10

	// HeartbeatACK
	//
	//	- Client Action: Receive.
	// 	- Description: Sent in response to a heartbeat to acknowledge that it has been received.
	gatewayOpcodeHeartbeatACK gatewayOpcode = 11

	// RequestSoundboardSounds
	//
	//	- Client Action: Send.
	//	- Description: Request information about soundboard sounds in a set of guilds.
	gatewayOpcodeRequestSoundboardSounds gatewayOpcode = 31

	// RequestChannelInfo
	//
	//	- Client Action: Send.
	//	- Description: Request ephemeral channel data for channels in a guild.
	gatewayOpcodeRequestChannelInfo gatewayOpcode = 43
)

// Is returns true if this gatewayOpcode matches the provided one.
func (v gatewayOpcode) Is(other gatewayOpcode) bool {
	return v == other
}

// ChannelInfoField represents an ephemeral channel data field that can be requested via Opcode 43 (Request Channel Info).
//
// Reference: https://discord.com/developers/docs/events/gateway-events#request-channel-info-request-channel-info-structure
type ChannelInfoField string

const (
	// ChannelInfoFieldStatus requests the status text for voice channels in the guild.
	ChannelInfoFieldStatus ChannelInfoField = "status"

	// ChannelInfoFieldVoiceStartTime requests the start timestamp of current active voice sessions.
	ChannelInfoFieldVoiceStartTime ChannelInfoField = "voice_start_time"
)

// Is returns true if this ChannelInfoField matches the provided one.
func (f ChannelInfoField) Is(other ChannelInfoField) bool {
	return f == other
}

// RequestChannelInfoOptions contains parameters for requesting ephemeral channel data.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#request-channel-info
type RequestChannelInfoOptions struct {
	// GuildID is the guild id to request channel info for.
	GuildID Snowflake `json:"guild_id"`
	// Fields is the list of fields to request (e.g. ChannelInfoFieldStatus, ChannelInfoFieldVoiceStartTime).
	Fields []ChannelInfoField `json:"fields"`
}

// gatewayPayload represents a single payload exchanged over the Discord Gateway WebSocket.
//
// Fields:
//   - op: Operation code indicating the type of payload (e.g., Dispatch, Heartbeat).
//   - d: Raw JSON-encoded event data or payload data.
//   - s: Sequence number of the event; only provided when 'op' is Dispatch (0).
//   - t: Event name; only provided when 'op' is Dispatch (0).
type gatewayPayload struct {
	Op gatewayOpcode   `json:"op"` // Operation code of the payload.
	D  json.RawMessage `json:"d"`  // Raw JSON payload data.
	S  int64           `json:"s"`  // Sequence number; present only if op == 0 (Dispatch).
	T  string          `json:"t"`  // Event name; present only if op == 0 (Dispatch).
}

// GatewayCloseEventCode represents Discord Gateway close event codes.
type GatewayCloseEventCode int

const (
	// UnknownError
	//
	//  - Explanation: We're not sure what went wrong. Try reconnecting?
	//  - Reconnect: true.
	GatewayCloseEventCodeUnknownError GatewayCloseEventCode = 4000

	// UnknownOpcode
	//
	//  - Explanation: You sent an invalid Gateway opcode or an invalid payload for an opcode. Don't do that!
	//  - Reconnect: true.
	GatewayCloseEventCodeUnknownOpcode GatewayCloseEventCode = 4001

	// DecodeError
	//
	//  - Explanation: You sent an invalid payload to Discord. Don't do that!
	//  - Reconnect: true.
	GatewayCloseEventCodeDecodeError GatewayCloseEventCode = 4002

	// NotAuthenticated
	//
	//  - Explanation: You sent a payload prior to identifying, or this session has been invalidated.
	//  - Reconnect: true.
	GatewayCloseEventCodeNotAuthenticated GatewayCloseEventCode = 4003

	// AuthenticationFailed
	//
	//  - Explanation: The account token sent with your identify payload is incorrect.
	//  - Reconnect: false.
	GatewayCloseEventCodeAuthenticationFailed GatewayCloseEventCode = 4004

	// AlreadyAuthenticated
	//
	//  - Explanation: You sent more than one identify payload. Don't do that!
	//  - Reconnect: true.
	GatewayCloseEventCodeAlreadyAuthenticated GatewayCloseEventCode = 4005

	// InvalidSeq
	//
	//  - Explanation: The sequence sent when resuming the session was invalid. Reconnect and start a new session.
	//  - Reconnect: true.
	GatewayCloseEventCodeInvalidSeq GatewayCloseEventCode = 4007

	// RateLimited
	//
	//  - Explanation: You're sending payloads too quickly. Slow down! You will be disconnected on receiving this.
	//  - Reconnect: true.
	GatewayCloseEventCodeRateLimited GatewayCloseEventCode = 4008

	// SessionTimedOut
	//
	//  - Explanation: Your session timed out. Reconnect and start a new one.
	//  - Reconnect: true.
	GatewayCloseEventCodeSessionTimedOut GatewayCloseEventCode = 4009

	// InvalidShard
	//
	//  - Explanation: You sent an invalid shard when identifying.
	//  - Reconnect: false.
	GatewayCloseEventCodeInvalidShard GatewayCloseEventCode = 4010

	// ShardingRequired
	//
	//  - Explanation: The session would have handled too many guilds - sharding is required.
	//  - Reconnect: false.
	GatewayCloseEventCodeShardingRequired GatewayCloseEventCode = 4011

	// InvalidAPIVersion
	//
	//  - Explanation: You sent an invalid version for the gateway.
	//  - Reconnect: false.
	GatewayCloseEventCodeInvalidAPIVersion GatewayCloseEventCode = 4012

	// InvalidIntents
	//
	//  - Explanation: You sent an invalid intent for a Gateway Intent. You may have incorrectly calculated the bitwise value.
	//  - Reconnect: false.
	GatewayCloseEventCodeInvalidIntents GatewayCloseEventCode = 4013

	// DisallowedIntents
	//
	//  - Explanation: You sent a disallowed intent for a Gateway Intent. You may have tried to specify an intent you are not approved for.
	//  - Reconnect: false.
	GatewayCloseEventCodeDisallowedIntents GatewayCloseEventCode = 4014
)

// Is returns true if this GatewayCloseEventCode matches the provided one.
func (v GatewayCloseEventCode) Is(other GatewayCloseEventCode) bool {
	return v == other
}

// GatewayBot is Discord Gateway Bot.
type GatewayBot struct {
	// WSS URL that can be used for connecting to the Gateway
	URL string `json:"url"`
	// Recommended number of shards to use when connecting
	Shards int `json:"shards"`
	// Information on the current session start limit
	SessionStartLimit struct {
		Total          int `json:"total"`
		Remaining      int `json:"remaining"`
		ResetAfter     int `json:"reset_after"`
		MaxConcurrency int `json:"max_concurrency"`
	} `json:"session_start_limit"`
}

// GatewayPresenceUpdate contains parameters for updating client presence.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#update-presence
type GatewayPresenceUpdate struct {
	// Since is unix time (in milliseconds) of when the client went idle, or nil if not idle.
	Since *int64 `json:"since"`

	// Activities is the user's activities.
	Activities []Activity `json:"activities"`

	// Status is the user's new status.
	Status UserStatus `json:"status"`

	// AFK is whether or not the client is afk.
	AFK bool `json:"afk"`
}

// GatewayVoiceStateUpdate contains parameters for updating client voice state (join/move/leave voice).
//
// Reference: https://discord.com/developers/docs/events/gateway-events#update-voice-state
type GatewayVoiceStateUpdate struct {
	// GuildID is the id of the guild.
	GuildID Snowflake `json:"guild_id"`

	// ChannelID is the id of the voice channel client wants to join (nil if disconnecting).
	ChannelID *Snowflake `json:"channel_id"`

	// SelfMute is whether the client is muted.
	SelfMute bool `json:"self_mute"`

	// SelfDeaf is whether the client is deafened.
	SelfDeaf bool `json:"self_deaf"`
}

// RequestGuildMembersOptions contains parameters for requesting guild members.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#request-guild-members
type RequestGuildMembersOptions struct {
	// GuildID is the id of the guild to get members for.
	GuildID Snowflake `json:"guild_id"`

	// Query is string that username starts with, or an empty string to return all members.
	Query string `json:"query,omitempty"`

	// Limit is maximum number of members to send matching the query (0 to return all members).
	Limit int `json:"limit"`

	// Presences specifies if we want presences of the matched members.
	Presences bool `json:"presences,omitempty"`

	// UserIDs specifies which users you wish to fetch.
	UserIDs []Snowflake `json:"user_ids,omitempty"`

	// Nonce is nonce to identify the Guild Members Chunk response.
	Nonce string `json:"nonce,omitempty"`
}

// RequestSoundboardSoundsOptions contains parameters for requesting soundboard sounds for a list of guilds.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#request-soundboard-sounds
type RequestSoundboardSoundsOptions struct {
	// GuildIDs is the list of guild IDs to get soundboard sounds for.
	GuildIDs []Snowflake `json:"guild_ids"`
}

// Has returns true if the flags contain all of the provided flags.
func (f GatewayIntent) Has(flags ...GatewayIntent) bool {
	return BitFieldHas(f, flags...)
}
