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
	"errors"
	"fmt"
	"time"
)

// ChannelType represents Discord channel types.
//
// Reference: https://discord.com/developers/docs/resources/channel#channel-object-channel-types
type ChannelType int

const (
	// GuildText is a text channel within a server.
	ChannelTypeGuildText ChannelType = 0

	// DM is a direct message between users.
	ChannelTypeDM ChannelType = 1

	// GuildVoice is a voice channel within a server.
	ChannelTypeGuildVoice ChannelType = 2

	// GroupDM is a direct message between multiple users.
	ChannelTypeGroupDM ChannelType = 3

	// GuildCategory is an organizational category that contains up to 50 channels.
	ChannelTypeGuildCategory ChannelType = 4

	// GuildAnnouncement is a channel that users can follow and crosspost into their own server (formerly news channels).
	ChannelTypeGuildAnnouncement ChannelType = 5

	// AnnouncementThread is a temporary sub-channel within a GuildAnnouncement channel.
	ChannelTypeAnnouncementThread ChannelType = 10

	// PublicThread is a temporary sub-channel within a GuildText or GuildForum channel.
	ChannelTypePublicThread ChannelType = 11

	// PrivateThread is a temporary sub-channel within a GuildText channel that is only viewable by those invited and those with the MANAGE_THREADS permission.
	ChannelTypePrivateThread ChannelType = 12

	// GuildStageVoice is a voice channel for hosting events with an audience.
	ChannelTypeGuildStageVoice ChannelType = 13

	// GuildDirectory is the channel in a hub containing the listed servers.
	ChannelTypeGuildDirectory ChannelType = 14

	// GuildForum is a channel that can only contain threads.
	ChannelTypeGuildForum ChannelType = 15

	// GuildMedia is a channel that can only contain threads, similar to GuildForum channels.
	ChannelTypeGuildMedia ChannelType = 16
)

// Is returns true if the channel's Type matches the provided one.
func (t ChannelType) Is(channelType ChannelType) bool {
	return t == channelType
}

// ChannelFlags represents Discord channel flags combined as a bitfield.
//
// Reference: https://discord.com/developers/docs/resources/channel#channel-object-channel-flags
type ChannelFlags int

const (
	// ChannelFlagPinned indicates that this thread is pinned to the top of its parent
	// GUILD_FORUM or GUILD_MEDIA channel.
	//
	// Applicable only to threads within forum or media channels.
	ChannelFlagPinned ChannelFlags = 1 << 1

	// ChannelFlagRequireTag indicates whether a tag is required to be specified when
	// creating a thread in a GUILD_FORUM or GUILD_MEDIA channel.
	//
	// Tags are specified in the AppliedTags field.
	ChannelFlagRequireTag ChannelFlags = 1 << 4

	// ChannelFlagHideMediaDownloadOptions, when set, hides the embedded media download options
	// for media channel attachments.
	//
	// Available only for media channels.
	ChannelFlagHideMediaDownloadOptions ChannelFlags = 1 << 15
)

// Has returns true if all provided flags are set.
func (f ChannelFlags) Has(flags ...ChannelFlags) bool {
	return BitFieldHas(f, flags...)
}

// PermissionOverwriteType defines the type of permission overwrite target.
//
// Reference: https://discord.com/developers/docs/resources/channel#overwrite-object-overwrite-structure
type PermissionOverwriteType int

const (
	// PermissionOverwriteTypeRole indicates the overwrite applies to a role.
	PermissionOverwriteTypeRole PermissionOverwriteType = 0

	// PermissionOverwriteTypeMember indicates the overwrite applies to a member.
	PermissionOverwriteTypeMember PermissionOverwriteType = 1
)

// Is returns true if the overWrite's Type matches the provided one.
func (t PermissionOverwriteType) Is(overWriteType PermissionOverwriteType) bool {
	return t == overWriteType
}

// ForumPostsSortOrder defines the sort order type used to order posts in forum/media channels.
//
// Reference: https://discord.com/developers/docs/resources/channel#channel-object-sort-order-types
type ForumPostsSortOrder int

const (
	// ForumPostsSortOrderLatestActivity sorts posts by latest activity (default).
	ForumPostsSortOrderLatestActivity ForumPostsSortOrder = 0

	// ForumPostsSortOrderCreationDate sorts posts by creation time (most recent to oldest).
	ForumPostsSortOrderCreationDate ForumPostsSortOrder = 1
)

// Is returns true if the channel's SortOrder type matches the provided one.
func (t ForumPostsSortOrder) Is(sortOrderType ForumPostsSortOrder) bool {
	return t == sortOrderType
}

// ForumLayout defines the layout type used to place posts in forum/media channels.
//
// Reference: https://discord.com/developers/docs/resources/channel#channel-object-forum-layout-types
type ForumLayout int

const (
	// ForumLayoutNotSet indicates no default has been set for forum channel.
	ForumLayoutNotSet ForumLayout = 0

	// ForumLayoutListView displays posts as a list.
	ForumLayoutListView ForumLayout = 1

	// ForumLayoutGalleryView displays posts as a collection of tiles.
	ForumLayoutGalleryView ForumLayout = 2
)

// Is returns true if the channel's PostsLayout type matches the provided one.
func (t ForumLayout) Is(layoutType ForumLayout) bool {
	return t == layoutType
}

// PermissionOverwrite represents a permission overwrite for a role or member.
//
// Used to grant or deny specific permissions in a channel.
//
// Reference: https://discord.com/developers/docs/resources/channel#overwrite-object-overwrite-structure
type PermissionOverwrite struct {
	// ID is the role or user ID the overwrite applies to.
	ID Snowflake `json:"id"`

	// Type specifies whether this overwrite is for a role or a member.
	Type PermissionOverwriteType `json:"type"`

	// Allow is the permission bit set explicitly allowed.
	Allow Permissions `json:"allow,omitempty"`

	// Deny is the permission bit set explicitly denied.
	Deny Permissions `json:"deny,omitempty"`
}

