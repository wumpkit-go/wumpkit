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
	"strconv"
	"time"
)

// UserFlags represents flags on a Discord user account.
//
// Reference: https://discord.com/developers/docs/resources/user#user-object-user-flags
type UserFlags int

const (
	// Discord Employee
	UserFlagStaff UserFlags = 1 << 0

	// Partnered Server Owner
	UserFlagPartner UserFlags = 1 << 1

	// HypeSquad Events Member
	UserFlagHypeSquad UserFlags = 1 << 2

	// Bug Hunter Level 1
	UserFlagBugHunterLevel1 UserFlags = 1 << 3

	// House Bravery Member
	UserFlagHypeSquadOnlineHouse1 UserFlags = 1 << 6

	// House Brilliance Member
	UserFlagHypeSquadOnlineHouse2 UserFlags = 1 << 7

	// House Balance Member
	UserFlagHypeSquadOnlineHouse3 UserFlags = 1 << 8

	// Early Nitro Supporter
	UserFlagPremiumEarlySupporter UserFlags = 1 << 9

	// User is a team
	UserFlagTeamPseudoUser UserFlags = 1 << 10

	// Bug Hunter Level 2
	UserFlagBugHunterLevel2 UserFlags = 1 << 14

	// Verified Bot
	UserFlagVerifiedBot UserFlags = 1 << 16

	// Early Verified Bot Developer
	UserFlagVerifiedDeveloper UserFlags = 1 << 17

	// Moderator Programs Alumni
	UserFlagCertifiedModerator UserFlags = 1 << 18

	// Bot uses only HTTP interactions and is shown in the online member list
	UserFlagBotHTTPInteractions UserFlags = 1 << 19

	// User is an Active Developer
	UserFlagActiveDeveloper UserFlags = 1 << 22
)

// Has returns true if all provided flags are set.
func (f UserFlags) Has(flags ...UserFlags) bool {
	return BitFieldHas(f, flags...)
}

// Nameplate represents the nameplate the user has.
//
// Reference: https://discord.com/developers/docs/resources/user#nameplate
type Nameplate struct {
	// SkuID is the Discord snowflake ID of the nameplate SKU.
	SkuID Snowflake `json:"sku_id"`

	// Asset is the path to the nameplate asset.
	Asset string `json:"asset"`

	// Label is the label of this nameplate.
	//
	// Optional:
	//  - May be empty string.
	Label string `json:"label"`

	// Palette is the background color of the nameplate.
	//
	// Values possible:
	// "crimson", "berry", "sky", "teal", "forest",
	// "bubble_gum", "violet", "cobalt", "clover", "lemon", "white"
	Palette string `json:"palette"`
}

// Collectibles represents collectibles the user owns,
// excluding avatar decorations and profile effects.
//
// Reference: https://discord.com/developers/docs/resources/user#collectibles
type Collectibles struct {
	// Nameplate is the user's nameplate collectible data.
	//
	// Optional:
	//  - May be nil if the user has no nameplate collectible.
	Nameplate *Nameplate `json:"nameplate,omitempty"`
}

// UserPrimaryGuild represents the user's primary guild info.
//
// Optionally included by Discord API.
//
// Reference: https://discord.com/developers/docs/resources/user#user-primary-guild-object
type UserPrimaryGuild struct {
	// IdentityGuildID is the Discord snowflake ID of the user's primary guild.
	//
	// Optional:
	//  - May be nil if the user has no primary guild set.
	//  - May be nil if the system cleared the identity due to guild tag support removal or privacy.
	IdentityGuildID *Snowflake `json:"identity_guild_id,omitempty"`

	// IdentityEnabled indicates if the user currently displays the primary guild's server tag.
	//
	// Optional:
	//  - May be nil if the identity was cleared by the system (e.g., guild tag disabled).
	//  - May be false if the user explicitly disabled showing the tag.
	IdentityEnabled *bool `json:"identity_enabled,omitempty"`

	// Tag is the text of the user's server tag.
	//
	// Optional:
	//  - May be nil or empty string if no tag is set.
	//  - Limited to 4 characters.
	//  - May be cleared if tag data is invalid or unavailable.
	Tag *string `json:"tag,omitempty"`

	// Badge is the hash string of the user's server tag badge.
	//
	// Optional:
	//  - May be nil if user has no badge or badge info unavailable.
	Badge *string `json:"badge,omitempty"`
}

