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
	"strconv"
)

// Component limits as defined by Discord's API.
const (
	// ActionRowMaxComponents is the maximum number of interactive components in an ActionRow.
	ActionRowMaxComponents = 5

	// StringSelectMenuMaxOptions is the maximum number of options in a StringSelectMenu.
	StringSelectMenuMaxOptions = 25

	// SectionMinComponents is the minimum number of components in a Section.
	SectionMinComponents = 1

	// SectionMaxComponents is the maximum number of components in a Section.
	SectionMaxComponents = 3

	// MediaGalleryMinItems is the minimum number of items in a MediaGallery.
	MediaGalleryMinItems = 1

	// MediaGalleryMaxItems is the maximum number of items in a MediaGallery.
	MediaGalleryMaxItems = 10
)

// ComponentType represents the type of a Discord component.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#component-object-component-types
type ComponentType int

const (
	// ComponentTypeActionRow is a container to display a row of interactive components.
	//
	// Style: Layout
	// Usage: Message
	ComponentTypeActionRow ComponentType = 1 + iota

	// ComponentTypeButton represents a button object.
	//
	// Style: Interactive
	// Usage: Message
	ComponentTypeButton

	// ComponentTypeStringSelect is a select menu for picking from defined text options.
	//
	// Style: Interactive
	// Usage: Message, Modal
	ComponentTypeStringSelect

	// ComponentTypeTextInput is a text input object.
	//
	// Style: Interactive
	// Usage: Modal
	ComponentTypeTextInput

	// ComponentTypeUserSelect is a select menu for users.
	//
	// Style: Interactive
	// Usage: Message
	ComponentTypeUserSelect

	// ComponentTypeRoleSelect is a select menu for roles.
	//
	// Style: Interactive
	// Usage: Message
	ComponentTypeRoleSelect

	// ComponentTypeMentionableSelect is a select menu for mentionables (users and roles).
	//
	// Style: Interactive
	// Usage: Message
	ComponentTypeMentionableSelect

	// ComponentTypeChannelSelect is a select menu for channels.
	//
	// Style: Interactive
	// Usage: Message
	ComponentTypeChannelSelect

	// ComponentTypeSection is a container to display text alongside an accessory component.
	//
	// Style: Layout
	// Usage: Message
	ComponentTypeSection

	// ComponentTypeTextDisplay is a markdown text display component.
	//
	// Style: Content
	// Usage: Message
	ComponentTypeTextDisplay

	// ComponentTypeThumbnail is a small image used as an accessory component.
	//
	// Style: Content
	// Usage: Message
	ComponentTypeThumbnail

	// ComponentTypeMediaGallery displays images and other media.
	//
	// Style: Content
	// Usage: Message
	ComponentTypeMediaGallery

	// ComponentTypeFile displays an attached file.
	//
	// Style: Content
	// Usage: Message
	ComponentTypeFile

	// ComponentTypeSeparator adds vertical padding between other components.
	//
	// Style: Layout
	// Usage: Message
	ComponentTypeSeparator

	_
	_

	// ComponentTypeContainer visually groups a set of components.
	//
	// Style: Layout
	// Usage: Message
	ComponentTypeContainer

	// ComponentTypeLabel associates a label and description with a component.
	//
	// Style: Layout
	// Usage: Modal
	ComponentTypeLabel
)

// Is returns true if the component's Type matches the provided one.
func (t ComponentType) Is(componentType ComponentType) bool {
	return t == componentType
}

// Component is an interface for all kind of components.
//
// ActionRowComponent, ButtonComponent
// StringSelectMenuComponent, TextInputComponent
// UserSelectMenuComponent, RoleSelectMenuComponent
// MentionableSelectMenuComponent, ChannelSelectMenuComponent
// SectionComponent, TextDisplayComponent
// ThumbnailComponent, MediaGalleryComponent
// FileComponent, SeparatorComponent
// ContainerComponent, UnknownComponent
type Component interface {
	GetID() int
	GetType() ComponentType
	json.Marshaler
}

// LayoutComponent is an interface for all components that can be present as a top level component in a Message.
//
// ActionRowComponent, SectionComponent, TextDisplayComponent,
// MediaGalleryComponent, FileComponent, SeparatorComponent,
// ContainerComponent, LabelComponent
type LayoutComponent interface {
	Component
	layoutComponent()
}

// InteractiveComponent is an interface for components that can be included in an ActionRowComponent.
//
// ButtonComponent, StringSelectMenuComponent, UserSelectMenuComponent, RoleSelectMenuComponent,
// MentionableSelectMenuComponent, ChannelSelectMenuComponent
type InteractiveComponent interface {
	Component
	GetCustomID() string
}

// SectionSubComponent is an interface for all components that can be present in a SectionComponent.
//
// TextDisplayComponent
type SectionSubComponent interface {
	Component
	sectionSubComponent()
}

// SectionAccessoryComponent is an interface for all components that can be present in a SectionComponent.
//
// ButtonComponent, ThumbnailComponent
type SectionAccessoryComponent interface {
	Component
	sectionAccessoryComponent()
}

// ContainerSubComponent is an interface for all components that can be present in a ContainerComponent.
//
// ActionRowComponent, TextDisplayComponent, MediaGalleryComponent,
// SeparatorComponent, FileComponent, SectionComponent
type ContainerSubComponent interface {
	Component
	containerSubComponent()
}

// LabelSubComponent is an interface for all components that can be present in a LabelComponent.
//
// TextInputComponent, StringSelectMenuComponent
type LabelSubComponent interface {
	Component
	labelSubComponent()
}

// ComponentFields holds common fields for all components.
//
// All components include a type field indicating the component type and an optional id field for identification in interaction responses.
// The id is a 32-bit integer, unique within the message, and is generated sequentially by the API if not specified.
// If set to 0, it is treated as empty and replaced by the API.
// The API ensures generated IDs do not conflict with other defined IDs in the message.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#component-object
type ComponentFields struct {
	// ID is an optional 32-bit integer identifier for the component, unique within the message.
	//
	// Note:
	//   - If not specified or set to 0, the API generates a sequential ID.
	//   - Generated IDs do not conflict with other defined IDs in the message.
	ID int `json:"id,omitempty"`

	// Type is the type of the component.
	Type ComponentType `json:"type"`
}

func (c *ComponentFields) GetID() int {
	return c.ID
}

func (c *ComponentFields) GetType() ComponentType {
	return c.Type
}

// InteractiveComponentFields holds common fields for interactive components, extending ComponentFields with a custom_id.
//
// Interactive components (e.g., buttons, select menus) require a custom_id, which is a developer-defined identifier returned in interaction payloads.
// The custom_id must be unique per component within the same message and is used to maintain state or pass data.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#component-object
type InteractiveComponentFields struct {
	ComponentFields

	// CustomID is a developer-defined identifier for interactive components, returned in interaction payloads.
	//
	// Note:
	//   - Max 100 characters.
	//   - Must be unique among components in the same message.
	//   - Used to maintain state or pass data when a user interacts with the component.
	CustomID string `json:"custom_id"`
}

func (c *InteractiveComponentFields) GetCustomID() string {
	return c.CustomID
}

// ActionRowComponent is a top-level layout component that organizes interactive components.
//
// It can contain either up to 5 ButtonComponent instances or a single select component (StringSelectMenuComponent,
// UserSelectMenuComponent, RoleSelectMenuComponent, MentionableSelectMenuComponent, or ChannelSelectMenuComponent).
// In modals, ActionRowComponent with TextInputComponent is deprecated; use LabelComponent instead.
//
// Note:
//   - Only one type of component (buttons or a single select) can be included at a time.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#action-row
type ActionRowComponent struct {
	ComponentFields

	// Components is an array of up to 5 interactive button components or a single select component.
	//
	// Valid components:
	//   - ButtonComponent (up to 5)
	//   - StringSelectMenuComponent (1)
	//   - UserSelectMenuComponent (1)
	//   - RoleSelectMenuComponent (1)
	//   - MentionableSelectMenuComponent (1)
	//   - ChannelSelectMenuComponent (1)
	Components []InteractiveComponent `json:"components"`
}

func (ActionRowComponent) layoutComponent()       {}
func (ActionRowComponent) containerSubComponent() {}

var _ json.Unmarshaler = (*ActionRowComponent)(nil)

func (c *ActionRowComponent) UnmarshalJSON(buf []byte) error {
	type tempActionRowComponent struct {
		ComponentFields
		Components []json.RawMessage `json:"components"`
	}

	var temp tempActionRowComponent
	if err := json.Unmarshal(buf, &temp); err != nil {
		return err
	}

	c.ComponentFields = temp.ComponentFields

	if temp.Components != nil {
		c.Components = make([]InteractiveComponent, 0, len(temp.Components))
		for i := range len(temp.Components) {
			if len(temp.Components[i]) == 0 || bytes.Equal(temp.Components[i], []byte("null")) {
				continue
			}
			component, err := UnmarshalComponent(temp.Components[i])
			if err != nil {
				return err
			}
			if interactiveComponent, ok := component.(InteractiveComponent); ok {
				c.Components = append(c.Components, interactiveComponent)
			} else {
				return errors.New(
					"cannot unmarshal non-InteractiveComponent into InteractiveComponent",
				)
			}
		}
	}

	return nil
}

var _ json.Marshaler = (*ActionRowComponent)(nil)

func (c *ActionRowComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeActionRow
	type NoMethod ActionRowComponent
	return json.Marshal((*NoMethod)(c))
}

