package wumpkit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissions_Bitwise(t *testing.T) {
	var perms Permissions
	perms.Add(PermissionSendMessages, PermissionEmbedLinks)
	assert.True(t, perms.Has(PermissionSendMessages))
	assert.True(t, perms.Has(PermissionEmbedLinks))

	perms.Remove(PermissionEmbedLinks)
	assert.False(t, perms.Has(PermissionEmbedLinks))
	assert.True(t, perms.Has(PermissionSendMessages))
}

func TestPermissions_JSON(t *testing.T) {
	perms := PermissionSendMessages | PermissionEmbedLinks
	data, err := json.Marshal(perms)
	require.NoError(t, err)

	var unmarshaled Permissions
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, perms, unmarshaled)
}

func TestPermissions_Names(t *testing.T) {
	perms := PermissionAddReactions | PermissionBanMembers
	names := perms.Names()

	assert.Len(t, names, 2)
	assert.Contains(t, names, PermissionNameAddReactions)
	assert.Contains(t, names, PermissionNameBanMembers)
}

func TestPermissions_Stringer(t *testing.T) {
	assert.Equal(t, "None", PermissionNone.String())
	assert.Equal(t, "Administrator", PermissionAdministrator.String())
	assert.Equal(t, "SendMessages", PermissionSendMessages.String())
	assert.Equal(
		t,
		"BanMembers, AddReactions",
		(PermissionBanMembers | PermissionAddReactions).String(),
	)
	assert.Equal(t, "ViewAuditLog", PermissionNameViewAuditLog.String())
	assert.Equal(t, "SetVoiceChannelStatus", PermissionNameSetVoiceChannelStatus.String())
	assert.True(t, PermissionAll.Has(PermissionSetVoiceChannelStatus))
	assert.Equal(
		t,
		PermissionNameSetVoiceChannelStatus,
		PermissionNames[PermissionSetVoiceChannelStatus],
	)
}
