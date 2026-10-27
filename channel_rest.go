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
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"
)

// ModifyGroupDMOptions contains parameters for modifying a Group DM channel.
type ModifyGroupDMOptions struct {
	// Name is the channel's name (1-100 characters).
	Name string `json:"name,omitempty"`

	// Icon sets the icon for Group DM.
	Icon *Base64Image `json:"icon,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// ModifyGuildChannelOptions contains parameters for modifying a Guild channel.
type ModifyGuildChannelOptions struct {
	// Name is the channel's name (1-100 characters).
	Name string `json:"name,omitempty"`

	// Type specifies the type of channel to update.
	//
	// Note:
	//  - Defaults to ChannelTypeGuildText if unset.
	//  - Valid values include ChannelTypeGuildText, ChannelTypeGuildVoice, ChannelTypeGuildForum, etc.
	//
	// Applies to All Channels.
	Type *ChannelType `json:"type,omitempty"`

	// Position determines the channel’s position in the server’s channel list (lower numbers appear higher).
	//
	// Note:
	//  - Channels with the same position are sorted by their internal ID.
	//
	// Applies to All Channels.
	Position *int `json:"position,omitempty"`

	// Topic is a description of the channel (0-1024 characters).
	Topic *string `json:"topic,omitempty"`

	// Nsfw marks the channel as Not Safe For Work, restricting it to 18+ users.
	Nsfw *bool `json:"nsfw,omitempty"`

	// RateLimitPerUser sets the seconds a user must wait before sending another message (0-21600).
	RateLimitPerUser *int `json:"rate_limit_per_user,omitempty"`

	// Bitrate sets the audio quality for voice or stage channels (in bits, minimum 8000).
	Bitrate *int `json:"bitrate,omitempty"`

	// UserLimit caps the number of users in a voice or stage channel (0 for unlimited, 1-99 for a limit).
	UserLimit *int `json:"user_limit,omitempty"`

	// PermissionOverwrites defines custom permissions for specific roles or users.
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites,omitempty"`

	// ParentID is the ID of the category to nest the channel under.
	ParentID *Snowflake `json:"parent_id,omitempty"`

	// RtcRegion sets the voice region id for the voice channel.
	RtcRegion *string `json:"rtc_region,omitempty"`

	// VideoQualityMode sets the camera video quality for voice or stage channels.
	VideoQualityMode *VideoQualityModes `json:"video_quality_mode,omitempty"`

	// DefaultAutoArchiveDuration sets the default auto archive duration for new threads in this channel.
	DefaultAutoArchiveDuration *AutoArchiveDuration `json:"default_auto_archive_duration,omitempty"`

	// Flags sets the channel flags.
	Flags *ChannelFlags `json:"flags,omitempty"`

	// AvailableTags sets the available tags for a forum/media channel.
	AvailableTags []ForumTag `json:"available_tags,omitempty"`

	// DefaultReactionEmoji sets the default reaction emoji for a forum/media channel.
	DefaultReactionEmoji *DefaultReactionEmoji `json:"default_reaction_emoji,omitempty"`

	// DefaultThreadRateLimitPerUser sets the default thread slowmode.
	DefaultThreadRateLimitPerUser *int `json:"default_thread_rate_limit_per_user,omitempty"`

	// DefaultSortOrder sets the default sort order for a forum/media channel.
	DefaultSortOrder *ForumPostsSortOrder `json:"default_sort_order,omitempty"`

	// DefaultForumLayout sets the default layout for a forum channel.
	DefaultForumLayout *ForumLayout `json:"default_forum_layout,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// ModifyGuildThreadOptions contains parameters for modifying a Guild thread's settings.
type ModifyGuildThreadOptions struct {
	// Name is the channel's name (1-100 characters).
	Name string `json:"name,omitempty"`

	// Archived sets the archived state of the thread.
	Archived *bool `json:"archived,omitempty"`

	// AutoArchiveDuration sets the duration after which the thread will automatically archive.
	AutoArchiveDuration *AutoArchiveDuration `json:"auto_archive_duration,omitempty"`

	// Locked sets the locked state of the thread.
	Locked *bool `json:"locked,omitempty"`

	// Invitable sets whether non-moderators can invite others to the thread (private threads only).
	Invitable *bool `json:"invitable,omitempty"`

	// RateLimitPerUser sets the seconds a user must wait before sending another message (0-21600).
	RateLimitPerUser *int `json:"rate_limit_per_user,omitempty"`

	// Flags sets the channel flags.
	Flags *ChannelFlags `json:"flags,omitempty"`

	// AppliedTags sets the tags applied to a thread.
	AppliedTags []Snowflake `json:"applied_tags,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// DeleteChannelOptions contains parameters for deleting a channel, or closing a private message.
type DeleteChannelOptions struct {
	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// EditChannelPermissionsOptions contains parameters for updating a channel overwrite permissions.
type EditChannelPermissionsOptions struct {
	// Allow is the permissions to allow for the overwrite.
	Allow *Permissions `json:"allow,omitempty"`

	// Deny is the permissions to deny for the overwrite.
	Deny *Permissions `json:"deny,omitempty"`

	// Type is the type of the overwrite.
	Type PermissionOverwriteType `json:"type"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// CreateChannelInviteOptions contains parameters for creating a new invite for a channel.
type CreateChannelInviteOptions struct {
	// MaxAge is the duration of invite in seconds before expiry, or 0 for never. between 0 and 604800 (7 days).
	MaxAge *int `json:"max_age,omitempty"`

	// MaxUses is the max number of uses or 0 for unlimited. between 0 and 100.
	MaxUses int `json:"max_uses,omitempty"`

	// Temporary is whether this invite only grants temporary membership.
	Temporary bool `json:"temporary,omitempty"`

	// Unique if true, don't try to reuse a similar invite (useful for creating many unique one time use invites)
	Unique bool `json:"unique,omitempty"`

	// TargetType is the type of target for this voice channel invite.
	TargetType InviteTargetType `json:"target_type,omitempty"`

	// TargetUserID is the id of the user whose stream to display for this invite,
	// required if TargetType is InviteTargetTypeStream, the user must be streaming in the channel.
	TargetUserID Snowflake `json:"target_user_id,omitempty"`

	// TargetApplicationID is the id of the embedded application to open for this invite,
	// required if TargetType is InviteTargetTypeEmbeddedApplication, the application must have the EMBEDDED flag.
	TargetApplicationID Snowflake `json:"target_application_id,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// DeleteChannelPermissionOptions contains parameters for deleting a channel permission.
type DeleteChannelPermissionOptions struct {
	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// FollowAnnouncementChannelOptions contains parameters for following an Announcement channel.
type FollowAnnouncementChannelOptions struct {
	// WebhookChannelID is the id of target channel.
	WebhookChannelID Snowflake `json:"webhook_channel_id"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// GroupDMAddRecipientOptions contains parameters for adding a recipient to a group dm channel.
type GroupDMAddRecipientOptions struct {
	// AccessToken is the access token of a user that has granted your app the 'gdm.join' scope.
	AccessToken string `json:"access_token"`

	// Nick is the nickname of the user being added.
	Nick string `json:"nick,omitempty"`
}

// StartThreadFromMessageOptions contains parameters for starting a thread from a message.
type StartThreadFromMessageOptions struct {
	// Name is a 1-100 character channel name
	Name string `json:"name"`

	// AutoArchiveDuration is the number of minutes of inactivity after which the
	// thread will be automatically archived and stop showing in the channel list.
	// Valid values are: 60, 1440, 4320, and 10080.
	AutoArchiveDuration AutoArchiveDuration `json:"auto_archive_duration,omitempty"`

	// RateLimitPerUser is the amount of seconds a user has to wait before sending another message (0-21600).
	RateLimitPerUser int `json:"rate_limit_per_user,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// StartThreadWithoutMessageOptions contains parameters for starting a thread without a message.
type StartThreadWithoutMessageOptions struct {
	// Type the type of thread to create
	//
	// Can be one of: ChannelTypeAnnouncementThread, ChannelTypePublicThread or ChannelTypePrivateThread.
	Type ChannelType `json:"type,omitempty"`

	// Name is a 1-100 character channel name
	Name string `json:"name"`

	// AutoArchiveDuration is the number of minutes of inactivity after which the
	// thread will be automatically archived and stop showing in the channel list.
	// Valid values are: 60, 1440, 4320, and 10080.
	AutoArchiveDuration AutoArchiveDuration `json:"auto_archive_duration,omitempty"`

	// RateLimitPerUser is the amount of seconds a user has to wait before sending another message (0-21600).
	RateLimitPerUser int `json:"rate_limit_per_user,omitempty"`

	// Invitable is whether non-moderators can add other non-moderators to a thread
	//
	// Note:
	//   - only available when creating a private thread.
	Invitable *bool `json:"invitable,omitempty"`

	// Reason specifies the audit log reason for this action.
	Reason string `json:"-"`
}

// FetchThreadMemberOptions contains parameters for fetching a thread member.
type FetchThreadMemberOptions struct {
	// WithMember is whether to include a guild member object for the thread member.
	WithMember bool `json:"member,omitempty"`
}

// ListThreadMembersOptions contains parameters for listing a thread members.
type ListThreadMembersOptions struct {
	// WithMember is whether to include a guild member object for the thread member.
	WithMember bool `json:"member,omitempty"`

	// Limit is the maximum number of members to return (1-100).
	//
	//  Note:
	//   - Defaults to 100 if not specified.
	Limit int `json:"limit,omitempty"`

	// After is the member ID to start after for pagination.
	// Used to get the next page of results.
	After Snowflake `json:"after,omitempty"`
}

// ListArchivedThreadsOptions contains parameters for listing archived threads.
type ListArchivedThreadsOptions struct {
	// Limit is the maximum number of members to return (1-100).
	//
	//  Note:
	//   - Defaults to 100 if not specified.
	Limit int `json:"limit,omitempty"`

	// Returns threads archived before this timestamp.
	Before time.Time `json:"before,omitempty"`
}

// ListArchivedThreadsResponse represents the response when listing archived threads.
type ListArchivedThreadsResponse struct {
	// Threads are the archived threads.
	Threads []ThreadChannel `json:"threads"`

	// A thread member object for each returned thread the current user has joined.
	Members []ThreadMember `json:"members"`

	// HasMore is whether there are potentially additional threads that could be returned on a subsequent call.
	HasMore bool `json:"has_more"`
}

// ListJoinedPrivateArchivedThreadsOptions contains parameters for listing joined private archived threads.
type ListJoinedPrivateArchivedThreadsOptions struct {
	// Limit is the maximum number of members to return (1-100).
	//
	//  Note:
	//   - Defaults to 100 if not specified.
	Limit int `json:"limit,omitempty"`

	// Returns threads archived before this id.
	Before Snowflake `json:"before,omitempty"`
}

// FetchChannel retrieves a channel by its ID.
//
// Returns a Channel interface which can be asserted to specific types.
func (r *requester) FetchChannel(ctx context.Context, channelID Snowflake) (Channel, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/channels/" + channelID.String(),
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
			Error("failed parsing response", "method", "GET",
				"url", "/channels/{id}",
				"error", err.Error(),
			)
		return nil, err
	}
	return channel, nil
}

func (r *requester) modifyChannel(ctx context.Context, channelID Snowflake, reqBody []byte, reason string) (Channel, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/channels/" + channelID.String(),
		Body:   reqBody,
		Reason: reason,
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
			Error("failed parsing response", "method", "PATCH",
				"url", "/channels/{id}",
				"error", err.Error(),
			)
		return nil, err
	}
	return channel, nil
}

// ModifyGroupDMChannel updates a group DM channel's settings.
func (r *requester) ModifyGroupDMChannel(ctx context.Context, channelID Snowflake, opts ModifyGroupDMOptions) (*GroupDMChannel, error) {
	reqBody, _ := json.Marshal(opts)
	channel, err := r.modifyChannel(ctx, channelID, reqBody, opts.Reason)
	if err != nil {
		return nil, err
	}
	if groupdmChannel, ok := channel.(*GroupDMChannel); ok {
		return groupdmChannel, nil
	}
	return nil, fmt.Errorf("ModifyGroupDMChannel: channel ID %v is not a Group DM", channelID)
}

// ModifyGuildChannel updates a guild channel's settings.
//
// Requires the PermissionManageChannels permission.
func (r *requester) ModifyGuildChannel(ctx context.Context, channelID Snowflake, opts ModifyGuildChannelOptions) (GuildChannel, error) {
	reqBody, _ := json.Marshal(opts)
	channel, err := r.modifyChannel(ctx, channelID, reqBody, opts.Reason)
	if err != nil {
		return nil, err
	}
	if gChannel, ok := channel.(GuildChannel); ok {
		return gChannel, nil
	}
	return nil, fmt.Errorf("ModifyGuildChannel: channel ID %v is not a Guild channel", channelID)
}

// ModifyGuildThread updates a guild thread channel's settings.
//
// Requires the PermissionManageThreads permission.
func (r *requester) ModifyGuildThread(ctx context.Context, channelID Snowflake, opts ModifyGuildThreadOptions) (*ThreadChannel, error) {
	reqBody, _ := json.Marshal(opts)
	channel, err := r.modifyChannel(ctx, channelID, reqBody, opts.Reason)
	if err != nil {
		return nil, err
	}
	if threadChannel, ok := channel.(*ThreadChannel); ok {
		return threadChannel, nil
	}
	return nil, fmt.Errorf("ModifyGuildThread: channel ID %v is not a Thread channel", channelID)
}

// DeleteChannel deletes/closes a channel.
//
// Requires the PermissionManageChannels permission for the guild, or PermissionManageThreads if the channel is a thread.
func (r *requester) DeleteChannel(ctx context.Context, channelID Snowflake, opts DeleteChannelOptions) (Channel, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/channels/" + channelID.String(),
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
			Error("failed parsing response", "method", "DELETE",
				"url", "/channels/{id}",
				"error", err.Error(),
			)
		return nil, err
	}
	return channel, nil
}

// EditChannelPermissions edits the channel permission overwrites for a user or role in a channel.
func (r *requester) EditChannelPermissions(ctx context.Context, channelID Snowflake, overwriteID Snowflake, opts EditChannelPermissionsOptions) error {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    "/channels/" + channelID.String() + "/permissions/" + overwriteID.String(),
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// FetchChannelInvites returns a list of invite objects for the channel.
//
// Requires the PermissionManageChannels permission.
func (r *requester) FetchChannelInvites(ctx context.Context, channelID Snowflake) ([]FullInvite, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/channels/" + channelID.String() + "/invites",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var invites []FullInvite
	if err := json.NewDecoder(body).Decode(&invites); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/channels/{id}/invites",
				"error", err.Error(),
			)
		return nil, err
	}
	return invites, nil
}

// CreateChannelInvite creates a new invite object for the channel.
//
// Requires the PermissionCreateInstantInvite permission.
func (r *requester) CreateChannelInvite(ctx context.Context, channelID Snowflake, opts CreateChannelInviteOptions) (Invite, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/channels/" + channelID.String() + "/invites",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return Invite{}, err
	}
	defer body.Close()

	var invite Invite
	if err := json.NewDecoder(body).Decode(&invite); err != nil {
		r.logger.
			Error("failed parsing response", "method", "POST",
				"url", "/channels/{id}/invites",
				"error", err.Error(),
			)
		return Invite{}, err
	}
	return invite, nil
}

