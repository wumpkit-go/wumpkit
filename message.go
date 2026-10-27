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
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"
)

// Embed field limits as defined by Discord's API.
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-limits
const (
	// EmbedTitleMaxLength is the maximum length for an embed title.
	EmbedTitleMaxLength = 256

	// EmbedDescriptionMaxLength is the maximum length for an embed description.
	EmbedDescriptionMaxLength = 4096

	// EmbedFieldsMaxCount is the maximum number of fields in an embed.
	EmbedFieldsMaxCount = 25

	// EmbedFieldNameMaxLength is the maximum length for an embed field name.
	EmbedFieldNameMaxLength = 256

	// EmbedFieldValueMaxLength is the maximum length for an embed field value.
	EmbedFieldValueMaxLength = 1024

	// EmbedFooterTextMaxLength is the maximum length for embed footer text.
	EmbedFooterTextMaxLength = 2048

	// EmbedAuthorNameMaxLength is the maximum length for an embed author name.
	EmbedAuthorNameMaxLength = 256

	// EmbedTotalCharacterLimit is the combined character limit across all embed fields.
	// This includes title, description, field names, field values, footer text, and author name.
	EmbedTotalCharacterLimit = 6000
)

// MessageType represents the type of a message on Discord.
//
// Reference: https://discord.com/developers/docs/resources/message#message-object-message-types
type MessageType int

const (
	// MessageTypeDefault is a standard user-sent message.
	// Deletable: true
	MessageTypeDefault MessageType = iota
	// MessageTypeRecipientAdd is a system message indicating a user was added to a group DM.
	// Deletable: false
	MessageTypeRecipientAdd
	// MessageTypeRecipientRemove is a system message indicating a user was removed from a group DM.
	// Deletable: false
	MessageTypeRecipientRemove
	// MessageTypeCall is a system message indicating a call was started.
	// Deletable: false
	MessageTypeCall
	// MessageTypeChannelNameChange is a system message indicating a channel's name was changed.
	// Deletable: false
	MessageTypeChannelNameChange
	// MessageTypeChannelIconChange is a system message indicating a channel's icon was changed.
	// Deletable: false
	MessageTypeChannelIconChange
	// MessageTypeChannelPinnedMessage is a system message indicating a message was pinned.
	// Deletable: true
	MessageTypeChannelPinnedMessage
	// MessageTypeUserJoin is a system message indicating a user joined a guild.
	// Deletable: true
	MessageTypeUserJoin
	// MessageTypeGuildBoost is a system message indicating a guild received a boost.
	// Deletable: true
	MessageTypeGuildBoost
	// MessageTypeGuildBoostTier1 is a system message indicating a guild reached boost tier 1.
	// Deletable: true
	MessageTypeGuildBoostTier1
	// MessageTypeGuildBoostTier2 is a system message indicating a guild reached boost tier 2.
	// Deletable: true
	MessageTypeGuildBoostTier2
	// MessageTypeGuildBoostTier3 is a system message indicating a guild reached boost tier 3.
	// Deletable: true
	MessageTypeGuildBoostTier3
	// MessageTypeChannelFollowAdd is a system message indicating a channel was followed.
	// Deletable: true
	MessageTypeChannelFollowAdd
	_
	// MessageTypeGuildDiscoveryDisqualified is a system message indicating a guild was disqualified from Discovery.
	// Deletable: true
	MessageTypeGuildDiscoveryDisqualified
	// MessageTypeGuildDiscoveryRequalified is a system message indicating a guild requalified for Discovery.
	// Deletable: true
	MessageTypeGuildDiscoveryRequalified
	// MessageTypeGuildDiscoveryGracePeriodInitialWarning is a system message for the initial Discovery grace period warning.
	// Deletable: true
	MessageTypeGuildDiscoveryGracePeriodInitialWarning
	// MessageTypeGuildDiscoveryGracePeriodFinalWarning is a system message for the final Discovery grace period warning.
	// Deletable: true
	MessageTypeGuildDiscoveryGracePeriodFinalWarning
	// MessageTypeThreadCreated is a system message indicating a thread was created.
	// Deletable: true
	MessageTypeThreadCreated
	// MessageTypeReply is a message sent as a reply to another message.
	// Deletable: true
	MessageTypeReply
	// MessageTypeSlashCommand is a message resulting from a slash command interaction.
	// Deletable: true
	MessageTypeSlashCommand
	// MessageTypeThreadStarterMessage is the initial message of a thread.
	// Deletable: false
	MessageTypeThreadStarterMessage
	// MessageTypeGuildInviteReminder is a system message reminding users to invite friends to the guild.
	// Deletable: true
	MessageTypeGuildInviteReminder
	// MessageTypeContextMenuCommand is a message resulting from a context menu command.
	// Deletable: true
	MessageTypeContextMenuCommand
	// MessageTypeAutoModerationAction is a system message generated by an AutoMod action.
	// Deletable: true
	MessageTypeAutoModerationAction
	// MessageTypeRoleSubscriptionPurchase is a system message for a role subscription purchase.
	// Deletable: true
	MessageTypeRoleSubscriptionPurchase
	// MessageTypeInteractionPremiumUpsell is a system message encouraging a premium subscription upsell.
	// Deletable: true
	MessageTypeInteractionPremiumUpsell
	// MessageTypeStageStart is a system message indicating a stage instance started.
	// Deletable: true
	MessageTypeStageStart
	// MessageTypeStageEnd is a system message indicating a stage instance ended.
	// Deletable: true
	MessageTypeStageEnd
	// MessageTypeStageSpeaker is a system message indicating a stage speaker was added.
	// Deletable: true
	MessageTypeStageSpeaker
	_
	// MessageTypeStageTopic is a system message indicating a stage topic was changed.
	// Deletable: true
	MessageTypeStageTopic
	// MessageTypeGuildApplicationPremiumSubscription is a system message for a guild premium subscription via an app.
	// Deletable: true
	MessageTypeGuildApplicationPremiumSubscription
	_
	_
	_
	// MessageTypeGuildIncidentAlertModeEnabled is a system message indicating incident alert mode was enabled.
	// Deletable: true
	MessageTypeGuildIncidentAlertModeEnabled
	// MessageTypeGuildIncidentAlertModeDisabled is a system message indicating incident alert mode was disabled.
	// Deletable: true
	MessageTypeGuildIncidentAlertModeDisabled
	// MessageTypeGuildIncidentReportRaid is a system message indicating a raid was reported.
	// Deletable: true
	MessageTypeGuildIncidentReportRaid
	// MessageTypeGuildIncidentReportFalseAlarm is a system message indicating a raid report was marked as false.
	// Deletable: true
	MessageTypeGuildIncidentReportFalseAlarm
	_
	_
	_
	_
	// MessageTypePurchaseNotification is a system message for a purchase notification.
	// Deletable: true
	MessageTypePurchaseNotification
	_
	// MessageTypePollResult is a system message indicating the results of a poll.
	// Deletable: true
	MessageTypePollResult
)

