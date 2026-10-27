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

// SendSoundboardSoundOptions contains parameters for playing a soundboard sound in a voice channel.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#send-soundboard-sound
type SendSoundboardSoundOptions struct {
	// SoundID is the id of the soundboard sound to play.
	SoundID Snowflake `json:"sound_id"`

	// SourceGuildID is the id of the guild the soundboard sound is from (required to play sounds from different servers).
	SourceGuildID Snowflake `json:"source_guild_id,omitempty"`
}

// GuildSoundboardSoundsResponse represents the response when listing guild soundboard sounds.
type GuildSoundboardSoundsResponse struct {
	// Items is the list of soundboard sounds.
	Items []SoundBoardSound `json:"items"`
}

// CreateGuildSoundboardSoundOptions contains parameters for creating a new guild soundboard sound.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#create-guild-soundboard-sound
type CreateGuildSoundboardSoundOptions struct {
	// Name is the name of the soundboard sound (2-32 characters).
	Name string `json:"name"`

	// Sound is the mp3 or ogg sound data, base64 encoded data URI.
	Sound string `json:"sound"`

	// Volume is the volume of the soundboard sound, from 0 to 1 (defaults to 1). Pointer allows setting 0.
	Volume *float64 `json:"volume,omitempty"`

	// EmojiID is the id of the custom emoji for the soundboard sound.
	EmojiID *Snowflake `json:"emoji_id,omitempty"`

	// EmojiName is the unicode character of a standard emoji for the soundboard sound.
	EmojiName *string `json:"emoji_name,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// ModifyGuildSoundboardSoundOptions contains parameters for modifying a guild soundboard sound.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#modify-guild-soundboard-sound
type ModifyGuildSoundboardSoundOptions struct {
	// Name is the name of the soundboard sound (2-32 characters). Zero value means not provided.
	Name string `json:"name,omitempty"`

	// Volume is the volume of the soundboard sound, from 0 to 1. Pointer allows setting 0.
	Volume *float64 `json:"volume,omitempty"`

	// EmojiID is the id of the custom emoji for the soundboard sound. Pointer allows setting nil to clear.
	EmojiID *Snowflake `json:"emoji_id,omitempty"`

	// EmojiName is the unicode character of a standard emoji for the soundboard sound. Pointer allows setting nil to clear.
	EmojiName *string `json:"emoji_name,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// SendSoundboardSound sends a soundboard sound to a voice channel the user is connected to.
//
// Requires the PermissionSpeak and PermissionUseSoundboard permissions.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#send-soundboard-sound
func (r *requester) SendSoundboardSound(ctx context.Context, channelID Snowflake, opts SendSoundboardSoundOptions) error {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/channels/" + channelID.String() + "/send-soundboard-sound",
		Body:   reqBody,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// ListDefaultSoundboardSounds returns an array of default soundboard sound objects available to all users.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#list-default-soundboard-sounds
func (r *requester) ListDefaultSoundboardSounds(ctx context.Context) ([]SoundBoardSound, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/soundboard-default-sounds",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var sounds []SoundBoardSound
	if err := json.NewDecoder(body).Decode(&sounds); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/soundboard-default-sounds",
			"error", err.Error(),
		)
		return nil, err
	}
	return sounds, nil
}

// ListGuildSoundboardSounds returns a list of the guild's soundboard sounds.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#list-guild-soundboard-sounds
func (r *requester) ListGuildSoundboardSounds(ctx context.Context, guildID Snowflake) ([]SoundBoardSound, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/soundboard-sounds",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp GuildSoundboardSoundsResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/soundboard-sounds",
				"error", err.Error(),
			)
		return nil, err
	}
	for i := range resp.Items {
		resp.Items[i].GuildID = guildID
	}
	return resp.Items, nil
}

// FetchGuildSoundboardSound returns a soundboard sound object for the given sound ID.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#get-guild-soundboard-sound
func (r *requester) FetchGuildSoundboardSound(ctx context.Context, guildID, soundID Snowflake) (SoundBoardSound, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/soundboard-sounds/" + soundID.String(),
	})
	if err != nil {
		return SoundBoardSound{}, err
	}
	defer body.Close()

	var sound SoundBoardSound
	if err := json.NewDecoder(body).Decode(&sound); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/soundboard-sounds/{id}",
				"error", err.Error(),
			)
		return SoundBoardSound{}, err
	}
	sound.GuildID = guildID
	return sound, nil
}

// CreateGuildSoundboardSound creates a new soundboard sound for the guild.
//
// Requires the PermissionCreateGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#create-guild-soundboard-sound
func (r *requester) CreateGuildSoundboardSound(ctx context.Context, guildID Snowflake, opts CreateGuildSoundboardSoundOptions) (SoundBoardSound, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/" + guildID.String() + "/soundboard-sounds",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return SoundBoardSound{}, err
	}
	defer body.Close()

	var sound SoundBoardSound
	if err := json.NewDecoder(body).Decode(&sound); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/guilds/{id}/soundboard-sounds",
				"error", err.Error(),
			)
		return SoundBoardSound{}, err
	}
	sound.GuildID = guildID
	return sound, nil
}

// ModifyGuildSoundboardSound modifies the given soundboard sound in a guild.
//
// Requires the PermissionManageGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#modify-guild-soundboard-sound
func (r *requester) ModifyGuildSoundboardSound(ctx context.Context, guildID, soundID Snowflake, opts ModifyGuildSoundboardSoundOptions) (SoundBoardSound, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/soundboard-sounds/" + soundID.String(),
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return SoundBoardSound{}, err
	}
	defer body.Close()

	var sound SoundBoardSound
	if err := json.NewDecoder(body).Decode(&sound); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/soundboard-sounds/{id}",
				"error", err.Error(),
			)
		return SoundBoardSound{}, err
	}
	sound.GuildID = guildID
	return sound, nil
}

// DeleteGuildSoundboardSound deletes the given soundboard sound from a guild.
//
// Requires the PermissionManageGuildExpressions permission.
//
// Reference: https://discord.com/developers/docs/resources/soundboard#delete-guild-soundboard-sound
func (r *requester) DeleteGuildSoundboardSound(ctx context.Context, guildID, soundID Snowflake, reason string) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/guilds/" + guildID.String() + "/soundboard-sounds/" + soundID.String(),
		Reason: reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}