// DeleteChannelPermission deletes a channel permission overwrite for a user or role in a channel.
//
// Requires the PermissionManageRoles permission.
func (r *requester) DeleteChannelPermission(ctx context.Context, channelID Snowflake, overwriteID Snowflake, opts DeleteChannelPermissionOptions) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/channels/" + channelID.String() + "/permissions/" + overwriteID.String(),
		Reason: opts.Reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// FollowAnnouncementChannel follows an Announcement Channel to send messages to a target channel.
//
// Requires the PermissionManageWebhooks permission in the target channel.
func (r *requester) FollowAnnouncementChannel(ctx context.Context, channelID Snowflake, opts FollowAnnouncementChannelOptions) (FollowedChannel, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/channels/" + channelID.String() + "/followers",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return FollowedChannel{}, err
	}
	defer body.Close()

	var followedChannel FollowedChannel
	if err := json.NewDecoder(body).Decode(&followedChannel); err != nil {
		r.logger.
			Error("failed parsing response", "method", "POST",
				"url", "/channels/{id}/followers",
				"error", err.Error(),
			)
		return FollowedChannel{}, err
	}
	return followedChannel, nil
}

// TriggerTypingIndicator posts a typing indicator for the specified channel.
func (r *requester) TriggerTypingIndicator(ctx context.Context, channelID Snowflake) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/channels/" + channelID.String() + "/typing",
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// GroupDMAddRecipient adds a recipient to a Group DM using their access token.
func (r *requester) GroupDMAddRecipient(ctx context.Context, channelID Snowflake, userID Snowflake, opts GroupDMAddRecipientOptions) error {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    "/channels/" + channelID.String() + "/recipients/" + userID.String(),
		Body:   reqBody,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// GroupDMRemoveRecipient removes a recipient from a Group DM.
func (r *requester) GroupDMRemoveRecipient(ctx context.Context, channelID Snowflake, userID Snowflake) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/channels/" + channelID.String() + "/recipients/" + userID.String(),
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// StartThreadFromMessage creates a new thread from an existing message.
func (r *requester) StartThreadFromMessage(ctx context.Context, channelID Snowflake, messageID Snowflake, opts StartThreadFromMessageOptions) (GuildChannel, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/channels/" + channelID.String() + "/messages/" + messageID.String() + "/threads",
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
				"url", "/channels/{id}/messages/{msg_id}/threads",
				"error", err.Error(),
			)
		return nil, err
	}
	if gc, ok := channel.(GuildChannel); ok {
		return gc, nil
	}
	return nil, fmt.Errorf("StartThreadFromMessage: created channel is not a GuildChannel")
}