// Is checks if the message type matches the provided type.
func (t MessageType) Is(messageType MessageType) bool {
	return t == messageType
}

// System indicates whether the message is a system-generated message.
func (t MessageType) System() bool {
	switch t {
	case MessageTypeDefault, MessageTypeReply, MessageTypeSlashCommand,
		MessageTypeThreadStarterMessage, MessageTypeContextMenuCommand:
		return false
	default:
		return true
	}
}

// Deletable indicates whether the message can be deleted.
func (t MessageType) Deletable() bool {
	switch t {
	case MessageTypeRecipientAdd, MessageTypeRecipientRemove, MessageTypeCall,
		MessageTypeChannelNameChange, MessageTypeChannelIconChange,
		MessageTypeThreadStarterMessage:
		return false
	default:
		return true
	}
}

// MessageReaction represents a reaction to a message on Discord.
//
// Reference: https://discord.com/developers/docs/resources/message#reaction-object
type MessageReaction struct {
	// Count is the total number of times this emoji has been used to react, including super reactions.
	Count int `json:"count"`
	// CountDetails provides detailed counts of normal and super reactions.
	CountDetails ReactionCountDetails `json:"count_details"`
	// Me indicates whether the current user reacted with this emoji.
	Me bool `json:"me"`
	// MeBurst indicates whether the current user used a super reaction with this emoji.
	MeBurst bool `json:"me_burst"`
	// Emoji is the emoji used for the reaction.
	Emoji Emoji `json:"emoji"`
	// BurstColors contains the HEX colors used for super reactions.
	//
	// Optional:
	//   - Will be empty if no super reactions are present.
	BurstColors []string `json:"burst_colors"`
}

// ReactionCountDetails provides counts for normal and super reactions to a message.
//
// Reference: https://discord.com/developers/docs/resources/message#reaction-count-details-object
type ReactionCountDetails struct {
	// Burst is the count of super reactions.
	Burst int `json:"burst"`
	// Normal is the count of normal reactions.
	Normal int `json:"normal"`
}

// MentionChannel represents a channel mentioned in a message.
//
// Reference: https://discord.com/developers/docs/resources/message#channel-mention-object
type MentionChannel struct {
	// ID is the unique Discord snowflake ID of the channel.
	ID Snowflake `json:"id"`
	// GuildID is the ID of the guild containing the channel.
	//
	// Optional:
	//   - Will be 0 if the channel is not in a guild (e.g., a DM channel).
	GuildID Snowflake `json:"guild_id"`
	// Type is the type of the channel.
	Type ChannelType `json:"type"`
	// Name is the name of the channel.
	Name string `json:"name"`
}

// Mention returns the Discord mention string for the channel in the format "<#channel_id>".
func (c *MentionChannel) Mention() string {
	return "<#" + c.ID.String() + ">"
}

// String implements the fmt.Stringer interface.
func (c *MentionChannel) String() string {
	return c.Mention()
}

// StickerItem represents a sticker included in a message.
//
// Reference: https://discord.com/developers/docs/resources/sticker#sticker-item-object
type StickerItem struct {
	// ID is the unique Discord snowflake ID of the sticker.
	ID Snowflake `json:"id"`
	// Name is the name of the sticker.
	Name string `json:"name"`
	// FormatType is the format type of the sticker (e.g., PNG, APNG, Lottie).
	FormatType StickerFormatType `json:"format_type"`
}

// MessageActivityType represents the type of activity associated with a message.
//
// Reference: https://discord.com/developers/docs/resources/message#message-object-message-activity-types
type MessageActivityType int

