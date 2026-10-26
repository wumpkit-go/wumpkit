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
	"errors"
	"fmt"
)

// InteractionType represents the type of an interaction in Discord.
// It indicates how the interaction was triggered.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-type
type InteractionType int

const (
	InteractionTypePing InteractionType = iota + 1
	// InteractionTypeApplicationCommand is triggered when a user invokes a slash command.
	InteractionTypeApplicationCommand

	// InteractionTypeComponent is triggered when a user interacts with a message component, like a button or select menu.
	InteractionTypeComponent

	// InteractionTypeAutocomplete is triggered when a user is typing in a command option that supports autocomplete.
	InteractionTypeAutocomplete

	// InteractionTypeModalSubmit is triggered when a user submits a modal dialog.
	InteractionTypeModalSubmit
)

// Is returns true if this InteractionType matches the provided one.
func (v InteractionType) Is(other InteractionType) bool {
	return v == other
}

// InteractionContextType is the context in Discord where an interaction can be used, or where it was triggered from.
// Details about using interaction contexts for application commands is in the commands context [documentation].
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-context-types
//
// [documentation]: https://discord.com/developers/docs/interactions/application-commands#interaction-contexts
type InteractionContextType int

const (
	// Interaction can be used within servers
	InteractionContextTypeGuild InteractionContextType = iota

	// Interaction can be used within DMs with the app's bot user
	InteractionContextTypeBotDM

	// Interaction can be used within Group DMs and DMs other than the app's bot user
	InteractionContextTypePrivateChannel
)

// Is returns true if the interaction context's Type matches the provided one.
func (t InteractionContextType) Is(interactionType InteractionContextType) bool {
	return t == interactionType
}

// ApplicationCommandInteractionDataFields holds fields common to all application command interaction data.
type ApplicationCommandInteractionDataFields struct {
	// ID is the unique ID of the invoked command.
	ID Snowflake `json:"id"`

	// Type is the type of the invoked command.
	Type ApplicationCommandType `json:"type"`

	// Name is the name of the invoked command.
	Name string `json:"name"`

	// GuildID is the ID of the guild the command is registered to.
	//
	// Optional:
	//   - Will be 0 for global commands.
	GuildID Snowflake `json:"guild_id"`
}

// ChatInputCommandResolvedInteractionData represents the resolved data inside
// Interaction.Data.Resolved for chat input command interactions. This includes
// users, members, roles, channels, messages, and attachments referenced by the command.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-resolved-data-structure
type ChatInputCommandResolvedInteractionData struct {
	// Users is a map of user IDs to User objects referenced by the command.
	Users map[Snowflake]User `json:"users,omitempty"`

	// Members is a map of user IDs to partial Member objects for the guild.
	Members map[Snowflake]PartialMember `json:"members,omitempty"`

	// Roles is a map of role IDs to Role objects referenced by the command.
	Roles map[Snowflake]Role `json:"roles,omitempty"`

	// Channels is a map of channel IDs to partial Channel objects referenced by the command.
	Channels map[Snowflake]PartialChannel `json:"channels,omitempty"`

	// Messages is a map of message IDs to partial Message objects referenced by the command.
	Messages map[Snowflake]Message `json:"messages,omitempty"`

	// Attachments is a map of attachment IDs to Attachment objects referenced by the command.
	Attachments map[Snowflake]Attachment `json:"attachments,omitempty"`
}

// ChatInputInteractionCommandOption represents a single option provided
// by a user when invoking a chat input command (slash command).
//
// Reference: https://discord.com/developers/docs/interactions/application-commands#interaction-object-application-command-interaction-data-option-object
type ChatInputInteractionCommandOption struct {
	// Name is the name of the option as defined in the command.
	Name string `json:"name"`

	// Type is the type of the option (string, integer, boolean, etc.).
	Type ApplicationCommandOptionType `json:"type"`

	// Value is the raw JSON value provided by the user for this option.
	//
	// Note: It's recommended to use the helper methods provided
	// (String, Int, Float, Bool, Snowflake) to extract the value.
	// These methods will panic if called on a type that doesn't match the
	// expected option type, so you must be sure of the type or check the
	// Type field before calling. For example, if Type is
	// ApplicationCommandOptionTypeString, calling String() is safe.
	Value json.RawMessage `json:"value,omitempty"`

	// Options is present for subcommands and subcommand groups.
	Options []ChatInputInteractionCommandOption `json:"options,omitempty"`

	// Focused indicates whether this option is the currently focused option for autocomplete.
	Focused bool `json:"focused,omitempty"`
}