// StartThreadWithoutMessage creates a new thread.
func (r *requester) StartThreadWithoutMessage(ctx context.Context, channelID Snowflake, opts StartThreadWithoutMessageOptions) (GuildChannel, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/channels/" + channelID.String() + "/threads",
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
				"url", "/channels/{id}/threads",
				"error", err.Error(),
			)
		return nil, err
	}
	if gc, ok := channel.(GuildChannel); ok {
		return gc, nil
	}
	return nil, fmt.Errorf("StartThreadWithoutMessage: created channel is not a GuildChannel")
}

// JoinThread adds the current user to a thread.
func (r *requester) JoinThread(ctx context.Context, channelID Snowflake) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    "/channels/" + channelID.String() + "/thread-members/@me",
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// AddThreadMember adds a member to a thread.
func (r *requester) AddThreadMember(ctx context.Context, channelID Snowflake, userID Snowflake) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    "/channels/" + channelID.String() + "/thread-members/" + userID.String(),
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// LeaveThread removes the current user from a thread.
func (r *requester) LeaveThread(ctx context.Context, channelID Snowflake) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/channels/" + channelID.String() + "/thread-members/@me",
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// RemoveThreadMember removes a member from a thread.
//
// Requires the PermissionManageThreads permission, or the creator of the thread if it is a 'PrivateThread'.
func (r *requester) RemoveThreadMember(ctx context.Context, channelID Snowflake, userID Snowflake) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/channels/" + channelID.String() + "/thread-members/" + userID.String(),
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// FetchThreadMember retrieves a thread member.
func (r *requester) FetchThreadMember(ctx context.Context, channelID Snowflake, userID Snowflake, opts ...FetchThreadMemberOptions) (ThreadMember, error) {
	var withMember bool
	if len(opts) > 0 {
		withMember = opts[0].WithMember
	}

	endpoint := "/channels/" + channelID.String() + "/thread-members/" + userID.String() + "?with_member=" + strconv.FormatBool(withMember)
	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return ThreadMember{}, err
	}
	defer body.Close()

	var member ThreadMember
	if err := json.NewDecoder(body).Decode(&member); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/channels/{id}/thread-members/{id}",
				"error", err.Error(),
			)
		return ThreadMember{}, err
	}
	return member, nil
}