const (
	// MessageActivityTypeJoin indicates a join activity (e.g., for Rich Presence).
	MessageActivityTypeJoin MessageActivityType = iota + 1
	// MessageActivityTypeSpectate indicates a spectate activity.
	MessageActivityTypeSpectate
	// MessageActivityTypeListen indicates a listen activity.
	MessageActivityTypeListen
	_
	// MessageActivityTypeJoinRequest indicates a join request activity.
	MessageActivityTypeJoinRequest
)

// Is checks if the message activity type matches the provided type.
func (t MessageActivityType) Is(messageActivityType MessageActivityType) bool {
	return t == messageActivityType
}

// MessageActivity represents an activity associated with a message, used for Rich Presence.
//
// Reference: https://discord.com/developers/docs/resources/message#message-object-message-activity-structure
type MessageActivity struct {
	// Type is the type of the activity (e.g., join, spectate).
	Type MessageActivityType `json:"type"`
	// PartyID is the party ID from a Rich Presence event.
	//
	// Optional:
	//   - Will be empty if no party is associated with the activity.
	PartyID string `json:"party_id"`
}

// MessageFlags represents a bitfield of flags for a message.
//
// Reference: https://discord.com/developers/docs/resources/message#message-object-message-flags
type MessageFlags int

const (
	// MessageFlagCrossposted indicates the message was published to subscribed channels via Channel Following.
	MessageFlagCrossposted MessageFlags = 1 << iota
	// MessageFlagIsCrosspost indicates the message originated from another channel via Channel Following.
	MessageFlagIsCrosspost
	// MessageFlagSuppressEmbeds indicates embeds should not be included when serializing the message.
	MessageFlagSuppressEmbeds
	// MessageFlagSourceMessageDeleted indicates the source message for a crosspost was deleted.
	MessageFlagSourceMessageDeleted
	// MessageFlagUrgent indicates the message came from the urgent message system.
	MessageFlagUrgent
	// MessageFlagHasThread indicates the message has an associated thread with the same ID.
	MessageFlagHasThread
	// MessageFlagEphemeral indicates the message is only visible to the user who invoked the interaction.
	MessageFlagEphemeral
	// MessageFlagLoading indicates the message is an interaction response and the bot is "thinking".
	MessageFlagLoading
	// MessageFlagFailedToMentionSomeRolesInThread indicates the message failed to mention some roles in a thread.
	MessageFlagFailedToMentionSomeRolesInThread
	_
	_
	_
	// MessageFlagSuppressNotifications indicates the message will not trigger push or desktop notifications.
	MessageFlagSuppressNotifications
	// MessageFlagIsVoiceMessage indicates the message is a voice message.
	MessageFlagIsVoiceMessage
	// MessageFlagHasSnapshot indicates the message has a snapshot via Message Forwarding.
	MessageFlagHasSnapshot
	// MessageFlagIsComponentsV2 indicates the message supports fully component-driven interactions.
	MessageFlagIsComponentsV2
	_
	// MessageFlagsNone represents no flags set.
	MessageFlagsNone MessageFlags = 0
)

// Has checks if all provided flags are set in the bitfield.
func (f MessageFlags) Has(flags ...MessageFlags) bool {
	return BitFieldHas(f, flags...)
}

// Add sets all provided flags in the bitfield.
func (f *MessageFlags) Add(flags ...MessageFlags) {
	*f = BitFieldAdd(*f, flags...)
}

// Remove clears all provided flags from the bitfield.
func (f *MessageFlags) Remove(flags ...MessageFlags) {
	*f = BitFieldRemove(*f, flags...)
}

// MessageReferenceType determines how associated data is populated for a message reference.
//
// Reference: https://discord.com/developers/docs/resources/message#message-reference-types
type MessageReferenceType int

const (
	// MessageReferenceTypeDefault is a standard message reference (e.g., for replies).
	MessageReferenceTypeDefault MessageReferenceType = iota
	// MessageReferenceTypeForward is a reference for forwarded messages.
	MessageReferenceTypeForward
)

// Is returns true if this MessageReferenceType matches the provided one.
func (v MessageReferenceType) Is(other MessageReferenceType) bool {
	return v == other
}

// MessageReference represents a reference to another message, such as for replies or crossposts.
//
// Reference: https://discord.com/developers/docs/resources/message#message-reference-structure
type MessageReference struct {
	// Type is the type of reference (e.g., default, forward).
	//
	// Optional:
	//   - Will be omitted if not specified when sending a message.
	Type MessageReferenceType `json:"type,omitempty"`

	// MessageID is the ID of the referenced message.
	MessageID Snowflake `json:"message_id"`

	// ChannelID is the ID of the channel containing the referenced message.
	//
	// Optional:
	//   - Will be omitted if not specified when sending a message.
	ChannelID Snowflake `json:"channel_id,omitempty"`

	// GuildID is the ID of the guild containing the referenced message.
	//
	// Optional:
	//   - Will be 0 if the referenced message is not in a guild.
	GuildID Snowflake `json:"guild_id,omitempty"`

	// FailIfNotExists determines whether to error if the referenced message does not exist.
	//
	// Optional:
	//   - Defaults to true when sending a message.
	FailIfNotExists bool `json:"fail_if_not_exists,omitempty"`
}