// AvatarDecorationData represents avatar decoration info.
//
// Reference: https://discord.com/developers/docs/resources/user#avatar-decoration-object
type AvatarDecorationData struct {
	// Asset is the avatar decoration hash.
	Asset string `json:"asset"`

	// SkuID is the Discord snowflake ID of the avatar decoration SKU.
	SkuID Snowflake `json:"sku_id"`
}

// UserPremiumType is the type of premium (nitro) subscription a user has (see UserPremiumType* consts).
// https://discord.com/developers/docs/resources/user#user-object-premium-types
type UserPremiumType int

// Valid UserPremiumType values.
const (
	UserPremiumTypeNone         UserPremiumType = 0
	UserPremiumTypeNitroClassic UserPremiumType = 1
	UserPremiumTypeNitro        UserPremiumType = 2
	UserPremiumTypeNitroBasic   UserPremiumType = 3
)

// Is returns true if the user's premium type matches the provided premium type.
func (t UserPremiumType) Is(premiumType UserPremiumType) bool {
	return t == premiumType
}

// User represents a Discord user object.
//
// Reference: https://discord.com/developers/docs/resources/user#user-object-user-structure
type User struct {
	// ID is the user's unique Discord snowflake ID.
	ID Snowflake `json:"id"`

	// Username is the user's username (not unique).
	Username string `json:"username"`

	// Discriminator is the user's 4-digit Discord tag suffix.
	Discriminator string `json:"discriminator"`

	// GlobalName is the user's display name.
	// For bots, this is the application name.
	//
	// Optional:
	//  - May be empty string if unset.
	GlobalName string `json:"global_name"`

	// Avatar is the user's avatar hash.
	//
	// Optional:
	//  - May be empty string if no avatar.
	Avatar string `json:"avatar"`

	// Bot indicates if the user is a bot account.
	Bot bool `json:"bot,omitempty"`

	// System indicates if the user is an official Discord system user.
	System bool `json:"system,omitempty"`

	// Banner is the user's banner hash.
	//
	// Optional:
	//  - May be empty string if no banner.
	Banner string `json:"banner"`

	// AccentColor is the user's banner color encoded as an integer.
	//
	// Optional:
	//  - May be nil if no accent color is set.
	AccentColor *Color `json:"accent_color"`

	// PremiumType is the Nitro subscription type.
	PremiumType UserPremiumType `json:"premium_type,omitempty"`

	// PublicFlags are the public flags on the user account.
	PublicFlags UserFlags `json:"public_flags,omitempty"`

	// AvatarDecorationData holds avatar decoration info.
	//
	// Optional:
	//  - May be nil if user has no avatar decoration.
	AvatarDecorationData *AvatarDecorationData `json:"avatar_decoration_data,omitempty"`

	// Collectibles holds user's collectibles.
	//
	// Optional:
	//  - May be nil if user has no collectibles.
	Collectibles *Collectibles `json:"collectibles,omitempty"`

	// PrimaryGuild holds the user's primary guild info.
	//
	// Optional:
	// - May be nil if user is a bot.
	// - May be nil if no primary guild set.
	// - May be nil if identity cleared due to guild tag or privacy settings.
	PrimaryGuild *UserPrimaryGuild `json:"primary_guild,omitempty"`
}

type OAuth2User struct {
	User

	// Flags are internal user account flags.
	Flags UserFlags `json:"flags"`

	// Locale is the user's chosen language/locale.
	Locale Locale `json:"locale"`

	// MFAEnabled indicates if the user has two-factor authentication enabled.
	MFAEnabled bool `json:"mfa_enabled"`

	// Verified indicates if the user's email is verified.
	Verified bool `json:"verified"`

	// Email is the user's email address.
	Email string `json:"email"`
}

