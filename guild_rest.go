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
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FetchGuildOptions contains parameters for fetching a guild.
type FetchGuildOptions struct {
	// When 'true', will return approximate member and presence counts for the guild
	WithCounts bool `json:"with_counts,omitempty"`
}

// FetchGuild retrieves a guild by its ID.
//
// Reference: https://discord.com/developers/docs/resources/guild#get-guild
func (r *requester) FetchGuild(ctx context.Context, guildID Snowflake, opts ...FetchGuildOptions) (RestGuild, error) {
	var first FetchGuildOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	endpoint := "/guilds/" + guildID.String() + "?with_counts=" + strconv.FormatBool(first.WithCounts)

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return RestGuild{}, err
	}
	defer body.Close()

	var guild RestGuild
	if err := json.NewDecoder(body).Decode(&guild); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}",
				"error", err.Error(),
			)
		return RestGuild{}, err
	}
	return guild, nil
}

// FetchGuildPreview retrieves a guild preview by its ID.
//
// Reference: https://discord.com/developers/docs/resources/guild#get-guild-preview
func (r *requester) FetchGuildPreview(ctx context.Context, guildID Snowflake) (GuildPreview, error) {
	endpoint := "/guilds/" + guildID.String() + "/preview"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return GuildPreview{}, err
	}
	defer body.Close()

	var preview GuildPreview
	if err := json.NewDecoder(body).Decode(&preview); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/preview",
				"error", err.Error(),
			)
		return GuildPreview{}, err
	}
	return preview, nil
}

type AfkTimeout int

const (
	AfkTimeout60   AfkTimeout = 60
	AfkTimeout300  AfkTimeout = 300
	AfkTimeout900  AfkTimeout = 900
	AfkTimeout1800 AfkTimeout = 1800
	AfkTimeout3600 AfkTimeout = 3600
)

// Is returns true if this AfkTimeout matches the provided one.
func (v AfkTimeout) Is(other AfkTimeout) bool {
	return v == other
}

// ModifyGuildOptions contains parameters for modifying a guild's settings.
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-guild
type ModifyGuildOptions struct {
	Name                        string                      `json:"name,omitempty"`
	VerificationLevel           *VerificationLevel          `json:"verification_level,omitempty"`
	DefaultMessageNotifications *MessageNotificationsLevel  `json:"default_message_notifications,omitempty"`
	ExplicitContentFilter       *ExplicitContentFilterLevel `json:"explicit_content_filter,omitempty"`
	AfkChannelID                *Snowflake                  `json:"afk_channel_id,omitempty"`
	AfkTimeout                  AfkTimeout                  `json:"afk_timeout,omitempty"`
	Icon                        *Base64Image                `json:"icon,omitempty"`
	Splash                      *Base64Image                `json:"splash,omitempty"`
	DiscoverySplash             *Base64Image                `json:"discovery_splash,omitempty"`
	Banner                      *Base64Image                `json:"banner,omitempty"`
	SystemChannelID             *Snowflake                  `json:"system_channel_id,omitempty"`
	SystemChannelFlags          *SystemChannelFlags         `json:"system_channel_flags,omitempty"`
	RulesChannelID              *Snowflake                  `json:"rules_channel_id,omitempty"`
	PublicUpdatesChannelID      *Snowflake                  `json:"public_updates_channel_id,omitempty"`
	PreferredLocale             Locale                      `json:"preferred_locale,omitempty"`
	Features                    []GuildFeature              `json:"features"`
	Description                 string                      `json:"description"`
	PremiumProgressBarEnabled   *bool                       `json:"premium_progress_bar_enabled,omitempty"`
	SafetyAlertsChannelID       *Snowflake                  `json:"safety_alerts_channel_id,omitempty"`

	Reason string `json:"-"`
}

// ModifyGuild modifies a guild's settings and returns the updated guild object.
//
// Requires the PermissionManageGuild permission.
//
// Note:
//   - Attempting to add or remove the GuildFeatureCommunity guild feature requires the PermissionAdministrator permission
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-guild
func (r *requester) ModifyGuild(ctx context.Context, guildID Snowflake, opts ModifyGuildOptions) (Guild, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String(),
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return Guild{}, err
	}
	defer body.Close()

	var guild Guild
	if err := json.NewDecoder(body).Decode(&guild); err != nil {
		return Guild{}, err
	}
	return guild, nil
}

// FetchGuildChannels returns a list of guild channel objects.
//
// Note:
//   - Does not include threads.
//
// Reference: https://discord.com/developers/docs/resources/guild#get-guild-channels
func (r *requester) FetchGuildChannels(ctx context.Context, guildID Snowflake) ([]GuildChannel, error) {
	endpoint := "/guilds/" + guildID.String() + "/channels"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var rawChannels []json.RawMessage
	if err := json.NewDecoder(body).Decode(&rawChannels); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/channels",
				"error", err.Error(),
			)
		return nil, err
	}

	channels := make([]GuildChannel, 0, len(rawChannels))
	for _, raw := range rawChannels {
		ch, err := UnmarshalChannel(raw)
		if err != nil {
			return nil, err
		}
		if guildCh, ok := ch.(GuildChannel); ok {
			channels = append(channels, guildCh)
		}
	}

	return channels, nil
}

