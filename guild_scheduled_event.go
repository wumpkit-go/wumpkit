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
	"fmt"
	"time"
)

// GuildScheduledEventPrivacyLevel represents Guild Scheduled Event Privacy Levels types.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-guild-scheduled-event-privacy-level
type GuildScheduledEventPrivacyLevel int

const (
	// GuildScheduledEventPrivacyLevelGuildOnly indicates the scheduled event is only accessible to guild members.
	GuildScheduledEventPrivacyLevelGuildOnly GuildScheduledEventPrivacyLevel = 2
)

// Is returns true if the guild scheduled event's level matches the provided one.
func (l GuildScheduledEventPrivacyLevel) Is(privacyLevel GuildScheduledEventPrivacyLevel) bool {
	return l == privacyLevel
}

// GuildScheduledEventEntityType represents entity types for a scheduled event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-guild-scheduled-event-entity-types
type GuildScheduledEventEntityType int

const (
	GuildScheduledEventEntityTypeStageInstance GuildScheduledEventEntityType = 1
	GuildScheduledEventEntityTypeVoice         GuildScheduledEventEntityType = 2
	GuildScheduledEventEntityTypeExternal      GuildScheduledEventEntityType = 3
)

// Is returns true if the guild scheduled event entity's Type matches the provided one.
func (t GuildScheduledEventEntityType) Is(typ GuildScheduledEventEntityType) bool {
	return t == typ
}

// GuildScheduledEventStatus represents the status of a guild scheduled event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-guild-scheduled-event-status
type GuildScheduledEventStatus int

const (
	GuildScheduledEventStatusScheduled GuildScheduledEventStatus = 1
	GuildScheduledEventStatusActive    GuildScheduledEventStatus = 2
	GuildScheduledEventStatusCompleted GuildScheduledEventStatus = 3
	GuildScheduledEventStatusCanceled  GuildScheduledEventStatus = 4
)

// Is returns true if the status matches the provided one.
func (s GuildScheduledEventStatus) Is(status GuildScheduledEventStatus) bool {
	return s == status
}

// GuildScheduledEventEntityMetadata represents additional metadata for external events.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-guild-scheduled-event-entity-metadata
type GuildScheduledEventEntityMetadata struct {
	// Location is the location of the event (1-100 characters). Required for EXTERNAL events.
	Location *string `json:"location,omitempty"`
}

// GuildScheduledEventRecurrenceRuleFrequency represents how often a recurrence rule repeats.
type GuildScheduledEventRecurrenceRuleFrequency int

const (
	RecurrenceFrequencyYearly  GuildScheduledEventRecurrenceRuleFrequency = 0
	RecurrenceFrequencyMonthly GuildScheduledEventRecurrenceRuleFrequency = 1
	RecurrenceFrequencyWeekly  GuildScheduledEventRecurrenceRuleFrequency = 2
	RecurrenceFrequencyDaily   GuildScheduledEventRecurrenceRuleFrequency = 3
)

// Is returns true if this GuildScheduledEventRecurrenceRuleFrequency matches the provided one.
func (v GuildScheduledEventRecurrenceRuleFrequency) Is(
	other GuildScheduledEventRecurrenceRuleFrequency,
) bool {
	return v == other
}

// GuildScheduledEventRecurrenceRuleWeekday represents a day of the week for recurrence rules.
type GuildScheduledEventRecurrenceRuleWeekday int

const (
	WeekdayMonday    GuildScheduledEventRecurrenceRuleWeekday = 0
	WeekdayTuesday   GuildScheduledEventRecurrenceRuleWeekday = 1
	WeekdayWednesday GuildScheduledEventRecurrenceRuleWeekday = 2
	WeekdayThursday  GuildScheduledEventRecurrenceRuleWeekday = 3
	WeekdayFriday    GuildScheduledEventRecurrenceRuleWeekday = 4
	WeekdaySaturday  GuildScheduledEventRecurrenceRuleWeekday = 5
	WeekdaySunday    GuildScheduledEventRecurrenceRuleWeekday = 6
)

// Is returns true if this GuildScheduledEventRecurrenceRuleWeekday matches the provided one.
func (v GuildScheduledEventRecurrenceRuleWeekday) Is(
	other GuildScheduledEventRecurrenceRuleWeekday,
) bool {
	return v == other
}

// GuildScheduledEventRecurrenceRuleNWeekday represents a specific week (1-5) and day to recur on.
type GuildScheduledEventRecurrenceRuleNWeekday struct {
	// N is the week to reoccur on (1-5).
	N int `json:"n"`

	// Day is the day within the week to reoccur on.
	Day GuildScheduledEventRecurrenceRuleWeekday `json:"day"`
}

// GuildScheduledEventRecurrenceRuleMonth represents a month of the year.
type GuildScheduledEventRecurrenceRuleMonth int

const (
	MonthJanuary   GuildScheduledEventRecurrenceRuleMonth = 1
	MonthFebruary  GuildScheduledEventRecurrenceRuleMonth = 2
	MonthMarch     GuildScheduledEventRecurrenceRuleMonth = 3
	MonthApril     GuildScheduledEventRecurrenceRuleMonth = 4
	MonthMay       GuildScheduledEventRecurrenceRuleMonth = 5
	MonthJune      GuildScheduledEventRecurrenceRuleMonth = 6
	MonthJuly      GuildScheduledEventRecurrenceRuleMonth = 7
	MonthAugust    GuildScheduledEventRecurrenceRuleMonth = 8
	MonthSeptember GuildScheduledEventRecurrenceRuleMonth = 9
	MonthOctober   GuildScheduledEventRecurrenceRuleMonth = 10
	MonthNovember  GuildScheduledEventRecurrenceRuleMonth = 11
	MonthDecember  GuildScheduledEventRecurrenceRuleMonth = 12
)