// ListThreadMembers retrieves a list of thread members.
func (r *requester) ListThreadMembers(ctx context.Context, channelID Snowflake, opts ...ListThreadMembersOptions) ([]ThreadMember, error) {
	var first ListThreadMembersOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	params := url.Values{}
	params.Set("with_member", strconv.FormatBool(first.WithMember))

	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if !first.After.UnSet() {
		params.Set("after", first.After.String())
	}

	endpoint := "/channels/" + channelID.String() + "/thread-members?" + params.Encode()

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var members []ThreadMember
	if err := json.NewDecoder(body).Decode(&members); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/channels/{id}/thread-members",
				"error", err.Error(),
			)
		return nil, err
	}
	return members, nil
}

func (r *requester) listArchivedThreads(ctx context.Context, channelID Snowflake, subEndPoint string, params url.Values) (ListArchivedThreadsResponse, error) {
	endpoint := "/channels/" + channelID.String() + subEndPoint
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{Method: "GET", URL: endpoint})
	if err != nil {
		return ListArchivedThreadsResponse{}, err
	}
	defer body.Close()

	var response ListArchivedThreadsResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/channels/{id}"+subEndPoint,
				"error", err.Error(),
			)
		return ListArchivedThreadsResponse{}, err
	}
	return response, nil
}