// CreateChannelOptions defines the configuration for creating a new Discord guild channel.
//
// Note:
//   - This struct configures properties for a new channel, such as text, voice, or forum.
//   - Only set fields applicable to the channel type to avoid errors.
//
// Reference: https://discord.com/developers/docs/resources/guild#create-guild-channel
type CreateChannelOptions struct {
	// Name is the channel's name (1-100 characters).
	//
	// Applies to All Channels.
	Name string `json:"name"`

	// Type specifies the type of channel to create.
	//
	// Note:
	//  - Defaults to ChannelTypeGuildText if unset.
	//
	// Applies to All Channels.
	Type ChannelType `json:"type"`

	// Topic is a description of the channel (0-1024 characters).
	//
	// Note:
	//  - This field is optional.
	//
	// Applies to Channels of Type: Text, Announcement, Forum, Media.
	Topic string `json:"topic,omitempty"`

	// Bitrate sets the audio quality for voice or stage channels (in bits, minimum 8000).
	//
	// Note:
	//  - This field is ignored for non-voice channels.
	//
	// Applies to Channels of Type: Voice, Stage.
	Bitrate Bitrate `json:"bitrate,omitempty"`

	// UserLimit caps the number of users in a voice or stage channel (0 for unlimited, 1-99 for a limit).
	//
	// Note:
	//  - Set to 0 to allow unlimited users.
	//
	// Applies to Channels of Type: Voice, Stage.
	UserLimit int `json:"user_limit,omitempty"`

	// RateLimitPerUser sets the seconds a user must wait before sending another message (0-21600).
	//
	// Note:
	//  - Bots and users with manage_messages or manage_channel permissions are unaffected.
	//
	// Applies to Channels of Type: Text, Voice, Stage, Forum, Media.
	RateLimitPerUser int `json:"rate_limit_per_user,omitempty"`

	// Position determines the channel’s position in the server’s channel list (lower numbers appear higher).
	//
	// Note:
	//  - Channels with the same position are sorted by their internal ID.
	//
	// Applies to All Channels.
	Position *int `json:"position,omitempty"`

	// PermissionOverwrites defines custom permissions for specific roles or users.
	//
	// Applies to All Channels.
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites,omitempty"`

	// ParentID is the ID of the category to nest the channel under.
	//
	// Applies to Channels of Type: Text, Voice, Announcement, Stage, Forum, Media.
	ParentID Snowflake `json:"parent_id,omitempty"`

	// Nsfw marks the channel as Not Safe For Work, restricting it to 18+ users.
	//
	// Note:
	//  - Set to true to enable the age restriction.
	//
	// Applies to Channels of Type: Text, Voice, Announcement, Stage, Forum.
	Nsfw bool `json:"nsfw,omitempty"`

	// RTCRegion sets the channel voice region id of the voice or stage channel.
	//
	// Applies to Channels of Type: Voice, Stage.
	RTCRegion string `json:"rtc_region,omitempty"`

	// VideoQualityMode sets the camera video quality for voice or stage channels.
	//
	// Applies to Channels of Type: Voice, Stage.
	VideoQualityMode VideoQualityModes `json:"video_quality_mode,omitempty"`

	// DefaultAutoArchiveDuration sets the default time (in minutes) before threads are archived.
	//
	// Applies to Channels of Type: Text, Announcement, Forum, Media.
	DefaultAutoArchiveDuration AutoArchiveDuration `json:"default_auto_archive_duration,omitempty"`

	// DefaultReactionEmoji is the default emoji for the add reaction button on threads.
	//
	// Applies to Channels of Type: Forum, Media.
	DefaultReactionEmoji *DefaultReactionEmoji `json:"default_reaction_emoji,omitempty"`

	// AvailableTags lists tags that can be applied to threads for organization.
	//
	// Note:
	//  - This field defines tags users can select for threads.
	//
	// Applies to Channels of Type: Forum, Media.
	AvailableTags []ForumTag `json:"available_tags,omitempty"`

	// DefaultSortOrder sets how threads are sorted by default.
	//
	// Note:
	//  - Valid options are defined in ForumPostsSortOrder.
	//
	// Applies to Channels of Type: Forum, Media.
	DefaultSortOrder *ForumPostsSortOrder `json:"default_sort_order,omitempty"`

	// DefaultForumLayout sets the default view for forum posts.
	//
	// Applies to Channels of Type: Forum.
	DefaultForumLayout ForumLayout `json:"default_forum_layout,omitempty"`

	// DefaultThreadRateLimitPerUser sets the default slow mode for messages in new threads.
	//
	// Note:
	//  - This value is copied to new threads at creation and does not update live.
	//
	// Applies to Channels of Type: Text, Announcement, Forum, Media.
	DefaultThreadRateLimitPerUser int `json:"default_thread_rate_limit_per_user,omitempty"`

	// Reason specifies the audit log reason for creating the channel.
	Reason string `json:"-"`
}

// CreateChannel creates a new channel for the guild.
//
// Requires the PermissionManageChannels permission.
//
// Reference: https://discord.com/developers/docs/resources/guild#create-guild-channel
func (r *requester) CreateChannel(ctx context.Context, guildID Snowflake, opts CreateChannelOptions) (GuildChannel, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/" + guildID.String() + "/channels",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	channel, err := UnmarshalChannel(bodyBytes)
	if err != nil {
		r.logger.
			Error("failed parsing response", "method", "POST",
				"url", "/guilds/{id}/channels",
				"error", err.Error(),
			)
		return nil, err
	}
	if gc, ok := channel.(GuildChannel); ok {
		return gc, nil
	}
	return nil, errors.New("created channel is not a guild channel")
}

type ChannelPosition struct {
	// Channel id
	ID Snowflake `json:"id"`

	// Position is the sorting position of the channel (channels with the same position are sorted by id).
	Position *int `json:"position,omitempty"`

	// Syncs the permission overwrites with the new parent, if moving to a new category.
	LockPermissions bool `json:"lock_permissions,omitempty"`

	// ParentID is the new parent ID for the channel that is moved.
	ParentID Snowflake `json:"parent_id,omitempty"`
}

// ModifyChannelPositionOptions contains parameters for modifying channel positions.
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-guild-channel-positions
type ModifyChannelPositionOptions struct {
	Channels []ChannelPosition
}

// ModifyChannelPositions modifies the positions of a set of channel objects for the guild.
//
// Requires the PermissionManageChannels permission.
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-guild-channel-positions
func (r *requester) ModifyChannelPositions(ctx context.Context, guildID Snowflake, opts ModifyChannelPositionOptions) error {
	reqBody, _ := json.Marshal(opts.Channels)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/channels",
		Body:   reqBody,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// ActiveThreadsResponse is the response for listing active threads.
//
// Reference: https://discord.com/developers/docs/resources/guild#list-active-guild-threads
type ActiveThreadsResponse struct {
	// Threads is a list of active threads.
	//
	//  Note:
	//   - Threads are ordered by their id, in descending order.
	Threads []ThreadChannel `json:"threads"`

	// Members is list of a thread member object for each returned thread the current user has joined.
	Members []ThreadMember `json:"members"`
}

// ListActiveGuildThreads returns all active threads in the guild.
//
// Reference: https://discord.com/developers/docs/resources/guild#list-active-guild-threads
func (r *requester) ListActiveGuildThreads(ctx context.Context, guildID Snowflake) (ActiveThreadsResponse, error) {
	endpoint := "/guilds/" + guildID.String() + "/threads/active"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return ActiveThreadsResponse{}, err
	}
	defer body.Close()

	var response ActiveThreadsResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/threads/active",
				"error", err.Error(),
			)
		return ActiveThreadsResponse{}, err
	}
	return response, nil
}