// MemberPermissionOverwrites contains the permission overwrites relevant to a specific member.
type MemberPermissionOverwrites struct {
	// Everyone contains the @everyone role overwrite if it exists.
	Everyone *PermissionOverwrite

	// Roles contains role-specific overwrites (excluding @everyone).
	Roles []PermissionOverwrite

	// Member contains the member-specific overwrite if it exists.
	Member *PermissionOverwrite
}

// permissionOverwritesFor is a helper function to compute member permission overwrites.
func permissionOverwritesFor(
	guildID Snowflake,
	overwrites []PermissionOverwrite,
	m Member,
) MemberPermissionOverwrites {
	var result MemberPermissionOverwrites

	roleMap := make(map[Snowflake]struct{}, len(m.RoleIDs))
	for _, roleID := range m.RoleIDs {
		roleMap[roleID] = struct{}{}
	}

	for _, overwrite := range overwrites {
		if overwrite.Type == PermissionOverwriteTypeMember && overwrite.ID == m.ID {
			o := overwrite
			result.Member = &o
			continue
		}

		if overwrite.Type == PermissionOverwriteTypeRole {
			if overwrite.ID == guildID {
				o := overwrite
				result.Everyone = &o
				continue
			}

			if _, hasMemberRole := roleMap[overwrite.ID]; hasMemberRole {
				result.Roles = append(result.Roles, overwrite)
			}
		}
	}

	return result
}

// permissionsFor is a helper function to compute member permissions in a channel.
func permissionsFor(basePerms Permissions, overwrites MemberPermissionOverwrites) Permissions {
	perms := basePerms

	if perms.Has(PermissionAdministrator) {
		return PermissionAll
	}

	if overwrites.Everyone != nil {
		perms.Remove(overwrites.Everyone.Deny)
		perms.Add(overwrites.Everyone.Allow)
	}

	var roleDeny, roleAllow Permissions
	for _, overwrite := range overwrites.Roles {
		roleDeny.Add(overwrite.Deny)
		roleAllow.Add(overwrite.Allow)
	}
	perms.Remove(roleDeny)
	perms.Add(roleAllow)

	if overwrites.Member != nil {
		perms.Remove(overwrites.Member.Deny)
		perms.Add(overwrites.Member.Allow)
	}

	return perms
}

// ForumTag represents a tag that can be applied to a thread
// in a GuildForum or GuildMedia channel.
//
// Reference: https://discord.com/developers/docs/resources/channel#forum-tag-object
type ForumTag struct {
	// ID is the id of the tag.
	ID Snowflake `json:"id"`

	// Name is the name of the tag (0-20 characters).
	Name string `json:"name"`

	// Moderated indicates whether this tag can only be added to or removed from
	// threads by a member with the ManageThreads permission.
	Moderated bool `json:"moderated"`

	// EmojiID is the ID of a guild's custom emoji.
	//
	// Optional:
	//  - May be equal 0.
	//
	// Note:
	//  - If EmojiName is empty (not set), then EmojiID must be set (non-zero).
	EmojiID Snowflake `json:"emoji_id,omitempty"`

	// EmojiName is the Unicode character of the emoji.
	//
	// Optional:
	//  - May be empty string.
	//
	// Note:
	//  - If EmojiName is empty (not set), then EmojiID must be set (non-zero).
	EmojiName string `json:"emoji_name,omitempty"`
}

// DefaultReactionEmoji represents a default reaction emoji for forum channels.
type DefaultReactionEmoji struct {
	// EmojiID is the ID of a guild's custom emoji.
	//
	// Optional:
	//  - May be equal to 0.
	//
	// Info:
	//  - If 0, EmojiName will be set instead.
	EmojiID Snowflake `json:"emoji_id"`

	// EmojiName is the Unicode character of the emoji.
	//
	// Optional:
	//  - May be empty string.
	//
	// Info:
	//  - If empty, EmojiID will be set instead.
	EmojiName string `json:"emoji_name"`
}

// AutoArchiveDuration represents the auto archive duration of a thread channel
//
// Reference: https://discord.com/developers/docs/resources/channel#thread-metadata-object
type AutoArchiveDuration int

const (
	AutoArchiveDuration1h  AutoArchiveDuration = 60
	AutoArchiveDuration24h AutoArchiveDuration = 1440
	AutoArchiveDuration3d  AutoArchiveDuration = 4320
	AutoArchiveDuration1w  AutoArchiveDuration = 10080
)

// Is returns true if the thread's auto archive duration matches the provided auto archive duration.
func (d AutoArchiveDuration) Is(duration AutoArchiveDuration) bool {
	return d == duration
}

// ThreadMetaData represents the metadata object that contains a number of thread-specific channel fields.
//
// Reference: https://discord.com/developers/docs/resources/channel#thread-metadata-object
type ThreadMetaData struct {
	// Archived is whether the thread is archived
	Archived bool `json:"archived"`

	// AutoArchiveDuration is the duration will thread need to stop showing in the channel list.
	AutoArchiveDuration AutoArchiveDuration `json:"auto_archive_duration"`

	// ArchiveTimestamp is the timestamp when the thread's archive status was last changed,
	// used for calculating recent activity
	ArchiveTimestamp time.Time `json:"archive_timestamp,omitzero"`

	// Locked is whether the thread is locked; when a thread is locked,
	// only users with MANAGE_THREADS can unarchive it
	Locked bool `json:"locked"`

	// Invitable is whether non-moderators can add other non-moderators to a thread.
	Invitable bool `json:"invitable"`
}