// ButtonStyle represents different styles for interactive buttons, each with specific actions and required fields.
//
// The style determines the button's appearance and behavior, such as triggering interactions or navigating to a URL.
// Non-link and non-premium buttons (Primary, Secondary, Success, Danger) require a custom_id and send interactions.
// Link buttons require a url, cannot have a custom_id, and do not send interactions.
// Premium buttons require a sku_id, cannot have a custom_id, label, url, or emoji, and do not send interactions.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#button-object-button-styles
type ButtonStyle int

const (
	// ButtonStylePrimary is the most important or recommended action in a group of options.
	//
	// Required Field: ButtonComponent.CustomID
	// Note: Sends an interaction when clicked.
	ButtonStylePrimary ButtonStyle = iota + 1

	// ButtonStyleSecondary represents alternative or supporting actions.
	//
	// Required Field: ButtonComponent.CustomID
	// Note: Sends an interaction when clicked.
	ButtonStyleSecondary

	// ButtonStyleSuccess represents positive confirmation or completion actions.
	//
	// Required Field: ButtonComponent.CustomID
	// Note: Sends an interaction when clicked.
	ButtonStyleSuccess

	// ButtonStyleDanger represents an action with irreversible consequences.
	//
	// Required Field: ButtonComponent.CustomID
	// Note: Sends an interaction when clicked.
	ButtonStyleDanger

	// ButtonStyleLink navigates to a URL.
	//
	// Required Field: ButtonComponent.URL
	// Note: Does not send an interaction when clicked.
	ButtonStyleLink

	// ButtonStylePremium represents a purchase action.
	//
	// Required Field: ButtonComponent.SkuID
	// Note: Does not send an interaction when clicked.
	ButtonStylePremium
)

// Is returns true if the button style matches the provided one.
func (s ButtonStyle) Is(style ButtonStyle) bool {
	return s == style
}

// ButtonComponent is an interactive component used in messages, creating clickable elements that users can interact with.
//
// Buttons must be placed inside an ActionRowComponent or a SectionComponent's accessory field.
// Non-link and non-premium buttons (Primary, Secondary, Success, Danger) send an interaction to the app when clicked.
// Link and Premium buttons do not send interactions.
//
// Note:
//   - Non-link and non-premium buttons require a custom_id and cannot have a url or sku_id.
//   - Link buttons require a url and cannot have a custom_id.
//   - Premium buttons require a sku_id and cannot have a custom_id, label, url, or emoji.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#button-object
type ButtonComponent struct {
	InteractiveComponentFields

	// Style is the button style, determining its appearance and behavior.
	Style ButtonStyle `json:"style"`

	// Label is the text that appears on the button.
	//
	// Note:
	//   - Max 80 characters.
	//   - Cannot be set for Premium buttons.
	Label string `json:"label,omitempty"`

	// Emoji is the emoji that appears on the button.
	//
	// Note:
	//   - Cannot be set for Premium buttons.
	Emoji *PartialEmoji `json:"emoji,omitempty"`

	// SkuID is the identifier for a purchasable SKU, used only for Premium buttons.
	//
	// Note:
	//   - Required for Premium buttons.
	//   - Cannot be set for non-Premium buttons.
	SkuID Snowflake `json:"sku_id,omitempty"`

	// URL is the URL for Link buttons.
	//
	// Note:
	//   - Max 512 characters.
	//   - Required for Link buttons.
	//   - Cannot be set for non-Link buttons.
	URL string `json:"url,omitempty"`

	// Disabled is whether the button is disabled.
	//
	// Note:
	//   - Defaults to false.
	Disabled bool `json:"disabled,omitempty"`
}

func (ButtonComponent) sectionAccessoryComponent() {}

var _ json.Marshaler = (*ButtonComponent)(nil)

func (c *ButtonComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeButton
	type NoMethod ButtonComponent
	return json.Marshal((*NoMethod)(c))
}

// SelectOptionStructure represents an option in a StringSelectMenuComponent.
//
// It defines the user-facing label, developer-defined value, and optional description and emoji for a selectable option.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#select-menu-object-select-option-structure
type SelectOptionStructure struct {
	// Label is the user-facing name of the option.
	//
	// Note:
	//   - Maximum of 100 characters.
	Label string `json:"label"`

	// Value is the developer-defined value of the option.
	//
	// Note:
	//   - Maximum of 100 characters.
	Value string `json:"value"`

	// Description is an additional description of the option.
	//
	// Note:
	//   - Maximum of 100 characters.
	Description string `json:"description,omitempty"`

	// Emoji is the emoji displayed alongside the option.
	Emoji *PartialEmoji `json:"emoji,omitempty"`

	// Default specifies whether this option is selected by default.
	Default bool `json:"default,omitempty"`
}

// StringSelectMenuComponent represents a string select menu, an interactive component allowing users to select one or more predefined text options.
//
// It supports both single-select and multi-select behavior, sending an interaction to the application when a user makes their selection(s).
// StringSelectMenuComponent must be placed inside an ActionRowComponent for messages or a LabelComponent for modals.
//
// Note:
//   - Maximum of 25 options can be provided.
//   - In messages, it must be the only component in an ActionRowComponent (cannot coexist with buttons).
//   - In modals, the `Disabled` field will cause an error if set to true, as modals do not support disabled components.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#select-menu-object-select-menu-structure
type StringSelectMenuComponent struct {
	InteractiveComponentFields

	// Options is an array of choices available in the select menu.
	//
	// Note:
	//   - Maximum of 25 options.
	Options []SelectOptionStructure `json:"options,omitempty"`

	// Placeholder is the custom placeholder text displayed when no option is selected.
	//
	// Note:
	//   - Maximum of 150 characters.
	Placeholder string `json:"placeholder,omitempty"`

	// MinValues is the minimum number of options that must be selected.
	//
	// Note:
	//   - Defaults to 1.
	//   - Minimum 0, maximum 25.
	MinValues *int `json:"min_values,omitempty"`

	// MaxValues is the maximum number of options that can be selected.
	//
	// Note:
	//   - Defaults to 1.
	//   - Maximum 25.
	MaxValues int `json:"max_values,omitempty"`

	// Required specifies whether the select menu must be filled in a modal.
	//
	// Note:
	//   - Defaults to true.
	//   - Only applicable in modals; ignored in messages.
	Required bool `json:"required,omitempty"`

	// Disabled specifies whether the select menu is disabled in a message.
	//
	// Note:
	//   - Defaults to false.
	//   - Causes an error if set to true in modals.
	Disabled bool `json:"disabled,omitempty"`
}

func (StringSelectMenuComponent) labelSubComponent() {}

var _ json.Marshaler = (*StringSelectMenuComponent)(nil)

func (c *StringSelectMenuComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeStringSelect
	type NoMethod StringSelectMenuComponent
	return json.Marshal((*NoMethod)(c))
}

// TextInputStyle represents the style of a TextInputComponent.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#text-inputs-text-input-styles
type TextInputStyle int

const (
	// TextInputStyleShort represents a single-line text input.
	TextInputStyleShort TextInputStyle = 1 + iota

	// TextInputStyleLong represents a multi-line text input.
	TextInputStyleLong
)

// Is returns true if the text input style matches the provided one.
func (s TextInputStyle) Is(style TextInputStyle) bool {
	return s == style
}

// TextInputComponent is an interactive component that allows users to enter free-form text responses in modals.
// It supports both short (single-line) and long (multi-line) input styles.
//
// TextInputComponent must be placed inside a LabelComponent in modals.
//
// Note:
//   - Only available in modals.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#text-inputs
type TextInputComponent struct {
	InteractiveComponentFields

	// Style specifies the text input style (short or paragraph).
	Style TextInputStyle `json:"style,omitempty"`

	// MinLength is the minimum input length for the text input.
	//
	// Note:
	//   - Minimum 0, maximum 4000.
	MinLength *int `json:"min_length,omitempty"`

	// MaxLength is the maximum input length for the text input.
	//
	// Note:
	//   - Minimum 1, maximum 4000.
	MaxLength int `json:"max_length,omitempty"`

	// Required specifies whether this component must be filled in a modal.
	//
	// Note:
	//   - Defaults to true.
	Required bool `json:"required,omitempty"`

	// Value is the pre-filled text for this component.
	//
	// Note:
	//   - Maximum of 4000 characters.
	Value string `json:"value,omitempty"`

	// Placeholder is the custom placeholder text displayed when the input is empty.
	//
	// Note:
	//   - Maximum of 100 characters.
	Placeholder string `json:"placeholder,omitempty"`
}

func (TextInputComponent) labelSubComponent() {}

var _ json.Marshaler = (*TextInputComponent)(nil)

func (c *TextInputComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeTextInput
	type NoMethod TextInputComponent
	return json.Marshal((*NoMethod)(c))
}

// SelectDefaultValueType represents the type of a default value in a select menu component.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#select-menu-object-select-default-value-structure
type SelectDefaultValueType string

const (
	// SelectDefaultValueTypeUser indicates the default value is a user ID.
	SelectDefaultValueTypeUser SelectDefaultValueType = "user"

	// SelectDefaultValueTypeRole indicates the default value is a role ID.
	SelectDefaultValueTypeRole SelectDefaultValueType = "role"

	// SelectDefaultValueTypeChannel indicates the default value is a channel ID.
	SelectDefaultValueTypeChannel SelectDefaultValueType = "channel"
)

