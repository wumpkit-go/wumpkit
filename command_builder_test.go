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

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlashCommandBuilder(t *testing.T) {
	cmd := NewSlashCommandBuilder("ban", "Ban a user from the guild").
		SetDefaultMemberPermissions(PermissionBanMembers).
		SetNSFW(false).
		AddUserOption("target", "User to ban", true).
		AddStringOption(
			"reason", "Reason for the ban", false,
			NewStringChoice("Spam", "spam"),
			NewStringChoice("Harassment", "harassment"),
		).
		AddIntegerOption(
			"delete_messages_days", "Days of messages to delete", false,
			NewIntegerChoice("None", 0),
			NewIntegerChoice("1 Day", 1),
			NewIntegerChoice("7 Days", 7),
		).
		AddChannelOption("log_channel", "Channel to log the ban", false, ChannelTypeGuildText).
		AddRoleOption("role_filter", "Filter by role", false).
		AddMentionableOption("notify", "Entity to notify", false).
		AddBooleanOption("silent", "Whether to ban silently", false).
		AddAttachmentOption("proof", "Evidence screenshot or file", false).
		AddFloatOption(
			"score", "Confidence score", false,
			NewFloatChoice("High", 1.0),
			NewFloatChoice("Low", 0.5),
		).
		Build()

	require.NotNil(t, cmd)
	assert.Equal(t, "ban", cmd.Name)
	assert.Equal(t, "Ban a user from the guild", cmd.Description)
	assert.Equal(t, ApplicationCommandTypeChatInput, cmd.Type)
	assert.Equal(t, PermissionBanMembers, cmd.DefaultMemberPermissions)
	assert.Len(t, cmd.Options, 9)

	// verify JSON marshaling roundtrip
	data, err := json.Marshal(cmd)
	require.NoError(t, err)

	unmarshaled, err := UnmarshalApplicationCommand(data)
	require.NoError(t, err)
	chatCmd, ok := unmarshaled.(*ChatInputCommand)
	require.True(t, ok)
	assert.Equal(t, "ban", chatCmd.Name)
	assert.Len(t, chatCmd.Options, 9)
}

func TestContextMenuCommandBuilders(t *testing.T) {
	userCmd := NewUserCommandBuilder("User Info").
		SetDefaultMemberPermissions(PermissionModerateMembers).
		Build()

	require.NotNil(t, userCmd)
	assert.Equal(t, "User Info", userCmd.Name)
	assert.Equal(t, ApplicationCommandTypeUser, userCmd.Type)

	msgCmd := NewMessageCommandBuilder("Report Message").
		SetDefaultMemberPermissions(PermissionSendMessages).
		Build()

	require.NotNil(t, msgCmd)
	assert.Equal(t, "Report Message", msgCmd.Name)
	assert.Equal(t, ApplicationCommandTypeMessage, msgCmd.Type)
}
