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
	"slices"
	"time"
)

// MemberFlags represents flags of a guild member.
//
// Reference: https://discord.com/developers/docs/resources/guild#guild-member-object-guild-member-flags
type MemberFlags int

const (
	// Member has left and rejoined the guild
	//  - Editable: false
	MemberFlagDidRejoin MemberFlags = 1 << iota

	// Member has completed onboarding
	//  - Editable: false
	MemberFlagCompletedOnboarding

	// Member is exempt from guild verification requirements
	//  - Editable: true
	MemberFlagBypassesVerification

	// Member has started onboarding
	//  - Editable: false
	MemberFlagStartedOnboarding

	// Member is a guest and can only access the voice channel they were invited to
	//  - Editable: false
	MemberFlagIsGuest

	// Member has started Server Guide new member actions
	//  - Editable: false
	MemberFlagStartedHomeActions

	// Member has completed Server Guide new member actions
	//  - Editable: false
	MemberFlagCompletedHomeActions

	// Member's username, display name, or nickname is blocked by AutoMod
	//  - Editable: false
	MemberFlagQuarantinedUsername
	_

	// Member has dismissed the DM settings upsell
	//  - Editable: false
	MemberFlagDMSettingsUpsellAcknowledged

	// Member's guild tag is blocked by AutoMod
	//  - Editable: false
	MemberFlagQuarantinedGuildTag
)

// Has returns true if all provided flags are set.
func (f MemberFlags) Has(flags ...MemberFlags) bool {
	return BitFieldHas(f, flags...)
}

// Member represents a user's membership in a specific guild.
//
// This contains guild-specific information like nickname, roles, and join date.
type Member struct {
	// ID is the user's unique Discord ID.
	ID Snowflake `json:"id"`

	// GuildID is the ID of the guild this member belongs to.
	GuildID Snowflake `json:"guild_id"`

	// Nickname is the member's guild-specific nickname.
	// Empty if no nickname is set.
	Nickname string `json:"nick"`

	// Avatar is the member's guild-specific avatar hash.
	// This is different from the user's global avatar.
	// Empty if the member hasn't set a guild avatar.
	Avatar string `json:"avatar"`

	// Banner is the member's guild-specific banner hash.
	// This is different from the user's global banner.
	// Empty if the member hasn't set a guild banner.
	Banner string `json:"banner"`

	// RoleIDs contains the IDs of all roles assigned to this member.
	RoleIDs []Snowflake `json:"roles,omitempty"`

	// JoinedAt is the timestamp when the user joined the guild.
	// May be nil if the member was invited as a guest.
	JoinedAt *time.Time `json:"joined_at,omitzero"`

	// PremiumSince is the timestamp when the user started boosting the guild.
	// Nil if the member is not currently boosting.
	PremiumSince *time.Time `json:"premium_since,omitzero"`

	// Deaf indicates whether the user is deafened in voice channels.
	Deaf bool `json:"deaf,omitempty"`

	// Mute indicates whether the user is muted in voice channels.
	Mute bool `json:"mute,omitempty"`

	// Flags contains the member's guild-specific flags.
	Flags MemberFlags `json:"flags"`

	// Pending indicates whether the user has completed the guild's Membership Screening.
	// True if they still need to complete screening.
	Pending bool `json:"pending"`

	// CommunicationDisabledUntil is when the member's timeout expires.
	// Nil or a past time if the member is not currently timed out.
	CommunicationDisabledUntil *time.Time `json:"communication_disabled_until,omitzero"`

	// AvatarDecorationData contains information about the member's avatar decoration.
	AvatarDecorationData *AvatarDecorationData `json:"avatar_decoration_data"`
}

// Mention returns a Discord mention string for the user.
//
// Example output: "<@123456789012345678>"
func (m *Member) Mention() string {
	return "<@" + m.ID.String() + ">"
}

// String implements the fmt.Stringer interface.
func (m *Member) String() string {
	return m.Mention()
}

// CreatedAt returns the time when this member account is created.
func (m *Member) CreatedAt() time.Time {
	return m.ID.Timestamp()
}