// Is returns true if the value Type matches the provided one.
func (t SelectDefaultValueType) Is(valueType SelectDefaultValueType) bool {
	return t == valueType
}

// SelectDefaultValue represents a default value in a select menu component
// (e.g., UserSelectMenuComponent, RoleSelectMenuComponent, ChannelSelectMenuComponent, MentionableSelectMenuComponent).
//
// It specifies the ID and type of the default selected entity (user, role, or channel).
//
// Reference: https://discord.com/developers/docs/interactions/message-components#select-menu-object-select-default-value-structure
type SelectDefaultValue struct {
	// ID is the identifier of the default value (e.g., user ID, role ID, or channel ID).
	ID Snowflake `json:"id"`

	// Type is the type of the default value (user, role, or channel).
	Type SelectDefaultValueType `json:"type"`
}

// UserSelectMenuComponent represents a user select menu component.
//
// Reference: https://discord.com/developers/docs/components/reference#user-select-user-select-structure
type UserSelectMenuComponent struct {
	InteractiveComponentFields

	// Placeholder is the custom placeholder text if the input is empty.
	//
	// Note:
	//   - Max 150 characters.
	Placeholder string `json:"placeholder,omitempty"`

	// DefaultValues is a list of default values for auto-populated select menu components;
	// number of default values must be in the range defined by min_values and max_values.
	DefaultValues []SelectDefaultValue `json:"default_values,omitempty"`

	// MinValues is the minimum number of items that must be chosen (defaults to 1).
	//
	// Note:
	//   - Min 0, max 25.
	MinValues *int `json:"min_values,omitempty"`

	// MaxValues is the maximum number of items that can be chosen (defaults to 1).
	//
	// Note:
	//   - Min 1, max 25.
	MaxValues int `json:"max_values,omitempty"`

	// Disabled is whether select menu is disabled (defaults to false).
	Disabled bool `json:"disabled,omitempty"`
}

var _ json.Marshaler = (*UserSelectMenuComponent)(nil)

func (c *UserSelectMenuComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeUserSelect
	type NoMethod UserSelectMenuComponent
	return json.Marshal((*NoMethod)(c))
}

// RoleSelectMenuComponent represents a role select menu, an interactive component allowing users to select one or more roles in a message.
// Options are automatically populated based on the server's available roles.
//
// It supports both single-select and multi-select behavior, sending an interaction to the application when a user makes their selection(s).
// RoleSelectMenuComponent must be placed inside an ActionRowComponent and is only available in messages.
// An ActionRowComponent containing a RoleSelectMenuComponent cannot include buttons.
//
// Note:
//   - Maximum of 25 selections can be allowed (via MaxValues).
//   - Options are auto-populated by Discord based on server roles.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#select-menu-object-select-menu-structure
type RoleSelectMenuComponent struct {
	InteractiveComponentFields

	// Placeholder is the custom placeholder text displayed when no role is selected.
	//
	// Note:
	//   - Maximum of 150 characters.
	Placeholder string `json:"placeholder,omitempty"`

	// DefaultValues is an array of default roles for the select menu.
	//
	// Note:
	//   - Number of default values must be within the range defined by MinValues and MaxValues.
	DefaultValues []SelectDefaultValue `json:"default_values,omitempty"`

	// MinValues is the minimum number of roles that must be selected.
	//
	// Note:
	//   - Defaults to 1.
	//   - Minimum 0, maximum 25.
	MinValues *int `json:"min_values,omitempty"`

	// MaxValues is the maximum number of roles that can be selected.
	//
	// Note:
	//   - Defaults to 1.
	//   - Maximum 25.
	MaxValues int `json:"max_values,omitempty"`

	// Disabled specifies whether the select menu is disabled in a message.
	//
	// Note:
	//   - Defaults to false.
	Disabled bool `json:"disabled,omitempty"`
}

var _ json.Marshaler = (*RoleSelectMenuComponent)(nil)

func (c *RoleSelectMenuComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeRoleSelect
	type NoMethod RoleSelectMenuComponent
	return json.Marshal((*NoMethod)(c))
}

// MentionableSelectMenuComponent represents a mentionable select menu component.
//
// Reference: https://discord.com/developers/docs/components/reference#mentionable-select
type MentionableSelectMenuComponent struct {
	InteractiveComponentFields

	// Placeholder is the text if nothing is selected.
	//
	// Note:
	//   - Max 150 characters.
	Placeholder string `json:"placeholder,omitempty"`

	// DefaultValues is a list of default values for auto-populated select menu components;
	// number of default values must be in the range defined by min_values and max_values.
	DefaultValues []SelectDefaultValue `json:"default_values,omitempty"`

	// MinValues is the minimum number of items that must be chosen (defaults to 1).
	//
	// Note:
	//   - Min 0, max 25.
	MinValues *int `json:"min_values,omitempty"`

	// MaxValues is the maximum number of items that can be chosen (defaults to 1).
	//
	// Note:
	//   - Min 1, max 25.
	MaxValues int `json:"max_values,omitempty"`

	// Disabled is whether select menu is disabled (defaults to false).
	Disabled bool `json:"disabled,omitempty"`
}

var _ json.Marshaler = (*MentionableSelectMenuComponent)(nil)

func (c *MentionableSelectMenuComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeMentionableSelect
	type NoMethod MentionableSelectMenuComponent
	return json.Marshal((*NoMethod)(c))
}

// ChannelSelectMenuComponent represents a channel select menu, an interactive component allowing users to select one or more channels in a message.
// Options are automatically populated based on the server's available channels and can be filtered by channel types.
//
// It supports both single-select and multi-select behavior, sending an interaction to the application when a user makes their selection(s).
// ChannelSelectMenuComponent must be placed inside an ActionRowComponent and is only available in messages.
// An ActionRowComponent containing a ChannelSelectMenuComponent cannot include buttons.
//
// Note:
//   - Maximum of 25 selections can be allowed (via MaxValues).
//   - Options are auto-populated by Discord based on server channels, filterable by ChannelTypes.
//
// Reference: https://discord.com/developers/docs/interactions/message-components#select-menu-object-select-menu-structure
type ChannelSelectMenuComponent struct {
	InteractiveComponentFields

	// ChannelTypes is an array of channel types to include in the select menu.
	//
	// Note:
	//   - Filters the channels shown in the select menu.
	ChannelTypes []ChannelType `json:"channel_types,omitempty"`

	// Placeholder is the custom placeholder text displayed when no channel is selected.
	//
	// Note:
	//   - Maximum of 150 characters.
	Placeholder string `json:"placeholder,omitempty"`

	// DefaultValues is an array of default channels for the select menu.
	//
	// Note:
	//   - Number of default values must be within the range defined by MinValues and MaxValues.
	DefaultValues []SelectDefaultValue `json:"default_values,omitempty"`

	// MinValues is the minimum number of channels that must be selected.
	//
	// Note:
	//   - Defaults to 1.
	//   - Minimum 0, maximum 25.
	MinValues *int `json:"min_values,omitempty"`

	// MaxValues is the maximum number of channels that can be selected.
	//
	// Note:
	//   - Defaults to 1.
	//   - Maximum 25.
	MaxValues int `json:"max_values,omitempty"`

	// Disabled specifies whether the select menu is disabled in a message.
	//
	// Note:
	//   - Defaults to false.
	Disabled bool `json:"disabled,omitempty"`
}

var _ json.Marshaler = (*ChannelSelectMenuComponent)(nil)

func (c *ChannelSelectMenuComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeChannelSelect
	type NoMethod ChannelSelectMenuComponent
	return json.Marshal((*NoMethod)(c))
}

// SectionComponent is a top-level layout component that contextually associates content with an accessory component.
//
// It is typically used to associate text content with an accessory, such as a button or thumbnail.
// Sections require the IS_COMPONENTS_V2 message flag (1 << 15) to be set when sending the message.
// Additional component types for content and accessories may be supported in the future.
//
// Note:
//   - Only available in messages.
//   - Requires the IS_COMPONENTS_V2 message flag (1 << 15).
//   - Contains one to three child components for content.
//
// Reference: https://discord.com/developers/docs/components/reference#section
type SectionComponent struct {
	ComponentFields

	// Components is an array of one to three child components representing the content of the section.
	//
	// Valid components:
	//   - TextDisplayComponent
	Components []SectionSubComponent `json:"components"`

	// Accessory is a component contextually associated with the content of the section.
	//
	// Valid components:
	//   - ButtonComponent
	//   - ThumbnailComponent
	Accessory SectionAccessoryComponent `json:"accessory,omitempty"`
}

func (SectionComponent) layoutComponent()       {}
func (SectionComponent) containerSubComponent() {}

var _ json.Unmarshaler = (*SectionComponent)(nil)

func (c *SectionComponent) UnmarshalJSON(buf []byte) error {
	type tempSectionComponent struct {
		ComponentFields
		Components []json.RawMessage `json:"components"`
		Accessory  json.RawMessage   `json:"accessory,omitempty"`
	}

	var temp tempSectionComponent
	if err := json.Unmarshal(buf, &temp); err != nil {
		return err
	}

	c.ComponentFields = temp.ComponentFields

	if temp.Components != nil {
		c.Components = make([]SectionSubComponent, 0, len(temp.Components))
		for i := range len(temp.Components) {
			if len(temp.Components[i]) == 0 || bytes.Equal(temp.Components[i], []byte("null")) {
				continue
			}
			component, err := UnmarshalComponent(temp.Components[i])
			if err != nil {
				return err
			}
			if sub, ok := component.(SectionSubComponent); ok {
				c.Components = append(c.Components, sub)
			} else {
				return errors.New(
					"cannot unmarshal non-SectionSubComponent into SectionSubComponent",
				)
			}
		}
	}

	accessory, err := UnmarshalComponent(temp.Accessory)
	if err != nil {
		return err
	}
	if sub, ok := accessory.(SectionAccessoryComponent); ok {
		c.Accessory = sub
	}

	return nil
}