// PartialMessage represents a lightweight version of a Discord message.
//
// Reference: https://discord.com/developers/docs/resources/message#message-reference-structure
type PartialMessage struct {
	// Type is the type of the message (e.g., default, reply).
	Type MessageType `json:"type"`

	// Content is the text content of the message.
	Content string `json:"content"`

	// Embeds is an array of embedded content in the message.
	//
	// Optional:
	//   - Will be empty if no embeds are included.
	Embeds []Embed `json:"embeds"`

	// Attachments is an array of file attachments in the message.
	//
	// Optional:
	//   - Will be empty if no attachments are included.
	Attachments []Attachment `json:"attachments"`

	// Timestamp is when the message was sent.
	Timestamp time.Time `json:"timestamp,omitzero"`

	// EditedTimestamp is when the message was last edited.
	//
	// Optional:
	//   - Will be nil if the message has never been edited.
	EditedTimestamp *time.Time `json:"edited_timestamp,omitzero"`

	// Flags is a bitfield of message flags (e.g., crossposted, ephemeral).
	Flags MessageFlags `json:"flags"`

	// Mentions is an array of users mentioned in the message.
	//
	// Optional:
	//   - Will be empty if no users are mentioned.
	Mentions []User `json:"mentions"`

	// MentionRoles is an array of roles mentioned in the message.
	//
	// Optional:
	//   - Will be empty if no roles are mentioned.
	MentionRoles []Snowflake `json:"mention_roles"`

	// Stickers is an array of stickers included in the message.
	//
	// Optional:
	//   - Will be empty if no stickers are included.
	Stickers []Sticker `json:"stickers"`
}

// MessageSnapshot represents a snapshot of a message, used for message forwarding.
//
// Reference: https://discord.com/developers/docs/resources/message#message-snapshot-object
type MessageSnapshot struct {
	// Message is the partial message data for the snapshot.
	Message PartialMessage `json:"message"`
}

// InteractionMetadata represents metadata about an interaction that generated a message.
//
// Reference: https://discord.com/developers/docs/resources/message#message-interaction-metadata-object
type InteractionMetadata struct {
	// ID is the unique Discord snowflake ID of the interaction.
	ID Snowflake `json:"id"`

	// Type is the type of the interaction (e.g., command, component).
	Type InteractionType `json:"type"`

	// User is the user who triggered the interaction.
	User User `json:"user"`

	// AuthorizingIntegrationOwners maps integration types to their owner IDs.
	AuthorizingIntegrationOwners map[ApplicationIntegrationType]Snowflake `json:"authorizing_integration_owners"`

	// OriginalResponseMessageID is the ID of the original response message for the interaction.
	//
	// Optional:
	//   - Will be 0 if no original response message exists.
	OriginalResponseMessageID Snowflake `json:"original_response_message_id"`

	// TargetUser is the user targeted by a user command interaction.
	//
	// Optional:
	//   - Will be nil unless the interaction is of type ApplicationCommandTypeUser.
	TargetUser *User `json:"target_user"`

	// TargetMessageID is the ID of the message targeted by a message command interaction.
	//
	// Optional:
	//   - Will be 0 unless the interaction is of type ApplicationCommandTypeMessage.
	TargetMessageID Snowflake `json:"target_message_id"`

	// InteractedMessageID is the ID of the message interacted with in a component interaction.
	//
	// Optional:
	//   - Will be 0 unless the interaction is of type InteractionTypeComponent.
	InteractedMessageID Snowflake `json:"interacted_message_id"`

	// TriggeringInteractionMetadata is the metadata of the interaction that triggered a modal submit.
	//
	// Optional:
	//   - Will be nil unless the interaction is of type InteractionTypeModalSubmit.
	TriggeringInteractionMetadata *InteractionMetadata `json:"triggering_interaction_metadata"`
}

// Nonce represents a string or integer used to validate message sending.
type Nonce string

var _ json.Unmarshaler = (*Nonce)(nil)

func (n *Nonce) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		return nil
	}
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		i, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			return err
		}
		*n = Nonce(strconv.FormatInt(i, 10))
	} else {
		*n = Nonce(unquoted)
	}
	return nil
}

// MessageCall represents a call associated with a message.
//
// Reference: https://discord.com/developers/docs/resources/message#message-call-object
type MessageCall struct {
	// Participants is an array of user IDs participating in the call.
	//
	// Optional:
	//   - Will be empty if no participants are present.
	Participants []Snowflake `json:"participants"`

	// EndedTimestamp is when the call ended.
	//
	// Optional:
	//   - Will be nil if the call is ongoing or not applicable.
	EndedTimestamp *time.Time `json:"ended_timestamp,omitzero"`
}

// MessageApplication represents a partial application object for Rich Presence or webhooks.
//
// Reference: https://discord.com/developers/docs/resources/application#application-object
type MessageApplication struct {
	// ID is the unique Discord snowflake ID of the application.
	ID Snowflake `json:"id"`

	// CoverImage is the cover image hash of the application.
	//
	// Optional:
	//   - Will be empty if no cover image is set.
	CoverImage string `json:"cover_image"`

	// Description is the description of the application.
	Description string `json:"description"`

	// Icon is the icon hash of the application.
	//
	// Optional:
	//   - Will be empty if no icon is set.
	Icon string `json:"icon"`

	// Name is the name of the application.
	Name string `json:"name"`
}