// Tag returns the user's tag in the format "username#discriminator".
//
// Example output: "bob#1337"
//
// Note: For users with no discriminator (new Discord accounts),
// this returns only the username (Example output: "bob").
func (u *User) Tag() string {
	if u.Discriminator != "0" {
		return u.Username + "#" + u.Discriminator
	}
	return u.Username
}

// Mention returns a Discord mention string for the user.
//
// Example output: "<@123456789012345678>"
func (u *User) Mention() string {
	return "<@" + u.ID.String() + ">"
}

// String implements the fmt.Stringer interface.
func (u *User) String() string {
	return u.Mention()
}

var _ fmt.Stringer = (*User)(nil)

// CreatedAt returns the time when this user account is created.
func (u *User) CreatedAt() time.Time {
	return u.ID.Timestamp()
}

// AvatarURL returns the URL to the user's avatar image.
//
// If the user has a custom avatar set, it returns the URL to that avatar, otherwise empty string.
// By default, it uses GIF format if the avatar is animated, otherwise PNG.
//
// If the user has no custom avatar, it returns the URL to their default avatar
// based on their discriminator or ID, using PNG format.
//
// Example usage:
//
//	url := user.AvatarURL()
func (u *User) AvatarURL() string {
	if u.Avatar != "" {
		return UserAvatarURL(u.ID, u.Avatar, ImageFormatDefault, ImageSizeDefault)
	}
	return DefaultUserAvatarURL(u.DefaultAvatarIndex())
}

// AvatarURLWith returns the URL to the user's avatar image,
// allowing explicit specification of image format and size.
//
// If the user has a custom avatar set, it returns the URL to that avatar
// using the provided format and size, otherwise empty string.
//
// If the user has no custom avatar, it returns the URL to their default avatar,
// using PNG format (size parameter is ignored for default avatars).
//
// Example usage:
//
//	url := user.AvatarURLWith(ImageFormatWebP, ImageSize1024)
func (u *User) AvatarURLWith(format ImageFormat, size ImageSize) string {
	if u.Avatar != "" {
		return UserAvatarURL(u.ID, u.Avatar, format, size)
	}
	return DefaultUserAvatarURL(u.DefaultAvatarIndex())
}

