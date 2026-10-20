package wumpkit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnowflake_GuildShardID(t *testing.T) {
	guildID := Snowflake(41771983423143937)

	assert.Equal(t, 0, GuildShardID(guildID, 1))

	expectedShard1 := int((uint64(guildID) >> 22) % 2)
	assert.Equal(t, expectedShard1, GuildShardID(guildID, 2))
}

func TestSnowflake_Utilities(t *testing.T) {
	// parse Snowflake
	sf, err := ParseSnowflake("175928847299117063")
	require.NoError(t, err)
	assert.Equal(t, Snowflake(175928847299117063), sf)
	assert.Equal(t, "175928847299117063", sf.String())

	// invalid Snowflake
	_, err = ParseSnowflake("invalid")
	assert.Error(t, err)

	// timestamp
	ts := sf.Timestamp()
	assert.True(t, ts.After(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
}

func TestParseAppIDFromToken(t *testing.T) {
	// 18-digit snowflake (0 padding in base64: 175928847299117063 -> MTc1OTI4ODQ3Mjk5MTE3MDYz)
	token18 := "MTc1OTI4ODQ3Mjk5MTE3MDYz.Gb78_q.xY98-zabcdef12345"
	id18, err := ParseAppIDFromToken(token18)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(175928847299117063), id18)

	// with "Bot " prefix
	idWithPrefix, err := ParseAppIDFromToken("Bot " + token18)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(175928847299117063), idWithPrefix)

	// 19-digit snowflake with base64 padding (1759288472991170634 -> MTc1OTI4ODQ3Mjk5MTE3MDYzNA==)
	token19Padded := "Bearer MTc1OTI4ODQ3Mjk5MTE3MDYzNA==.Gb78_q.xY98-z"
	id19Padded, err := ParseAppIDFromToken(token19Padded)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(1759288472991170634), id19Padded)

	// 19-digit snowflake raw base64 without padding (1759288472991170634 -> MTc1OTI4ODQ3Mjk5MTE3MDYzNA)
	token19Raw := "MTc1OTI4ODQ3Mjk5MTE3MDYzNA.Gb78_q.xY98-z"
	id19Raw, err := ParseAppIDFromToken(token19Raw)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(1759288472991170634), id19Raw)

	// alias TokenAppID
	idAlias, err := TokenAppID(token18)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(175928847299117063), idAlias)

	// invalid tokens
	_, err = ParseAppIDFromToken("invalid-token")
	assert.Error(t, err)

	_, err = ParseAppIDFromToken("!!!.abc.def")
	assert.Error(t, err)
}