// TypeMismatchError represents an unrecoverable programmer contract violation where an option
// was accessed with an incompatible type getter or failed type unmarshaling.
type TypeMismatchError struct {
	Option   string
	Expected string
	Cause    error
}

func (e *TypeMismatchError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("wumpkit: type mismatch on option %q (expected %s): %v", e.Option, e.Expected, e.Cause)
	}
	return fmt.Sprintf("wumpkit: type mismatch on option %q (expected %s)", e.Option, e.Expected)
}

func (e *TypeMismatchError) Unwrap() error {
	return e.Cause
}

func newOptionTypeMismatchError(option, expected string, cause error) *TypeMismatchError {
	return &TypeMismatchError{
		Option:   option,
		Expected: expected,
		Cause:    cause,
	}
}

// String returns the option's value as a string.
//
// Panics with *TypeMismatchError if the underlying JSON value cannot be unmarshaled into a string.
func (o *ChatInputInteractionCommandOption) String() string {
	var v string
	if err := json.Unmarshal(o.Value, &v); err != nil {
		panic(newOptionTypeMismatchError(o.Name, "string", err))
	}
	return v
}

// Int returns the option's value as an int.
//
// Panics with *TypeMismatchError if the underlying JSON value cannot be unmarshaled into an int.
func (o *ChatInputInteractionCommandOption) Int() int {
	var v int
	if err := json.Unmarshal(o.Value, &v); err != nil {
		panic(newOptionTypeMismatchError(o.Name, "int", err))
	}
	return v
}

// Float returns the option's value as a float64.
//
// Panics with *TypeMismatchError if the underlying JSON value cannot be unmarshaled into a float64.
func (o *ChatInputInteractionCommandOption) Float() float64 {
	var v float64
	if err := json.Unmarshal(o.Value, &v); err != nil {
		panic(newOptionTypeMismatchError(o.Name, "float64", err))
	}
	return v
}

// Bool returns the option's value as a boolean.
//
// Panics with *TypeMismatchError if the underlying JSON value cannot be unmarshaled into a bool.
func (o *ChatInputInteractionCommandOption) Bool() bool {
	var v bool
	if err := json.Unmarshal(o.Value, &v); err != nil {
		panic(newOptionTypeMismatchError(o.Name, "bool", err))
	}
	return v
}

// Snowflake returns the option's value as a Snowflake.
//
// Panics with *TypeMismatchError if the underlying JSON value cannot be unmarshaled into a Snowflake.
func (o *ChatInputInteractionCommandOption) Snowflake() Snowflake {
	var v Snowflake
	if err := json.Unmarshal(o.Value, &v); err != nil {
		panic(newOptionTypeMismatchError(o.Name, "Snowflake", err))
	}
	return v
}

// ChatInputCommandInteractionData represents the data payload for a chat input
// command interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type ChatInputCommandInteractionData struct {
	ApplicationCommandInteractionDataFields

	// Resolved contains the resolved objects referenced by this command.
	Resolved ChatInputCommandResolvedInteractionData `json:"resolved"`

	// Options contains the parameters and values provided by the user for this command.
	Options []ChatInputInteractionCommandOption `json:"options"`
}

// Option returns the command option with the given name.
// It searches top-level options as well as nested subcommand options.
func (d *ChatInputCommandInteractionData) Option(name string) *ChatInputInteractionCommandOption {
	return findChatInputInteractionCommandOption(d.Options, name)
}