// FetchMember retrieves a guild member object for the specified user.
func (r *requester) FetchMember(ctx context.Context, guildID, userID Snowflake) (FullMember, error) {
	endpoint := "/guilds/" + guildID.String() + "/members/" + userID.String()

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return FullMember{}, err
	}
	defer body.Close()

	var member FullMember
	if err := json.NewDecoder(body).Decode(&member); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/members/{user_id}",
				"error", err.Error(),
			)
		return FullMember{}, err
	}
	member.GuildID = guildID
	if member.ID == 0 && member.User.ID != 0 {
		member.ID = member.User.ID
	}
	if member.ID == 0 {
		member.ID = userID
	}
	return member, nil
}

// ListMembersOptions contains parameters for paginating through guild members.
type ListMembersOptions struct {
	// Limit is the maximum number of members to return (1-1000).
	//
	//  Note:
	//   - Defaults to 1 if not specified.
	Limit int `json:"limit,omitempty"`

	// After is the user ID to start after for pagination.
	// Used to get the next page of results.
	After Snowflake `json:"after,omitempty"`
}

// ListMembers retrieves a paginated list of members in a guild.
//
//	Note:
//	 - This endpoint is restricted according to whether the GUILD_MEMBERS Privileged Intent is enabled for your application.
func (r *requester) ListMembers(ctx context.Context, guildID Snowflake, opts ...ListMembersOptions) ([]FullMember, error) {
	var first ListMembersOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	endpoint := "/guilds/" + guildID.String() + "/members"

	params := url.Values{}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if first.After != 0 {
		params.Set("after", first.After.String())
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var members []FullMember
	if err := json.NewDecoder(body).Decode(&members); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/members",
				"error", err.Error(),
			)
		return nil, err
	}
	for i := range members {
		members[i].GuildID = guildID
		if members[i].ID == 0 && members[i].User.ID != 0 {
			members[i].ID = members[i].User.ID
		}
	}
	return members, nil
}

// SearchMembersOptions contains parameters for searching members by name.
type SearchMembersOptions struct {
	// Query is the text to search for in usernames and nicknames.
	//
	//  Note:
	//   - Query is required to be set
	Query string `json:"query"`

	// Limit is the maximum number of members to return (1-1000).
	// Defaults to 1 if not specified.
	Limit int `json:"limit,omitempty"`
}

// SearchMembers returns a list of guild member objects whose username or nickname starts with a provided string.
func (r *requester) SearchMembers(ctx context.Context, guildID Snowflake, opts SearchMembersOptions) ([]FullMember, error) {
	endpoint := "/guilds/" + guildID.String() + "/members/search"

	params := url.Values{}
	params.Set("query", opts.Query)
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var members []FullMember
	if err := json.NewDecoder(body).Decode(&members); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/members/search",
				"error", err.Error(),
			)
		return nil, err
	}
	for i := range members {
		members[i].GuildID = guildID
		if members[i].ID == 0 && members[i].User.ID != 0 {
			members[i].ID = members[i].User.ID
		}
	}
	return members, nil
}

// AddMemberOptions contains parameters for adding a user to a guild.
//
// Requires a valid OAuth2 access token with the guilds.join scope.
type AddMemberOptions struct {
	// AccessToken is the OAuth2 access token for the user you want to add.
	// This must have the guilds.join scope.
	AccessToken string `json:"access_token"`

	// Nick sets the member's initial nickname in the guild.
	//
	// Requires the PermissionManageNicknames permission.
	Nick string `json:"nick,omitempty"`

	// Roles is a list of role IDs to assign to the member initially.
	//
	// Requires the PermissionManageRoles permission.
	Roles []Snowflake `json:"roles,omitempty"`

	// Mute sets whether the user is muted in voice channels.
	//
	// Requires the PermissionMuteMember permission.
	Mute bool `json:"mute,omitempty"`

	// Deaf sets whether the user is deafened in voice channels.
	//
	// Requires the PermissionDeafenMembers permission.
	Deaf bool `json:"deaf,omitempty"`
}

// AddGuildMember adds a user to the guild using their OAuth2 access token.
//
// The access token must have the guilds.join scope.
// If the user is already in the guild, this has no effect.
//
//	Note:
//	 - The bot must be a member of the guild with PermissionCreateInstantInvite permission.
//	 - For guilds with Membership Screening enabled, this endpoint will default to adding new members as
//	   pending in the guild member object. Members that are pending will have to complete membership screening
//	   before they become full members that can talk.
//
// Reference: https://discord.com/developers/docs/resources/guild#add-guild-member
func (r *requester) AddGuildMember(ctx context.Context, guildID, userID Snowflake, opts AddMemberOptions) (FullMember, error) {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/members/" + userID.String()

	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    endpoint,
		Body:   reqBody,
	})
	if err != nil {
		return FullMember{}, err
	}
	defer body.Close()

	var member FullMember
	if err := json.NewDecoder(body).Decode(&member); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PUT",
				"url", "/guilds/{id}/members/{user_id}",
				"error", err.Error(),
			)
		return FullMember{}, err
	}
	member.GuildID = guildID
	if member.ID == 0 {
		if member.User.ID != 0 {
			member.ID = member.User.ID
		} else {
			member.ID = userID
		}
	}
	return member, nil
}

