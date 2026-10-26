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

// MembershipState represent a team member MembershipState.
//
// Reference: https://discord.com/developers/docs/topics/teams#data-models-membership-state-enum
type MembershipState int

const (
	MembershipStateInvited MembershipState = iota + 1
	MembershipStateAccepted
)

// Is returns true if the team member's membership state matches the provided membership state.
func (s MembershipState) Is(memberShipState MembershipState) bool {
	return s == memberShipState
}

// TeamRole represent a team member role.
//
// Reference: https://discord.com/developers/docs/topics/teams#team-member-roles-team-member-role-types
type TeamRole string

const (
	// Admins have similar access as owners, except they cannot take
	// destructive actions on the team or team-owned apps.
	TeamRoleAdmin TeamRole = "admin"

	// Developers can access information about team-owned apps, like the client secret or public key.
	// They can also take limited actions on team-owned apps, like configuring interaction endpoints
	// or resetting the bot token. Members with the Developer role cannot manage the team or its members,
	// or take destructive actions on team-owned apps.
	TeamRoleDeveloper TeamRole = "developer"

	// Read-only members can access information about a team and any team-owned apps.
	// Some examples include getting the IDs of applications and exporting payout records.
	// Members can also invite bots associated with team-owned apps that are marked private.
	TeamRoleReadOnly TeamRole = "read_only"
)

// Is returns true if the team member's role matches the provided role.
func (r TeamRole) Is(teamRole TeamRole) bool {
	return r == teamRole
}

// TeamMember represent a member of a Discord team.
//
// Reference: https://discord.com/developers/docs/topics/teams#data-models-team-member-object
type TeamMember struct {
	// MembershipState is the user's membership state on the team.
	MembershipState MembershipState `json:"membership_state"`

	// TeamID is the team member's unique Discord snowflake ID.
	TeamID Snowflake `json:"team_id"`

	// User is the partial user object of the team member.
	// Avatar, discriminator, ID, and Username of the user.
	User User `json:"user"`

	// Role is the role of the team member.
	Role TeamRole `json:"role"`
}

// Team represent a Discord team object.
//
// Reference: https://discord.com/developers/docs/topics/teams#data-models-team-object
type Team struct {
	// ID is the team's unique Discord snowflake ID.
	ID Snowflake `json:"id"`

	// Icon is the team's icon hash.
	//
	// Optional:
	//  - May be empty string if no icon.
	Icon string `json:"icon"`

	// Members are the members of the team.
	Members []TeamMember `json:"members"`

	// Name is the name of the team.
	Name string `json:"name"`

	// OwnerID is the user ID of the current team owner.
	OwnerID Snowflake `json:"owner_user_id"`
}