// SubcommandGroup returns the name of the active subcommand group if present.
func (d *ChatInputCommandInteractionData) SubcommandGroup() (string, bool) {
	for _, opt := range d.Options {
		if opt.Type == ApplicationCommandOptionTypeSubCommandGroup {
			return opt.Name, true
		}
	}
	return "", false
}

// Subcommand returns the name of the active subcommand if present.
func (d *ChatInputCommandInteractionData) Subcommand() (string, bool) {
	for _, opt := range d.Options {
		if opt.Type == ApplicationCommandOptionTypeSubCommand {
			return opt.Name, true
		}
		if opt.Type == ApplicationCommandOptionTypeSubCommandGroup {
			for _, subOpt := range opt.Options {
				if subOpt.Type == ApplicationCommandOptionTypeSubCommand {
					return subOpt.Name, true
				}
			}
		}
	}
	return "", false
}

func findChatInputInteractionCommandOption(options []ChatInputInteractionCommandOption, name string) *ChatInputInteractionCommandOption {
	for i := range options {
		if options[i].Name == name {
			return &options[i]
		}
		if len(options[i].Options) > 0 {
			if opt := findChatInputInteractionCommandOption(options[i].Options, name); opt != nil {
				return opt
			}
		}
	}
	return nil
}

// User returns the resolved User for the option with the given name.
func (d *ChatInputCommandInteractionData) User(optionName string) (User, bool) {
	opt := d.Option(optionName)
	if opt == nil {
		return User{}, false
	}
	id := opt.Snowflake()
	u, ok := d.Resolved.Users[id]
	return u, ok
}

// Member returns the resolved PartialMember for the option with the given name.
func (d *ChatInputCommandInteractionData) Member(optionName string) (PartialMember, bool) {
	opt := d.Option(optionName)
	if opt == nil {
		return PartialMember{}, false
	}
	id := opt.Snowflake()
	m, ok := d.Resolved.Members[id]
	return m, ok
}

// Role returns the resolved Role for the option with the given name.
func (d *ChatInputCommandInteractionData) Role(optionName string) (Role, bool) {
	opt := d.Option(optionName)
	if opt == nil {
		return Role{}, false
	}
	id := opt.Snowflake()
	r, ok := d.Resolved.Roles[id]
	return r, ok
}

// Channel returns the resolved Channel for the option with the given name.
func (d *ChatInputCommandInteractionData) Channel(optionName string) (PartialChannel, bool) {
	opt := d.Option(optionName)
	if opt == nil {
		return PartialChannel{}, false
	}
	id := opt.Snowflake()
	ch, ok := d.Resolved.Channels[id]
	return ch, ok
}

// Attachment returns the resolved Attachment for the option with the given name.
func (d *ChatInputCommandInteractionData) Attachment(optionName string) (Attachment, bool) {
	opt := d.Option(optionName)
	if opt == nil {
		return Attachment{}, false
	}
	id := opt.Snowflake()
	att, ok := d.Resolved.Attachments[id]
	return att, ok
}

// String returns the string value for the option with the given name.
func (d *ChatInputCommandInteractionData) String(optionName string) (string, bool) {
	opt := d.Option(optionName)
	if opt == nil || opt.Type != ApplicationCommandOptionTypeString {
		return "", false
	}
	return opt.String(), true
}

// Int returns the int value for the option with the given name.
func (d *ChatInputCommandInteractionData) Int(optionName string) (int, bool) {
	opt := d.Option(optionName)
	if opt == nil || opt.Type != ApplicationCommandOptionTypeInteger {
		return 0, false
	}
	return opt.Int(), true
}

// Float returns the float64 value for the option with the given name.
func (d *ChatInputCommandInteractionData) Float(optionName string) (float64, bool) {
	opt := d.Option(optionName)
	if opt == nil || opt.Type != ApplicationCommandOptionTypeFloat {
		return 0, false
	}
	return opt.Float(), true
}