// ListPublicArchivedThreads retrieves a list of public archived threads.
func (r *requester) ListPublicArchivedThreads(ctx context.Context, channelID Snowflake, opts ...ListArchivedThreadsOptions) (ListArchivedThreadsResponse, error) {
	var first ListArchivedThreadsOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	params := url.Values{}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if !first.Before.IsZero() {
		params.Set("before", first.Before.Format(time.RFC3339))
	}
	return r.listArchivedThreads(ctx, channelID, "/threads/archived/public", params)
}

// ListPrivateArchivedThreads retrieves a list of private archived threads.
func (r *requester) ListPrivateArchivedThreads(ctx context.Context, channelID Snowflake, opts ...ListArchivedThreadsOptions) (ListArchivedThreadsResponse, error) {
	var first ListArchivedThreadsOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	params := url.Values{}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if !first.Before.IsZero() {
		params.Set("before", first.Before.Format(time.RFC3339))
	}
	return r.listArchivedThreads(ctx, channelID, "/threads/archived/private", params)
}

// ListJoinedPrivateArchivedThreads retrieves a list of private archived threads that the current user has joined.
func (r *requester) ListJoinedPrivateArchivedThreads(ctx context.Context, channelID Snowflake, opts ...ListJoinedPrivateArchivedThreadsOptions) (ListArchivedThreadsResponse, error) {
	var first ListJoinedPrivateArchivedThreadsOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	params := url.Values{}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if !first.Before.UnSet() {
		params.Set("before", first.Before.String())
	}
	return r.listArchivedThreads(ctx, channelID, "/users/@me/threads/archived/private", params)
}