// ChannelFields contains only fields present in all channel types.
//
// Reference: https://discord.com/developers/docs/resources/channel#channel-object-channel-structure
type ChannelFields struct {
	// ID is the unique Discord snowflake ID of the channel.
	ID Snowflake `json:"id"`

	// Type is the type of the channel.
	Type ChannelType `json:"type"`
}

func (c *ChannelFields) GetID() Snowflake {
	return c.ID
}

func (c *ChannelFields) GetType() ChannelType {
	return c.Type
}

func (c *ChannelFields) CreatedAt() time.Time {
	return c.ID.Timestamp()
}

// Mention returns a Discord mention string for the channel.
//
// Example output: "<#123456789012345678>"
func (c *ChannelFields) Mention() string {
	return "<#" + c.ID.String() + ">"
}

// String implements the fmt.Stringer interface.
func (c *ChannelFields) String() string {
	return c.Mention()
}

// GuildChannelFields embeds BaseChannel and adds fields common to guild channels except threads.
//
// Used by guild-specific channel types like TextChannel, VoiceChannel, ForumChannel, etc.
type GuildChannelFields struct {
	ChannelFields

	// GuildID is the id of the guild.
	GuildID Snowflake `json:"guild_id"`

	// Name is the name of the channel.
	//
	// Info:
	//  - can be 1 to 100 characters.
	Name string `json:"name,omitempty"`

	// Position is the sorting position of the channel.
	Position int `json:"position,omitempty"`

	// PermissionOverwrites are explicit permission overwrites for members and roles.
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites,omitempty"`

	// Flags are combined channel flags.
	Flags ChannelFlags `json:"flags,omitempty"`
}

func (c *GuildChannelFields) GetGuildID() Snowflake {
	return c.GuildID
}

func (c *GuildChannelFields) GetName() string {
	return c.Name
}

func (c *GuildChannelFields) GetPosition() int {
	return c.Position
}

func (c *GuildChannelFields) GetPermissionOverwrites() []PermissionOverwrite {
	return c.PermissionOverwrites
}

func (c *GuildChannelFields) GetFlags() ChannelFlags {
	return c.Flags
}

func (c *GuildChannelFields) JumpURL() string {
	return "https://discord.com/channels/" + c.GuildID.String() + "/" + c.ID.String()
}

// PermissionOverwritesFor returns all permission overwrites relevant to the member.
func (c *GuildChannelFields) PermissionOverwritesFor(m Member) MemberPermissionOverwrites {
	return permissionOverwritesFor(c.GuildID, c.PermissionOverwrites, m)
}

// PermissionsFor returns the member's calculated permissions in the channel given their base permissions.
func (c *GuildChannelFields) PermissionsFor(basePerms Permissions, m Member) Permissions {
	overwrites := c.PermissionOverwritesFor(m)
	return permissionsFor(basePerms, overwrites)
}

// ThreadChannelFields embeds BaseChannel and adds fields common to thread channels.
type ThreadChannelFields struct {
	ChannelFields

	// GuildID is the id of the guild.
	GuildID Snowflake `json:"guild_id"`

	// Name is the name of the channel.
	//
	// Info:
	//  - can be 1 to 100 characters.
	Name string `json:"name,omitempty"`

	// PermissionOverwrites are explicit permission overwrites for members and roles.
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites,omitempty"`

	// Flags are combined channel flags.
	Flags ChannelFlags `json:"flags,omitempty"`
}

func (c *ThreadChannelFields) GetGuildID() Snowflake {
	return c.GuildID
}

func (c *ThreadChannelFields) GetName() string {
	return c.Name
}

func (c *ThreadChannelFields) GetPermissionOverwrites() []PermissionOverwrite {
	return c.PermissionOverwrites
}

func (c *ThreadChannelFields) GetFlags() ChannelFlags {
	return c.Flags
}

func (c *ThreadChannelFields) JumpURL() string {
	return "https://discord.com/channels/" + c.GuildID.String() + "/" + c.ID.String()
}

// PermissionOverwritesFor returns all permission overwrites relevant to the member.
func (c *ThreadChannelFields) PermissionOverwritesFor(m Member) MemberPermissionOverwrites {
	return permissionOverwritesFor(c.GuildID, c.PermissionOverwrites, m)
}

// PermissionsFor returns the member's calculated permissions in the channel given their base permissions.
func (c *ThreadChannelFields) PermissionsFor(basePerms Permissions, m Member) Permissions {
	overwrites := c.PermissionOverwritesFor(m)
	return permissionsFor(basePerms, overwrites)
}

// CategorizedChannelFields holds the parent category field for categorized guild channels.
type CategorizedChannelFields struct {
	// ParentID is the id of the parent category for this channel.
	//
	// Info:
	//  - Each parent category can contain up to 50 channels.
	//
	// Optional:
	//  - May be equal 0 if the channel is not in a category.
	ParentID Snowflake `json:"parent_id"`
}

func (c *CategorizedChannelFields) GetParentID() Snowflake {
	return c.ParentID
}

// MessageChannelFields holds fields related to text-based features like messaging.
type MessageChannelFields struct {
	// LastMessageID is the id of the last message sent in this channel.
	LastMessageID Snowflake `json:"last_message_id"`
}

func (t *MessageChannelFields) GetLastMessageID() Snowflake {
	return t.LastMessageID
}

// GuildMessageChannelFields holds fields related to text-based features like messaging.
type GuildMessageChannelFields struct {
	MessageChannelFields
	// RateLimitPerUser is the amount of seconds a user has to wait before sending another message.
	// Bots, as well as users with the permission manageMessages or manageChannel, are unaffected.
	RateLimitPerUser time.Duration `json:"rate_limit_per_user"`
}

func (t *GuildMessageChannelFields) GetRateLimitPerUser() time.Duration {
	return t.RateLimitPerUser
}