// AvatarURL returns the URL to the members's avatar image.
//
// If the member has a custom avatar set, it returns the URL to that avatar, otherwise empty string.
// By default, it uses GIF format if the avatar is animated, otherwise PNG.
//
// Example usage:
//
//	url := member.AvatarURL()
func (m *Member) AvatarURL() string {
	if m.Avatar != "" {
		return GuildMemberAvatarURL(m.GuildID, m.ID, m.Avatar, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// AvatarURLWith returns the URL to the member's avatar image,
// allowing explicit specification of image format and size.
//
// If the user has a custom avatar set, it returns the URL to that avatar, otherwise empty string.
//
// Example usage:
//
//	url := member.AvatarURLWith(ImageFormatWebP, ImageSize512)
func (m *Member) AvatarURLWith(format ImageFormat, size ImageSize) string {
	if m.Avatar != "" {
		return GuildMemberAvatarURL(m.GuildID, m.ID, m.Avatar, format, size)
	}
	return ""
}

// BannerURL returns the URL to the member's banner image.
//
// If the member has a custom banner set, it returns the URL to that banner, otherwise empty string.
// BannerURL returns the URL to the member's banner image.
//
// If the member has a custom banner set, it returns the URL to that banner, otherwise empty string.
// By default, it uses GIF format if the banner is animated, otherwise PNG.
//
// Example usage:
//
//	url := member.BannerURL()
func (m *Member) BannerURL() string {
	if m.Banner != "" {
		return GuildMemberBannerURL(m.GuildID, m.ID, m.Banner, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// BannerURLWith returns the URL to the member's banner image,
// allowing explicit specification of image format and size.
//
// If the user has a custom banner set, it returns the URL to that banner, otherwise empty string.
//
// Example usage:
//
//	url := member.BannerURLWith(ImageFormatWebP, ImageSize512)
func (m *Member) BannerURLWith(format ImageFormat, size ImageSize) string {
	if m.Banner != "" {
		return GuildMemberBannerURL(m.GuildID, m.ID, m.Banner, format, size)
	}
	return ""
}

// AvatarDecorationURL returns the URL to the member's avatar decoration image.
//
// If the member has no avatar decoration, it returns an empty string.
//
// Example usage:
//
//	url := member.AvatarDecorationURL()
func (m *Member) AvatarDecorationURL() string {
	if m.AvatarDecorationData != nil {
		return AvatarDecorationURL(m.AvatarDecorationData.Asset, ImageSizeDefault)
	}
	return ""
}

// AvatarDecorationURLWith returns the URL to the member's avatar decoration image,
// allowing explicit specification of image size.
//
// If the member has no avatar decoration, it returns an empty string.
//
// Example usage:
//
//	url := member.AvatarDecorationURLWith(ImageSize512)
func (m *Member) AvatarDecorationURLWith(size ImageSize) string {
	if m.AvatarDecorationData != nil {
		return AvatarDecorationURL(m.AvatarDecorationData.Asset, size)
	}
	return ""
}

// HasRole returns true if the member has the specified role.
func (m *Member) HasRole(roleID Snowflake) bool {
	return slices.Contains(m.RoleIDs, roleID)
}

// IsTimedOut returns true if the member is currently timed out.
func (m *Member) IsTimedOut() bool {
	return m.CommunicationDisabledUntil != nil && m.CommunicationDisabledUntil.After(time.Now())
}

// IsBoosting returns true if the member is currently server boosting.
func (m *Member) IsBoosting() bool {
	return m.PremiumSince != nil
}

// PartialMember returns a PartialMember representation of the member.
func (m *Member) PartialMember() PartialMember {
	return PartialMember{
		ID:                         m.ID,
		GuildID:                    m.GuildID,
		Nickname:                   m.Nickname,
		Avatar:                     m.Avatar,
		Banner:                     m.Banner,
		RoleIDs:                    m.RoleIDs,
		JoinedAt:                   m.JoinedAt,
		PremiumSince:               m.PremiumSince,
		Flags:                      m.Flags,
		Pending:                    m.Pending,
		CommunicationDisabledUntil: m.CommunicationDisabledUntil,
		AvatarDecorationData:       m.AvatarDecorationData,
	}
}

// FullMember represents a complete guild member with their associated user information.
//
// This extends Member with the full User object, combining both
// guild-specific data (roles, nickname) and global user data (username, avatar).
type FullMember struct {
	Member

	// User contains the member's global Discord user information.
	User User `json:"user"`
}

// DisplayName returns the member's nickname if set,
// otherwise it returns their global display name if set,
// otherwise it falls back to their username.
//
// - Nickname: a guild-specific name set by the user or server mods.
// - Globalname: the name shown across Discord (can differ from username).
// - Username: the original account username.
//
// Example usage:
//
//	name := member.DisplayName()
func (m *FullMember) DisplayName() string {
	if m.Nickname != "" {
		return m.Nickname
	}
	return m.User.DisplayName()
}

// AvatarURL returns the URL to the members's avatar image.
//
// If the member has a custom avatar set, it returns the URL to that avatar.
// Otherwise it returns their global user avatar URL,
// By default, it uses GIF format if the avatar is animated, otherwise PNG.
//
// Example usage:
//
//	url := member.AvatarURL()
func (m *FullMember) AvatarURL() string {
	if m.Avatar != "" {
		return GuildMemberAvatarURL(m.GuildID, m.ID, m.Avatar, ImageFormatDefault, ImageSizeDefault)
	}
	return m.User.AvatarURL()
}

// AvatarURLWith returns the URL to the member's avatar image,
// allowing explicit specification of image format and size.
//
// If the user has a custom avatar set, it returns the URL to that avatar.
// Otherwise it returns their global user avatar URL using the provided format and size.
//
// Example usage:
//
//	url := member.AvatarURLWith(ImageFormatWebP, ImageSize512)
func (m *FullMember) AvatarURLWith(format ImageFormat, size ImageSize) string {
	if m.Avatar != "" {
		return GuildMemberAvatarURL(m.GuildID, m.ID, m.Avatar, format, size)
	}
	return m.User.AvatarURLWith(format, size)
}

// BannerURL returns the URL to the member's banner image.
//
// If the member has a custom banner set, it returns the URL to that banner.
// Otherwise it returns their global user banner URL,
// By default, it uses GIF format if the banner is animated, otherwise PNG.
//
// Example usage:
//
//	url := member.BannerURL()
func (m *FullMember) BannerURL() string {
	if m.Banner != "" {
		return GuildMemberBannerURL(m.GuildID, m.ID, m.Banner, ImageFormatDefault, ImageSizeDefault)
	}
	return m.User.BannerURL()
}

// BannerURLWith returns the URL to the member's banner image,
// allowing explicit specification of image format and size.
//
// If the user has a custom banner set, it returns the URL to that avatar.
// Otherwise it returns their global user banner URL using the provided format and size.
//
// Example usage:
//
//	url := member.BannerURLWith(ImageFormatWebP, ImageSize512)
func (m *FullMember) BannerURLWith(format ImageFormat, size ImageSize) string {
	if m.Banner != "" {
		return GuildMemberBannerURL(m.GuildID, m.ID, m.Banner, format, size)
	}
	return m.User.BannerURLWith(format, size)
}

// InteractionMember represents the invoking guild member sent at the root of an interaction.
// It contains full member details, the nested User object, and channel permissions.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object
type InteractionMember struct {
	FullMember

	// Permissions contains all permissions the member has in the channel where the interaction occurred.
	Permissions Permissions `json:"permissions,omitempty"`
}

// PartialMember represents a partial member object in interaction resolved data (data.resolved.members).
//
// Discord explicitly omits user, deaf, and mute fields from resolved.members.
// The corresponding user object is provided separately in resolved.users.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-resolved-data-structure
type PartialMember struct {
	// ID is the unique ID of the user this member represents.
	ID Snowflake `json:"id,omitzero"`

	// GuildID is the ID of the guild this member belongs to.
	GuildID Snowflake `json:"guild_id,omitzero"`

	// Nickname is the member's guild-specific nickname.
	// Empty if no nickname is set.
	Nickname string `json:"nick"`

	// Avatar is the member's guild-specific avatar hash.
	// Empty if the member hasn't set a guild avatar.
	Avatar string `json:"avatar"`

	// Banner is the member's guild-specific banner hash.
	// Empty if the member hasn't set a guild banner.
	Banner string `json:"banner"`

	// RoleIDs contains the IDs of all roles assigned to this member.
	RoleIDs []Snowflake `json:"roles,omitempty"`

	// JoinedAt is the timestamp when the user joined the guild.
	// May be nil if the member was invited as a guest.
	JoinedAt *time.Time `json:"joined_at,omitzero"`

	// PremiumSince is the timestamp when the user started boosting the guild.
	// Nil if the member is not currently boosting.
	PremiumSince *time.Time `json:"premium_since,omitzero"`

	// Permissions contains all permissions the member has in the channel where the interaction occurred.
	Permissions Permissions `json:"permissions,omitempty"`

	// Flags contains the member's guild-specific flags.
	Flags MemberFlags `json:"flags"`

	// Pending indicates whether the user has completed the guild's Membership Screening.
	Pending bool `json:"pending"`

	// CommunicationDisabledUntil is when the member's timeout expires.
	// Nil or a past time if the member is not currently timed out.
	CommunicationDisabledUntil *time.Time `json:"communication_disabled_until,omitzero"`

	// AvatarDecorationData contains information about the member's avatar decoration.
	AvatarDecorationData *AvatarDecorationData `json:"avatar_decoration_data"`
}

// DisplayName returns the member's nickname if set, or empty string.
func (m *PartialMember) DisplayName() string {
	return m.Nickname
}

// HasRole returns true if the member has the specified role ID.
func (m *PartialMember) HasRole(roleID Snowflake) bool {
	return slices.Contains(m.RoleIDs, roleID)
}

// IsTimedOut returns true if the member is currently timed out.
func (m *PartialMember) IsTimedOut() bool {
	return m.CommunicationDisabledUntil != nil && m.CommunicationDisabledUntil.After(time.Now())
}

// IsBoosting returns true if the member is currently server boosting.
func (m *PartialMember) IsBoosting() bool {
	return m.PremiumSince != nil
}

// HasPermission returns true if the member has the specified permission in the interaction channel.
func (m *PartialMember) HasPermission(perm Permissions) bool {
	return m.Permissions.Has(perm)
}

// AvatarURL returns the URL to the member's guild avatar image.
func (m *PartialMember) AvatarURL() string {
	if m.Avatar != "" {
		return GuildMemberAvatarURL(m.GuildID, m.ID, m.Avatar, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// AvatarURLWith returns the URL to the member's guild avatar image with custom format and size.
func (m *PartialMember) AvatarURLWith(format ImageFormat, size ImageSize) string {
	if m.Avatar != "" {
		return GuildMemberAvatarURL(m.GuildID, m.ID, m.Avatar, format, size)
	}
	return ""
}

// BannerURL returns the URL to the member's guild banner image.
func (m *PartialMember) BannerURL() string {
	if m.Banner != "" {
		return GuildMemberBannerURL(m.GuildID, m.ID, m.Banner, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// BannerURLWith returns the URL to the member's guild banner image with custom format and size.
func (m *PartialMember) BannerURLWith(format ImageFormat, size ImageSize) string {
	if m.Banner != "" {
		return GuildMemberBannerURL(m.GuildID, m.ID, m.Banner, format, size)
	}
	return ""
}

// AvatarDecorationURL returns the URL to the member's avatar decoration image.
func (m *PartialMember) AvatarDecorationURL() string {
	if m.AvatarDecorationData != nil {
		return AvatarDecorationURL(m.AvatarDecorationData.Asset, ImageSizeDefault)
	}
	return ""
}

// AvatarDecorationURLWith returns the URL to the member's avatar decoration image with custom size.
func (m *PartialMember) AvatarDecorationURLWith(size ImageSize) string {
	if m.AvatarDecorationData != nil {
		return AvatarDecorationURL(m.AvatarDecorationData.Asset, size)
	}
	return ""
}

var (
	_ fmt.Stringer = (*Member)(nil)
	_ fmt.Stringer = (*FullMember)(nil)
)
