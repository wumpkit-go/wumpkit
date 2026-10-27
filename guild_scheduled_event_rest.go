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
	"net/url"
	"strconv"
	"time"
)

// ListGuildScheduledEventsOptions contains query parameters for listing scheduled events.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#list-scheduled-events-for-guild
type ListGuildScheduledEventsOptions struct {
	// WithUserCount includes number of users subscribed to each event.
	WithUserCount bool `json:"with_user_count,omitempty"`
}

// CreateGuildScheduledEventOptions contains parameters for creating a scheduled event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#create-guild-scheduled-event
type CreateGuildScheduledEventOptions struct {
	// ChannelID is the channel id of the scheduled event.
	ChannelID *Snowflake `json:"channel_id,omitempty"`

	// EntityMetadata is the entity metadata of the scheduled event.
	EntityMetadata *GuildScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`

	// Name is the name of the scheduled event (1-100 characters).
	Name string `json:"name"`

	// PrivacyLevel is the privacy level of the scheduled event.
	PrivacyLevel GuildScheduledEventPrivacyLevel `json:"privacy_level"`

	// ScheduledStartTime is the time to schedule the scheduled event.
	ScheduledStartTime time.Time `json:"scheduled_start_time"`

	// ScheduledEndTime is the time when the scheduled event is scheduled to end.
	ScheduledEndTime *time.Time `json:"scheduled_end_time,omitempty"`

	// Description is the description of the scheduled event (1-1000 characters).
	Description *string `json:"description,omitempty"`

	// EntityType is the entity type of the scheduled event.
	EntityType GuildScheduledEventEntityType `json:"entity_type"`

	// Image is the cover image of the scheduled event.
	Image *Base64Image `json:"image,omitempty"`

	// RecurrenceRule is the definition for how often this event should recur.
	RecurrenceRule *GuildScheduledEventRecurrenceRule `json:"recurrence_rule,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// ModifyGuildScheduledEventOptions contains parameters for modifying a scheduled event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#modify-guild-scheduled-event
type ModifyGuildScheduledEventOptions struct {
	// ChannelID is the channel id of the scheduled event. Pointer allows setting nil to clear.
	ChannelID *Snowflake `json:"channel_id,omitempty"`

	// EntityMetadata is the entity metadata of the scheduled event.
	EntityMetadata *GuildScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`

	// Name is the name of the scheduled event (1-100 characters). Zero value means not provided.
	Name string `json:"name,omitempty"`

	// PrivacyLevel is the privacy level of the scheduled event.
	PrivacyLevel *GuildScheduledEventPrivacyLevel `json:"privacy_level,omitempty"`

	// ScheduledStartTime is the time to schedule the scheduled event.
	ScheduledStartTime *time.Time `json:"scheduled_start_time,omitempty"`

	// ScheduledEndTime is the time when the scheduled event is scheduled to end.
	ScheduledEndTime *time.Time `json:"scheduled_end_time,omitempty"`

	// Description is the description of the scheduled event (1-1000 characters). Pointer allows setting nil to clear.
	Description *string `json:"description,omitempty"`

	// EntityType is the entity type of the scheduled event.
	EntityType *GuildScheduledEventEntityType `json:"entity_type,omitempty"`

	// Status is the status of the scheduled event.
	Status *GuildScheduledEventStatus `json:"status,omitempty"`

	// Image is the cover image of the scheduled event. Pointer allows setting nil to clear.
	Image *Base64Image `json:"image,omitempty"`

	// RecurrenceRule is the definition for how often this event should recur.
	RecurrenceRule *GuildScheduledEventRecurrenceRule `json:"recurrence_rule,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// FetchGuildScheduledEventUsersOptions contains parameters for fetching users subscribed to an event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#get-guild-scheduled-event-users
type FetchGuildScheduledEventUsersOptions struct {
	// Limit is the number of users to return (up to maximum 100, default 100).
	Limit int `json:"limit,omitempty"`

	// WithMember includes guild member data if it exists.
	WithMember bool `json:"with_member,omitempty"`

	// Before considers only users before given user id.
	Before Snowflake `json:"before,omitempty"`

	// After considers only users after given user id.
	After Snowflake `json:"after,omitempty"`
}

// ListGuildScheduledEvents returns a list of scheduled event objects for the given guild.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#list-scheduled-events-for-guild
func (r *requester) ListGuildScheduledEvents(
	ctx context.Context,
	guildID Snowflake,
	opts ...ListGuildScheduledEventsOptions,
) ([]GuildScheduledEvent, error) {
	params := url.Values{}
	if len(opts) > 0 && opts[0].WithUserCount {
		params.Set("with_user_count", "true")
	}

	endpoint := "/guilds/" + guildID.String() + "/scheduled-events"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var events []GuildScheduledEvent
	if err := json.NewDecoder(body).Decode(&events); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/scheduled-events",
				"error", err.Error(),
			)
		return nil, err
	}
	for i := range events {
		events[i].GuildID = guildID
	}
	return events, nil
}