// NsfwChannelFields holds the NSFW indicator field.
type NsfwChannelFields struct {
	// Nsfw indicates whether the channel is NSFW.
	Nsfw bool `json:"nsfw"`
}

// TopicChannelFields holds the topic field.
type TopicChannelFields struct {
	// Topic is the channel topic.
	//
	// Length:
	//  - 0-1024 characters for text, announcement, and stage voice channels.
	//  - 0-4096 characters for forum and media channels.
	//
	// Optional:
	//  - May be empty string if the channel has no topic.
	Topic string `json:"topic"`
}

type Bitrate int

const (
	BitrateMin                 Bitrate = 8000
	BitrateMaxForStageChannels Bitrate = 64000
	BitrateMaxForNormalGuilds  Bitrate = 96000
	BitrateMaxForLevel1Guilds  Bitrate = 128000
	BitrateMaxForLevel2Guilds  Bitrate = 256000
	BitrateMaxForLevel3Guilds  Bitrate = 384000
)

// AudioChannelFields holds voice-related configuration fields.
type AudioChannelFields struct {
	// Bitrate is the bitrate (in bits) of the voice channel.
	Bitrate int `json:"bitrate"`

	// UserLimit is the user limit of the voice channel.
	UserLimit int `json:"user_limit"`

	// RtcRegion is the voice region id for the voice channel. Automatic when set to empty string.
	RtcRegion string `json:"rtc_region"`
}

func (c *AudioChannelFields) GetBitrate() int {
	return c.Bitrate
}

func (c *AudioChannelFields) GetUserLimit() int {
	return c.UserLimit
}

func (c *AudioChannelFields) GetRtcRegion() string {
	return c.RtcRegion
}

// ForumChannelFields holds forum and media channel specific fields.
type ForumChannelFields struct {
	// AvailableTags is the set of tags that can be used in this channel.
	AvailableTags []ForumTag `json:"available_tags"`

	// DefaultReactionEmoji specifies the emoji used as the default way to react to a forum post.
	DefaultReactionEmoji DefaultReactionEmoji `json:"default_reaction_emoji"`

	// DefaultSortOrder is the default sort order type used to order posts
	// in GuildForum and GuildMedia channels. Defaults to PostsSortOrderLatestActivity.
	DefaultSortOrder ForumPostsSortOrder `json:"default_sort_order"`

	// DefaultForumLayout is the default forum layout view used to display posts
	// in GuildForum channels. Defaults to ForumLayoutNotSet.
	DefaultForumLayout ForumLayout `json:"default_forum_layout"`
}

// CategoryChannel represents a guild category channel.
type CategoryChannel struct {
	GuildChannelFields
}

func (c *CategoryChannel) MarshalJSON() ([]byte, error) {
	type NoMethod CategoryChannel
	return json.Marshal((*NoMethod)(c))
}

// TextChannel represents a guild text channel.
type TextChannel struct {
	GuildChannelFields
	CategorizedChannelFields
	GuildMessageChannelFields
	NsfwChannelFields
	TopicChannelFields
}

func (c *TextChannel) MarshalJSON() ([]byte, error) {
	type NoMethod TextChannel
	return json.Marshal((*NoMethod)(c))
}

// VoiceChannel represents a guild voice channel.
type VoiceChannel struct {
	GuildChannelFields
	CategorizedChannelFields
	GuildMessageChannelFields
	NsfwChannelFields
	AudioChannelFields
}

func (c *VoiceChannel) MarshalJSON() ([]byte, error) {
	type NoMethod VoiceChannel
	return json.Marshal((*NoMethod)(c))
}

// AnnouncementChannel represents an announcement channel.
type AnnouncementChannel struct {
	GuildChannelFields
	CategorizedChannelFields
	GuildMessageChannelFields
	NsfwChannelFields
	TopicChannelFields
}

func (c *AnnouncementChannel) MarshalJSON() ([]byte, error) {
	type NoMethod AnnouncementChannel
	return json.Marshal((*NoMethod)(c))
}

// StageVoiceChannel represents a stage voice channel.
type StageVoiceChannel struct {
	GuildChannelFields
	CategorizedChannelFields
	GuildMessageChannelFields
	NsfwChannelFields
	AudioChannelFields
	TopicChannelFields
}

func (c *StageVoiceChannel) MarshalJSON() ([]byte, error) {
	type NoMethod StageVoiceChannel
	return json.Marshal((*NoMethod)(c))
}

// ForumChannel represents a guild forum channel.
type ForumChannel struct {
	GuildChannelFields
	CategorizedChannelFields
	GuildMessageChannelFields
	NsfwChannelFields
	TopicChannelFields
	ForumChannelFields
}

func (c *ForumChannel) MarshalJSON() ([]byte, error) {
	type NoMethod ForumChannel
	return json.Marshal((*NoMethod)(c))
}

// MediaChannel represents a media channel.
type MediaChannel struct {
	ForumChannel
}

type ThreadMemberFlags int

// ThreadMember represents Discord thread channel member.
//
// Reference: https://discord.com/developers/docs/resources/channel#channel-object-channel-types
type ThreadMember struct {
	// ThreadID is the id of the thread.
	ThreadID Snowflake `json:"id"`

	// UserID is the id of the member.
	UserID Snowflake `json:"user_id"`

	// JoinTimestamp is the time the user last joined the thread.
	JoinTimestamp time.Time `json:"join_timestamp,omitzero"`

	// Flags are any user-thread settings, currently only used for notifications.
	Flags ThreadMemberFlags `json:"flags"`

	// Member is the guild member object of this thread member.
	//
	// Optional:
	//   - This field is only present when 'with_member' is set to true when calling [ListThreadMembers] or [GetThreadMember].
	//
	// [ListThreadMembers]: https://discord.com/developers/docs/resources/channel#list-thread-members
	// [GetThreadMember]: https://discord.com/developers/docs/resources/channel#get-thread-member
	Member *Member `json:"member"`
}