// Bool returns the boolean value for the option with the given name.
func (d *ChatInputCommandInteractionData) Bool(optionName string) (bool, bool) {
	opt := d.Option(optionName)
	if opt == nil || opt.Type != ApplicationCommandOptionTypeBool {
		return false, false
	}
	return opt.Bool(), true
}

// MessageContextMenuCommandInteractionDataResolved represents the resolved data for a
// message context menu command interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type MessageContextMenuCommandInteractionDataResolved struct {
	// Messages is a map of message IDs to partial Message objects
	// referenced by the command.
	Messages map[Snowflake]Message `json:"messages,omitempty"`

	// Users is a map of user IDs to User objects referenced by the command.
	Users map[Snowflake]User `json:"users,omitempty"`

	// Members is a map of user IDs to partial Member objects for the guild.
	Members map[Snowflake]PartialMember `json:"members,omitempty"`
}

// MessageContextMenuCommandInteractionData represents the data for a message context menu command interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type MessageContextMenuCommandInteractionData struct {
	ApplicationCommandInteractionDataFields

	// Resolved contains the resolved objects referenced by the command, messages in this case.
	Resolved MessageContextMenuCommandInteractionDataResolved `json:"resolved"`

	// TargetID is the id of the message targeted by the command.
	TargetID Snowflake `json:"target_id"`
}

// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type UserContextMenuCommandInteractionDataResolved struct {
	// Users is a map of user IDs to User objects referenced by the command.
	Users map[Snowflake]User `json:"users,omitempty"`

	// Members is a map of user IDs to partial Member objects for the guild.
	Members map[Snowflake]PartialMember `json:"members,omitempty"`
}

// UserContextMenuCommandInteractionData represents the data for a user context menu command interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type UserContextMenuCommandInteractionData struct {
	ApplicationCommandInteractionDataFields

	// Resolved contains the resolved objects referenced by the command, members and users in this case.
	Resolved UserContextMenuCommandInteractionDataResolved `json:"resolved"`

	// TargetID is the id of the user targeted by the command.
	TargetID Snowflake `json:"target_id"`
}

// InteractionFields holds fields common to all Discord interactions.
type InteractionFields struct {
	// ID is the unique ID of the interaction.
	ID Snowflake `json:"id"`

	// Type is the type of the interaction.
	Type InteractionType `json:"type"`

	// ApplicationID is the ID of the application/bot this interaction belongs to.
	ApplicationID Snowflake `json:"application_id"`

	// Token is a token used to respond to the interaction.
	Token string `json:"token"`

	// Version is the read-only version property, always 1.
	Version int `json:"version"`
}

func (i *InteractionFields) GetID() Snowflake {
	return i.ID
}

func (i *InteractionFields) GetType() InteractionType {
	return i.Type
}

func (i *InteractionFields) GetApplicationID() Snowflake {
	return i.ApplicationID
}

func (i *InteractionFields) GetToken() string {
	return i.Token
}