// ModifyMemberOptions contains parameters for modifying a guild member.
//
// All fields are optional. Only provide the fields you want to change.
type ModifyMemberOptions struct {
	// Nickname sets the member's guild nickname.
	// Set to empty string to remove the nickname.
	//
	// Requires the PermissionManageNicknames permission.
	Nickname *string `json:"nick,omitempty"`

	// Roles sets the complete list of role IDs for the member.
	// This replaces all existing roles with the provided list.
	//
	// Requires the PermissionManageRoles permission.
	Roles []Snowflake `json:"roles,omitempty"`

	// Mute sets whether the member is muted in voice channels.
	// The member must be in a voice channel for this to work.
	//
	// Requires the PermissionMuteMembers permission.
	Mute *bool `json:"mute,omitempty"`

	// Deaf sets whether the member is deafened in voice channels.
	// The member must be in a voice channel for this to work.
	//
	// Requires the PermissionDeafenMembers permission.
	Deaf *bool `json:"deaf,omitempty"`

	// ChannelID moves the member to a different voice channel.
	// The member must be connected to voice for this to work.
	//
	// Requires the PermissionMoveMembers permission.
	ChannelID Snowflake `json:"channel_id,omitempty"`

	// CommunicationDisabledUntil sets when the member's timeout expires.
	// Can be up to 28 days in the future.
	//
	// Note: Supplying 'optional.Nil[time.Time]()' disables the action.
	//
	// Requires the PermissionModerateMembers permission.
	CommunicationDisabledUntil *time.Time `json:"communication_disabled_until,omitempty"`

	// Flags sets the member's guild-specific flags.
	//
	// Requires the PermissionManageGuild or PermissionManageRoles or (PermissionModerateMembers and PermissionKickMembers and PermissionBanMembers).
	Flags MemberFlags `json:"flags,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// ModifyMember updates a member's properties in a guild.
func (r *requester) ModifyMember(ctx context.Context, guildID, userID Snowflake, opts ModifyMemberOptions) (FullMember, error) {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/members/" + userID.String()

	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    endpoint,
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return FullMember{}, err
	}
	defer body.Close()

	var member FullMember
	if err := json.NewDecoder(body).Decode(&member); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/members/{user_id}",
				"error", err.Error(),
			)
		return FullMember{}, err
	}
	member.GuildID = guildID
	if member.ID == 0 {
		if member.User.ID != 0 {
			member.ID = member.User.ID
		} else {
			member.ID = userID
		}
	}
	return member, nil
}

// ModifyCurrentMemberOptions contains parameters for modifying your bot's member profile.
//
// This allows you to update your bot's nickname, avatar, banner, and bio in a specific guild.
type ModifyCurrentMemberOptions struct {
	// Nick sets your bot's nickname in this guild.
	//
	// Requires the PermissionManageNicknames permission.
	Nick *string `json:"nick,omitempty"`

	// Banner sets your bot's guild-specific banner image.
	Banner *Base64Image `json:"banner,omitempty"`

	// Avatar sets your bot's guild-specific avatar image.
	Avatar *Base64Image `json:"avatar,omitempty"`

	// Bio sets your bot's bio text for this guild.
	Bio *string `json:"bio,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// ModifyCurrentMember updates your bot's member properties in a guild.
func (r *requester) ModifyCurrentMember(ctx context.Context, guildID Snowflake, opts ModifyCurrentMemberOptions) (FullMember, error) {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/members/@me"

	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    endpoint,
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return FullMember{}, err
	}
	defer body.Close()

	var member FullMember
	if err := json.NewDecoder(body).Decode(&member); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/members/@me",
				"error", err.Error(),
			)
		return FullMember{}, err
	}
	member.GuildID = guildID
	if member.ID == 0 && member.User.ID != 0 {
		member.ID = member.User.ID
	}
	return member, nil
}

// AddMemberRoleOptions contains parameters for adding a role to a member.
type AddMemberRoleOptions struct {
	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// AddMemberRole assigns a role to a member in a guild.
//
// Requires the PermissionManageRoles permission.
func (r *requester) AddMemberRole(ctx context.Context, guildID, userID, roleID Snowflake, opts AddMemberRoleOptions) error {
	endpoint := "/guilds/" + guildID.String() + "/members/" + userID.String() + "/roles/" + roleID.String()
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    endpoint,
		Reason: opts.Reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// RemoveMemberRoleOptions contains parameters for removing a role from a member.
type RemoveMemberRoleOptions struct {
	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// RemoveMemberRole unassigns a role from a member in a guild.
//
// Requires the PermissionManageRoles permission.
func (r *requester) RemoveMemberRole(ctx context.Context, guildID, userID, roleID Snowflake, opts RemoveMemberRoleOptions) error {
	endpoint := "/guilds/" + guildID.String() + "/members/" + userID.String() + "/roles/" + roleID.String()
	body, err := r.DoRequest(ctx, Request{Method: "DELETE", URL: endpoint, Reason: opts.Reason})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// KickMemberOptions contains parameters for kicking a member from a guild.
type KickMemberOptions struct {
	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// KickMember kicks a member from a guild.
//
// Requires the PermissionKickMembers permission.
func (r *requester) KickMember(ctx context.Context, guildID, userID Snowflake, opts KickMemberOptions) error {
	endpoint := "/guilds/" + guildID.String() + "/members/" + userID.String()
	body, err := r.DoRequest(ctx, Request{Method: "DELETE", URL: endpoint, Reason: opts.Reason})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// FetchGuildBansOptions contains parameters for fetching guild bans.
//
// Reference: https://discord.com/developers/docs/resources/guild#get-guild-bans
type FetchGuildBansOptions struct {
	// Limit is the number of users to return (up to maximum 1000)
	//
	// Default to 1000 if not spesified
	Limit int `json:"limit,omitempty"`

	// Before consider only users before given user id.
	Before Snowflake `json:"before,omitempty"`

	// After consider only users after given user id.
	After Snowflake `json:"after,omitempty"`
}

// FetchGuildBans returns a list of ban objects for the users banned from this guild.
//
//	Note:
//	 - Provide a user id to before and after for pagination. Users will always be returned in ascending order by 'user.id'.
//	   If both before and after are provided, only before is respected.
//
// Requires the PermissionBanMembers permission.
func (r *requester) FetchGuildBans(ctx context.Context, guildID Snowflake, opts ...FetchGuildBansOptions) ([]Ban, error) {
	var first FetchGuildBansOptions
	if len(opts) > 0 {
		first = opts[0]
	}
	endpoint := "/guilds/" + guildID.String() + "/bans"

	params := url.Values{}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if first.Before != 0 {
		params.Set("before", first.Before.String())
	}
	if first.After != 0 {
		params.Set("after", first.After.String())
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var bans []Ban
	if err := json.NewDecoder(body).Decode(&bans); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/bans",
				"error", err.Error(),
			)
		return nil, err
	}
	return bans, nil
}

// FetchGuildBan returns a ban object for the given user.
// FetchGuildBan returns a ban object for the given user, or nil if not banned.
//
// Requires the PermissionBanMembers permission.
func (r *requester) FetchGuildBan(ctx context.Context, guildID, userID Snowflake) (*Ban, error) {
	endpoint := "/guilds/" + guildID.String() + "/bans/" + userID.String()

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var ban Ban
	if err := json.NewDecoder(body).Decode(&ban); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}

		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/bans/{user_id}",
				"error", err.Error(),
			)
		return nil, err
	}
	return &ban, nil
}

// BanMemberOptions contains parameters for banning a guild member.
type BanMemberOptions struct {
	// DeleteMessageSeconds is the number of seconds to delete messages for, between 0 and 604800 (7 days)
	DeleteMessageSeconds int `json:"delete_message_seconds,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// BanMember ban's a guild member.
//
// Requires the PermissionBanMembers permission.
func (r *requester) BanMember(ctx context.Context, guildID, userID Snowflake, opts BanMemberOptions) error {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/bans/" + userID.String()
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    endpoint,
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// UnbanMemberOptions contains parameters for unbanning a guild member.
type UnbanMemberOptions struct {
	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// UnbanMember removes the ban for a user.
//
// Requires the PermissionBanMembers permission.
func (r *requester) UnbanMember(ctx context.Context, guildID, userID Snowflake, opts UnbanMemberOptions) error {
	endpoint := "/guilds/" + guildID.String() + "/bans/" + userID.String()
	body, err := r.DoRequest(ctx, Request{Method: "DELETE", URL: endpoint, Reason: opts.Reason})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// BulkBanMembersOptions contains parameters for bulk banning guild members.
//
// Reference: https://discord.com/developers/docs/resources/guild#bulk-guild-ban
type BulkBanMembersOptions struct {
	// UserIDs is a list of user ids to ban (max 200).
	UserIDs []Snowflake `json:"user_ids"`

	// DeleteMessageSeconds is the number of seconds to delete messages for, between 0 and 604800 (7 days).
	DeleteMessageSeconds int `json:"delete_message_seconds,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// BulkBanResponse is the response for bulk banning members.
//
// Reference: https://discord.com/developers/docs/resources/guild#bulk-guild-ban
type BulkBanResponse struct {
	// BannedUsers is a list of user ids, that were successfully banned.
	BannedUsers []Snowflake `json:"banned_users"`

	// FailedUsers is a list of user ids, that were not banned.
	FailedUsers []Snowflake `json:"failed_users"`
}

// BulkBanMembers bans up to 200 users from a guild.
//
// Reference: https://discord.com/developers/docs/resources/guild#bulk-guild-ban
func (r *requester) BulkBanMembers(ctx context.Context, guildID Snowflake, opts BulkBanMembersOptions) (BulkBanResponse, error) {
	if len(opts.UserIDs) > 200 {
		panic("BulkBanMembers: UserIDs exceeds Discord limit of 200 users")
	}
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/" + guildID.String() + "/bulk-ban",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return BulkBanResponse{}, err
	}
	defer body.Close()

	var response BulkBanResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		r.logger.
			Error("failed parsing response", "method", "POST",
				"url", "/guilds/{id}/bulk-ban",
				"error", err.Error(),
			)
		return BulkBanResponse{}, err
	}
	return response, nil
}

// FetchRoles returns a list of role objects for the guild.
func (r *requester) FetchRoles(ctx context.Context, guildID Snowflake) ([]Role, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/roles",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var roles []Role
	if err := json.NewDecoder(body).Decode(&roles); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/roles",
				"error", err.Error(),
			)
		return nil, err
	}
	for i := range len(roles) {
		roles[i].GuildID = guildID
	}
	return roles, nil
}