// ThreadChannel represents the base for thread channels.
type ThreadChannel struct {
	ThreadChannelFields
	CategorizedChannelFields
	GuildMessageChannelFields
	// OwnerID is the id of this thread owner
	OwnerID Snowflake `json:"owner_id"`
	// ThreadMetadata is the metadata that contains a number of thread-specific channel fields.
	ThreadMetadata ThreadMetaData `json:"thread_metadata"`
}

func (c *ThreadChannel) MarshalJSON() ([]byte, error) {
	type NoMethod ThreadChannel
	return json.Marshal((*NoMethod)(c))
}

// DMChannelFields contains fields common to DM and Group DM channels.
type DMChannelFields struct {
	ChannelFields
	MessageChannelFields
}

// ThreadChannel represents a DM channel between the currect user and other user.
type DMChannel struct {
	DMChannelFields
	// Recipients is the list of users participating in the group DM channel.
	//
	// Info:
	//   - Contains the users involved in the group DM, excluding the current user or bot.
	Recipients []User `json:"recipients"`
}

func (c *DMChannel) MarshalJSON() ([]byte, error) {
	type NoMethod DMChannel
	return json.Marshal((*NoMethod)(c))
}

// group dm channel between multiple users
type GroupDMChannel struct {
	DMChannelFields
	// users participating in the group dm channel
	Recipients []User `json:"recipients"`
	// Icon is the custom icon for the group DM channel.
	Icon string `json:"icon"`
	// id of the creator of the group dm
	OwnerID Snowflake `json:"owner_id"`
	// application id of the creator if bot-created
	ApplicationID Snowflake `json:"application_id,omitempty"`
}

func (c *GroupDMChannel) MarshalJSON() ([]byte, error) {
	type NoMethod GroupDMChannel
	return json.Marshal((*NoMethod)(c))
}

// Channel is the interface representing a Discord channel.
//
// This interface can represent any type of channel returned by Discord,
// including text channels, voice channels, thread channels, forum channels, etc.
//
// Use this interface when you want to handle channels generically without knowing
// the specific concrete type in advance.
//
// You can convert (assert) it to a specific channel type using a type assertion or
// a type switch, as described in the official Go documentation:
//   - https://go.dev/ref/spec#Type_assertions
//   - https://go.dev/doc/effective_go#type_switch
//
// Example usage:
//
//	var myChannel Channel
//
//	switch c := ch.(type) {
//	case *TextChannel:
//	    fmt.Println("Text channel name:", c.Name)
//	case *VoiceChannel:
//	    fmt.Println("Voice channel bitrate:", c.Bitrate)
//	case *ForumChannel:
//	    fmt.Println("Forum channel tags:", c.AvailableTags)
//	default:
//	    fmt.Println("Other channel type:", c.GetType())
//	}
//
// You can also use an if-condition to check a specific type:
//
//	if textCh, ok := ch.(*TextChannel); ok {
//	    fmt.Println("Text channel:", textCh.Name)
//	}
type Channel interface {
	json.Marshaler
	fmt.Stringer
	GetID() Snowflake
	GetType() ChannelType
	CreatedAt() time.Time
	Mention() string
}

var (
	_ Channel = (*CategoryChannel)(nil)
	_ Channel = (*TextChannel)(nil)
	_ Channel = (*VoiceChannel)(nil)
	_ Channel = (*AnnouncementChannel)(nil)
	_ Channel = (*StageVoiceChannel)(nil)
	_ Channel = (*ForumChannel)(nil)
	_ Channel = (*MediaChannel)(nil)
	_ Channel = (*ThreadChannel)(nil)
	_ Channel = (*DMChannel)(nil)
	_ Channel = (*GroupDMChannel)(nil)
	_ Channel = (*FollowedChannel)(nil)
)

// MessageChannel represents a Discord text channel.
//
// This interface extends the Channel interface and adds text-channel-specific fields,
// such as the ID of the last message and the rate limit (slowmode) per user.
//
// Use this interface when you want to handle text channels specifically.
//
// You can convert (assert) it to a concrete type using a type assertion or type switch:
//
// Example usage:
//
//	var ch MessageChannel
//
//	switch c := ch.(type) {
//	case *TextChannel:
//	    fmt.Println("Text channel name:", c.GetName())
//	    fmt.Println("Last message ID:", c.GetLastMessageID())
//	case *VoiceChannel:
//	    fmt.Println("Voiec channel name:", c.GetName())
//	    fmt.Println("Last message ID:", c.GetLastMessageID())
//	case *DMChannel:
//	    fmt.Println("DM channel name:", c.GetName())
//	    fmt.Println("Last message ID:", c.GetLastMessageID())
//	default:
//	    fmt.Println("Other text channel type:", c.GetType())
//	}
//
// You can also use an if-condition to check a specific type:
//
//	if textCh, ok := ch.(*TextChannel); ok {
//	    fmt.Println("Text channel:", textCh.GetName())
//	}
type MessageChannel interface {
	Channel
	// GetLastMessageID returns the Snowflake ID to the last message sent in this channel.
	//
	// Note:
	//   - Will always return 0 if no Message has been sent yet.
	GetLastMessageID() Snowflake
}

var (
	_ MessageChannel = (*TextChannel)(nil)
	_ MessageChannel = (*VoiceChannel)(nil)
	_ MessageChannel = (*AnnouncementChannel)(nil)
	_ MessageChannel = (*StageVoiceChannel)(nil)
	_ MessageChannel = (*ForumChannel)(nil)
	_ MessageChannel = (*MediaChannel)(nil)
	_ MessageChannel = (*ThreadChannel)(nil)
	_ MessageChannel = (*DMChannel)(nil)
	_ MessageChannel = (*GroupDMChannel)(nil)
)