// ApplicationCommandInteractionFields holds fields common to all application command interactions
// (chat input slash commands, user commands, message commands, and autocomplete).
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object
type ApplicationCommandInteractionFields struct {
	InteractionFields

	// GuildID is the ID of the guild the interaction happened in.
	//
	// Optional:
	//   - Will be 0 if the interaction occurred in a DM.
	GuildID Snowflake `json:"guild_id,omitempty"`

	// Guild is the guild the interaction happened in.
	//
	// Optional:
	//   - Will be nil if the interaction occurred in a DM.
	Guild *PartialGuild `json:"guild,omitempty"`

	// ChannelID is the ID of the channel the interaction happened in.
	ChannelID Snowflake `json:"channel_id,omitempty"`

	// Channel is the partial channel the interaction happened in.
	Channel PartialChannel `json:"channel,omitzero"`

	// Member is the guild member data for the invoking user.
	//
	// Optional:
	//   - Present when the interaction is invoked in a guild.
	Member *InteractionMember `json:"member,omitempty"`

	// User is the user object for the invoking user, if invoked in a DM.
	//
	// Optional:
	//   - Present only when the interaction is invoked in a DM.
	User *User `json:"user,omitempty"`

	// Locale is the selected language of the invoking user.
	Locale Locale `json:"locale,omitempty"`

	// GuildLocale is the guild's preferred locale, if invoked in a guild.
	GuildLocale Locale `json:"guild_locale,omitempty"`

	// AppPermissions is a bitwise set of permissions the app has in the source location of the interaction.
	AppPermissions Permissions `json:"app_permissions,omitempty"`

	// Entitlements is a list of entitlements for the invoking user.
	Entitlements []Entitlement `json:"entitlements,omitempty"`

	// AuthorizingIntegrationOwners maps installation contexts that the interaction was authorized for
	// to related user or guild IDs.
	AuthorizingIntegrationOwners map[ApplicationIntegrationType]Snowflake `json:"authorizing_integration_owners,omitempty"`

	// Context indicates where the interaction was triggered from.
	Context InteractionContextType `json:"context,omitempty"`

	// AttachmentSizeLimit is the maximum size of attachments in bytes for this interaction.
	AttachmentSizeLimit int `json:"attachment_size_limit,omitempty"`
}

// Author returns the user who triggered the interaction.
// It returns &Member.User if the interaction was triggered in a guild,
// or User if triggered in a DM. Returns nil if neither is present.
func (i *ApplicationCommandInteractionFields) Author() *User {
	if i.Member != nil {
		return &i.Member.User
	}
	return i.User
}

// PingInteraction represents a Discord Ping interaction.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding
type PingInteraction struct {
	InteractionFields
}

// ChatInputCommandInteraction represents an interaction triggered by a chat input (slash) command.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands
type ChatInputCommandInteraction struct {
	ApplicationCommandInteractionFields

	// Data contains the payload of the interaction specific to chat input commands.
	Data ChatInputCommandInteractionData `json:"data"`
}

// UserContextMenuCommandInteraction represents an interaction triggered by a user context menu command.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands
type UserContextMenuCommandInteraction struct {
	ApplicationCommandInteractionFields

	// Data contains the payload of the interaction specific to user commands.
	Data UserContextMenuCommandInteractionData `json:"data"`
}

// MessageContextMenuCommandInteraction represents an interaction triggered by a message context menu command.
//
// Reference: https://discord.com/developers/docs/interactions/application-commands
type MessageContextMenuCommandInteraction struct {
	ApplicationCommandInteractionFields

	// Data contains the payload of the interaction specific to message commands.
	Data MessageContextMenuCommandInteractionData `json:"data"`
}

// ComponentInteractionFields holds fields common to message component interactions.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object
type ComponentInteractionFields struct {
	ApplicationCommandInteractionFields

	// Message is the message the component was attached to.
	Message Message `json:"message"`
}

// ComponentInteractionData represents the data payload for message component interactions (buttons, select menus).
type ComponentInteractionData struct {
	CustomID      string                                  `json:"custom_id"`
	ComponentType ComponentType                           `json:"component_type"`
	Values        []string                                `json:"values,omitempty"`
	Resolved      ChatInputCommandResolvedInteractionData `json:"resolved,omitzero"`
}

// ComponentInteraction represents an interaction triggered by a message component (button, select menu).
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object
type ComponentInteraction struct {
	ComponentInteractionFields

	// Data contains the component interaction payload (custom_id, component_type, values, resolved).
	Data ComponentInteractionData `json:"data"`
}

// AutoCompleteInteractionData represents the data payload for autocomplete interactions.
type AutoCompleteInteractionData struct {
	ApplicationCommandInteractionDataFields
	Options []ChatInputInteractionCommandOption `json:"options"`
}

// AutoCompleteInteraction represents an interaction triggered by an autocomplete event.
type AutoCompleteInteraction struct {
	ApplicationCommandInteractionFields
	Data AutoCompleteInteractionData `json:"data"`
}

