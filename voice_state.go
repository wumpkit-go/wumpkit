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

// VoiceState represents a user's voice connection status in a guild.
//
// Reference: https://discord.com/developers/docs/resources/voice#voice-state-object
type VoiceState struct {
	// GuildID is the ID of the guild this voice state is for.
	GuildID Snowflake `json:"guild_id"`

	// ChannelID is the ID of the voice channel the user is connected to.
	ChannelID Snowflake `json:"channel_id,omitempty"`

	// UserID is the ID of the user this voice state is for.
	UserID Snowflake `json:"user_id"`

	// SessionID is the session identifier for this voice state.
	SessionID string `json:"session_id"`

	// GuildDeaf indicates whether the user is deafened by the server.
	GuildDeaf bool `json:"deaf"`

	// GuildMute indicates whether the user is muted by the server.
	GuildMute bool `json:"mute"`

	// SelfDeaf indicates whether the user is locally deafened.
	SelfDeaf bool `json:"self_deaf"`

	// SelfMute indicates whether the user is locally muted.
	SelfMute bool `json:"self_mute"`

	// SelfStream indicates whether the user is streaming using "Go Live".
	SelfStream bool `json:"self_stream,omitempty"`

	// SelfVideo indicates whether the user's camera is enabled.
	SelfVideo bool `json:"self_video"`

	// Suppress indicates whether the user's permission to speak is denied.
	Suppress bool `json:"suppress"`

	// RequestToSpeakTimestamp is the time at which the user requested to speak.
	//
	// Optional:
	//  - May be nil if the user has not requested to speak.
	RequestToSpeakTimestamp *time.Time `json:"request_to_speak_timestamp,omitempty"`
}

type GatewayVoiceState struct {
	VoiceState

	Member FullMember `json:"member"`
}

func (g *GatewayVoiceState) UnmarshalJSON(buf []byte) error {
	type localState GatewayVoiceState
	var v localState
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	*g = GatewayVoiceState(v)
	if g.GuildID != 0 {
		g.Member.GuildID = g.GuildID
	}
	if g.UserID != 0 {
		g.Member.ID = g.UserID
	} else if g.Member.User.ID != 0 {
		g.Member.ID = g.Member.User.ID
	}
	return nil
}

// VoiceRegion represents a Discord voice region.
//
// Reference: https://discord.com/developers/docs/resources/voice#voice-region-object
type VoiceRegion struct {
	// ID is the unique identifier for the voice region.
	ID string `json:"id"`

	// Name is the name of the voice region.
	Name string `json:"name"`

	// Optimal indicates whether this region is optimal for the current user.
	//
	// True if this is the single server closest to the current user's client.
	Optimal bool `json:"optimal"`

	// Deprecated indicates whether this voice region is deprecated.
	//
	// Avoid switching to these regions.
	Deprecated bool `json:"deprecated"`

	// Custom indicates whether this is a custom voice region.
	//
	// Used for special events or similar use cases.
	Custom bool `json:"custom"`
}