// NamedChannel represents a Discord channel that has a name.
//
// This interface is used for channel types that expose a name, such as text channels,
// voice channels, forum channels, thread channels, DM channels, and Group DM channels.
//
// Use this interface when you want to handle channels generically by their name without
// knowing the specific concrete type in advance.
//
// You can convert (assert) it to a specific channel type using a type assertion or a type
// switch, as described in the official Go documentation:
//   - https://go.dev/ref/spec#Type_assertions
//   - https://go.dev/doc/effective_go#type_switch
//
// Example usage:
//
//	var ch NamedChannel
//
//	// Using a type switch to handle specific channel types
//	switch c := ch.(type) {
//	case *TextChannel:
//	    fmt.Println("Text channel name:", c.GetName())
//	case *VoiceChannel:
//	    fmt.Println("Voice channel name:", c.GetName())
//	default:
//	    fmt.Println("Other named channel type:", c.GetType())
//	}
//
//	// Using a type assertion to check a specific type
//	if textCh, ok := ch.(*TextChannel); ok {
//	    fmt.Println("Text channel name:", textCh.GetName())
//	}
type NamedChannel interface {
	Channel
	GetName() string
}

var (
	_ NamedChannel = (*CategoryChannel)(nil)
	_ NamedChannel = (*TextChannel)(nil)
	_ NamedChannel = (*VoiceChannel)(nil)
	_ NamedChannel = (*AnnouncementChannel)(nil)
	_ NamedChannel = (*StageVoiceChannel)(nil)
	_ NamedChannel = (*ForumChannel)(nil)
	_ NamedChannel = (*MediaChannel)(nil)
	_ NamedChannel = (*ThreadChannel)(nil)
)

// GuildChannel represents a guild-specific Discord channel.
//
// This interface extends the Channel interface and adds guild-specific fields,
// such as the guild ID, channel name, permission overwrites, flags, and jump URL.
//
// Use this interface when you want to handle guild channels generically without
// knowing the specific concrete type (TextChannel, VoiceChannel, ForumChannel, etc.).
//
// You can convert (assert) it to a specific guild channel type using a type assertion
// or a type switch, as described in the official Go documentation:
//   - https://go.dev/ref/spec#Type_assertions
//   - https://go.dev/doc/effective_go#type_switch
//
// Example usage:
//
//	var myGuildChannel GuildChannel
//
//	switch c := ch.(type) {
//	case *TextChannel:
//	    fmt.Println("Text channel name:", c.Name)
//	case *VoiceChannel:
//	    fmt.Println("Voice channel bitrate:", c.Bitrate)
//	case *ForumChannel:
//	    fmt.Println("Forum channel tags:", c.AvailableTags)
//	default:
//	    fmt.Println("Other guild channel type:", c.GetType())
//	}
//
// You can also use an if-condition to check a specific type:
//
//	if textCh, ok := ch.(*TextChannel); ok {
//	    fmt.Println("Text channel:", textCh.Name)
//	}
type GuildChannel interface {
	Channel
	NamedChannel
	GetGuildID() Snowflake
	GetPermissionOverwrites() []PermissionOverwrite
	GetFlags() ChannelFlags
	JumpURL() string
	PermissionOverwritesFor(member Member) MemberPermissionOverwrites
	PermissionsFor(basePerms Permissions, member Member) Permissions
}

var (
	_ GuildChannel = (*CategoryChannel)(nil)
	_ GuildChannel = (*TextChannel)(nil)
	_ GuildChannel = (*VoiceChannel)(nil)
	_ GuildChannel = (*AnnouncementChannel)(nil)
	_ GuildChannel = (*StageVoiceChannel)(nil)
	_ GuildChannel = (*ForumChannel)(nil)
	_ GuildChannel = (*MediaChannel)(nil)
	_ GuildChannel = (*ThreadChannel)(nil)
)

// GuildMessageChannel represents a Discord text channel.
//
// This interface extends the Channel interface and adds text-channel-specific fields,
// such as the ID of the last message and the rate limit (slowmode) per user.
//
// Use this interface when you want to handle text channels specifically.
//
// You can convert (assert) it to a concrete type using a type assertion or type switch:
//
// Example usage:
//
//	var ch GuildMessageChannel
//
//	switch c := ch.(type) {
//	case *TextChannel:
//	    fmt.Println("Text channel name:", c.GetName())
//	    fmt.Println("Last message ID:", c.GetLastMessageID())
//	    fmt.Println("Rate limit per user:", c.GetRateLimitPerUser())
//	case *VoiceChannel:
//	    fmt.Println("Voiec channel name:", c.GetName())
//	    fmt.Println("Last message ID:", c.GetLastMessageID())
//	    fmt.Println("Rate limit per user:", c.GetRateLimitPerUser())
//	default:
//	    fmt.Println("Other text channel type:", c.GetType())
//	}
//
// You can also use an if-condition to check a specific type:
//
//	if textCh, ok := ch.(*TextChannel); ok {
//	    fmt.Println("Text channel:", textCh.GetName())
//	}
type GuildMessageChannel interface {
	GuildChannel
	MessageChannel
	GetRateLimitPerUser() time.Duration
}

var (
	_ GuildMessageChannel = (*TextChannel)(nil)
	_ GuildMessageChannel = (*VoiceChannel)(nil)
	_ GuildMessageChannel = (*AnnouncementChannel)(nil)
	_ GuildMessageChannel = (*StageVoiceChannel)(nil)
	_ GuildMessageChannel = (*ForumChannel)(nil)
	_ GuildMessageChannel = (*MediaChannel)(nil)
	_ GuildMessageChannel = (*ThreadChannel)(nil)
)