// ModalSubmitInteractionData represents the data payload for modal submit interactions.
type ModalSubmitInteractionData struct {
	CustomID   string                                  `json:"custom_id"`
	Components []Component                             `json:"components"`
	Resolved   ChatInputCommandResolvedInteractionData `json:"resolved,omitzero"`
}

// unmarshal modal submit interaction data with concrete component types
func (m *ModalSubmitInteractionData) UnmarshalJSON(data []byte) error {
	type rawModalSubmitData struct {
		CustomID   string                                  `json:"custom_id"`
		Components []json.RawMessage                       `json:"components"`
		Resolved   ChatInputCommandResolvedInteractionData `json:"resolved,omitzero"`
	}

	var raw rawModalSubmitData
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.CustomID = raw.CustomID
	m.Resolved = raw.Resolved

	if raw.Components != nil {
		m.Components = make([]Component, 0, len(raw.Components))
		for _, rawComp := range raw.Components {
			if len(rawComp) == 0 || bytes.Equal(rawComp, []byte("null")) {
				continue
			}
			comp, err := UnmarshalComponent(rawComp)
			if err != nil {
				return err
			}
			m.Components = append(m.Components, comp)
		}
	}

	return nil
}

// ModalSubmitInteraction represents an interaction triggered by a modal submission.
//
// Reference: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object
type ModalSubmitInteraction struct {
	ApplicationCommandInteractionFields
	Data    ModalSubmitInteractionData `json:"data"`
	Message *Message                   `json:"message,omitempty"`
}

// Interaction is the interface representing a Discord interaction.
//
// This interface can represent any type of interaction returned by Discord,
// including ping interactions, chat input commands, user commands, message commands,
// component interactions, autocomplete interactions, and modal submits.
//
// Use this interface when you want to handle interactions generically without knowing
// the specific concrete type in advance.
//
// You can convert (assert) it to a specific interaction type using a type assertion or
// a type switch, as described in the official Go documentation:
//   - https://go.dev/ref/spec#Type_assertions
//   - https://go.dev/doc/effective_go#type_switch
//
// Example usage:
//
//	var i Interaction
//
//	switch in := i.(type) {
//	case *PingInteraction:
//	    fmt.Println("Ping interaction ID:", in.ID)
//	case *ChatInputCommandInteraction:
//	    fmt.Println("Chat input command name:", in.Data.Name)
//	case *UserContextMenuCommandInteraction:
//	    fmt.Println("User command target ID:", in.Data.TargetID)
//	default:
//	    fmt.Println("Other interaction type:", in.GetType())
//	}
//
// You can also use an if-condition to check a specific type:
//
//	if chatInputIn, ok := i.(*ChatInputCommandInteraction); ok {
//	    fmt.Println("Chat input command:", chatInputIntr.Data.Name)
//	}
type Interaction interface {
	GetID() Snowflake
	GetType() InteractionType
	GetApplicationID() Snowflake
	GetToken() string
}

var (
	_ Interaction = (*PingInteraction)(nil)
	_ Interaction = (*ChatInputCommandInteraction)(nil)
	_ Interaction = (*UserContextMenuCommandInteraction)(nil)
	_ Interaction = (*MessageContextMenuCommandInteraction)(nil)
	_ Interaction = (*ComponentInteraction)(nil)
	_ Interaction = (*AutoCompleteInteraction)(nil)
	_ Interaction = (*ModalSubmitInteraction)(nil)
)