var _ json.Marshaler = (*SectionComponent)(nil)

func (c *SectionComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeSection
	type NoMethod SectionComponent
	return json.Marshal((*NoMethod)(c))
}

// TextDisplayComponent is a content component that displays markdown-formatted text, including mentions and emojis.
//
// It behaves similarly to the content field of a message, allowing multiple text components to control message layout.
// Pingable mentions (@user, @role, etc.) in this component will trigger notifications based on the message's allowed_mentions field.
// Text Displays require the IS_COMPONENTS_V2 message flag (1 << 15) to be set when sending the message.
//
// Note:
//   - Only available in messages.
//   - Requires the IS_COMPONENTS_V2 message flag (1 << 15).
//   - Supports markdown formatting, user/role mentions, and emojis.
//
// Reference: https://discord.com/developers/docs/components/reference#text-display
type TextDisplayComponent struct {
	ComponentFields

	// Content is the markdown-formatted text to be displayed, similar to a message's content field.
	Content string `json:"content,omitempty"`
}

func (TextDisplayComponent) layoutComponent()       {}
func (TextDisplayComponent) sectionSubComponent()   {}
func (TextDisplayComponent) containerSubComponent() {}

var _ json.Marshaler = (*TextDisplayComponent)(nil)

func (c *TextDisplayComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeTextDisplay
	type NoMethod TextDisplayComponent
	return json.Marshal((*NoMethod)(c))
}

type UnfurledMediaItemLoadingState int

const (
	UnfurledMediaItemLoadingStateUnknown UnfurledMediaItemLoadingState = iota
	UnfurledMediaItemLoadingStateLoading
	UnfurledMediaItemLoadingStateLoadedSuccess
	UnfurledMediaItemLoadingStateLoadedNotFound
)

// Is returns true if this UnfurledMediaItemLoadingState matches the provided one.
func (v UnfurledMediaItemLoadingState) Is(other UnfurledMediaItemLoadingState) bool {
	return v == other
}

// UnfurledMediaItem represents an unfurled media item.
//
// Reference: https://discord.com/developers/docs/components/reference#unfurled-media-item
type UnfurledMediaItem struct {
	// URL is the url of the media item.
	//
	// Note:
	//   - Supports arbitrary urls and 'attachment://filename' references.
	URL string `json:"url"`

	// ProxyURL is the proxied url of the media item. This field is ignored
	// and provided by the API as part of the response.
	ProxyURL string `json:"proxy_url"`

	// Height is the height of the media item. This field is ignored
	// and provided by the API as part of the response.
	Height int `json:"height,omitempty"`

	// Width is the width of the media item. This field is ignored
	// and provided by the API as part of the response.
	Width int `json:"width,omitempty"`

	// ContentType is the [media type] of the content. This field is ignored
	// and provided by the API as part of the response.
	//
	// [media type]: https://en.wikipedia.org/wiki/Media_type
	ContentType string `json:"content_type,omitempty"`

	// AttachmentID is the id of the uploaded attachment. This field is ignored
	// and provided by the API as part of the response.
	AttachmentID Snowflake `json:"attachment_id,omitempty"`
}

// ThumbnailComponent represents a Thumbnail component.
//
// Reference: https://discord.com/developers/docs/components/reference#thumbnail
type ThumbnailComponent struct {
	ComponentFields

	// Description is an alt text for the media.
	//
	// Note:
	//   - Max 1024 characters.
	Description string `json:"description,omitempty"`

	// Media is a url or attachment provided as an unfurled media item.
	Media UnfurledMediaItem `json:"media"`

	// Spoiler is whether the thumbnail should be a spoiler (or blurred out). Defaults to false.
	Spoiler bool `json:"spoiler,omitempty"`
}

func (ThumbnailComponent) sectionAccessoryComponent() {}

var _ json.Marshaler = (*ThumbnailComponent)(nil)

func (c *ThumbnailComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeThumbnail
	type NoMethod ThumbnailComponent
	return json.Marshal((*NoMethod)(c))
}

// MediaGalleryItem represents an item in a MediaGallery component.
//
// It contains a single media item with an optional description and spoiler flag.
//
// Note:
//   - Maximum of 1024 characters for the description.
//
// Reference: https://discord.com/developers/docs/components/reference#media-gallery-media-gallery-item-structure
type MediaGalleryItem struct {
	// Media is a url or attachment provided as an unfurled media item.
	Media UnfurledMediaItem `json:"media"`

	// Description is an alt text for the media.
	//
	// Note:
	//   - Max 1024 characters.
	Description string `json:"description,omitempty"`

	// Spoiler is whether the media should be a spoiler (or blurred out). Defaults to false.
	Spoiler bool `json:"spoiler,omitempty"`
}

// MediaGalleryComponent is a top-level content component that displays 1-10 media attachments in an organized gallery format.
//
// Each item in the gallery can have an optional description and can be marked as a spoiler.
// Media Galleries require the IS_COMPONENTS_V2 message flag (1 << 15) to be set when sending the message.
//
// Note:
//   - Only available in messages.
//   - Requires the IS_COMPONENTS_V2 message flag (1 << 15).
//   - Contains 1 to 10 media gallery items.
//
// Reference: https://discord.com/developers/docs/components/reference#media-gallery
type MediaGalleryComponent struct {
	ComponentFields

	// Items is an array of 1 to 10 media gallery items.
	//
	// Valid components:
	//   - MediaGalleryItem (1 to 10)
	Items []MediaGalleryItem `json:"items"`
}

func (MediaGalleryComponent) layoutComponent()       {}
func (MediaGalleryComponent) containerSubComponent() {}

var _ json.Marshaler = (*MediaGalleryComponent)(nil)

func (c *MediaGalleryComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeMediaGallery
	type NoMethod MediaGalleryComponent
	return json.Marshal((*NoMethod)(c))
}

// FileComponent is a top-level content component that displays an uploaded file as an attachment to the message.
//
// Each file component can only display one attached file, but multiple files can be uploaded and added to different file components within a payload.
// The file must use the attachment://filename syntax in the unfurled media item.
// File components require the IS_COMPONENTS_V2 message flag (1 << 15) to be set when sending the message.
//
// Note:
//   - Only available in messages.
//   - Requires the IS_COMPONENTS_V2 message flag (1 << 15).
//   - Only supports attachment references using the attachment://filename syntax.
//
// Reference: https://discord.com/developers/docs/components/reference#file
type FileComponent struct {
	ComponentFields

	// File is an unfurled media item that only supports attachment references using the attachment://filename syntax.
	File UnfurledMediaItem `json:"file"`

	// Spoiler is whether the media should be a spoiler (or blurred out). Defaults to false.
	Spoiler bool `json:"spoiler,omitempty"`

	// Name is the name of the file. This field is ignored and provided by the API as part of the response.
	Name string `json:"name,omitempty"`

	// Size is the size of the file in bytes. This field is ignored and provided by the API as part of the response.
	Size int `json:"size,omitempty"`
}

func (FileComponent) layoutComponent()       {}
func (FileComponent) containerSubComponent() {}

var _ json.Marshaler = (*FileComponent)(nil)

func (c *FileComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeFile
	type NoMethod FileComponent
	return json.Marshal((*NoMethod)(c))
}

type SeperatorComponentSpacing int

const (
	SeperatorComponentSpacingSmall SeperatorComponentSpacing = 1 + iota
	SeperatorComponentSpacingLarge
)

// Is returns true if this SeperatorComponentSpacing matches the provided one.
func (v SeperatorComponentSpacing) Is(other SeperatorComponentSpacing) bool {
	return v == other
}

// SeparatorComponent is a top-level layout component that adds vertical padding and an optional visual divider between other components.
//
// It is used to create spacing or visual separation in messages.
// Separators require the IS_COMPONENTS_V2 message flag (1 << 15) to be set when sending the message.
//
// Note:
//   - Only available in messages.
//   - Requires the IS_COMPONENTS_V2 message flag (1 << 15).
//   - The divider field defaults to true, indicating whether a visual divider is displayed.
//   - The spacing field defaults to 1 (small padding), with 2 indicating large padding.
//
// Reference: https://discord.com/developers/docs/components/reference#separator
type SeparatorComponent struct {
	ComponentFields

	// Divider indicates whether a visual divider line should be displayed.
	//
	// Note:
	//   - Defaults to true.
	Divider bool `json:"divider,omitempty"`

	// Spacing determines the size of the vertical padding.
	//
	// Note:
	//   - 1 for small padding, 2 for large padding.
	//   - Defaults to 1.
	Spacing SeperatorComponentSpacing `json:"spacing,omitempty"`
}

func (SeparatorComponent) layoutComponent()       {}
func (SeparatorComponent) containerSubComponent() {}

var _ json.Marshaler = (*SeparatorComponent)(nil)

func (c *SeparatorComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeSeparator
	type NoMethod SeparatorComponent
	return json.Marshal((*NoMethod)(c))
}

