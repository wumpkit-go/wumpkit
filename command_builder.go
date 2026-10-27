/************************************************************************************
 *
 * wumpkit, A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD-3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package wumpkit

// NewStringChoice creates a string option choice.
func NewStringChoice(name, value string) ApplicationCommandOptionChoiceString {
	return ApplicationCommandOptionChoiceString{
		ChoiceOptionBase: ChoiceOptionBase{Name: name},
		Value:            value,
	}
}

// NewIntegerChoice creates an integer option choice.
func NewIntegerChoice(name string, value int) ApplicationCommandOptionChoiceInteger {
	return ApplicationCommandOptionChoiceInteger{
		ChoiceOptionBase: ChoiceOptionBase{Name: name},
		Value:            value,
	}
}

// NewFloatChoice creates a float option choice.
func NewFloatChoice(name string, value float64) ApplicationCommandOptionChoiceFloat {
	return ApplicationCommandOptionChoiceFloat{
		ChoiceOptionBase: ChoiceOptionBase{Name: name},
		Value:            value,
	}
}

// SlashCommandBuilder provides a fluent, chainable API for constructing ChatInputCommand objects.
type SlashCommandBuilder struct {
	cmd ChatInputCommand
}

// NewSlashCommandBuilder creates a new SlashCommandBuilder with the given name and description.
func NewSlashCommandBuilder(name, description string) *SlashCommandBuilder {
	b := &SlashCommandBuilder{}
	b.cmd.Type = ApplicationCommandTypeChatInput
	b.cmd.Name = name
	b.cmd.Description = description
	return b
}

// SetName sets the name of the slash command.
func (b *SlashCommandBuilder) SetName(name string) *SlashCommandBuilder {
	b.cmd.Name = name
	return b
}

// SetDescription sets the description of the slash command.
func (b *SlashCommandBuilder) SetDescription(description string) *SlashCommandBuilder {
	b.cmd.Description = description
	return b
}

// SetDefaultMemberPermissions sets the required permissions for members to run this command.
func (b *SlashCommandBuilder) SetDefaultMemberPermissions(perms Permissions) *SlashCommandBuilder {
	b.cmd.DefaultMemberPermissions = perms
	return b
}

// SetNSFW sets whether the command is age-restricted (NSFW).
func (b *SlashCommandBuilder) SetNSFW(nsfw bool) *SlashCommandBuilder {
	b.cmd.NSFW = nsfw
	return b
}

// SetIntegrationTypes sets the installation contexts where the command is available.
func (b *SlashCommandBuilder) SetIntegrationTypes(
	types ...ApplicationIntegrationType,
) *SlashCommandBuilder {
	b.cmd.IntegrationTypes = types
	return b
}

// SetContexts sets the interaction contexts where the command can be used.
func (b *SlashCommandBuilder) SetContexts(contexts ...InteractionContextType) *SlashCommandBuilder {
	b.cmd.Contexts = contexts
	return b
}

// AddStringOption adds a string option parameter.
func (b *SlashCommandBuilder) AddStringOption(
	name, description string,
	required bool,
	choices ...ApplicationCommandOptionChoiceString,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionString{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeString,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
		ChoiceFieldsString: ChoiceFieldsString{
			Choices: choices,
		},
	})
	return b
}

// AddIntegerOption adds an integer option parameter.
func (b *SlashCommandBuilder) AddIntegerOption(
	name, description string,
	required bool,
	choices ...ApplicationCommandOptionChoiceInteger,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionInteger{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeInteger,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
		ChoiceFieldsInteger: ChoiceFieldsInteger{
			Choices: choices,
		},
	})
	return b
}

// AddFloatOption adds a float (number) option parameter.
func (b *SlashCommandBuilder) AddFloatOption(
	name, description string,
	required bool,
	choices ...ApplicationCommandOptionChoiceFloat,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionFloat{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeFloat,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
		ChoiceFieldsFloat: ChoiceFieldsFloat{
			Choices: choices,
		},
	})
	return b
}

// AddBooleanOption adds a boolean (true/false) option parameter.
func (b *SlashCommandBuilder) AddBooleanOption(
	name, description string,
	required bool,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionBool{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeBool,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
	})
	return b
}

// AddUserOption adds a user option parameter.
func (b *SlashCommandBuilder) AddUserOption(
	name, description string,
	required bool,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionUser{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeUser,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
	})
	return b
}

// AddChannelOption adds a channel option parameter with optional channel type restrictions.
func (b *SlashCommandBuilder) AddChannelOption(
	name, description string,
	required bool,
	channelTypes ...ChannelType,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionChannel{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeChannel,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
		ChannelConstraints: ChannelConstraints{
			ChannelTypes: channelTypes,
		},
	})
	return b
}

// AddRoleOption adds a role option parameter.
func (b *SlashCommandBuilder) AddRoleOption(
	name, description string,
	required bool,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionRole{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeRole,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
	})
	return b
}

// AddMentionableOption adds a mentionable (user or role) option parameter.
func (b *SlashCommandBuilder) AddMentionableOption(
	name, description string,
	required bool,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionMentionable{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeMentionable,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
	})
	return b
}

// AddAttachmentOption adds an attachment (file upload) option parameter.
func (b *SlashCommandBuilder) AddAttachmentOption(
	name, description string,
	required bool,
) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &ApplicationCommandOptionAttachment{
		OptionBase: OptionBase{
			Type:        ApplicationCommandOptionTypeAttachment,
			Name:        name,
			Description: description,
		},
		RequiredBase: RequiredBase{
			Required: required,
		},
	})
	return b
}

// AddOption adds any custom ApplicationCommandOption directly.
func (b *SlashCommandBuilder) AddOption(opt ApplicationCommandOption) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, opt)
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *SlashCommandBuilder) Reset() {
	b.cmd = ChatInputCommand{
		ApplicationCommandBase: ApplicationCommandBase{
			Type: ApplicationCommandTypeChatInput,
		},
	}
}

// Build returns the constructed ChatInputCommand.
func (b *SlashCommandBuilder) Build() *ChatInputCommand {
	return &b.cmd
}

// UserCommandBuilder provides a fluent API for constructing user context menu commands.
type UserCommandBuilder struct {
	cmd ApplicationUserCommand
}

// NewUserCommandBuilder creates a new UserCommandBuilder.
func NewUserCommandBuilder(name string) *UserCommandBuilder {
	b := &UserCommandBuilder{}
	b.cmd.Type = ApplicationCommandTypeUser
	b.cmd.Name = name
	return b
}

// SetName sets the name of the user command.
func (b *UserCommandBuilder) SetName(name string) *UserCommandBuilder {
	b.cmd.Name = name
	return b
}

// SetDefaultMemberPermissions sets the required permissions for members to run this command.
func (b *UserCommandBuilder) SetDefaultMemberPermissions(perms Permissions) *UserCommandBuilder {
	b.cmd.DefaultMemberPermissions = perms
	return b
}

// SetIntegrationTypes sets the installation contexts where the command is available.
func (b *UserCommandBuilder) SetIntegrationTypes(
	types ...ApplicationIntegrationType,
) *UserCommandBuilder {
	b.cmd.IntegrationTypes = types
	return b
}

// SetContexts sets the interaction contexts where the command can be used.
func (b *UserCommandBuilder) SetContexts(contexts ...InteractionContextType) *UserCommandBuilder {
	b.cmd.Contexts = contexts
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *UserCommandBuilder) Reset() {
	b.cmd = ApplicationUserCommand{
		ApplicationCommandBase: ApplicationCommandBase{
			Type: ApplicationCommandTypeUser,
		},
	}
}

// Build returns the constructed ApplicationUserCommand.
func (b *UserCommandBuilder) Build() *ApplicationUserCommand {
	return &b.cmd
}

// MessageCommandBuilder provides a fluent API for constructing message context menu commands.
type MessageCommandBuilder struct {
	cmd ApplicationMessageCommand
}

// NewMessageCommandBuilder creates a new MessageCommandBuilder.
func NewMessageCommandBuilder(name string) *MessageCommandBuilder {
	b := &MessageCommandBuilder{}
	b.cmd.Type = ApplicationCommandTypeMessage
	b.cmd.Name = name
	return b
}

// SetName sets the name of the message command.
func (b *MessageCommandBuilder) SetName(name string) *MessageCommandBuilder {
	b.cmd.Name = name
	return b
}

// SetDefaultMemberPermissions sets the required permissions for members to run this command.
func (b *MessageCommandBuilder) SetDefaultMemberPermissions(
	perms Permissions,
) *MessageCommandBuilder {
	b.cmd.DefaultMemberPermissions = perms
	return b
}

// SetIntegrationTypes sets the installation contexts where the command is available.
func (b *MessageCommandBuilder) SetIntegrationTypes(
	types ...ApplicationIntegrationType,
) *MessageCommandBuilder {
	b.cmd.IntegrationTypes = types
	return b
}

// SetContexts sets the interaction contexts where the command can be used.
func (b *MessageCommandBuilder) SetContexts(
	contexts ...InteractionContextType,
) *MessageCommandBuilder {
	b.cmd.Contexts = contexts
	return b
}

// Reset clears the builder state, allowing it to be reused.
func (b *MessageCommandBuilder) Reset() {
	b.cmd = ApplicationMessageCommand{
		ApplicationCommandBase: ApplicationCommandBase{
			Type: ApplicationCommandTypeMessage,
		},
	}
}

// Build returns the constructed ApplicationMessageCommand.
func (b *MessageCommandBuilder) Build() *ApplicationMessageCommand {
	return &b.cmd
}