// RoleSubscriptionData represents data for a role subscription purchase.
type RoleSubscriptionData struct {
	// RoleSubscriptionListingID is the ID of the role subscription listing.
	RoleSubscriptionListingID Snowflake `json:"role_subscription_listing_id"`

	// TierName is the name of the subscription tier.
	TierName string `json:"tier_name"`

	// TotalMonthsSubscribed is the total number of months subscribed.
	TotalMonthsSubscribed int `json:"total_months_subscribed"`

	// IsRenewal indicates whether the subscription is a renewal.
	IsRenewal bool `json:"is_renewal"`
}

// EmbedType represents the type of an embed.
type EmbedType string

const (
	EmbedTypeRich       EmbedType = "rich"
	EmbedTypeImage      EmbedType = "image"
	EmbedTypeVideo      EmbedType = "video"
	EmbedTypeGifv       EmbedType = "gifv"
	EmbedTypeArticle    EmbedType = "article"
	EmbedTypeLink       EmbedType = "link"
	EmbedTypePollResult EmbedType = "poll_result"
)

// Is returns true if this EmbedType matches the provided one.
func (v EmbedType) Is(other EmbedType) bool {
	return v == other
}

// Embed represents a Discord embed object.
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-object
//
// Limits:
//   - The combined sum of characters in all title, description, field.name, field.value,
//     footer.text, and author.name fields across all embeds in a message must not exceed 6000.
type Embed struct {
	// Title is the title of the embed.
	//
	// Optional, max 256 characters, empty string if not set.
	Title string `json:"title,omitempty"`

	// Type is the type of the embed.
	//
	// Optional, always "rich" for webhook embeds.
	Type EmbedType `json:"type,omitempty"`

	// Description is the description text of the embed.
	//
	// Optional, max 4096 characters, empty string if not set.
	Description string `json:"description,omitempty"`

	// URL is the URL of the embed.
	//
	// Optional, empty string if not set.
	URL string `json:"url,omitempty"`

	// Timestamp is the timestamp of the embed content in ISO8601 format.
	//
	// Optional, zero value if not set.
	Timestamp *time.Time `json:"timestamp,omitzero"`

	// Color is the color code of the embed (decimal integer).
	//
	// Optional, 0 if not set.
	Color Color `json:"color,omitempty"`

	// Footer contains footer information.
	//
	// Optional, may be nil if not set.
	Footer *EmbedFooter `json:"footer,omitempty"`

	// Image contains image information.
	//
	// Optional, may be nil if not set.
	Image *EmbedImage `json:"image,omitempty"`

	// Thumbnail contains thumbnail information.
	//
	// Optional, may be nil if not set.
	Thumbnail *EmbedThumbnail `json:"thumbnail,omitempty"`

	// Video contains video information.
	//
	// Optional, may be nil if not set.
	Video *EmbedVideo `json:"video,omitempty"`

	// Provider contains provider information.
	//
	// Optional, may be nil if not set.
	Provider *EmbedProvider `json:"provider,omitempty"`

	// Author contains author information.
	//
	// Optional, may be nil if not set.
	// author.name max 256 characters
	Author *EmbedAuthor `json:"author,omitempty"`

	// Fields contains an array of embed fields.
	//
	// Optional, max 25 fields.
	// field.name max 256 characters, field.value max 1024 characters
	Fields []EmbedField `json:"fields,omitempty"`
}

// Builder returns a new EmbedBuilder initialized with a copy of the current embed.
func (e *Embed) Builder() EmbedBuilder {
	return EmbedBuilder{embed: *e}
}