// ContainerComponent is a top-level layout component that visually encapsulates a collection of child components with an optional customizable accent color bar.
//
// It is used to group components in messages, providing a visual container with an optional colored accent.
// Containers require the IS_COMPONENTS_V2 message flag (1 << 15) to be set when sending the message.
//
// Note:
//   - Only available in messages.
//   - Requires the IS_COMPONENTS_V2 message flag (1 << 15).
//   - The accent_color is an optional RGB color value (0x000000 to 0xFFFFFF).
//   - The spoiler field defaults to false, indicating whether the container content is blurred out.
//
// Reference: https://discord.com/developers/docs/components/reference#container
type ContainerComponent struct {
	ComponentFields

	// Components is an array of child components encapsulated within the container.
	//
	// Valid components:
	//   - ActionRowComponent
	//   - TextDisplayComponent
	//   - SectionComponent
	//   - MediaGalleryComponent
	//   - SeparatorComponent
	//   - FileComponent
	Components []ContainerSubComponent `json:"components"`

	// AccentColor is an optional RGB color for the accent bar on the container.
	//
	// Note:
	//   - Represented as an integer (0x000000 to 0xFFFFFF).
	AccentColor Color `json:"accent_color,omitempty"`

	// Spoiler indicates whether the container content should be blurred out as a spoiler.
	//
	// Note:
	//   - Defaults to false.
	Spoiler bool `json:"spoiler,omitempty"`
}

func (ContainerComponent) layoutComponent() {}

var _ json.Unmarshaler = (*ContainerComponent)(nil)

func (c *ContainerComponent) UnmarshalJSON(buf []byte) error {
	type tempContainerComponent struct {
		ComponentFields
		Components  []json.RawMessage `json:"components"`
		AccentColor Color             `json:"accent_color,omitempty"`
		Spoiler     bool              `json:"spoiler,omitempty"`
	}

	var temp tempContainerComponent
	if err := json.Unmarshal(buf, &temp); err != nil {
		return err
	}

	c.ComponentFields = temp.ComponentFields
	c.AccentColor = temp.AccentColor
	c.Spoiler = temp.Spoiler

	if temp.Components != nil {
		c.Components = make([]ContainerSubComponent, 0, len(temp.Components))
		for i := range len(temp.Components) {
			if len(temp.Components[i]) == 0 || bytes.Equal(temp.Components[i], []byte("null")) {
				continue
			}
			component, err := UnmarshalComponent(temp.Components[i])
			if err != nil {
				return err
			}
			if sub, ok := component.(ContainerSubComponent); ok {
				c.Components = append(c.Components, sub)
			} else {
				return errors.New(
					"cannot unmarshal non-ContainerSubComponent into ContainerSubComponent",
				)
			}
		}
	}

	return nil
}

var _ json.Marshaler = (*ContainerComponent)(nil)

func (c *ContainerComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeContainer
	type NoMethod ContainerComponent
	return json.Marshal((*NoMethod)(c))
}

// LabelComponent is a top-level layout component that wraps modal components with a label and an optional description.
//
// It is used to provide context to a single child component (e.g., TextInputComponent or StringSelectMenuComponent) in modals.
// The description may appear above or below the component depending on the platform.
//
// Note:
//   - Only available in modals.
//   - Supports only one child component.
//   - The label field is required and must not exceed 45 characters.
//   - The description field is optional and must not exceed 100 characters.
//
// Reference: https://discord.com/developers/docs/components/reference#label
type LabelComponent struct {
	ComponentFields

	// Label is the text displayed as the label for the component.
	//
	// Note:
	//   - Required.
	//   - Maximum of 45 characters.
	Label string `json:"label"`

	// Description is an optional text providing additional context for the component.
	//
	// Note:
	//   - Maximum of 100 characters.
	//   - May appear above or below the component depending on the platform.
	Description string `json:"description,omitempty"`

	// Components is an array containing a single child component.
	//
	// Valid components:
	//   - TextInputComponent
	//   - StringSelectMenuComponent
	Components []LabelSubComponent `json:"components"`
}

func (LabelComponent) layoutComponent() {}

var _ json.Unmarshaler = (*LabelComponent)(nil)

func (c *LabelComponent) UnmarshalJSON(buf []byte) error {
	type tempLabelComponent struct {
		ComponentFields
		Label       string            `json:"label"`
		Description string            `json:"description,omitempty"`
		Components  []json.RawMessage `json:"components"`
	}

	var temp tempLabelComponent
	if err := json.Unmarshal(buf, &temp); err != nil {
		return err
	}

	c.ComponentFields = temp.ComponentFields
	c.Label = temp.Label
	c.Description = temp.Description

	if temp.Components != nil {
		c.Components = make([]LabelSubComponent, 0, len(temp.Components))
		for i := range len(temp.Components) {
			if len(temp.Components[i]) == 0 || bytes.Equal(temp.Components[i], []byte("null")) {
				continue
			}
			component, err := UnmarshalComponent(temp.Components[i])
			if err != nil {
				return err
			}
			if sub, ok := component.(LabelSubComponent); ok {
				c.Components = append(c.Components, sub)
			} else {
				return errors.New("cannot unmarshal non-LabelSubComponent into LabelSubComponent")
			}
		}
	}

	return nil
}

var _ json.Marshaler = (*LabelComponent)(nil)

func (c *LabelComponent) MarshalJSON() ([]byte, error) {
	c.Type = ComponentTypeLabel
	type NoMethod LabelComponent
	return json.Marshal((*NoMethod)(c))
}

func UnmarshalComponent(buf []byte) (Component, error) {
	var meta struct {
		Type ComponentType `json:"type"`
	}
	if err := json.Unmarshal(buf, &meta); err != nil {
		return nil, err
	}

	switch meta.Type {
	case ComponentTypeActionRow:
		var c ActionRowComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeButton:
		var c ButtonComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeStringSelect:
		var c StringSelectMenuComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeTextInput:
		var c TextInputComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeUserSelect:
		var c UserSelectMenuComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeRoleSelect:
		var c RoleSelectMenuComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeMentionableSelect:
		var c MentionableSelectMenuComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeChannelSelect:
		var c ChannelSelectMenuComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeSection:
		var c SectionComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeTextDisplay:
		var c TextDisplayComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeThumbnail:
		var c ThumbnailComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeMediaGallery:
		var c MediaGalleryComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeFile:
		var c FileComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeSeparator:
		var c SeparatorComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeContainer:
		var c ContainerComponent
		return &c, json.Unmarshal(buf, &c)
	case ComponentTypeLabel:
		var c LabelComponent
		return &c, json.Unmarshal(buf, &c)
	default:
		return nil, errors.New("unknown component type")
	}
}

/////////////

// ButtonBuilder helps build a ButtonComponent with chainable methods.
type ActionRowBuilder struct {
	actionRow ActionRowComponent
}

// NewActionRowBuilder creates a new ActionRowBuilder instance.
func NewActionRowBuilder() *ActionRowBuilder {
	actionRowBuilder := &ActionRowBuilder{}
	actionRowBuilder.actionRow.Type = ComponentTypeActionRow
	return actionRowBuilder
}

// SetComponent sets the action row components.
func (b *ActionRowBuilder) SetComponent(components ...InteractiveComponent) *ActionRowBuilder {
	if len(components) > ActionRowMaxComponents {
		panic(
			"ActionRow can contain a maximum of " + strconv.Itoa(
				ActionRowMaxComponents,
			) + " components, got " + strconv.Itoa(
				len(components),
			),
		)
	}
	b.actionRow.Components = components
	return b
}

// AddComponent add's a component to the action row components.
func (b *ActionRowBuilder) AddComponent(component InteractiveComponent) *ActionRowBuilder {
	if len(b.actionRow.Components) >= ActionRowMaxComponents {
		panic(
			"ActionRow can contain a maximum of " + strconv.Itoa(
				ActionRowMaxComponents,
			) + " components",
		)
	}
	b.actionRow.Components = append(b.actionRow.Components, component)
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *ActionRowBuilder) Reset() {
	b.actionRow = ActionRowComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeActionRow},
		Components:      nil,
	}
}

// Build returns the final ActionRowComponent.
func (b *ActionRowBuilder) Build() *ActionRowComponent {
	return &b.actionRow
}

// ButtonBuilder helps build a ButtonComponent with chainable methods.
type ButtonBuilder struct {
	button ButtonComponent
}

// NewButtonBuilder creates a new ButtonBuilder instance.
func NewButtonBuilder() *ButtonBuilder {
	buttonBuilder := &ButtonBuilder{}
	buttonBuilder.button.Type = ComponentTypeButton
	return buttonBuilder
}

// SetLabel sets the button label.
func (b *ButtonBuilder) SetLabel(label string) *ButtonBuilder {
	b.button.Label = label
	return b
}

// SetStyle sets the button style.
func (b *ButtonBuilder) SetStyle(style ButtonStyle) *ButtonBuilder {
	b.button.Style = style
	return b
}

// SetCustomID sets the button custom ID.
func (b *ButtonBuilder) SetCustomID(customID string) *ButtonBuilder {
	b.button.CustomID = customID
	return b
}

// SetURL sets the button URL.
func (b *ButtonBuilder) SetURL(url string) *ButtonBuilder {
	b.button.URL = url
	return b
}

// SetSkuID sets the button Sku ID.
func (b *ButtonBuilder) SetSkuID(skuID Snowflake) *ButtonBuilder {
	b.button.SkuID = skuID
	return b
}