// CreateGuildScheduledEvent creates a guild scheduled event in the guild.
//
// Requires the PermissionCreateEvents permission.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#create-guild-scheduled-event
func (r *requester) CreateGuildScheduledEvent(
	ctx context.Context,
	guildID Snowflake,
	opts CreateGuildScheduledEventOptions,
) (GuildScheduledEvent, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/" + guildID.String() + "/scheduled-events",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return GuildScheduledEvent{}, err
	}
	defer body.Close()

	var event GuildScheduledEvent
	if err := json.NewDecoder(body).Decode(&event); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/guilds/{id}/scheduled-events",
				"error", err.Error(),
			)
		return GuildScheduledEvent{}, err
	}
	event.GuildID = guildID
	return event, nil
}

// FetchGuildScheduledEvent returns a guild scheduled event by ID.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#get-guild-scheduled-event
func (r *requester) FetchGuildScheduledEvent(
	ctx context.Context,
	guildID, eventID Snowflake,
	withUserCount bool,
) (GuildScheduledEvent, error) {
	endpoint := "/guilds/" + guildID.String() + "/scheduled-events/" + eventID.String()
	if withUserCount {
		endpoint += "?with_user_count=true"
	}

	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
	})
	if err != nil {
		return GuildScheduledEvent{}, err
	}
	defer body.Close()

	var event GuildScheduledEvent
	if err := json.NewDecoder(body).Decode(&event); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/scheduled-events/{id}",
				"error", err.Error(),
			)
		return GuildScheduledEvent{}, err
	}
	event.GuildID = guildID
	return event, nil
}

// ModifyGuildScheduledEvent modifies a guild scheduled event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#modify-guild-scheduled-event
func (r *requester) ModifyGuildScheduledEvent(
	ctx context.Context,
	guildID, eventID Snowflake,
	opts ModifyGuildScheduledEventOptions,
) (GuildScheduledEvent, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/scheduled-events/" + eventID.String(),
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return GuildScheduledEvent{}, err
	}
	defer body.Close()

	var event GuildScheduledEvent
	if err := json.NewDecoder(body).Decode(&event); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/scheduled-events/{id}",
				"error", err.Error(),
			)
		return GuildScheduledEvent{}, err
	}
	event.GuildID = guildID
	return event, nil
}

// DeleteGuildScheduledEvent deletes a guild scheduled event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#delete-guild-scheduled-event
func (r *requester) DeleteGuildScheduledEvent(
	ctx context.Context,
	guildID, eventID Snowflake,
) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/guilds/" + guildID.String() + "/scheduled-events/" + eventID.String(),
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// FetchGuildScheduledEventUsers gets a list of users subscribed to a guild scheduled event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#get-guild-scheduled-event-users
func (r *requester) FetchGuildScheduledEventUsers(
	ctx context.Context,
	guildID, eventID Snowflake,
	opts ...FetchGuildScheduledEventUsersOptions,
) ([]GuildScheduledEventUser, error) {
	var first FetchGuildScheduledEventUsersOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	params := url.Values{}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if first.WithMember {
		params.Set("with_member", "true")
	}
	if first.Before != 0 {
		params.Set("before", first.Before.String())
	}
	if first.After != 0 {
		params.Set("after", first.After.String())
	}

	endpoint := "/guilds/" + guildID.String() + "/scheduled-events/" + eventID.String() + "/users"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var users []GuildScheduledEventUser
	if err := json.NewDecoder(body).Decode(&users); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/guilds/{id}/scheduled-events/{id}/users",
				"error", err.Error(),
			)
		return nil, err
	}
	return users, nil
}