// BannerURL returns the URL to the user's banner image.
//
// If the user has a custom banner set, it returns the URL to that banner.
// By default, it uses GIF format if the banner is animated, otherwise PNG.
//
// If the user has no custom banner, it returns an empty string.
//
// Example usage:
//
//	url := user.BannerURL()
func (u *User) BannerURL() string {
	if u.Avatar != "" {
		return UserBannerURL(u.ID, u.Avatar, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// BannerURLWith returns the URL to the member's avatar image,
// allowing explicit specification of image format and size.
//
// If the user has no custom banner, it returns an empty string.
//
// Example usage:
//
//	url := user.BannerURLWith(ImageFormatWebP, ImageSize1024)
func (u *User) BannerURLWith(format ImageFormat, size ImageSize) string {
	if u.Avatar != "" {
		return UserBannerURL(u.ID, u.Avatar, format, size)
	}
	return ""
}

// DisplayName returns the user's global name if set,
// otherwise it returns their username.
//
// The global name is their profile display name visible across Discord,
// while Username is their original account username.
//
// Example usage:
//
//	name := user.DisplayName()
func (u *User) DisplayName() string {
	if u.GlobalName != "" {
		return u.GlobalName
	}
	return u.Username
}

// DefaultAvatarIndex returns the index (0-5) used to determine
// which default avatar is assigned to the user.
//
// For users with discriminator "0" (new Discord usernames),
// it uses the user's snowflake ID shifted right by 22 bits modulo 6.
//
// For legacy users with a numeric discriminator, it parses the discriminator
// as an integer and returns modulo 5.
//
// This logic follows Discord's default avatar assignment rules.
//
// Example usage:
//
//	index := user.DefaultAvatarIndex()
func (u *User) DefaultAvatarIndex() int {
	if u.Discriminator == "0" {
		return int((u.ID >> 22) % 6)
	}
	id, _ := strconv.Atoi(u.Discriminator)
	return id % 5
}

// AvatarDecorationURL returns the URL to the user's avatar decoration.
//
// If the user has no avatar decoration, it returns an empty string.
//
// Example usage:
//
//	url := user.AvatarDecorationURL()
func (u *User) AvatarDecorationURL() string {
	if u.AvatarDecorationData != nil {
		AvatarDecorationURL(u.AvatarDecorationData.Asset, ImageSizeDefault)
	}
	return ""
}

// AvatarDecorationURLWith returns the URL to the user's avatar decoration,
// allowing explicit specification of image size.
//
// If the user has no avatar decoration, it returns an empty string.
//
// Example usage:
//
//	url := user.AvatarDecorationURL(ImageSize512)
func (u *User) AvatarDecorationURLWith(size ImageSize) string {
	if u.AvatarDecorationData != nil {
		AvatarDecorationURL(u.AvatarDecorationData.Asset, size)
	}
	return ""
}

// GuildTagBadgeURL returns the URL to the user's PrimaryGuild badge image.
//
// If the user has no PrimaryGuild badge, it returns an empty string.
//
// Example usage:
//
//	url := user.GuildTagBadgeURL()
func (u *User) GuildTagBadgeURL() string {
	if u.PrimaryGuild != nil && u.PrimaryGuild.IdentityGuildID != nil &&
		u.PrimaryGuild.Badge != nil {
		return GuildTagBadgeURL(
			*u.PrimaryGuild.IdentityGuildID, *u.PrimaryGuild.Badge,
			ImageFormatDefault, ImageSizeDefault,
		)
	}
	return ""
}

// ConnectionVisibilityType represents the visibility of a user connection.
//
// Reference: https://discord.com/developers/docs/resources/user#connection-object-visibility-types
type ConnectionVisibilityType int

const (
	// Invisible to everyone except the user themselves.
	ConnectionVisibilityNone ConnectionVisibilityType = 0

	// Visible to everyone.
	ConnectionVisibilityEveryone ConnectionVisibilityType = 1
)

// Is returns true if this ConnectionVisibilityType matches the provided one.
func (v ConnectionVisibilityType) Is(other ConnectionVisibilityType) bool {
	return v == other
}

// Connection represents a connection to a third-party service.
//
// Reference: https://discord.com/developers/docs/resources/user#connection-object
type Connection struct {
	// ID is the id of the connection account.
	ID string `json:"id"`

	// Name is the username of the connection account.
	Name string `json:"name"`

	// Type is the service of this connection.
	Type string `json:"type"`

	// Revoked is whether the connection is revoked.
	Revoked *bool `json:"revoked,omitempty"`

	// Integrations is an array of partial server integrations.
	Integrations []Integration `json:"integrations,omitempty"`

	// Verified is whether the connection is verified.
	Verified bool `json:"verified"`

	// FriendSync is whether friend sync is enabled for this connection.
	FriendSync bool `json:"friend_sync"`

	// ShowActivity is whether activities related to this connection will be shown in presence updates.
	ShowActivity bool `json:"show_activity"`

	// TwoWayLink is whether this connection has a corresponding third party OAuth2 token.
	TwoWayLink bool `json:"two_way_link"`

	// Visibility is the visibility of this connection.
	Visibility ConnectionVisibilityType `json:"visibility"`
}

// ApplicationRoleConnection represents the role connection object that an application has attached to a user.
//
// Reference: https://discord.com/developers/docs/resources/user#application-role-connection-object
type ApplicationRoleConnection struct {
	// PlatformName is the vanity name of the platform a bot has connected (max 50 characters).
	PlatformName *string `json:"platform_name,omitempty"`

	// PlatformUsername is the username on the platform a bot has connected (max 100 characters).
	PlatformUsername *string `json:"platform_username,omitempty"`

	// Metadata is object mapping application role connection metadata keys to their string-ified values (max 100 characters).
	Metadata map[string]string `json:"metadata,omitempty"`
}