// SetDisabled sets the button disabled state.
func (b *ButtonBuilder) SetDisabled(disabled bool) *ButtonBuilder {
	b.button.Disabled = disabled
	return b
}

// Enable enables the button.
func (b *ButtonBuilder) Enable() *ButtonBuilder {
	b.button.Disabled = false
	return b
}

// Disable disables the button.
func (b *ButtonBuilder) Disable() *ButtonBuilder {
	b.button.Disabled = true
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *ButtonBuilder) Reset() {
	b.button = ButtonComponent{
		InteractiveComponentFields: InteractiveComponentFields{
			ComponentFields: ComponentFields{Type: ComponentTypeButton},
		},
	}
}

// Build returns the final ButtonComponent.
func (b *ButtonBuilder) Build() *ButtonComponent {
	return &b.button
}

// StringSelectMenuBuilder helps build a StringSelectMenuComponent with chainable methods.
type StringSelectMenuBuilder struct {
	selectMenu StringSelectMenuComponent
}

// NewStringSelectMenuBuilder creates a new StringSelectMenuBuilder instance.
func NewStringSelectMenuBuilder() *StringSelectMenuBuilder {
	builder := &StringSelectMenuBuilder{}
	builder.selectMenu.Type = ComponentTypeStringSelect
	return builder
}

// SetCustomID sets the custom ID.
func (b *StringSelectMenuBuilder) SetCustomID(customID string) *StringSelectMenuBuilder {
	b.selectMenu.CustomID = customID
	return b
}

// SetPlaceholder sets the placeholder text.
func (b *StringSelectMenuBuilder) SetPlaceholder(placeholder string) *StringSelectMenuBuilder {
	b.selectMenu.Placeholder = placeholder
	return b
}

// SetOptions sets all options at once.
func (b *StringSelectMenuBuilder) SetOptions(
	options []SelectOptionStructure,
) *StringSelectMenuBuilder {
	if len(options) > StringSelectMenuMaxOptions {
		panic(
			"StringSelectMenu can contain a maximum of " + strconv.Itoa(
				StringSelectMenuMaxOptions,
			) + " options, got " + strconv.Itoa(
				len(options),
			),
		)
	}
	b.selectMenu.Options = options
	return b
}

// AddOption adds a single option.
func (b *StringSelectMenuBuilder) AddOption(option SelectOptionStructure) *StringSelectMenuBuilder {
	if len(b.selectMenu.Options) >= StringSelectMenuMaxOptions {
		panic(
			"StringSelectMenu can contain a maximum of " + strconv.Itoa(
				StringSelectMenuMaxOptions,
			) + " options",
		)
	}
	b.selectMenu.Options = append(b.selectMenu.Options, option)
	return b
}

// SetMinValues sets the minimum number of selections.
func (b *StringSelectMenuBuilder) SetMinValues(minValues *int) *StringSelectMenuBuilder {
	b.selectMenu.MinValues = minValues
	return b
}

// SetMaxValues sets the maximum number of selections.
func (b *StringSelectMenuBuilder) SetMaxValues(maxValues int) *StringSelectMenuBuilder {
	b.selectMenu.MaxValues = maxValues
	return b
}

// SetRequired sets whether the select menu is required.
func (b *StringSelectMenuBuilder) SetRequired(required bool) *StringSelectMenuBuilder {
	b.selectMenu.Required = required
	return b
}

// SetDisabled sets whether the select menu is disabled.
func (b *StringSelectMenuBuilder) SetDisabled(disabled bool) *StringSelectMenuBuilder {
	b.selectMenu.Disabled = disabled
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *StringSelectMenuBuilder) Reset() {
	b.selectMenu = StringSelectMenuComponent{
		InteractiveComponentFields: InteractiveComponentFields{
			ComponentFields: ComponentFields{Type: ComponentTypeStringSelect},
		},
	}
}

// Build returns the final StringSelectMenuComponent.
func (b *StringSelectMenuBuilder) Build() *StringSelectMenuComponent {
	return &b.selectMenu
}

// TextInputBuilder helps build a TextInputComponent with chainable methods.
type TextInputBuilder struct {
	textInput TextInputComponent
}

// NewTextInputBuilder creates a new TextInputBuilder instance.
func NewTextInputBuilder() *TextInputBuilder {
	builder := &TextInputBuilder{}
	builder.textInput.Type = ComponentTypeTextInput
	return builder
}

// SetCustomID sets the custom ID.
func (b *TextInputBuilder) SetCustomID(customID string) *TextInputBuilder {
	b.textInput.CustomID = customID
	return b
}

// SetStyle sets the text input style.
func (b *TextInputBuilder) SetStyle(style TextInputStyle) *TextInputBuilder {
	b.textInput.Style = style
	return b
}

// SetMinLength sets the minimum input length.
func (b *TextInputBuilder) SetMinLength(minLength *int) *TextInputBuilder {
	b.textInput.MinLength = minLength
	return b
}

// SetMaxLength sets the maximum input length.
func (b *TextInputBuilder) SetMaxLength(maxLength int) *TextInputBuilder {
	b.textInput.MaxLength = maxLength
	return b
}

// SetValue sets the pre-filled value.
func (b *TextInputBuilder) SetValue(value string) *TextInputBuilder {
	b.textInput.Value = value
	return b
}

// SetPlaceholder sets the placeholder text.
func (b *TextInputBuilder) SetPlaceholder(placeholder string) *TextInputBuilder {
	b.textInput.Placeholder = placeholder
	return b
}

// SetRequired sets whether the input is required.
func (b *TextInputBuilder) SetRequired(required bool) *TextInputBuilder {
	b.textInput.Required = required
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *TextInputBuilder) Reset() {
	b.textInput = TextInputComponent{
		InteractiveComponentFields: InteractiveComponentFields{
			ComponentFields: ComponentFields{Type: ComponentTypeTextInput},
		},
	}
}

// Build returns the final TextInputComponent.
func (b *TextInputBuilder) Build() *TextInputComponent {
	return &b.textInput
}

// UserSelectMenuBuilder helps build a UserSelectMenuComponent with chainable methods.
type UserSelectMenuBuilder struct {
	selectMenu UserSelectMenuComponent
}

// NewUserSelectMenuBuilder creates a new UserSelectMenuBuilder instance.
func NewUserSelectMenuBuilder() *UserSelectMenuBuilder {
	builder := &UserSelectMenuBuilder{}
	builder.selectMenu.Type = ComponentTypeUserSelect
	return builder
}

// SetCustomID sets the custom ID.
func (b *UserSelectMenuBuilder) SetCustomID(customID string) *UserSelectMenuBuilder {
	b.selectMenu.CustomID = customID
	return b
}

// SetPlaceholder sets the placeholder text.
func (b *UserSelectMenuBuilder) SetPlaceholder(placeholder string) *UserSelectMenuBuilder {
	b.selectMenu.Placeholder = placeholder
	return b
}

// SetDefaultValues sets all default values at once.
func (b *UserSelectMenuBuilder) SetDefaultValues(
	defaults []SelectDefaultValue,
) *UserSelectMenuBuilder {
	b.selectMenu.DefaultValues = defaults
	return b
}

// AddDefaultValue adds a single default value.
func (b *UserSelectMenuBuilder) AddDefaultValue(
	defaultValue SelectDefaultValue,
) *UserSelectMenuBuilder {
	b.selectMenu.DefaultValues = append(b.selectMenu.DefaultValues, defaultValue)
	return b
}

// SetMinValues sets the minimum number of selections.
func (b *UserSelectMenuBuilder) SetMinValues(minValues *int) *UserSelectMenuBuilder {
	b.selectMenu.MinValues = minValues
	return b
}

// SetMaxValues sets the maximum number of selections.
func (b *UserSelectMenuBuilder) SetMaxValues(maxValues int) *UserSelectMenuBuilder {
	b.selectMenu.MaxValues = maxValues
	return b
}

// SetDisabled sets whether the select menu is disabled.
func (b *UserSelectMenuBuilder) SetDisabled(disabled bool) *UserSelectMenuBuilder {
	b.selectMenu.Disabled = disabled
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *UserSelectMenuBuilder) Reset() {
	b.selectMenu = UserSelectMenuComponent{
		InteractiveComponentFields: InteractiveComponentFields{
			ComponentFields: ComponentFields{Type: ComponentTypeUserSelect},
		},
	}
}

// Build returns the final UserSelectMenuComponent.
func (b *UserSelectMenuBuilder) Build() *UserSelectMenuComponent {
	return &b.selectMenu
}

// RoleSelectMenuBuilder helps build a RoleSelectMenuComponent with chainable methods.
type RoleSelectMenuBuilder struct {
	selectMenu RoleSelectMenuComponent
}

// NewRoleSelectMenuBuilder creates a new RoleSelectMenuBuilder instance.
func NewRoleSelectMenuBuilder() *RoleSelectMenuBuilder {
	builder := &RoleSelectMenuBuilder{}
	builder.selectMenu.Type = ComponentTypeRoleSelect
	return builder
}

// SetCustomID sets the custom ID.
func (b *RoleSelectMenuBuilder) SetCustomID(customID string) *RoleSelectMenuBuilder {
	b.selectMenu.CustomID = customID
	return b
}

// SetPlaceholder sets the placeholder text.
func (b *RoleSelectMenuBuilder) SetPlaceholder(placeholder string) *RoleSelectMenuBuilder {
	b.selectMenu.Placeholder = placeholder
	return b
}