// FetchRole returns a role object for the specified role id.
func (r *requester) FetchRole(ctx context.Context, guildID, roleID Snowflake) (Role, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/roles/" + roleID.String(),
	})
	if err != nil {
		return Role{}, err
	}
	defer body.Close()

	var role Role
	if err := json.NewDecoder(body).Decode(&role); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/roles/{id}",
				"error", err.Error(),
			)
		return Role{}, err
	}
	role.GuildID = guildID
	return role, nil
}

// FetchRolesMemberCount returns a map of role IDs to the number of members with the role.
//
//	Note:
//	 - Does not include the @everyone role.
func (r *requester) FetchRolesMemberCount(ctx context.Context, guildID Snowflake) (map[Snowflake]int, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/roles/member-counts",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var rolesMemberCount map[Snowflake]int
	if err := json.NewDecoder(body).Decode(&rolesMemberCount); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/roles/member-counts",
				"error", err.Error(),
			)
		return nil, err
	}
	return rolesMemberCount, nil
}

// CreateRoleOptions defines the configuration for creating a new Discord guild role.
type CreateRoleOptions struct {
	// Name is the channel's name (max 100 characters).
	//
	// Default to "new role" if unspesified.
	Name string `json:"name,omitempty"`

	// Permissions is the permission to set for the role.
	Permissions Permissions `json:"permissions,omitempty"`

	// Colors are the colors to set for the role.
	Colors RoleColors `json:"colors"`

	// Hoist is whether the role should be displayed separately in the sidebar.
	Hoist bool `json:"hoist,omitempty"`

	// Mentionable is whether the role should be displayed separately in the sidebar.
	Mentionable bool `json:"mentionable,omitempty"`

	// Icon is the role's icon image (if the guild has the GuildFeatureRoleIcons feature).
	Icon Base64Image `json:"icon,omitempty"`

	// UnicodeEmoji is the role's unicode emoji as a standard emoji (if the guild has the GuildFeatureRoleIcons feature).
	UnicodeEmoji string `json:"unicode_emoji,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// CreateRole creates a new role for the guild.
//
// Requires the PermissionManageRoles permission.
func (r *requester) CreateRole(ctx context.Context, guildID Snowflake, opts CreateRoleOptions) (Role, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/" + guildID.String() + "/roles",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return Role{}, err
	}
	defer body.Close()

	var role Role
	if err := json.NewDecoder(body).Decode(&role); err != nil {
		r.logger.
			Error("failed parsing response", "method", "POST",
				"url", "/guilds/{id}/roles",
				"error", err.Error(),
			)
		return Role{}, err
	}
	return role, nil
}

type RolePosition struct {
	// Channel id
	ID Snowflake `json:"id"`

	// Position is the sorting position of the channel (channels with the same position are sorted by id).
	Position *int `json:"position,omitempty"`
}

// ModifyRolePositionsOptions defines the configuration for modifying roles positions.
type ModifyRolePositionsOptions struct {
	Roles []RolePosition

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// ModifyRolePositions modifies the positions of a set of roles.
//
// Requires the PermissionManageRoles permission.
func (r *requester) ModifyRolePositions(ctx context.Context, guildID Snowflake, opts ModifyRolePositionsOptions) ([]Role, error) {
	reqBody, _ := json.Marshal(opts.Roles)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/roles",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var roles []Role
	if err := json.NewDecoder(body).Decode(&roles); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/roles",
				"error", err.Error(),
			)
		return nil, err
	}
	return roles, nil
}

// ModifyRoleOptions contains parameters for modifying a guild role.
type ModifyRoleOptions struct {
	// Name is the channel's name (max 100 characters).
	Name string `json:"name,omitempty"`

	// Permissions is the permission to set for the role.
	Permissions *Permissions `json:"permissions,omitempty"`

	// Colors are the colors to set for the role.
	Colors *RoleColors `json:"colors,omitempty"`

	// Hoist is whether the role should be displayed separately in the sidebar.
	Hoist *bool `json:"hoist,omitempty"`

	// Mentionable is whether the role should be displayed separately in the sidebar.
	Mentionable *bool `json:"mentionable,omitempty"`

	// Icon is the role's icon image (if the guild has the GuildFeatureRoleIcons feature).
	Icon *Base64Image `json:"icon,omitempty"`

	// UnicodeEmoji is the role's unicode emoji as a standard emoji (if the guild has the GuildFeatureRoleIcons feature).
	UnicodeEmoji *string `json:"unicode_emoji,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// ModifyRole updates a member's properties in a guild.
//
// Requires the PermissionManageRoles permission.
func (r *requester) ModifyRole(ctx context.Context, guildID, roleID Snowflake, opts ModifyRoleOptions) (Role, error) {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/roles/" + roleID.String()

	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    endpoint,
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return Role{}, err
	}
	defer body.Close()

	var role Role
	if err := json.NewDecoder(body).Decode(&role); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/roles/{role_id}",
				"error", err.Error(),
			)
		return Role{}, err
	}
	role.GuildID = guildID
	return role, nil
}

