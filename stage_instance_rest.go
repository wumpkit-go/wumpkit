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

// CreateStageInstanceOptions contains parameters for creating a new stage instance.
//
// Reference: https://discord.com/developers/docs/resources/stage-instance#create-stage-instance
type CreateStageInstanceOptions struct {
	// ChannelID is the id of the Stage channel.
	ChannelID Snowflake `json:"channel_id"`

	// Topic is the topic of the Stage instance (1-120 characters).
	Topic string `json:"topic"`

	// PrivacyLevel is the privacy level of the Stage instance (default StagePrivacyLevelGuildOnly).
	PrivacyLevel *StagePrivacyLevel `json:"privacy_level,omitempty"`

	// SendStartNotification notifies @everyone that a Stage instance has started.
	SendStartNotification *bool `json:"send_start_notification,omitempty"`

	// GuildScheduledEventID is the guild scheduled event associated with this Stage instance.
	GuildScheduledEventID *Snowflake `json:"guild_scheduled_event_id,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// ModifyStageInstanceOptions contains parameters for modifying a stage instance.
//
// Reference: https://discord.com/developers/docs/resources/stage-instance#modify-stage-instance
type ModifyStageInstanceOptions struct {
	// Topic is the topic of the Stage instance (1-120 characters). Zero value means not provided.
	Topic string `json:"topic,omitempty"`

	// PrivacyLevel is the privacy level of the Stage instance.
	PrivacyLevel *StagePrivacyLevel `json:"privacy_level,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// CreateStageInstance creates a new Stage instance associated with a Stage channel.
//
// Requires the user to be a moderator of the Stage channel.
//
// Reference: https://discord.com/developers/docs/resources/stage-instance#create-stage-instance
func (r *requester) CreateStageInstance(ctx context.Context, opts CreateStageInstanceOptions) (StageInstance, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/stage-instances",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return StageInstance{}, err
	}
	defer body.Close()

	var stage StageInstance
	if err := json.NewDecoder(body).Decode(&stage); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "POST",
			"url", "/stage-instances",
			"error", err.Error(),
		)
		return StageInstance{}, err
	}
	return stage, nil
}

// FetchStageInstance gets the stage instance associated with the Stage channel, if it exists.
//
// Reference: https://discord.com/developers/docs/resources/stage-instance#get-stage-instance
func (r *requester) FetchStageInstance(ctx context.Context, channelID Snowflake) (StageInstance, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/stage-instances/" + channelID.String(),
	})
	if err != nil {
		return StageInstance{}, err
	}
	defer body.Close()

	var stage StageInstance
	if err := json.NewDecoder(body).Decode(&stage); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/stage-instances/{id}",
				"error", err.Error(),
			)
		return StageInstance{}, err
	}
	return stage, nil
}

// ModifyStageInstance updates fields of an existing Stage instance.
//
// Requires the user to be a moderator of the Stage channel.
//
// Reference: https://discord.com/developers/docs/resources/stage-instance#modify-stage-instance
func (r *requester) ModifyStageInstance(ctx context.Context, channelID Snowflake, opts ModifyStageInstanceOptions) (StageInstance, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/stage-instances/" + channelID.String(),
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return StageInstance{}, err
	}
	defer body.Close()

	var stage StageInstance
	if err := json.NewDecoder(body).Decode(&stage); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/stage-instances/{id}",
				"error", err.Error(),
			)
		return StageInstance{}, err
	}
	return stage, nil
}

// DeleteStageInstance deletes the Stage instance.
//
// Requires the user to be a moderator of the Stage channel.
//
// Reference: https://discord.com/developers/docs/resources/stage-instance#delete-stage-instance
func (r *requester) DeleteStageInstance(ctx context.Context, channelID Snowflake, reason string) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/stage-instances/" + channelID.String(),
		Reason: reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}