// SetDefaultValues sets all default values at once.
func (b *RoleSelectMenuBuilder) SetDefaultValues(
	defaults []SelectDefaultValue,
) *RoleSelectMenuBuilder {
	b.selectMenu.DefaultValues = defaults
	return b
}

// AddDefaultValue adds a single default value.
func (b *RoleSelectMenuBuilder) AddDefaultValue(
	defaultValue SelectDefaultValue,
) *RoleSelectMenuBuilder {
	b.selectMenu.DefaultValues = append(b.selectMenu.DefaultValues, defaultValue)
	return b
}

// SetMinValues sets the minimum number of selections.
func (b *RoleSelectMenuBuilder) SetMinValues(minValues *int) *RoleSelectMenuBuilder {
	b.selectMenu.MinValues = minValues
	return b
}

// SetMaxValues sets the maximum number of selections.
func (b *RoleSelectMenuBuilder) SetMaxValues(maxValues int) *RoleSelectMenuBuilder {
	b.selectMenu.MaxValues = maxValues
	return b
}

// SetDisabled sets whether the select menu is disabled.
func (b *RoleSelectMenuBuilder) SetDisabled(disabled bool) *RoleSelectMenuBuilder {
	b.selectMenu.Disabled = disabled
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *RoleSelectMenuBuilder) Reset() {
	b.selectMenu = RoleSelectMenuComponent{
		InteractiveComponentFields: InteractiveComponentFields{
			ComponentFields: ComponentFields{Type: ComponentTypeRoleSelect},
		},
	}
}

// Build returns the final RoleSelectMenuComponent.
func (b *RoleSelectMenuBuilder) Build() *RoleSelectMenuComponent {
	return &b.selectMenu
}

// MentionableSelectMenuBuilder helps build a MentionableSelectMenuComponent with chainable methods.
type MentionableSelectMenuBuilder struct {
	selectMenu MentionableSelectMenuComponent
}

// NewMentionableSelectMenuBuilder creates a new MentionableSelectMenuBuilder instance.
func NewMentionableSelectMenuBuilder() *MentionableSelectMenuBuilder {
	builder := &MentionableSelectMenuBuilder{}
	builder.selectMenu.Type = ComponentTypeMentionableSelect
	return builder
}

// SetCustomID sets the custom ID.
func (b *MentionableSelectMenuBuilder) SetCustomID(customID string) *MentionableSelectMenuBuilder {
	b.selectMenu.CustomID = customID
	return b
}

// SetPlaceholder sets the placeholder text.
func (b *MentionableSelectMenuBuilder) SetPlaceholder(
	placeholder string,
) *MentionableSelectMenuBuilder {
	b.selectMenu.Placeholder = placeholder
	return b
}

// SetDefaultValues sets all default values at once.
func (b *MentionableSelectMenuBuilder) SetDefaultValues(
	defaults []SelectDefaultValue,
) *MentionableSelectMenuBuilder {
	b.selectMenu.DefaultValues = defaults
	return b
}

// AddDefaultValue adds a single default value.
func (b *MentionableSelectMenuBuilder) AddDefaultValue(
	defaultValue SelectDefaultValue,
) *MentionableSelectMenuBuilder {
	b.selectMenu.DefaultValues = append(b.selectMenu.DefaultValues, defaultValue)
	return b
}

// SetMinValues sets the minimum number of selections.
func (b *MentionableSelectMenuBuilder) SetMinValues(minValues *int) *MentionableSelectMenuBuilder {
	b.selectMenu.MinValues = minValues
	return b
}

// SetMaxValues sets the maximum number of selections.
func (b *MentionableSelectMenuBuilder) SetMaxValues(maxValues int) *MentionableSelectMenuBuilder {
	b.selectMenu.MaxValues = maxValues
	return b
}

// SetDisabled sets whether the select menu is disabled.
func (b *MentionableSelectMenuBuilder) SetDisabled(disabled bool) *MentionableSelectMenuBuilder {
	b.selectMenu.Disabled = disabled
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *MentionableSelectMenuBuilder) Reset() {
	b.selectMenu = MentionableSelectMenuComponent{
		InteractiveComponentFields: InteractiveComponentFields{
			ComponentFields: ComponentFields{Type: ComponentTypeMentionableSelect},
		},
	}
}

// Build returns the final MentionableSelectMenuComponent.
func (b *MentionableSelectMenuBuilder) Build() *MentionableSelectMenuComponent {
	return &b.selectMenu
}

// ChannelSelectMenuBuilder helps build a ChannelSelectMenuComponent with chainable methods.
type ChannelSelectMenuBuilder struct {
	selectMenu ChannelSelectMenuComponent
}

// NewChannelSelectMenuBuilder creates a new ChannelSelectMenuBuilder instance.
func NewChannelSelectMenuBuilder() *ChannelSelectMenuBuilder {
	builder := &ChannelSelectMenuBuilder{}
	builder.selectMenu.Type = ComponentTypeChannelSelect
	return builder
}

// SetCustomID sets the custom ID.
func (b *ChannelSelectMenuBuilder) SetCustomID(customID string) *ChannelSelectMenuBuilder {
	b.selectMenu.CustomID = customID
	return b
}

// SetChannelTypes sets all channel types at once.
func (b *ChannelSelectMenuBuilder) SetChannelTypes(types []ChannelType) *ChannelSelectMenuBuilder {
	b.selectMenu.ChannelTypes = types
	return b
}

// AddChannelType adds a single channel type.
func (b *ChannelSelectMenuBuilder) AddChannelType(
	channelType ChannelType,
) *ChannelSelectMenuBuilder {
	b.selectMenu.ChannelTypes = append(b.selectMenu.ChannelTypes, channelType)
	return b
}

// SetPlaceholder sets the placeholder text.
func (b *ChannelSelectMenuBuilder) SetPlaceholder(placeholder string) *ChannelSelectMenuBuilder {
	b.selectMenu.Placeholder = placeholder
	return b
}

// SetDefaultValues sets all default values at once.
func (b *ChannelSelectMenuBuilder) SetDefaultValues(
	defaults []SelectDefaultValue,
) *ChannelSelectMenuBuilder {
	b.selectMenu.DefaultValues = defaults
	return b
}

// AddDefaultValue adds a single default value.
func (b *ChannelSelectMenuBuilder) AddDefaultValue(
	defaultValue SelectDefaultValue,
) *ChannelSelectMenuBuilder {
	b.selectMenu.DefaultValues = append(b.selectMenu.DefaultValues, defaultValue)
	return b
}

// SetMinValues sets the minimum number of selections.
func (b *ChannelSelectMenuBuilder) SetMinValues(minValues *int) *ChannelSelectMenuBuilder {
	b.selectMenu.MinValues = minValues
	return b
}

// SetMaxValues sets the maximum number of selections.
func (b *ChannelSelectMenuBuilder) SetMaxValues(maxValues int) *ChannelSelectMenuBuilder {
	b.selectMenu.MaxValues = maxValues
	return b
}

// SetDisabled sets whether the select menu is disabled.
func (b *ChannelSelectMenuBuilder) SetDisabled(disabled bool) *ChannelSelectMenuBuilder {
	b.selectMenu.Disabled = disabled
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *ChannelSelectMenuBuilder) Reset() {
	b.selectMenu = ChannelSelectMenuComponent{
		InteractiveComponentFields: InteractiveComponentFields{
			ComponentFields: ComponentFields{Type: ComponentTypeChannelSelect},
		},
	}
}

// Build returns the final ChannelSelectMenuComponent.
func (b *ChannelSelectMenuBuilder) Build() *ChannelSelectMenuComponent {
	return &b.selectMenu
}

// SectionBuilder helps build a SectionComponent with chainable methods.
type SectionBuilder struct {
	section SectionComponent
}

// NewSectionBuilder creates a new SectionBuilder instance.
func NewSectionBuilder() *SectionBuilder {
	builder := &SectionBuilder{}
	builder.section.Type = ComponentTypeSection
	return builder
}

// SetID sets the component ID.
func (b *SectionBuilder) SetID(id int) *SectionBuilder {
	b.section.ID = id
	return b
}

// SetComponents sets all components at once.
func (b *SectionBuilder) SetComponents(components []SectionSubComponent) *SectionBuilder {
	if len(components) < SectionMinComponents || len(components) > SectionMaxComponents {
		panic(
			"Section must contain between " + strconv.Itoa(
				SectionMinComponents,
			) + " and " + strconv.Itoa(
				SectionMaxComponents,
			) + " components, got " + strconv.Itoa(
				len(components),
			),
		)
	}
	b.section.Components = components
	return b
}

// AddComponent adds a single component.
func (b *SectionBuilder) AddComponent(component SectionSubComponent) *SectionBuilder {
	if len(b.section.Components) >= SectionMaxComponents {
		panic(
			"Section can contain a maximum of " + strconv.Itoa(
				SectionMaxComponents,
			) + " components",
		)
	}
	b.section.Components = append(b.section.Components, component)
	return b
}

// SetAccessory sets the accessory component.
func (b *SectionBuilder) SetAccessory(accessory SectionAccessoryComponent) *SectionBuilder {
	b.section.Accessory = accessory
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *SectionBuilder) Reset() {
	b.section = SectionComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeSection},
		Components:      nil,
		Accessory:       nil,
	}
}

// Build returns the final SectionComponent.
func (b *SectionBuilder) Build() *SectionComponent {
	return &b.section
}