// DeleteRoleOptions contains parameters for deleting a guild role.
type DeleteRoleOptions struct {
	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// DeleteRole delete's a guild role.
//
// Requires the PermissionManageRoles permission.
func (r *requester) DeleteRole(ctx context.Context, guildID, roleID Snowflake, opts DeleteRoleOptions) error {
	endpoint := "/guilds/" + guildID.String() + "/roles/" + roleID.String()
	body, err := r.DoRequest(ctx, Request{Method: "DELETE", URL: endpoint, Reason: opts.Reason})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// PruneCount contains parameters for fetching guild prune count.
type FetchGuildPruneCountOptions struct {
	// Pruned is the number of days to count prune for (1-30).
	Days int `json:"days,omitempty"`

	// By default, prune will not remove users with roles. You can optionally include specific roles in your prune by them in this field.
	// Any inactive user that has a subset of the provided role(s) will be counted in the prune and users with additional roles will not.
	IncludeRoles []Snowflake `json:"include_roles,omitempty"`
}

// PruneCount represents the result of a prune count.
type PruneCount struct {
	// Pruned is the number of members that would be/were removed in a prune operation.
	Pruned int `json:"pruned"`
}

// FetchGuildPruneCount returns an object with one pruned key indicating the number
// of members that would be removed in a prune operation.
//
// Requires the PermissionManageGuild and PermissionKickMembers permission.
func (r *requester) FetchGuildPruneCount(ctx context.Context, guildID Snowflake, opts ...FetchGuildPruneCountOptions) (PruneCount, error) {
	var first FetchGuildPruneCountOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	endpoint := "/guilds/" + guildID.String() + "/prune"

	params := url.Values{}
	if first.Days > 0 {
		params.Set("days", strconv.Itoa(first.Days))
	}
	if len(first.IncludeRoles) > 0 {
		var b strings.Builder
		for i, snowflake := range first.IncludeRoles {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(snowflake.String())
		}
		params.Set("include_roles", b.String())
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return PruneCount{}, err
	}
	defer body.Close()

	var count PruneCount
	if err := json.NewDecoder(body).Decode(&count); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/prune",
				"error", err.Error(),
			)
		return PruneCount{}, err
	}
	return count, nil
}

// BeginGuildPruneOptions contains parameters for begining guild prune.
type BeginGuildPruneOptions struct {
	// Pruned is the number of days to count prune for (1-30).
	//
	// Default 7 if not set.
	Days int `json:"days,omitempty"`

	// ComputePruneCount is whether pruned is returned, discouraged for large guilds.
	ComputePruneCount *bool `json:"compute_prune_count,omitempty"`

	// IncludeRoles are the roles to include.
	IncludeRoles []Snowflake `json:"include_roles,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// BeginGuildPrune Begin a prune operation.
//
// Requires the PermissionManageGuild and PermissionKickMembers permission.
func (r *requester) BeginGuildPrune(ctx context.Context, guildID Snowflake, opts BeginGuildPruneOptions) (PruneCount, error) {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/prune"
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    endpoint,
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return PruneCount{}, err
	}
	defer body.Close()

	var count PruneCount
	if err := json.NewDecoder(body).Decode(&count); err != nil {
		r.logger.
			Error("failed parsing response", "method", "POST",
				"url", "/guilds/{id}/prune",
				"error", err.Error(),
			)
		return PruneCount{}, err
	}
	return count, nil
}

// FetchGuildVoiceRegions returns a list of voice region objects for the guild.
// Unlike the similar /voice route, this returns VIP servers when the guild is VIP-enabled.
func (r *requester) FetchGuildVoiceRegions(ctx context.Context, guildID Snowflake) ([]VoiceRegion, error) {
	endpoint := "/guilds/" + guildID.String() + "/regions"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var regions []VoiceRegion
	if err := json.NewDecoder(body).Decode(&regions); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/regions",
				"error", err.Error(),
			)
		return nil, err
	}
	return regions, nil
}

// FetchGuildVoiceRegions returns a list of voice region objects for the guild.
// Unlike the similar /voice route, this returns VIP servers when the guild is VIP-enabled.
//
// Requires the PermissionManageGuild and PermissionKickMembers permission.
func (r *requester) FetchGuildInvites(ctx context.Context, guildID Snowflake) ([]FullInvite, error) {
	endpoint := "/guilds/" + guildID.String() + "/invites"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var invites []FullInvite
	if err := json.NewDecoder(body).Decode(&invites); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/invites",
				"error", err.Error(),
			)
		return nil, err
	}
	return invites, nil
}

// FetchGuildIntegrations returns a list of integration objects for the guild.
//
// Note:
//   - This endpoint returns a maximum of 50 integrations. If a guild has more integrations, they cannot be accessed.
//
// Requires the PermissionManageGuild permission.
func (r *requester) FetchGuildIntegrations(ctx context.Context, guildID Snowflake) ([]Integration, error) {
	endpoint := "/guilds/" + guildID.String() + "/integrations"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var integrations []Integration
	if err := json.NewDecoder(body).Decode(&integrations); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/integrations",
				"error", err.Error(),
			)
		return nil, err
	}
	return integrations, nil
}

// DeleteGuildIntegrationOptions contains parameters for deleting a guild integration.
type DeleteGuildIntegrationOptions struct {
	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// DeleteGuildIntegration deletes the attached integration object for the guild.
// Deletes any associated webhooks and kicks the associated bot if there is one.
//
// Requires the PermissionManageGuild permission.
func (r *requester) DeleteGuildIntegration(ctx context.Context, guildID, integrationID Snowflake, opts DeleteGuildIntegrationOptions) error {
	endpoint := "/guilds/" + guildID.String() + "/integrations/" + integrationID.String()
	body, err := r.DoRequest(ctx, Request{Method: "DELETE", URL: endpoint, Reason: opts.Reason})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// FetchGuildWidgetSettings returns a guild widget settings object.
//
// Requires the PermissionManageGuild permission.
func (r *requester) FetchGuildWidgetSettings(ctx context.Context, guildID Snowflake) (GuildWidgetSettings, error) {
	endpoint := "/guilds/" + guildID.String() + "/widget"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return GuildWidgetSettings{}, err
	}
	defer body.Close()

	var settings GuildWidgetSettings
	if err := json.NewDecoder(body).Decode(&settings); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/widget",
				"error", err.Error(),
			)
		return GuildWidgetSettings{}, err
	}
	return settings, nil
}

// ModifyGuildWidgetOptions contains parameters for modifying a guild widget.
type ModifyGuildWidgetOptions struct {
	Enabled   *bool     `json:"enabled,omitempty"`
	ChannelID Snowflake `json:"channel_id,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// ModifyGuildWidget modifies a guild widget settings object for the guild.
func (r *requester) ModifyGuildWidget(ctx context.Context, guildID Snowflake, opts ModifyGuildWidgetOptions) (GuildWidgetSettings, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/widget",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return GuildWidgetSettings{}, err
	}
	defer body.Close()

	var settings GuildWidgetSettings
	if err := json.NewDecoder(body).Decode(&settings); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/widget",
				"error", err.Error(),
			)
		return GuildWidgetSettings{}, err
	}
	return settings, nil
}