// PositionedChannel represents a Discord channel that has a sorting position within its parent category.
//
// This interface is used for guild channels that have a defined position, such as category channels, text channels,
// voice channels, announcement channels, stage voice channels, forum channels, and media channels.
// The position determines the order in which channels appear within their parent category in the
// Discord client. If the channel is not under a parent category, the position is relative to other
// top-level channels in the guild.
//
// Use this interface when you want to handle channels generically by their position without knowing
// the specific concrete type in advance.
//
// You can convert (assert) it to a specific channel type using a type assertion or a type switch,
// as described in the official Go documentation:
//   - https://go.dev/ref/spec#Type_assertions
//   - https://go.dev/doc/effective_go#type_switch
//
// Example usage:
//
//	var ch PositionedChannel
//
//	// Using a type switch to handle specific channel types
//	switch c := ch.(type) {
//	case *TextChannel:
//	    fmt.Println("Text channel position:", c.GetPosition())
//	case *VoiceChannel:
//	    fmt.Println("Voice channel position:", c.GetPosition())
//	case *ForumChannel:
//	    fmt.Println("Forum channel position:", c.GetPosition())
//	default:
//	    fmt.Println("Other positioned channel type:", c.GetType())
//	}
//
//	// Using a type assertion to check a specific type
//	if textCh, ok := ch.(*TextChannel); ok {
//	    fmt.Println("Text channel position:", textCh.GetPosition())
//	}
type PositionedChannel interface {
	NamedChannel
	GetPosition() int
}

var (
	_ PositionedChannel = (*CategoryChannel)(nil)
	_ PositionedChannel = (*TextChannel)(nil)
	_ PositionedChannel = (*VoiceChannel)(nil)
	_ PositionedChannel = (*AnnouncementChannel)(nil)
	_ PositionedChannel = (*StageVoiceChannel)(nil)
	_ PositionedChannel = (*ForumChannel)(nil)
	_ PositionedChannel = (*MediaChannel)(nil)
)

// CategorizedChannel represents a Discord channel that can be placed under a parent category channel within a guild.
//
// This interface is used for guild channels that can be organized under a category, such as text channels,
// voice channels, announcement channels, stage voice channels, forum channels, media channels, and thread channels.
//
// Use this interface when you want to handle channels generically by their parent category without knowing
// the specific concrete type in advance.
//
// You can convert (assert) it to a specific channel type using a type assertion or a type switch,
// as described in the official Go documentation:
//   - https://go.dev/ref/spec#Type_assertions
//   - https://go.dev/doc/effective_go#type_switch
//
// Example usage:
//
//	var ch CategorizedChannel
//
//	// Using a type switch to handle specific channel types
//	switch c := ch.(type) {
//	case *TextChannel:
//	    fmt.Println("Text channel parent ID:", c.GetParentID())
//	case *VoiceChannel:
//	    fmt.Println("Voice channel parent ID:", c.GetParentID())
//	case *ThreadChannel:
//	    fmt.Println("Thread channel parent ID:", c.GetParentID())
//	default:
//	    fmt.Println("Other categorized channel type:", c.GetType())
//	}
//
//	// Using a type assertion to check a specific type
//	if textCh, ok := ch.(*TextChannel); ok {
//	    fmt.Println("Text channel parent ID:", textCh.GetParentID())
//	}
type CategorizedChannel interface {
	NamedChannel
	GetParentID() Snowflake
}

var (
	_ CategorizedChannel = (*TextChannel)(nil)
	_ CategorizedChannel = (*VoiceChannel)(nil)
	_ CategorizedChannel = (*AnnouncementChannel)(nil)
	_ CategorizedChannel = (*StageVoiceChannel)(nil)
	_ CategorizedChannel = (*ForumChannel)(nil)
	_ CategorizedChannel = (*MediaChannel)(nil)
	_ CategorizedChannel = (*ThreadChannel)(nil)
)

// AudioChannel represents a Discord channel that supports voice or audio functionality.
//
// This interface is used for guild channels that have voice-related features, such as voice channels
// and stage voice channels. It provides access to audio-specific properties like bitrate, user limit,
// and RTC region.
//
// Note:
//   - DM channels (ChannelTypeDM) and Group DM channels (ChannelTypeGroupDM) support audio features
//     like calls, streams, and webcams for users. However, for bots, these channels are treated as
//     text channels, as bots cannot interact with their audio features (e.g., bots cannot initiate calls in them).
//
// Use this interface when you want to handle audio channels generically without knowing
// the specific concrete type in advance.
//
// You can convert (assert) it to a specific channel type using a type assertion or a type switch,
// as described in the official Go documentation:
//   - https://go.dev/ref/spec#Type_assertions
//   - https://go.dev/doc/effective_go#type_switch
//
// Example usage:
//
//	var ch AudioChannel
//
//	// Using a type switch to handle specific channel types
//	switch c := ch.(type) {
//	case *VoiceChannel:
//	    fmt.Println("Voice channel bitrate:", c.GetBitrate())
//	    fmt.Println("Voice channel user limit:", c.GetUserLimit())
//	    fmt.Println("Voice channel RTC region:", c.GetRtcRegion())
//	case *StageVoiceChannel:
//	    fmt.Println("Stage voice channel bitrate:", c.GetBitrate())
//	    fmt.Println("Stage voice channel user limit:", c.GetUserLimit())
//	    fmt.Println("Stage voice channel RTC region:", c.GetRtcRegion())
//	}
//
//	// Using a type assertion to check a specific type
//	if voiceCh, ok := ch.(*VoiceChannel); ok {
//	    fmt.Println("Voice channel bitrate:", voiceCh.GetBitrate())
//	}
type AudioChannel interface {
	GuildChannel
	GuildMessageChannel
	GetBitrate() int
	GetUserLimit() int
	GetRtcRegion() string
}