// TextDisplayBuilder helps build a TextDisplayComponent with chainable methods.
type TextDisplayBuilder struct {
	textDisplay TextDisplayComponent
}

// NewTextDisplayBuilder creates a new TextDisplayBuilder instance.
func NewTextDisplayBuilder() *TextDisplayBuilder {
	builder := &TextDisplayBuilder{}
	builder.textDisplay.Type = ComponentTypeTextDisplay
	return builder
}

// SetID sets the component ID.
func (b *TextDisplayBuilder) SetID(id int) *TextDisplayBuilder {
	b.textDisplay.ID = id
	return b
}

// SetContent sets the markdown content.
func (b *TextDisplayBuilder) SetContent(content string) *TextDisplayBuilder {
	b.textDisplay.Content = content
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *TextDisplayBuilder) Reset() {
	b.textDisplay = TextDisplayComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeTextDisplay},
	}
}

// Build returns the final TextDisplayComponent.
func (b *TextDisplayBuilder) Build() *TextDisplayComponent {
	return &b.textDisplay
}

// ThumbnailBuilder helps build a ThumbnailComponent with chainable methods.
type ThumbnailBuilder struct {
	thumbnail ThumbnailComponent
}

// NewThumbnailBuilder creates a new ThumbnailBuilder instance.
func NewThumbnailBuilder() *ThumbnailBuilder {
	builder := &ThumbnailBuilder{}
	builder.thumbnail.Type = ComponentTypeThumbnail
	return builder
}

// SetID sets the component ID.
func (b *ThumbnailBuilder) SetID(id int) *ThumbnailBuilder {
	b.thumbnail.ID = id
	return b
}

// SetMedia sets the media item.
func (b *ThumbnailBuilder) SetMedia(media UnfurledMediaItem) *ThumbnailBuilder {
	b.thumbnail.Media = media
	return b
}

// SetDescription sets the alt text description.
func (b *ThumbnailBuilder) SetDescription(description string) *ThumbnailBuilder {
	b.thumbnail.Description = description
	return b
}

// SetSpoiler sets whether the thumbnail is a spoiler.
func (b *ThumbnailBuilder) SetSpoiler(spoiler bool) *ThumbnailBuilder {
	b.thumbnail.Spoiler = spoiler
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *ThumbnailBuilder) Reset() {
	b.thumbnail = ThumbnailComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeThumbnail},
	}
}

// Build returns the final ThumbnailComponent.
func (b *ThumbnailBuilder) Build() *ThumbnailComponent {
	return &b.thumbnail
}

// MediaGalleryBuilder helps build a MediaGalleryComponent with chainable methods.
type MediaGalleryBuilder struct {
	gallery MediaGalleryComponent
}

// NewMediaGalleryBuilder creates a new MediaGalleryBuilder instance.
func NewMediaGalleryBuilder() *MediaGalleryBuilder {
	builder := &MediaGalleryBuilder{}
	builder.gallery.Type = ComponentTypeMediaGallery
	return builder
}

// SetID sets the component ID.
func (b *MediaGalleryBuilder) SetID(id int) *MediaGalleryBuilder {
	b.gallery.ID = id
	return b
}

// SetItems sets all gallery items at once.
func (b *MediaGalleryBuilder) SetItems(items []MediaGalleryItem) *MediaGalleryBuilder {
	if len(items) < MediaGalleryMinItems || len(items) > MediaGalleryMaxItems {
		panic(
			"MediaGallery must contain between " + strconv.Itoa(
				MediaGalleryMinItems,
			) + " and " + strconv.Itoa(
				MediaGalleryMaxItems,
			) + " items, got " + strconv.Itoa(
				len(items),
			),
		)
	}
	b.gallery.Items = items
	return b
}

// AddItem adds a single gallery item.
func (b *MediaGalleryBuilder) AddItem(item MediaGalleryItem) *MediaGalleryBuilder {
	if len(b.gallery.Items) >= MediaGalleryMaxItems {
		panic(
			"MediaGallery can contain a maximum of " + strconv.Itoa(
				MediaGalleryMaxItems,
			) + " items",
		)
	}
	b.gallery.Items = append(b.gallery.Items, item)
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *MediaGalleryBuilder) Reset() {
	b.gallery = MediaGalleryComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeMediaGallery},
		Items:           nil,
	}
}

// Build returns the final MediaGalleryComponent.
func (b *MediaGalleryBuilder) Build() *MediaGalleryComponent {
	return &b.gallery
}

// FileBuilder helps build a FileComponent with chainable methods.
type FileBuilder struct {
	file FileComponent
}

// NewFileBuilder creates a new FileBuilder instance.
func NewFileBuilder() *FileBuilder {
	builder := &FileBuilder{}
	builder.file.Type = ComponentTypeFile
	return builder
}

// SetID sets the component ID.
func (b *FileBuilder) SetID(id int) *FileBuilder {
	b.file.ID = id
	return b
}

// SetFile sets the file media item.
func (b *FileBuilder) SetFile(file UnfurledMediaItem) *FileBuilder {
	b.file.File = file
	return b
}

// SetSpoiler sets whether the file is a spoiler.
func (b *FileBuilder) SetSpoiler(spoiler bool) *FileBuilder {
	b.file.Spoiler = spoiler
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *FileBuilder) Reset() {
	b.file = FileComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeFile},
	}
}

// Build returns the final FileComponent.
func (b *FileBuilder) Build() *FileComponent {
	return &b.file
}

// SeparatorBuilder helps build a SeparatorComponent with chainable methods.
type SeparatorBuilder struct {
	separator SeparatorComponent
}

// NewSeparatorBuilder creates a new SeparatorBuilder instance.
func NewSeparatorBuilder() *SeparatorBuilder {
	builder := &SeparatorBuilder{}
	builder.separator.Type = ComponentTypeSeparator
	return builder
}

// SetID sets the component ID.
func (b *SeparatorBuilder) SetID(id int) *SeparatorBuilder {
	b.separator.ID = id
	return b
}

// SetDivider sets whether a visual divider is shown.
func (b *SeparatorBuilder) SetDivider(divider bool) *SeparatorBuilder {
	b.separator.Divider = divider
	return b
}

// SetSpacing sets the vertical padding size.
func (b *SeparatorBuilder) SetSpacing(spacing SeperatorComponentSpacing) *SeparatorBuilder {
	b.separator.Spacing = spacing
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *SeparatorBuilder) Reset() {
	b.separator = SeparatorComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeSeparator},
	}
}

// Build returns the final SeparatorComponent.
func (b *SeparatorBuilder) Build() *SeparatorComponent {
	return &b.separator
}

// ContainerBuilder helps build a ContainerComponent with chainable methods.
type ContainerBuilder struct {
	container ContainerComponent
}

// NewContainerBuilder creates a new ContainerBuilder instance.
func NewContainerBuilder() *ContainerBuilder {
	builder := &ContainerBuilder{}
	builder.container.Type = ComponentTypeContainer
	return builder
}

// SetID sets the component ID.
func (b *ContainerBuilder) SetID(id int) *ContainerBuilder {
	b.container.ID = id
	return b
}

// SetComponents sets all components at once.
func (b *ContainerBuilder) SetComponents(components []ContainerSubComponent) *ContainerBuilder {
	b.container.Components = components
	return b
}

// AddComponent adds a single component.
func (b *ContainerBuilder) AddComponent(component ContainerSubComponent) *ContainerBuilder {
	b.container.Components = append(b.container.Components, component)
	return b
}

// SetAccentColor sets the accent color.
func (b *ContainerBuilder) SetAccentColor(color Color) *ContainerBuilder {
	b.container.AccentColor = color
	return b
}

// SetSpoiler sets whether the container is a spoiler.
func (b *ContainerBuilder) SetSpoiler(spoiler bool) *ContainerBuilder {
	b.container.Spoiler = spoiler
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *ContainerBuilder) Reset() {
	b.container = ContainerComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeContainer},
		Components:      nil,
	}
}

// Build returns the final ContainerComponent.
func (b *ContainerBuilder) Build() *ContainerComponent {
	return &b.container
}

// LabelBuilder helps build a LabelComponent with chainable methods.
type LabelBuilder struct {
	label LabelComponent
}

// NewLabelBuilder creates a new LabelBuilder instance.
func NewLabelBuilder() *LabelBuilder {
	builder := &LabelBuilder{}
	builder.label.Type = ComponentTypeLabel
	return builder
}

// SetID sets the component ID.
func (b *LabelBuilder) SetID(id int) *LabelBuilder {
	b.label.ID = id
	return b
}

// SetLabel sets the label text.
func (b *LabelBuilder) SetLabel(label string) *LabelBuilder {
	b.label.Label = label
	return b
}

// SetDescription sets the description text.
func (b *LabelBuilder) SetDescription(description string) *LabelBuilder {
	b.label.Description = description
	return b
}

// SetComponents sets all components at once.
func (b *LabelBuilder) SetComponents(components []LabelSubComponent) *LabelBuilder {
	b.label.Components = components
	return b
}

// AddComponent adds a single component.
func (b *LabelBuilder) AddComponent(component LabelSubComponent) *LabelBuilder {
	b.label.Components = append(b.label.Components, component)
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *LabelBuilder) Reset() {
	b.label = LabelComponent{
		ComponentFields: ComponentFields{Type: ComponentTypeLabel},
		Components:      nil,
	}
}

// Build returns the final LabelComponent.
func (b *LabelBuilder) Build() *LabelComponent {
	return &b.label
}