// IconURL returns the URL to the team's icon image.
//
// If the team has a custom icon set, it returns the URL to that icon, otherwise empty string.
// By default, it uses PNG format.
//
// Example usage:
//
//	url := team.IconURL()
func (t *Team) IconURL() string {
	if t.Icon != "" {
		return TeamIconURL(t.ID, t.Icon, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// IconURLWith returns the URL to the team's icon image,
// allowing explicit specification of image format and size.
//
// If the team has a custom icon set, it returns the URL to that icon (otherwise empty string)
// using the provided format and size.
//
// Example usage:
//
//	url := team.IconURLWith(ImageFormatWebP, ImageSize512)
func (t *Team) IconURLWith(format ImageFormat, size ImageSize) string {
	if t.Icon != "" {
		return TeamIconURL(t.ID, t.Icon, format, size)
	}
	return ""
}

// CreatedAt returns the time when this team is created.
func (t *Team) CreatedAt() time.Time {
	return t.ID.Timestamp()
}

// ApplicationFlags represent a Discord application flags.
//
// Reference: https://discord.com/developers/docs/resources/application#application-object-application-flags
type ApplicationFlags int

const (
	// Indicates if an app uses the Auto Moderation API.
	//
	// See: https://discord.com/developers/docs/resources/auto-moderation
	ApplicationFlagAutoModerationRuleCreateBadge = 1 << (iota + 6)

	_
	_
	_
	_
	_

	// Intent required for bots in 100 or more servers to receive presence_update events
	//
	// See: https://discord.com/developers/docs/events/gateway-events#presence-update
	ApplicationFlagGatewayPresence

	// Intent required for bots in under 100 servers to receive presence_update events,
	// found on the Bot page in your app's settings.
	//
	// See: https://discord.com/developers/docs/events/gateway-events#presence-update
	ApplicationFlagGatewayPresenceLimited

	// Intent required for bots in 100 or more servers to receive member-related
	// events like guild_member_add. See the list of member-related events under wumpkit.GatewayIntentGuildMembers
	ApplicationFlagGatewayGuildMembers

	// Intent required for bots in under 100 servers to receive member-related
	// events like guild_member_add, found on the Bot page in your app's settings.
	// See the list of member-related events under wumpkit.GatewayIntentGuildMembers
	ApplicationFlagGatewayGuildMemberLimited

	// Indicates unusual growth of an app that prevents verification.
	ApplicationFlagVerificationPendingGuildLimit

	// Indicates if an app is embedded within the Discord client (currently unavailable publicly).
	ApplicationFlagEmbedded

	// Intent required for bots in 100 or more servers to receive message content.
	//
	// See: https://support-dev.discord.com/hc/en-us/articles/4404772028055-Message-Content-Privileged-Intent-FAQ
	ApplicationFlagGatewayMessageContent

	// Intent required for bots in under 100 servers to receive message content, found on the Bot page in your app's settings.
	//
	// See: https://support-dev.discord.com/hc/en-us/articles/4404772028055-Message-Content-Privileged-Intent-FAQ
	ApplicationFlagGatewayMessageContentLimited

	_
	_
	_

	// Indicates if an app has registered global application commands.
	//
	// See: https://discord.com/developers/docs/interactions/application-commands
	ApplicationFlagApplicationCommandBadge
)

// Has returns true if all provided flags are set.
func (f ApplicationFlags) Has(flags ...ApplicationFlags) bool {
	return BitFieldHas(f, flags...)
}

// ApplicationEventWebhookStatus represent a Discord application application event webhook status.
//
// Reference: https://discord.com/developers/docs/resources/application#application-object-application-event-webhook-status
type ApplicationEventWebhookStatus int

const (
	// Webhook events are disabled by developer.
	ApplicationEventWebhookStatusDisabled ApplicationEventWebhookStatus = 1 + iota
	// Webhook events are enabled by developer.
	ApplicationEventWebhookStatusEnabled
	// Webhook events are disabled by Discord, usually due to inactivity.
	ApplicationEventWebhookStatusDisabledByDiscord
)

// Is returns true if the app webhook event status matches the provided status.
func (s ApplicationEventWebhookStatus) Is(status ApplicationEventWebhookStatus) bool {
	return s == status
}

// WebhookEventTypes represent the webhook event types your app can subscribe to.
//
// Reference: https://discord.com/developers/docs/events/webhook-events#event-types
type WebhookEventTypes string

const (
	// Sent when an app was authorized by a user to a server or their account.
	//
	// See: https://discord.com/developers/docs/events/webhook-events#application-authorized
	WebhookEventTypesApplicationAuthorized WebhookEventTypes = "APPLICATION_AUTHORIZED"

	// Sent when an app was deauthorized by a user.
	//
	// See: https://discord.com/developers/docs/events/webhook-events#application-deauthorized
	WebhookEventTypesApplicationApplicationDeauthorized WebhookEventTypes = "APPLICATION_DEAUTHORIZED"

	// Entitlement was created.
	//
	// See: https://discord.com/developers/docs/events/webhook-events#entitlement-create
	WebhookEventTypesApplicationApplicationEntitlementCreate WebhookEventTypes = "ENTITLEMENT_CREATE"

	// User was added to a Quest (currently unavailable).
	//
	// See: https://discord.com/developers/docs/events/webhook-events#quest-user-enrollment
	WebhookEventTypesApplicationApplicationQuestUserEnrollment WebhookEventTypes = "QUEST_USER_ENROLLMENT"
)

// Is returns true if the webhook event type matches the provided webhook event type.
func (t WebhookEventTypes) Is(webhookEventType WebhookEventTypes) bool {
	return t == webhookEventType
}

// OAuth2Scope represent the scopes you can request in the OAuth2 flow.
//
// Referenceflag flag://discord.com/developers/docs/topics/oauth2#shared-resources-oauth2-flag
type OAuth2Scope string

const (
	// OAuth2ScopeActivitiesRead allows your app to fetch data from a
	// user's "Now Playing/Recently Played" list - requires Discord approval.
	OAuth2ScopeActivitiesRead OAuth2Scope = "activities.read"
	// OAuth2ScopeActivitiesWrite allows your app to update a user's activity - requires
	// Discord approval (NOT REQUIRED FOR GAMESDK ACTIVITY MANAGER).
	OAuth2ScopeActivitiesWrite OAuth2Scope = "activities.write"
	// OAuth2ScopeApplicationsBuildsRead allows your app to read build data for a user's applications.
	OAuth2ScopeApplicationsBuildsRead OAuth2Scope = "applications.builds.read"
	// OAuth2ScopeApplicationsBuildsUpload allows your app to upload/update builds for a user's
	// applications - requires Discord approval.
	OAuth2ScopeApplicationsBuildsUpload OAuth2Scope = "applications.builds.upload"
	// OAuth2ScopeApplicationsCommands allows your app to add commands to a
	// guild - included by default with the bot scope.
	OAuth2ScopeApplicationsCommands OAuth2Scope = "applications.commands"
	// OAuth2ScopeApplicationsCommandsUpdate allows your app to update its commands
	// using a Bearer token - client credentials grant only.
	OAuth2ScopeApplicationsCommandsUpdate OAuth2Scope = "applications.commands.update"
	// OAuth2ScopeApplicationsCommandsPermissionsUpdate allows your app to update permissions
	// for its commands in a guild a user has permissions to.
	OAuth2ScopeApplicationsCommandsPermissionsUpdate OAuth2Scope = "applications.commands.permissions.update"
	// OAuth2ScopeApplicationsEntitlements allows your app to read entitlements for a user's applications.
	OAuth2ScopeApplicationsEntitlements OAuth2Scope = "applications.entitlements"
	// OAuth2ScopeApplicationsStoreUpdate allows your app to read and update store data
	// (SKUs, store listings, achievements, etc.) for a user's applications.
	OAuth2ScopeApplicationsStoreUpdate OAuth2Scope = "applications.store.update"
	// OAuth2ScopeBot for oauth2 bots, this puts the bot in the user's selected guild by default.
	OAuth2ScopeBot OAuth2Scope = "bot"
	// OAuth2ScopeConnections allows /users/@me/connections to return linked third-party accounts.
	OAuth2ScopeConnections OAuth2Scope = "connections"
	// OAuth2ScopeDMChannelsRead allows your app to see information about the user's DMs and
	// group DMs - requires Discord approval.
	OAuth2ScopeDMChannelsRead OAuth2Scope = "dm_channels.read"
	// OAuth2ScopeEmail enables /users/@me to return an email.
	OAuth2ScopeEmail OAuth2Scope = "email"
	// OAuth2ScopeGDMJoin allows your app to join users to a group dm.
	OAuth2ScopeGDMJoin OAuth2Scope = "gdm.join"
	// OAuth2ScopeGuilds allows /users/@me/guilds to return basic information about all of a user's guilds.
	OAuth2ScopeGuilds OAuth2Scope = "guilds"
	// OAuth2ScopeGuildsJoin allows /guilds/{guild.id}/members/{user.id} to be used for joining users to a guild.
	OAuth2ScopeGuildsJoin OAuth2Scope = "guilds.join"
	// OAuth2ScopeGuildsMembersRead allows /users/@me/guilds/{guild.id}/member to return
	// a user's member information in a guild.
	OAuth2ScopeGuildsMembersRead OAuth2Scope = "guilds.members.read"
	// OAuth2ScopeIdentify allows /users/@me without email.
	OAuth2ScopeIdentify OAuth2Scope = "identify"
	// OAuth2ScopeIdentifyPremium allows your app to read a user's Nitro subscription type as defined by premium_type.
	OAuth2ScopeIdentifyPremium OAuth2Scope = "identify.premium"
	// OAuth2ScopeMessagesRead for local rpc server api access, this allows you to read messages
	// from all client channels (otherwise restricted to channels/guilds your app creates).
	OAuth2ScopeMessagesRead OAuth2Scope = "messages.read"
	// OAuth2ScopeRelationshipsRead allows your app to know a user's friends and implicit
	// relationships - requires Discord approval.
	OAuth2ScopeRelationshipsRead OAuth2Scope = "relationships.read"
	// OAuth2ScopeRoleConnectionsWrite allows your app to update a user's connection and metadata for the app.
	OAuth2ScopeRoleConnectionsWrite OAuth2Scope = "role_connections.write"
	// OAuth2ScopeRPC for local rpc server access, this allows you to control a user's local
	// Discord client - requires Discord approval.
	OAuth2ScopeRPC OAuth2Scope = "rpc"
	// OAuth2ScopeRPCActivitiesWrite for local rpc server access, this allows you to update
	// a user's activity - requires Discord approval.
	OAuth2ScopeRPCActivitiesWrite OAuth2Scope = "rpc.activities.write"
	// OAuth2ScopeRPCNotificationsRead for local rpc server access, this allows you to
	// receive notifications pushed out to the user - requires Discord approval.
	OAuth2ScopeRPCNotificationsRead OAuth2Scope = "rpc.notifications.read"
	// OAuth2ScopeRPCVoiceRead for local rpc server access, this allows you to read
	// a user's voice settings and listen for voice events - requires Discord approval.
	OAuth2ScopeRPCVoiceRead OAuth2Scope = "rpc.voice.read"
	// OAuth2ScopeRPCVoiceWrite for local rpc server access, this allows you to update
	// a user's voice settings - requires Discord approval.
	OAuth2ScopeRPCVoiceWrite OAuth2Scope = "rpc.voice.write"
	// OAuth2ScopeVoice allows your app to connect to voice on user's behalf
	// and see all the voice members - requires Discord approval.
	OAuth2ScopeVoice OAuth2Scope = "voice"
	// OAuth2ScopeWebhookIncoming this generates a webhook that is returned in
	// the oauth token response for authorization code grants.
	OAuth2ScopeWebhookIncoming OAuth2Scope = "webhook.incoming"
)

// Is returns true if this OAuth2Scope matches the provided one.
func (v OAuth2Scope) Is(other OAuth2Scope) bool {
	return v == other
}

// ApplicationInstallParams represent a Discord application install params object.
//
// Reference: https://discord.com/developers/docs/resources/application#install-params-object
type ApplicationInstallParams struct {
	// Scopes are scopes to add the application to the server with.
	Scopes []OAuth2Scope `json:"scopes"`

	// Permissions are permissions to request for the bot role
	Permissions Permissions `json:"permissions"`
}

// ApplicationIntegrationType represent where an app can be installed, also called its supported installation contexts.
//
// See: https://discord.com/developers/docs/resources/application#installation-context
//
// Reference: https://discord.com/developers/docs/resources/application#application-object-application-integration-types
type ApplicationIntegrationType int

const (
	// ApplicationIntegrationTypeGuildInstall if app is installable to servers.
	ApplicationIntegrationTypeGuildInstall ApplicationIntegrationType = iota
	// ApplicationIntegrationTypeUserInstall if app is installable to users.
	ApplicationIntegrationTypeUserInstall
)

// Is returns true if this ApplicationIntegrationType matches the provided one.
func (v ApplicationIntegrationType) Is(other ApplicationIntegrationType) bool {
	return v == other
}

type ApplicationIntegrationTypesConfig map[ApplicationIntegrationType]ApplicationIntegrationTypeConfiguration

// ApplicationIntegrationTypeConfiguration object.
//
// Reference: https://discord.com/developers/docs/resources/application#application-object-application-integration-type-configuration-object
type ApplicationIntegrationTypeConfiguration struct {
	// OAuth2InstallParams are the install params for each installation context's default in-app authorization link.
	OAuth2InstallParams *ApplicationInstallParams `json:"oauth2_install_params"`
}

// Application represent a Discord application object.
//
// Reference: https://discord.com/developers/docs/resources/application#application-object
type Application struct {
	// ID is the applications's unique Discord snowflake ID.
	ID Snowflake `json:"id"`

	// Name is the applications's name.
	Name string `json:"name"`

	// Icon is the application's icon hash.
	//
	// Optional:
	//  - May be empty string if no icon.
	Icon string `json:"icon"`

	// Description is the description of a application.
	Description string `json:"description"`

	// RPCOrigins is a List of RPC origin URLs, if RPC is enabled.
	RPCOrigins []string `json:"rpc_origins"`

	// BotPublic When false, only the app owner can add the app to guilds.
	BotPublic bool `json:"bot_public"`

	// BotRequireCodeGrant When true, the app's bot will only join upon completion
	// of the full OAuth2 code grant flow.
	BotRequireCodeGrant bool `json:"bot_require_code_grant"`

	// Bot is a partial user object for the bot user associated with the app
	//
	// Optional.
	Bot *User `json:"bot"`

	// TermsOfServiceURL is the URL of the app's Terms of Service.
	TermsOfServiceURL string `json:"terms_of_service_url"`

	// PrivacyPolicyURL is the URL of the app's Privacy Policy.
	PrivacyPolicyURL string `json:"privacy_policy_url"`

	// Owner is a partial user object for the owner of the app.
	//
	// Optional.
	Owner *User `json:"owner,omitempty"`

	// VerifyKey is the Hex encoded key for verification in interactions and the GameSDK's GetTicket.
	VerifyKey string `json:"verify_key"`

	// Team is the team this app belongs to.
	//
	// Optional:
	//   - Will be nil if app do not belong to any team.
	Team *Team `json:"team"`

	// GuildID is the id of the guild associated with the app. For example, a developer support server.
	//
	// Optional:
	//   - Will be equal 0 if no associated guild is set.
	GuildID Snowflake `json:"guild_id"`

	// Guild is a the partial guild associated with the app. For example, a developer support server.
	//
	// Optional:
	//   - Will be equal nil if no associated guild is set.
	Guild *Guild `json:"guild"`

	// PrimarySkuID If this app is a game sold on Discord, this field will
	// be the id of the "Game SKU" that is created, if exists
	//
	// Optional:
	//   - Will be equal 0 if this app is not a game sold on discord.
	PrimarySkuID Snowflake `json:"primary_sku_id"`

	// Slug If this app is a game sold on Discord, this field will be the URL slug that links to the store page.
	//
	// Optional:
	//   - Will be empty string if this app is not a game sold on discord.
	Slug string `json:"slug"`

	// CoverImage is the app's default rich presence invite cover image hash.
	//
	// Optional:
	//   - Will be empty string if no cover image is set.
	CoverImage string `json:"cover_image"`

	// Flags are the app's public flags.
	Flags ApplicationFlags `json:"flags"`

	// flags_new is the string-serialized integer of public flags for flags > 30
	FlagsNew string `json:"flags_new,omitempty"`

	// ApproximateGuildCount is the approximate count of guilds the app has been added to.
	ApproximateGuildCount *int `json:"approximate_guild_count,omitempty"`

	// ApproximateUserInstallCount is the Approximate count of users that have installed
	// the app (authorized with application.commands as a scope).
	ApproximateUserInstallCount *int `json:"approximate_user_install_count,omitempty"`

	// ApproximateUserAuthorizationCount is the approximate count of users that have OAuth2 authorizations for the app.
	ApproximateUserAuthorizationCount *int `json:"approximate_user_authorization_count,omitempty"`

	// RedirectURIs is an array of redirect URIs for the app.
	RedirectURIs []string `json:"redirect_uris"`

	// RedirectURIs is an array of redirect URIs for the app.
	//
	// See: https://discord.com/developers/docs/interactions/receiving-and-responding#receiving-an-interaction
	//
	// Optional:
	//   - Will be empty string if not set.
	InteractionsEndpointURL string `json:"interactions_endpoint_url"`

	// RoleConnectionsVerificationURL is the role connection verification URL for the app.
	//
	// Optional:
	//   - Will be empty string if not set.
	RoleConnectionsVerificationURL string `json:"role_connections_verification_url"`

	// EventWebhooksURL is the event webhooks URL for the app to receive webhook events
	//
	// See: https://discord.com/developers/docs/events/webhook-events#preparing-for-events
	//
	// Optional:
	//   - Will be empty string if not set.
	EventWebhooksURL string `json:"event_webhooks_url"`

	// EventWebhooksStatus is the app event webhook status.
	EventWebhooksStatus ApplicationEventWebhookStatus `json:"event_webhooks_status"`

	// EventWebhooksTypes is a list of Webhook event types the app subscribes to.
	EventWebhooksTypes []WebhookEventTypes `json:"event_webhooks_types"`

	// Tags is a list of tags describing the content and functionality of the app. Max of 5 tags.
	Tags []string `json:"tags"`

	// InstallParams are the settings for the app's default in-app authorization link, if enabled.
	InstallParams *ApplicationInstallParams `json:"install_params"`

	// IntegrationTypesConfig are the default scopes and permissions for each supported installation context.
	// Value for each key is an integration type configuration object
	IntegrationTypesConfig ApplicationIntegrationTypesConfig `json:"integration_types_config"`

	// CustomInstallURL is the default custom authorization URL for the app, if enabled.
	CustomInstallURL string `json:"custom_install_url"`
}

// IconURL returns the URL to the app's icon image.
//
// If the application has a custom icon set, it returns the URL to that icon, otherwise empty string.
// By default, it uses PNG format.
//
// Example usage:
//
//	url := application.IconURL()
func (a *Application) IconURL() string {
	if a.Icon != "" {
		return ApplicationIconURL(a.ID, a.Icon, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// IconURLWith returns the URL to the app's icon image,
// allowing explicit specification of image format and size.
//
// If the app has a custom icon set, it returns the URL to that icon (otherwise empty string)
// using the provided format and size.
//
// Example usage:
//
//	url := team.IconURLWith(ImageFormatWebP, ImageSize512)
func (a *Application) IconURLWith(format ImageFormat, size ImageSize) string {
	if a.Icon != "" {
		return ApplicationIconURL(a.ID, a.Icon, format, size)
	}
	return ""
}

// CoverImageURL returns the URL to the app's cover image.
//
// If the application has a custom cover image set, it returns the URL to that image, otherwise empty string.
// By default, it uses PNG format.
//
// Example usage:
//
//	url := application.CoverImageURL()
func (a *Application) CoverImageURL() string {
	if a.CoverImage != "" {
		return ApplicationCoverURL(a.ID, a.CoverImage, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// CoverImageURLWith returns the URL to the app's cover image,
// allowing explicit specification of image format and size.
//
// If the app has a custom cover image set, it returns the URL to that image (otherwise empty string)
// using the provided format and size.
//
// Example usage:
//
//	url := team.CoverImageURLWith(ImageFormatWebP, ImageSize512)
func (a *Application) CoverImageURLWith(format ImageFormat, size ImageSize) string {
	if a.Icon != "" {
		return ApplicationCoverURL(a.ID, a.Icon, format, size)
	}
	return ""
}

// String implements the fmt.Stringer interface.
func (a *Application) String() string {
	return a.Name
}

// PartialApplication represent a Discord partial application object.
//
// Reference: https://discord.com/developers/docs/resources/application#application-object
type PartialApplication struct {
	// ID is the applications's unique Discord snowflake ID.
	ID Snowflake `json:"id"`

	// Name is the applications's name.
	Name string `json:"name"`

	// Icon is the application's icon hash.
	//
	// Optional:
	//  - May be empty string if no icon.
	Icon string `json:"icon"`

	// Description is the description of a application.
	Description string `json:"description"`
}

// IconURL returns the URL to the app's icon image.
//
// If the application has a custom icon set, it returns the URL to that icon, otherwise empty string.
// By default, it uses PNG format.
//
// Example usage:
//
//	url := application.IconURL()
func (a *PartialApplication) IconURL() string {
	if a.Icon != "" {
		return ApplicationIconURL(a.ID, a.Icon, ImageFormatDefault, ImageSizeDefault)
	}
	return ""
}

// String implements the fmt.Stringer interface.
func (a *PartialApplication) String() string {
	return a.Name
}

var (
	_ fmt.Stringer = (*Application)(nil)
	_ fmt.Stringer = (*PartialApplication)(nil)
)

// ActivityLocationKind represents the kind of location an activity instance is running in.
type ActivityLocationKind string

const (
	ActivityLocationKindGuildChannel   ActivityLocationKind = "gc"
	ActivityLocationKindPrivateChannel ActivityLocationKind = "pc"
)

// Is returns true if this ActivityLocationKind matches the provided one.
func (v ActivityLocationKind) Is(other ActivityLocationKind) bool {
	return v == other
}

// ActivityLocation represents the location where an activity instance is running.
//
// Reference: https://discord.com/developers/docs/resources/application#get-application-activity-instance-activity-location-object
type ActivityLocation struct {
	// ID is the unique identifier for the location.
	ID string `json:"id"`

	// Kind describes the kind of location.
	Kind ActivityLocationKind `json:"kind"`

	// ChannelID is the ID of the channel.
	ChannelID Snowflake `json:"channel_id"`

	// GuildID is the ID of the guild, if applicable.
	GuildID *Snowflake `json:"guild_id,omitempty"`
}

// ApplicationActivityInstance represents a serialized activity instance.
//
// Reference: https://discord.com/developers/docs/resources/application#get-application-activity-instance-activity-instance-object
type ApplicationActivityInstance struct {
	// ApplicationID is the application ID.
	ApplicationID Snowflake `json:"application_id"`

	// InstanceID is the activity instance ID.
	InstanceID string `json:"instance_id"`

	// LaunchID is the unique identifier for the launch.
	LaunchID Snowflake `json:"launch_id"`

	// Location is the location the instance is running in.
	Location ActivityLocation `json:"location"`

	// Users is the IDs of the users currently connected to the instance.
	Users []Snowflake `json:"users"`
}

// ApplicationRoleConnectionMetadataType represents the type of metadata value.
//
// Reference: https://discord.com/developers/docs/resources/application-role-connection-metadata#application-role-connection-metadata-object-application-role-connection-metadata-type
type ApplicationRoleConnectionMetadataType int

const (
	// ApplicationRoleConnectionMetadataTypeIntegerLessThanOrEqual metadata value is less than or equal to configured value.
	ApplicationRoleConnectionMetadataTypeIntegerLessThanOrEqual ApplicationRoleConnectionMetadataType = 1

	// ApplicationRoleConnectionMetadataTypeIntegerGreaterThanOrEqual metadata value is greater than or equal to configured value.
	ApplicationRoleConnectionMetadataTypeIntegerGreaterThanOrEqual ApplicationRoleConnectionMetadataType = 2

	// ApplicationRoleConnectionMetadataTypeIntegerEqual metadata value is equal to configured value.
	ApplicationRoleConnectionMetadataTypeIntegerEqual ApplicationRoleConnectionMetadataType = 3

	// ApplicationRoleConnectionMetadataTypeIntegerNotEqual metadata value is not equal to configured value.
	ApplicationRoleConnectionMetadataTypeIntegerNotEqual ApplicationRoleConnectionMetadataType = 4

	// ApplicationRoleConnectionMetadataTypeDatetimeLessThanOrEqual metadata value is less than or equal to configured value.
	ApplicationRoleConnectionMetadataTypeDatetimeLessThanOrEqual ApplicationRoleConnectionMetadataType = 5

	// ApplicationRoleConnectionMetadataTypeDatetimeGreaterThanOrEqual metadata value is greater than or equal to configured value.
	ApplicationRoleConnectionMetadataTypeDatetimeGreaterThanOrEqual ApplicationRoleConnectionMetadataType = 6

	// ApplicationRoleConnectionMetadataTypeBooleanEqual metadata value is equal to configured value.
	ApplicationRoleConnectionMetadataTypeBooleanEqual ApplicationRoleConnectionMetadataType = 7

	// ApplicationRoleConnectionMetadataTypeBooleanNotEqual metadata value is not equal to configured value.
	ApplicationRoleConnectionMetadataTypeBooleanNotEqual ApplicationRoleConnectionMetadataType = 8
)

// Is returns true if this ApplicationRoleConnectionMetadataType matches the provided one.
func (v ApplicationRoleConnectionMetadataType) Is(other ApplicationRoleConnectionMetadataType) bool {
	return v == other
}

// ApplicationRoleConnectionMetadata represents role connection metadata for an application.
//
// Reference: https://discord.com/developers/docs/resources/application-role-connection-metadata#application-role-connection-metadata-object
type ApplicationRoleConnectionMetadata struct {
	// Type is the type of metadata value.
	Type ApplicationRoleConnectionMetadataType `json:"type"`

	// Key is the dictionary key for the metadata field (1-50 characters).
	Key string `json:"key"`

	// Name is the name of the metadata field (1-100 characters).
	Name string `json:"name"`

	// NameLocalizations is translations of the name.
	NameLocalizations map[Locale]string `json:"name_localizations,omitempty"`

	// Description is the description of the metadata field (1-200 characters).
	Description string `json:"description"`

	// DescriptionLocalizations is translations of the description.
	DescriptionLocalizations map[Locale]string `json:"description_localizations,omitempty"`
}