// FetchGuildWidget returns the widget for the guild.
func (r *requester) FetchGuildWidget(ctx context.Context, guildID Snowflake) (GuildWidget, error) {
	endpoint := "/guilds/" + guildID.String() + "/widget.json"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return GuildWidget{}, err
	}
	defer body.Close()

	var widget GuildWidget
	if err := json.NewDecoder(body).Decode(&widget); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/widget.json",
				"error", err.Error(),
			)
		return GuildWidget{}, err
	}
	return widget, nil
}

// PartialInvite represents a guild's partial object.
//
// Reference: https://discord.com/developers/docs/resources/guild#get-guild-vanity-url
type PartialInvite struct {
	// Code is the vanity url code.
	Code string `json:"code,omitempty"`

	// Uses is the number of times this invite has been used.
	Uses int `json:"uses"`
}

// FetchGuildVanityURL returns a partial invite object for guilds with the feature enabled.
//
// Requires the PermissionManageGuild permission.
func (r *requester) FetchGuildVanityURL(ctx context.Context, guildID Snowflake) (PartialInvite, error) {
	endpoint := "/guilds/" + guildID.String() + "/vanity-url"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return PartialInvite{}, err
	}
	defer body.Close()

	var partialInvite PartialInvite
	if err := json.NewDecoder(body).Decode(&partialInvite); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/vanity-url",
				"error", err.Error(),
			)
		return PartialInvite{}, err
	}
	return partialInvite, nil
}

// GuildFeature represents the style of a Discord guild widget.
//
// Reference: https://discord.com/developers/docs/resources/guild#get-guild-widget-image-widget-style-options
type GuildWidgetStyle string

const (
	// Shield style widget with Discord icon and guild members online count
	GuildWidgetStyleShield GuildFeature = "shield"
	// Large image with guild icon, name and online count. "POWERED BY DISCORD" as the footer of the widget
	GuildWidgetStyleBanner1 GuildFeature = "banner1"
	// Smaller widget style with guild icon, name and online count. Split on the right with Discord logo
	GuildWidgetStyleBanner2 GuildFeature = "banner2"
	// Large image with guild icon, name and online count. In the footer, Discord logo on the left and "Chat Now" on the right
	GuildWidgetStyleBanner3 GuildFeature = "banner3"
	// large Discord logo at the top of the widget. Guild icon, name and online count in the middle portion of the widget
	// and a "JOIN MY SERVER" button at the bottom
	GuildWidgetStyleBanner4 GuildFeature = "banner4"
)

// FetchGuildWidgetImageOptions contains parameters for fetching guild widget image.
//
// Reference: https://discord.com/developers/docs/resources/guild#get-guild-widget-image
type FetchGuildWidgetImageOptions struct {
	// Style is the style of the widget image returned.
	Style GuildWidgetStyle `json:"style,omitempty"`
}

// FetchGuildWidgetImage returns a URL for a PNG image widget for the guild.
func (r *requester) FetchGuildWidgetImage(ctx context.Context, guildID Snowflake, opts ...FetchGuildWidgetImageOptions) string {
	url := "https://discord.com/api/v10/guilds/" + guildID.String() + "/widget.png"
	if len(opts) > 0 && opts[0].Style != "" {
		url += "?style=" + string(opts[0].Style)
	}
	return url
}

// FetchGuildWelcomeScreen returns the Welcome Screen object for the guild.
//
// Requires the PermissionManageGuild permission.
func (r *requester) FetchGuildWelcomeScreen(ctx context.Context, guildID Snowflake) (GuildWelcomeScreen, error) {
	endpoint := "/guilds/" + guildID.String() + "/welcome-screen"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return GuildWelcomeScreen{}, err
	}
	defer body.Close()

	var screen GuildWelcomeScreen
	if err := json.NewDecoder(body).Decode(&screen); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/welcome-screen",
				"error", err.Error(),
			)
		return GuildWelcomeScreen{}, err
	}
	return screen, nil
}

