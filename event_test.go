package wumpkit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvent_JSONUnmarshaling(t *testing.T) {
	// guildMemberAddEvent
	memberAddJSON := `{
		"guild_id": "111",
		"roles": ["222"],
		"user": {
			"id": "333",
			"username": "wumpkit_user",
			"discriminator": "0"
		}
	}`

	var gma GuildMemberAddEvent
	err := json.Unmarshal([]byte(memberAddJSON), &gma)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(111), gma.GuildID)
	assert.Equal(t, "wumpkit_user", gma.Member.User.Username)
	assert.Equal(t, Snowflake(111), gma.Member.GuildID)

	// threadCreateEvent
	threadJSON := `{
		"id": "555",
		"type": 11,
		"guild_id": "111",
		"name": "cool-thread",
		"newly_created": true
	}`
	var tc ThreadCreateEvent
	err = json.Unmarshal([]byte(threadJSON), &tc)
	require.NoError(t, err)
	require.NotNil(t, tc.Thread)
	assert.Equal(t, Snowflake(555), tc.Thread.ID)
	assert.Equal(t, "cool-thread", tc.Thread.Name)
	assert.True(t, tc.NewlyCreated)

	// channelUpdateEvent
	channelUpdateJSON := `{
		"id": "777",
		"type": 0,
		"name": "general-updated"
	}`
	var cu ChannelUpdateEvent
	err = json.Unmarshal([]byte(channelUpdateJSON), &cu)
	require.NoError(t, err)
	require.NotNil(t, cu.NewChannel)
	assert.Equal(t, Snowflake(777), cu.NewChannel.GetID())
	textCh, ok := cu.NewChannel.(*TextChannel)
	require.True(t, ok)
	assert.Equal(t, "general-updated", textCh.Name)

	// threadUpdateEvent
	threadUpdateJSON := `{
		"id": "555",
		"type": 11,
		"name": "thread-updated"
	}`
	var tu ThreadUpdateEvent
	err = json.Unmarshal([]byte(threadUpdateJSON), &tu)
	require.NoError(t, err)
	require.NotNil(t, tu.NewThread)
	assert.Equal(t, Snowflake(555), tu.NewThread.ID)
	assert.Equal(t, "thread-updated", tu.NewThread.Name)

	// guildRoleUpdateEvent
	roleUpdateJSON := `{
		"guild_id": "111",
		"role": {
			"id": "888",
			"name": "Admin-Updated",
			"color": 255
		}
	}`
	var ru GuildRoleUpdateEvent
	err = json.Unmarshal([]byte(roleUpdateJSON), &ru)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(111), ru.GuildID)
	assert.Equal(t, Snowflake(888), ru.NewRole.ID)
	assert.Equal(t, "Admin-Updated", ru.NewRole.Name)

	// rateLimitedEvent
	rateLimitedJSON := `{
		"opcode": 8,
		"retry_after": 2.5,
		"meta": {
			"guild_id": "111",
			"nonce": "test-nonce"
		}
	}`
	var rle RateLimitedEvent
	err = json.Unmarshal([]byte(rateLimitedJSON), &rle)
	require.NoError(t, err)
	assert.Equal(t, gatewayOpcode(8), rle.Opcode)
	assert.Equal(t, 2.5, rle.RetryAfter)
	assert.Equal(t, Snowflake(111), rle.Meta.GuildID)
	require.NotNil(t, rle.Meta.Nonce)
	assert.Equal(t, "test-nonce", *rle.Meta.Nonce)

	// lobbyMessageCreateEvent
	lobbyMsgJSON := `{
		"id": "1397729799727878254",
		"type": 0,
		"content": "welcome to the party!",
		"lobby_id": "1397729744753266719",
		"channel_id": "1397729744753266719",
		"author": {
			"id": "111",
			"username": "gamer"
		}
	}`
	var lmc LobbyMessageCreateEvent
	err = json.Unmarshal([]byte(lobbyMsgJSON), &lmc)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(1397729799727878254), lmc.ID)
	assert.Equal(t, Snowflake(1397729744753266719), lmc.LobbyID)
	assert.Equal(t, "welcome to the party!", lmc.Content)

	// gameDirectMessageCreateEvent
	gdmJSON := `{
		"id": "1405614357781545021",
		"type": 0,
		"content": "get in friend, we are raiding",
		"channel_id": "1405604229820715098",
		"author": {
			"id": "222",
			"username": "raider"
		}
	}`
	var gdm GameDirectMessageCreateEvent
	err = json.Unmarshal([]byte(gdmJSON), &gdm)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(1405614357781545021), gdm.ID)
	assert.Equal(t, Snowflake(1405604229820715098), gdm.ChannelID)
	assert.Equal(t, "get in friend, we are raiding", gdm.Content)
}