// EmbedFooter represents the footer object of an embed.
//
// Limits:
// - text max 2048 characters
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-object-embed-footer-structure
type EmbedFooter struct {
	// Text is the footer text.
	//
	// Always present, max 2048 characters.
	Text string `json:"text"`

	// IconURL is the URL of the footer icon.
	//
	// Optional, may be empty string if not set.
	IconURL string `json:"icon_url,omitempty"`

	// ProxyIconURL is a proxied URL of the footer icon.
	//
	// Optional, may be empty string if not set.
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedImage represents the image object of an embed.
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-object-embed-image-structure
type EmbedImage struct {
	// URL is the source URL of the image.
	//
	// Always present. Supports only http(s) and attachments.
	URL string `json:"url"`

	// ProxyURL is a proxied URL of the image.
	//
	// Optional, may be empty string if not set.
	ProxyURL string `json:"proxy_url,omitempty"`

	// Height is the height of the image.
	//
	// Optional, 0 if not set.
	Height int `json:"height,omitempty"`

	// Width is the width of the image.
	//
	// Optional, 0 if not set.
	Width int `json:"width,omitempty"`
}

// EmbedThumbnail represents the thumbnail object of an embed.
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-object-embed-thumbnail-structure
type EmbedThumbnail struct {
	// URL is the source URL of the thumbnail.
	//
	// Always present. Supports only http(s) and attachments.
	URL string `json:"url"`

	// ProxyURL is a proxied URL of the thumbnail.
	//
	// Optional, may be empty string if not set.
	ProxyURL string `json:"proxy_url,omitempty"`

	// Height is the height of the thumbnail.
	//
	// Optional, 0 if not set.
	Height int `json:"height,omitempty"`

	// Width is the width of the thumbnail.
	//
	// Optional, 0 if not set.
	Width int `json:"width,omitempty"`
}

// EmbedVideo represents the video object of an embed.
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-object-embed-video-structure
type EmbedVideo struct {
	// URL is the source URL of the video.
	//
	// Optional, may be empty string if not set.
	URL string `json:"url,omitempty"`

	// ProxyURL is a proxied URL of the video.
	//
	// Optional, may be empty string if not set.
	ProxyURL string `json:"proxy_url,omitempty"`

	// Height is the height of the video.
	//
	// Optional, 0 if not set.
	Height int `json:"height,omitempty"`

	// Width is the width of the video.
	//
	// Optional, 0 if not set.
	Width int `json:"width,omitempty"`
}

// EmbedProvider represents the provider object of an embed.
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-object-embed-provider-structure
type EmbedProvider struct {
	// Name is the name of the provider.
	//
	// Optional, may be empty string if not set.
	Name string `json:"name,omitempty"`

	// URL is the URL of the provider.
	//
	// Optional, may be empty string if not set.
	URL string `json:"url,omitempty"`
}

// EmbedAuthor represents the author object of an embed.
//
// Limits:
// - name max 256 characters
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-object-embed-author-structure
type EmbedAuthor struct {
	// Name is the name of the author.
	//
	// Always present, max 256 characters.
	Name string `json:"name"`

	// URL is the URL of the author.
	//
	// Optional, may be empty string if not set.
	URL string `json:"url,omitempty"`

	// IconURL is the URL of the author icon.
	//
	// Optional, may be empty string if not set.
	IconURL string `json:"icon_url,omitempty"`

	// ProxyIconURL is a proxied URL of the author icon.
	//
	// Optional, may be empty string if not set.
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedField represents a field object in an embed.
//
// Limits:
// - name max 256 characters
// - value max 1024 characters
//
// Reference: https://discord.com/developers/docs/resources/channel#embed-object-embed-field-structure
type EmbedField struct {
	// Name is the name of the field.
	//
	// Always present, max 256 characters.
	Name string `json:"name"`

	// Value is the value of the field.
	//
	// Always present, max 1024 characters.
	Value string `json:"value"`

	// Inline indicates whether this field should display inline.
	//
	// Optional, false if not set.
	Inline bool `json:"inline,omitempty"`
}

// Message represents a message sent in a channel within Discord.
//
// Reference: https://discord.com/developers/docs/resources/message#message-object
type Message struct {
	// ID is the unique Discord snowflake ID of the message.
	ID Snowflake `json:"id"`

	// ChannelID is the ID of the channel where the message was sent.
	ChannelID Snowflake `json:"channel_id"`

	// Author is the user who authored the message.
	//
	// Note:
	//   - For webhook messages, this field reflects the webhook's ID, username, and avatar.
	Author User `json:"author"`

	// Content is the text content of the message.
	//
	// Note:
	//   - Will be empty if the bot lacks the GatewayIntentMessageContent privileged intent (1 << 15).
	Content string `json:"content"`

	// Timestamp is when the message was sent.
	Timestamp time.Time `json:"timestamp,omitzero"`

	// EditedTimestamp is when the message was last edited.
	//
	// Optional:
	//   - Will be nil if the message has never been edited.
	EditedTimestamp *time.Time `json:"edited_timestamp,omitempty"`

	// TTS indicates whether the message is a text-to-speech message.
	TTS bool `json:"tts"`

	// MentionEveryone indicates whether the message mentions @everyone or @here.
	MentionEveryone bool `json:"mention_everyone"`

	// Mentions is an array of users specifically mentioned in the message.
	Mentions []User `json:"mentions"`

	// MentionRoles is an array of roles specifically mentioned in the message.
	MentionRoles []Role `json:"mention_roles"`

	// MentionChannels is an array of channels specifically mentioned in the message.
	//
	// Note:
	//   - Not all channel mentions in this message will appear in this slice. Only textual
	//     channels that are visible to everyone in a lurkable guild will ever be included. Only
	//     crossposted messages (via Channel Following) currently include MentionChannels at all.
	//     If no mentions in the message meet these requirements, this field will be empty.
	MentionChannels []MentionChannel `json:"mention_channels"`

	// Attachments is an array of file attachments included with the message.
	//
	// Optional:
	//   - Will be empty if no attachments are included or if the bot lacks the GatewayIntentMessageContent privileged intent (1 << 15).
	Attachments []Attachment `json:"attachments"`

	// Embeds is an array of embedded content included with the message.
	//
	// Optional:
	//   - Will be empty if no embeds are included or if the bot lacks the GatewayIntentMessageContent privileged intent (1 << 15).
	Embeds []Embed `json:"embeds"`

	// Reactions is an array of reactions to the message.
	Reactions []MessageReaction `json:"reactions"`

	// Nonce is a value used to validate that a message was sent successfully.
	Nonce Nonce `json:"nonce"`

	// Pinned indicates whether the message is pinned in the channel.
	Pinned bool `json:"pinned"`

	// WebhookID is the ID of the webhook that generated the message.
	//
	// Optional:
	//   - Will be 0 if the message is not sent by a webhook.
	WebhookID Snowflake `json:"webhook_id"`

	// Type is the type of the message.
	Type MessageType `json:"type"`

	// Activity is the activity associated with the message, used for Rich Presence embeds.
	//
	// Optional:
	//   - Will be nil if the message is not associated with a Rich Presence activity.
	Activity *MessageActivity `json:"activity"`

	// Application is a partial application object sent with Rich Presence-related chat embeds.
	//
	// Optional:
	//   - Will be nil if the message is not associated with a Rich Presence-related chat embed.
	Application *MessageApplication `json:"application"`

	// ApplicationID is the unique Discord snowflake ID of the application.
	//
	// Optional:
	//   - Will be 0 if the message is not an interaction or application-owned webhook.
	ApplicationID Snowflake `json:"application_id"`

	// Flags is a bitfield of message flags.
	Flags MessageFlags `json:"flags"`

	// MessageReference contains data showing the source of a crosspost, channel follow add, pin, or reply message.
	//
	// Optional:
	//   - Will be nil if the message is not a crosspost, channel follow add, pin, or reply.
	MessageReference *MessageReference `json:"message_reference"`

	// MessageSnapshots is an array of message snapshot objects associated with the message_reference,
	// containing a minimal subset of message fields (e.g., excludes Author).
	//
	// Optional:
	//   - Will be empty if the message is not associated with message forwarding.
	MessageSnapshots []MessageSnapshot `json:"message_snapshots"`

	// ReferencedMessage is the message associated with the message_reference.
	//
	// Optional:
	//   - Will be nil if the message is not of type MessageTypeReply, MessageTypeThreadStarterMessage,
	//     or MessageTypeContextMenuCommand, or if the referenced message was deleted.
	//
	// Note:
	//   - Only included for messages of type MessageTypeReply, MessageTypeThreadStarterMessage, or MessageTypeContextMenuCommand.
	ReferencedMessage *Message `json:"referenced_message"`

	// InteractionMetadata contains metadata about the interaction that generated the message.
	//
	// Optional:
	//   - Will be nil if the message is not generated by an interaction.
	InteractionMetadata *InteractionMetadata `json:"interaction_metadata"`

	// Thread is the thread started from this message, including thread member data.
	//
	// Optional:
	//   - Will be nil if the message did not start a thread.
	Thread *ResolvedThread `json:"thread"`

	// Components is an array of interactive components (e.g., buttons, select menus) included with the message.
	//
	// Optional:
	//   - Will be empty if no components are included or if the bot lacks the GatewayIntentMessageContent privileged intent (1 << 15).
	Components []LayoutComponent `json:"components"`

	// StickerItems is an array of sticker items included with the message.
	StickerItems []StickerItem `json:"sticker_items"`

	// Position is the approximate position of the message in a thread.
	//
	// Optional:
	//   - Will be nil if the message is not in a thread.
	Position *int `json:"position,omitempty"`

	// GuildID is the ID of the guild where the message was sent.
	//
	// Optional:
	//   - Will be 0 if the message is not in a guild (e.g., a DM).
	GuildID Snowflake `json:"guild_id"`

	// Poll is the poll object included with the message.
	//
	// Optional:
	//   - Will be nil if no poll is included or if the bot lacks the GatewayIntentMessageContent privileged intent (1 << 15).
	Poll *Poll `json:"poll"`

	// Call is the call associated with the message (e.g., voice or video call).
	//
	// Optional:
	//   - Will be nil if the message is not associated with a call.
	Call *MessageCall `json:"call"`
}

var _ json.Unmarshaler = (*Message)(nil)

// UnmarshalJSON unmarshals a JSON message into a Message struct, handling component deserialization.
func (m *Message) UnmarshalJSON(buf []byte) error {
	type tempMessage Message
	var temp struct {
		Components []json.RawMessage `json:"components"`
		tempMessage
	}
	if err := json.Unmarshal(buf, &temp); err != nil {
		return err
	}
	*m = Message(temp.tempMessage)
	if temp.Components != nil {
		m.Components = make([]LayoutComponent, 0, len(temp.Components))
		for i := range len(temp.Components) {
			if len(temp.Components[i]) == 0 || bytes.Equal(temp.Components[i], []byte("null")) {
				continue
			}
			component, err := UnmarshalComponent(temp.Components[i])
			if err != nil {
				return err
			}
			if lComponent, ok := component.(LayoutComponent); ok {
				m.Components = append(m.Components, lComponent)
			}
		}
	}
	return nil
}

// URL returns the Discord URL for the message in the format "https://discord.com/channels/{guild_id}/{channel_id}/{message_id}".
// If GuildID is 0 (e.g., for DMs), "@me" is used instead of a guild ID.
func (m *Message) URL() string {
	source := m.GuildID.String()
	if m.GuildID == 0 {
		source = "@me"
	}
	return "https://discord.com/channels/" + source + "/" + m.ChannelID.String() + "/" + m.ID.String()
}

/////////////

// EmbedBuilder helps build an Embed with chainable methods.
type EmbedBuilder struct {
	embed Embed
}

// NewEmbedBuilder creates a new EmbedBuilder instance.
func NewEmbedBuilder() *EmbedBuilder {
	return &EmbedBuilder{}
}

// SetTitle sets the embed title (max 256 chars).
func (b *EmbedBuilder) SetTitle(title string) *EmbedBuilder {
	if len(title) > EmbedTitleMaxLength {
		panic(
			"Embed title must be max " + strconv.Itoa(
				EmbedTitleMaxLength,
			) + " characters, got " + strconv.Itoa(
				len(title),
			),
		)
	}
	b.embed.Title = title
	return b
}

// SetDescription sets the embed description (max 4096 chars).
func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder {
	if len(desc) > EmbedDescriptionMaxLength {
		panic(
			"Embed description must be max " + strconv.Itoa(
				EmbedDescriptionMaxLength,
			) + " characters, got " + strconv.Itoa(
				len(desc),
			),
		)
	}
	b.embed.Description = desc
	return b
}

// SetURL sets the embed SetURL.
func (b *EmbedBuilder) SetURL(url string) *EmbedBuilder {
	b.embed.URL = url
	return b
}

// SetTimestamp sets the embed timestamp.
func (b *EmbedBuilder) SetTimestamp(t time.Time) *EmbedBuilder {
	b.embed.Timestamp = &t
	return b
}

// SetColor sets the embed color.
func (b *EmbedBuilder) SetColor(color Color) *EmbedBuilder {
	b.embed.Color = color
	return b
}

// SetFooter sets the embed footer text and optional icon URL.
func (b *EmbedBuilder) SetFooter(text, iconURL string) *EmbedBuilder {
	if len(text) > EmbedFooterTextMaxLength {
		panic(
			"Embed footer text must be max " + strconv.Itoa(
				EmbedFooterTextMaxLength,
			) + " characters, got " + strconv.Itoa(
				len(text),
			),
		)
	}
	b.embed.Footer = &EmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return b
}

// SetImage sets the embed image URL.
func (b *EmbedBuilder) SetImage(url string) *EmbedBuilder {
	b.embed.Image = &EmbedImage{URL: url}
	return b
}

// SetThumbnail sets the embed thumbnail URL.
func (b *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder {
	b.embed.Thumbnail = &EmbedThumbnail{URL: url}
	return b
}

// SetAuthor sets the embed author name and optional URL/icon.
func (b *EmbedBuilder) SetAuthor(name, url, iconURL string) *EmbedBuilder {
	if len(name) > EmbedAuthorNameMaxLength {
		panic(
			"Embed author name must be max " + strconv.Itoa(
				EmbedAuthorNameMaxLength,
			) + " characters, got " + strconv.Itoa(
				len(name),
			),
		)
	}
	b.embed.Author = &EmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return b
}

// AddField appends a field to the embed fields slice.
func (b *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder {
	if len(b.embed.Fields) >= EmbedFieldsMaxCount {
		panic("Embed can contain a maximum of " + strconv.Itoa(EmbedFieldsMaxCount) + " fields")
	}
	if len(name) > EmbedFieldNameMaxLength {
		panic(
			"Embed field name must be max " + strconv.Itoa(
				EmbedFieldNameMaxLength,
			) + " characters, got " + strconv.Itoa(
				len(name),
			),
		)
	}
	if len(value) > EmbedFieldValueMaxLength {
		panic(
			"Embed field value must be max " + strconv.Itoa(
				EmbedFieldValueMaxLength,
			) + " characters, got " + strconv.Itoa(
				len(value),
			),
		)
	}
	b.embed.Fields = append(b.embed.Fields, EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return b
}

// AddBlankField appends a blank field to the embed fields slice.
func (b *EmbedBuilder) AddBlankField(inline bool) *EmbedBuilder {
	if len(b.embed.Fields) >= EmbedFieldsMaxCount {
		panic("Embed can contain a maximum of " + strconv.Itoa(EmbedFieldsMaxCount) + " fields")
	}
	b.embed.Fields = append(b.embed.Fields, EmbedField{
		Name:   "\u200E",
		Value:  "\u200E",
		Inline: inline,
	})
	return b
}

// SetFields sets all embed fields at once.
func (e *EmbedBuilder) SetFields(fields ...EmbedField) *EmbedBuilder {
	if len(fields) > EmbedFieldsMaxCount {
		panic(
			"Embed can contain a maximum of " + strconv.Itoa(
				EmbedFieldsMaxCount,
			) + " fields, got " + strconv.Itoa(
				len(fields),
			),
		)
	}
	for i, field := range fields {
		if len(field.Name) > EmbedFieldNameMaxLength {
			panic(
				"Embed field[" + strconv.Itoa(
					i,
				) + "] name must be max " + strconv.Itoa(
					EmbedFieldNameMaxLength,
				) + " characters, got " + strconv.Itoa(
					len(field.Name),
				),
			)
		}
		if len(field.Value) > EmbedFieldValueMaxLength {
			panic(
				"Embed field[" + strconv.Itoa(
					i,
				) + "] value must be max " + strconv.Itoa(
					EmbedFieldValueMaxLength,
				) + " characters, got " + strconv.Itoa(
					len(field.Value),
				),
			)
		}
	}
	e.embed.Fields = fields
	return e
}

// RemoveField removes a field from the EmbedBuilder
func (b *EmbedBuilder) RemoveField(i int) *EmbedBuilder {
	if len(b.embed.Fields) > i {
		b.embed.Fields = slices.Delete(b.embed.Fields, i, i+1)
	}
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *EmbedBuilder) Reset() {
	b.embed = Embed{}
}

// Build returns the final Embed object ready to send.
func (b *EmbedBuilder) Build() Embed {
	return b.embed
}

// String implements the fmt.Stringer interface.
func (m *Message) String() string {
	return m.Content
}

// String implements the fmt.Stringer interface.
func (m *PartialMessage) String() string {
	return m.Content
}

var (
	_ fmt.Stringer = (*Message)(nil)
	_ fmt.Stringer = (*PartialMessage)(nil)
)