// ModifyGuildWelcomeScreenOptions contains parameters for modifying a guild welcome screen.
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-guild-welcome-screen
type ModifyGuildWelcomeScreenOptions struct {
	// Enabled is whether the welcome screen is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// WelcomeChannels are the channels linked in the welcome screen and their display options.
	WelcomeChannels []GuildWelcomeChannel `json:"welcome_channels,omitempty"`

	// Description is the the server description to show in the welcome screen.
	Description *string `json:"description,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// ModifyGuildWelcomeScreen modifies the guild's Welcome Screen.
//
// Requires the PermissionManageGuild permission.
func (r *requester) ModifyGuildWelcomeScreen(ctx context.Context, guildID Snowflake, opts ModifyGuildWelcomeScreenOptions) (GuildWelcomeScreen, error) {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/welcome-screen"
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    endpoint,
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return GuildWelcomeScreen{}, err
	}
	defer body.Close()

	var screen GuildWelcomeScreen
	if err := json.NewDecoder(body).Decode(&screen); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PATCH",
				"url", "/guilds/{id}/welcome-screen",
				"error", err.Error(),
			)
		return GuildWelcomeScreen{}, err
	}
	return screen, nil
}

// FetchGuildOnboarding returns the Onboarding object for the guild.
//
// Requires the PermissionManageGuild and PermissionManageRoles permissions.
func (r *requester) FetchGuildOnboarding(ctx context.Context, guildID Snowflake) (GuildOnboarding, error) {
	endpoint := "/guilds/" + guildID.String() + "/onboarding"

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return GuildOnboarding{}, err
	}
	defer body.Close()

	var onboarding GuildOnboarding
	if err := json.NewDecoder(body).Decode(&onboarding); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/guilds/{id}/onboarding",
				"error", err.Error(),
			)
		return GuildOnboarding{}, err
	}
	return onboarding, nil
}

// ModifyGuildOnboardingOptions contains parameters for modifying guild onboarding.
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-guild-onboarding
type ModifyGuildOnboardingOptions struct {
	// Prompts are the prompts shown during onboarding and in customize community.
	Prompts []OnboardingPrompt `json:"prompts,omitempty"`

	// DefaultChannelIDs are the channel IDs that members get opted into automatically.
	DefaultChannelIDs []Snowflake `json:"default_channel_ids,omitempty"`

	// Enabled is whether onboarding is enabled in the guild.
	Enabled *bool `json:"enabled,omitempty"`

	// Mode is the current mode of onboarding.
	Mode OnboardingMode `json:"mode,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// ModifyGuildOnboarding modifies the onboarding configuration of the guild.
//
// Note:
//   - Onboarding enforces constraints when enabled. These constraints are that there must be at
//     least 7 Default Channels and at least 5 of them must allow sending messages to the @everyone role.
//     The mode field modifies what is considered when enforcing these constraints.
//
// Requires the PermissionManageGuild and PermissionManageRoles permissions.
func (r *requester) ModifyGuildOnboarding(ctx context.Context, guildID Snowflake, opts ModifyGuildOnboardingOptions) (GuildOnboarding, error) {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/onboarding"
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    endpoint,
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return GuildOnboarding{}, err
	}
	defer body.Close()

	var onboarding GuildOnboarding
	if err := json.NewDecoder(body).Decode(&onboarding); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PUT",
				"url", "/guilds/{id}/onboarding",
				"error", err.Error(),
			)
		return GuildOnboarding{}, err
	}
	return onboarding, nil
}

// ModifyGuildIncidentActionsOptions contains parameters for modifying guild incident actions.
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-guild-incident-actions
type ModifyGuildIncidentActionsOptions struct {
	// InvitesDisabledUntil is when invites will be enabled again.
	//
	// Note: Supplying 'optional.Nil[time.Time]()' disables the action.
	InvitesDisabledUntil *time.Time `json:"invites_disabled_until,omitempty"`

	// DMsDisabledUntil is when direct messages will be enabled again.
	//
	// Note: Supplying 'optional.Nil[time.Time]()' disables the action.
	DMsDisabledUntil *time.Time `json:"dms_disabled_until,omitempty"`

	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// ModifyGuildIncidentActions modifies the incident actions of the guild.
//
// Requires the PermissionManageGuild permission.
func (r *requester) ModifyGuildIncidentActions(ctx context.Context, guildID Snowflake, opts ModifyGuildIncidentActionsOptions) (GuildIncidentsData, error) {
	reqBody, _ := json.Marshal(opts)
	endpoint := "/guilds/" + guildID.String() + "/incident-actions"
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    endpoint,
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return GuildIncidentsData{}, err
	}
	defer body.Close()

	var incidentsData GuildIncidentsData
	if err := json.NewDecoder(body).Decode(&incidentsData); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PUT",
				"url", "/guilds/{id}/incident-actions",
				"error", err.Error(),
			)
		return GuildIncidentsData{}, err
	}
	return incidentsData, nil
}

// ModifyCurrentUserVoiceStateOptions contains parameters for modifying the current user's voice state.
type ModifyCurrentUserVoiceStateOptions struct {
	// ChannelID is the ID of the channel the user is currently in.
	ChannelID Snowflake `json:"channel_id"`

	// Suppress toggles the user's suppress state.
	Suppress *bool `json:"suppress,omitempty"`

	// RequestToSpeakTimestamp sets the user's request to speak timestamp.
	RequestToSpeakTimestamp *time.Time `json:"request_to_speak_timestamp,omitempty"`
}

// ModifyCurrentUserVoiceState updates the current user's voice state in a guild.
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-current-user-voice-state
func (r *requester) ModifyCurrentUserVoiceState(ctx context.Context, guildID Snowflake, opts ModifyCurrentUserVoiceStateOptions) error {
	payload, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	endpoint := "/guilds/" + guildID.String() + "/voice-states/@me"
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    endpoint,
		Body:   payload,
	})
	if err != nil {
		return err
	}
	defer body.Close()
	return nil
}

// ModifyUserVoiceStateOptions contains parameters for modifying another user's voice state.
type ModifyUserVoiceStateOptions struct {
	// ChannelID is the ID of the channel the user is currently in.
	ChannelID Snowflake `json:"channel_id"`

	// Suppress toggles the user's suppress state.
	Suppress *bool `json:"suppress,omitempty"`
}

// ModifyUserVoiceState updates another user's voice state in a guild.
//
// Reference: https://discord.com/developers/docs/resources/guild#modify-user-voice-state
func (r *requester) ModifyUserVoiceState(ctx context.Context, guildID, userID Snowflake, opts ModifyUserVoiceStateOptions) error {
	payload, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	endpoint := "/guilds/" + guildID.String() + "/voice-states/" + userID.String()
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    endpoint,
		Body:   payload,
	})
	if err != nil {
		return err
	}
	defer body.Close()
	return nil
}