// Helper func to Unmarshal any interaction type to a Interaction interface.
func UnmarshalInteraction(buf []byte) (Interaction, error) {
	var meta struct {
		Type InteractionType `json:"type"`
		Data struct {
			Type ApplicationCommandType `json:"type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &meta); err != nil {
		return nil, err
	}

	var in Interaction
	switch meta.Type {
	case InteractionTypePing:
		var i PingInteraction
		if err := json.Unmarshal(buf, &i); err != nil {
			return nil, err
		}
		in = &i
	case InteractionTypeApplicationCommand:
		switch meta.Data.Type {
		case 0, ApplicationCommandTypeChatInput:
			var i ChatInputCommandInteraction
			if err := json.Unmarshal(buf, &i); err != nil {
				return nil, err
			}
			in = &i
		case ApplicationCommandTypeUser:
			var i UserContextMenuCommandInteraction
			if err := json.Unmarshal(buf, &i); err != nil {
				return nil, err
			}
			in = &i
		case ApplicationCommandTypeMessage:
			var i MessageContextMenuCommandInteraction
			if err := json.Unmarshal(buf, &i); err != nil {
				return nil, err
			}
			in = &i
		default:
			return nil, errors.New("unknown application command interaction type")
		}
	case InteractionTypeComponent:
		var i ComponentInteraction
		if err := json.Unmarshal(buf, &i); err != nil {
			return nil, err
		}
		in = &i
	case InteractionTypeAutocomplete:
		var i AutoCompleteInteraction
		if err := json.Unmarshal(buf, &i); err != nil {
			return nil, err
		}
		in = &i
	case InteractionTypeModalSubmit:
		var i ModalSubmitInteraction
		if err := json.Unmarshal(buf, &i); err != nil {
			return nil, err
		}
		in = &i
	default:
		return nil, errors.New("unknown interaction type")
	}

	hydrateInteractionMembers(in)
	return in, nil
}

func hydrateInteractionMembers(in Interaction) {
	switch it := in.(type) {
	case *ChatInputCommandInteraction:
		if it.Member != nil {
			if it.GuildID != 0 {
				it.Member.GuildID = it.GuildID
			}
			if it.Member.User.ID != 0 {
				it.Member.ID = it.Member.User.ID
			}
		}
		for id, m := range it.Data.Resolved.Members {
			m.ID = id
			if it.GuildID != 0 {
				m.GuildID = it.GuildID
			}
			it.Data.Resolved.Members[id] = m
		}
	case *UserContextMenuCommandInteraction:
		if it.Member != nil {
			if it.GuildID != 0 {
				it.Member.GuildID = it.GuildID
			}
			if it.Member.User.ID != 0 {
				it.Member.ID = it.Member.User.ID
			}
		}
		for id, m := range it.Data.Resolved.Members {
			m.ID = id
			if it.GuildID != 0 {
				m.GuildID = it.GuildID
			}
			it.Data.Resolved.Members[id] = m
		}
	case *MessageContextMenuCommandInteraction:
		if it.Member != nil {
			if it.GuildID != 0 {
				it.Member.GuildID = it.GuildID
			}
			if it.Member.User.ID != 0 {
				it.Member.ID = it.Member.User.ID
			}
		}
		for id, m := range it.Data.Resolved.Members {
			m.ID = id
			if it.GuildID != 0 {
				m.GuildID = it.GuildID
			}
			it.Data.Resolved.Members[id] = m
		}
	case *ComponentInteraction:
		if it.Member != nil {
			if it.GuildID != 0 {
				it.Member.GuildID = it.GuildID
			}
			if it.Member.User.ID != 0 {
				it.Member.ID = it.Member.User.ID
			}
		}
		for id, m := range it.Data.Resolved.Members {
			m.ID = id
			if it.GuildID != 0 {
				m.GuildID = it.GuildID
			}
			it.Data.Resolved.Members[id] = m
		}
	case *AutoCompleteInteraction:
		if it.Member != nil {
			if it.GuildID != 0 {
				it.Member.GuildID = it.GuildID
			}
			if it.Member.User.ID != 0 {
				it.Member.ID = it.Member.User.ID
			}
		}
	case *ModalSubmitInteraction:
		if it.Member != nil {
			if it.GuildID != 0 {
				it.Member.GuildID = it.GuildID
			}
			if it.Member.User.ID != 0 {
				it.Member.ID = it.Member.User.ID
			}
		}
		for id, m := range it.Data.Resolved.Members {
			m.ID = id
			if it.GuildID != 0 {
				m.GuildID = it.GuildID
			}
			it.Data.Resolved.Members[id] = m
		}
	}
}