var (
	_ AudioChannel = (*VoiceChannel)(nil)
	_ AudioChannel = (*StageVoiceChannel)(nil)
)

// Helper func to Unmarshal any channel type to a Channel interface.
func UnmarshalChannel(buf []byte) (Channel, error) {
	var meta struct {
		Type ChannelType `json:"type"`
	}
	if err := json.Unmarshal(buf, &meta); err != nil {
		return nil, err
	}

	switch meta.Type {
	case ChannelTypeGuildCategory:
		var c CategoryChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeGuildText:
		var c TextChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeGuildVoice:
		var c VoiceChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeGuildAnnouncement:
		var c AnnouncementChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeGuildStageVoice:
		var c StageVoiceChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeGuildForum:
		var c ForumChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeGuildMedia:
		var c MediaChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeAnnouncementThread,
		ChannelTypePrivateThread,
		ChannelTypePublicThread:
		var c ThreadChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeDM:
		var c DMChannel
		return &c, json.Unmarshal(buf, &c)
	case ChannelTypeGroupDM:
		var c GroupDMChannel
		return &c, json.Unmarshal(buf, &c)
	default:
		return nil, errors.New("unknown channel type")
	}
}

// PartialChannel represents a partial channel object returned in interaction payloads,
// interaction resolved data (data.resolved.channels), and invites.
//
// Partial Channel objects only have id, name, type, permissions, app_permissions,
// last_message_id, last_pin_timestamp, nsfw, parent_id, guild_id, flags, rate_limit_per_user,
// topic and position fields. Threads will also have the thread_metadata field.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-resolved-data-structure
type PartialChannel struct {
	ChannelFields

	// Name is the name of the channel.
	Name string `json:"name,omitempty"`

	// Permissions contains the computed permissions for the invoking user in this channel.
	Permissions Permissions `json:"permissions,omitempty"`

	// AppPermissions contains the permissions the app has in this channel.
	AppPermissions Permissions `json:"app_permissions,omitempty"`

	// LastMessageID is the ID of the last message sent in this channel.
	LastMessageID Snowflake `json:"last_message_id,omitempty"`

	// LastPinTimestamp is when the last pinned message was pinned.
	LastPinTimestamp *time.Time `json:"last_pin_timestamp,omitzero"`

	// NSFW indicates whether the channel is marked as NSFW.
	NSFW bool `json:"nsfw,omitempty"`

	// ParentID is the ID of the parent category or parent text channel (for threads).
	ParentID Snowflake `json:"parent_id,omitempty"`

	// GuildID is the ID of the guild the channel belongs to.
	GuildID Snowflake `json:"guild_id,omitempty"`

	// Flags contains channel flags.
	Flags ChannelFlags `json:"flags,omitempty"`

	// RateLimitPerUser is the amount of seconds a user has to wait before sending another message.
	RateLimitPerUser int `json:"rate_limit_per_user,omitempty"`

	// Topic is the channel topic.
	Topic string `json:"topic,omitempty"`

	// Position is the explicit position of the channel in the channel list.
	Position int `json:"position,omitempty"`

	// ThreadMetadata contains thread-specific metadata if this channel is a thread.
	ThreadMetadata *ThreadMetaData `json:"thread_metadata,omitempty"`
}

func (c *PartialChannel) GetName() string {
	return c.Name
}

func (c *PartialChannel) MarshalJSON() ([]byte, error) {
	type alias PartialChannel
	return json.Marshal((*alias)(c))
}

var _ Channel = (*PartialChannel)(nil)

// ResolvedChannel is an alias for PartialChannel for backwards compatibility.
type ResolvedChannel = PartialChannel

type ResolvedThread struct {
	ThreadChannel
	Permissions Permissions `json:"permissions"`
}

type VideoQualityModes int

const (
	VideoQualityModesAuto VideoQualityModes = iota + 1
	VideoQualityModesFull
)

// Is returns true if this VideoQualityModes matches the provided one.
func (v VideoQualityModes) Is(other VideoQualityModes) bool {
	return v == other
}

// FollowedChannel represents a channel that is followed.
type FollowedChannel struct {
	// ChannelID is the source channel id.
	ChannelID Snowflake `json:"channel_id"`

	// WebhookID is the created target webhook id.
	WebhookID Snowflake `json:"webhook_id"`
}

var _ Channel = (*FollowedChannel)(nil)

// GetID returns the ID of the channel.
func (c *FollowedChannel) GetID() Snowflake {
	return c.ChannelID
}

// GetType returns the ChannelType of this channel.
func (c *FollowedChannel) GetType() ChannelType {
	return ChannelTypeGuildAnnouncement
}

// CreatedAt returns the timestamp when the channel was created.
func (c *FollowedChannel) CreatedAt() time.Time {
	return c.ChannelID.Timestamp()
}

// FollowedAt returns the timestamp when the channel was followed.
func (c *FollowedChannel) FollowedAt() time.Time {
	return c.WebhookID.Timestamp()
}

// Mention returns a Discord mention string for the channel.
//
// Example output: "<#123456789012345678>"
func (c *FollowedChannel) Mention() string {
	return "<#" + c.ChannelID.String() + ">"
}

// String implements the fmt.Stringer interface.
func (c *FollowedChannel) String() string {
	return c.Mention()
}

// MarshalJSON marshals FollowedChannel to JSON.
func (c *FollowedChannel) MarshalJSON() ([]byte, error) {
	type NoMethod FollowedChannel
	return json.Marshal((*NoMethod)(c))
}

// Has returns true if the flags contain all of the provided flags.
func (f ThreadMemberFlags) Has(flags ...ThreadMemberFlags) bool {
	return BitFieldHas(f, flags...)
}