// Is returns true if this GuildScheduledEventRecurrenceRuleMonth matches the provided one.
func (v GuildScheduledEventRecurrenceRuleMonth) Is(
	other GuildScheduledEventRecurrenceRuleMonth,
) bool {
	return v == other
}

// GuildScheduledEventRecurrenceRule represents how often an event should recur.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-recurrence-rule-object
type GuildScheduledEventRecurrenceRule struct {
	// Start is the starting time of the recurrence interval.
	Start time.Time `json:"start"`

	// End is the ending time of the recurrence interval.
	End *time.Time `json:"end,omitempty"`

	// Frequency is how often the event occurs.
	Frequency GuildScheduledEventRecurrenceRuleFrequency `json:"frequency"`

	// Interval is the spacing between the events defined by frequency.
	Interval int `json:"interval"`

	// ByWeekday is the set of specific days within a week for the event to recur on.
	ByWeekday []GuildScheduledEventRecurrenceRuleWeekday `json:"by_weekday,omitempty"`

	// ByNWeekday is the list of specific days within a specific week (1-5) to recur on.
	ByNWeekday []GuildScheduledEventRecurrenceRuleNWeekday `json:"by_n_weekday,omitempty"`

	// ByMonth is the set of specific months to recur on.
	ByMonth []GuildScheduledEventRecurrenceRuleMonth `json:"by_month,omitempty"`

	// ByMonthDay is the set of specific dates within a month to recur on.
	ByMonthDay []int `json:"by_month_day,omitempty"`

	// ByYearDay is the set of days within a year to recur on (1-364).
	ByYearDay []int `json:"by_year_day,omitempty"`

	// Count is the total amount of times that the event is allowed to recur before stopping.
	Count *int `json:"count,omitempty"`
}

// GuildScheduledEvent is a representation of a scheduled event in a guild.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object
type GuildScheduledEvent struct {
	// ID is the id of the scheduled event.
	ID Snowflake `json:"id"`

	// GuildID is the guild id which the scheduled event belongs to.
	GuildID Snowflake `json:"guild_id"`

	// ChannelID is the channel id in which the scheduled event will be hosted, or null if entity_type is EXTERNAL.
	ChannelID *Snowflake `json:"channel_id,omitempty"`

	// CreatorID is the id of the user that created the scheduled event.
	CreatorID *Snowflake `json:"creator_id,omitempty"`

	// Name is the name of the scheduled event (1-100 characters).
	Name string `json:"name"`

	// Description is the description of the scheduled event (1-1000 characters).
	Description *string `json:"description,omitempty"`

	// ScheduledStartTime is the time the scheduled event will start.
	ScheduledStartTime time.Time `json:"scheduled_start_time"`

	// ScheduledEndTime is the time the scheduled event will end, required if entity_type is EXTERNAL.
	ScheduledEndTime *time.Time `json:"scheduled_end_time,omitempty"`

	// PrivacyLevel is the privacy level of the scheduled event.
	PrivacyLevel GuildScheduledEventPrivacyLevel `json:"privacy_level"`

	// Status is the status of the scheduled event.
	Status GuildScheduledEventStatus `json:"status"`

	// EntityType is the type of the scheduled event.
	EntityType GuildScheduledEventEntityType `json:"entity_type"`

	// EntityID is the id of an entity associated with a guild scheduled event.
	EntityID *Snowflake `json:"entity_id,omitempty"`

	// EntityMetadata is additional metadata for the guild scheduled event.
	EntityMetadata *GuildScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`

	// Creator is the user that created the scheduled event.
	Creator *User `json:"creator,omitempty"`

	// UserCount is the number of users subscribed to the scheduled event.
	UserCount *int `json:"user_count,omitempty"`

	// Image is the cover image hash of the scheduled event.
	Image *string `json:"image,omitempty"`

	// RecurrenceRule is the definition for how often this event should recur.
	RecurrenceRule *GuildScheduledEventRecurrenceRule `json:"recurrence_rule,omitempty"`
}

// CreatedAt returns the creation timestamp of the scheduled event.
func (e *GuildScheduledEvent) CreatedAt() time.Time {
	return e.ID.Timestamp()
}

// GuildScheduledEventUser represents a user subscribed to a scheduled event.
//
// Reference: https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-user-object
type GuildScheduledEventUser struct {
	// GuildScheduledEventID is the scheduled event id which the user subscribed to.
	GuildScheduledEventID Snowflake `json:"guild_scheduled_event_id"`

	// User is the user which subscribed to an event.
	User User `json:"user"`

	// Member is the guild member data for this user in the guild, if requested.
	Member *FullMember `json:"member,omitempty"`
}

// String implements the fmt.Stringer interface.
func (e *GuildScheduledEvent) String() string {
	return e.Name
}

var _ fmt.Stringer = (*GuildScheduledEvent)(nil)
