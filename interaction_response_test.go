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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInteractionResponse_Helpers(t *testing.T) {
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/interactions/999/test-token/callback")

		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(
		context.Background(),
		WithToken("test-token"),
		WithRESTConfig(RequesterConfig{
			BaseURL: server.URL,
		}),
	)

	interaction := &ChatInputCommandInteraction{
		ApplicationCommandInteractionFields: ApplicationCommandInteractionFields{
			InteractionFields: InteractionFields{
				ID:            999,
				ApplicationID: 111,
				Token:         "test-token",
				Type:          InteractionTypeApplicationCommand,
			},
		},
	}

	ctx := context.Background()

	// 1. respondMessage
	err := client.RespondMessage(ctx, interaction, "Hello World", RespondMessageOptions{
		Ephemeral: true,
	})
	require.NoError(t, err)
	assert.Equal(t, float64(4), receivedBody["type"])

	data, ok := receivedBody["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Hello World", data["content"])
	assert.Equal(t, float64(MessageFlagEphemeral), data["flags"])

	// 2. deferResponse
	err = client.DeferResponse(ctx, interaction, false)
	require.NoError(t, err)
	assert.Equal(t, float64(5), receivedBody["type"])

	// 3. deferComponentUpdate
	err = client.DeferComponentUpdate(ctx, interaction)
	require.NoError(t, err)
	assert.Equal(t, float64(6), receivedBody["type"])
}

func TestInteraction_Author(t *testing.T) {
	// 1. guild interaction (Member populated)
	guildInteraction := &ChatInputCommandInteraction{
		ApplicationCommandInteractionFields: ApplicationCommandInteractionFields{
			Member: &InteractionMember{
				FullMember: FullMember{
					Member: Member{
						ID:       123456789,
						GuildID:  987654321,
						Nickname: "GuildNickname",
					},
					User: User{
						ID:            123456789,
						Username:      "GuildUser",
						Discriminator: "0001",
						GlobalName:    "GlobalName",
						Avatar:        "avatarhash",
						Bot:           true,
					},
				},
			},
		},
	}

	author := guildInteraction.Author()
	require.NotNil(t, author)
	assert.Equal(t, Snowflake(123456789), author.ID)
	assert.Equal(t, "GuildUser", author.Username)
	assert.Equal(t, "0001", author.Discriminator)
	assert.Equal(t, "GlobalName", author.DisplayName())
	assert.True(t, author.Bot)
	assert.Equal(t, "<@123456789>", author.Mention())
	assert.Contains(t, author.AvatarURL(), "avatarhash")

	// 2. DM interaction (User populated)
	dmInteraction := &ChatInputCommandInteraction{
		ApplicationCommandInteractionFields: ApplicationCommandInteractionFields{
			User: &User{
				ID:            555666777,
				Username:      "DMUser",
				Discriminator: "0002",
				GlobalName:    "DMGlobal",
				Bot:           false,
			},
		},
	}

	dmAuthor := dmInteraction.Author()
	require.NotNil(t, dmAuthor)
	assert.Equal(t, Snowflake(555666777), dmAuthor.ID)
	assert.Equal(t, "DMUser", dmAuthor.Username)
	assert.Equal(t, "0002", dmAuthor.Discriminator)
	assert.Equal(t, "DMGlobal", dmAuthor.DisplayName())
	assert.False(t, dmAuthor.Bot)
	assert.Equal(t, "<@555666777>", dmAuthor.Mention())

	// 3. empty interaction (Neither populated)
	emptyInteraction := &ChatInputCommandInteraction{}
	assert.Nil(t, emptyInteraction.Author())
}

func TestChatInputCommandInteractionData_Accessors(t *testing.T) {
	data := ChatInputCommandInteractionData{
		Options: []ChatInputInteractionCommandOption{
			{
				Name:  "target",
				Type:  ApplicationCommandOptionTypeUser,
				Value: []byte(`"111222"`),
			},
			{
				Name:  "role",
				Type:  ApplicationCommandOptionTypeRole,
				Value: []byte(`"333444"`),
			},
			{
				Name:  "channel",
				Type:  ApplicationCommandOptionTypeChannel,
				Value: []byte(`"555666"`),
			},
			{
				Name:  "doc",
				Type:  ApplicationCommandOptionTypeAttachment,
				Value: []byte(`"777888"`),
			},
			{
				Name:  "title",
				Type:  ApplicationCommandOptionTypeString,
				Value: []byte(`"MyTitle"`),
			},
			{
				Name:  "count",
				Type:  ApplicationCommandOptionTypeInteger,
				Value: []byte(`42`),
			},
			{
				Name:  "rating",
				Type:  ApplicationCommandOptionTypeFloat,
				Value: []byte(`4.5`),
			},
			{
				Name:  "active",
				Type:  ApplicationCommandOptionTypeBool,
				Value: []byte(`true`),
			},
		},
		Resolved: ChatInputCommandResolvedInteractionData{
			Users: map[Snowflake]User{
				111222: {ID: 111222, Username: "Alice"},
			},
			Members: map[Snowflake]PartialMember{
				111222: {Nickname: "Ali"},
			},
			Roles: map[Snowflake]Role{
				333444: {ID: 333444, Name: "Admin"},
			},
			Channels: map[Snowflake]ResolvedChannel{
				555666: {
					ChannelFields: ChannelFields{ID: 555666, Type: ChannelTypeGuildText},
					Name:          "general",
				},
			},
			Attachments: map[Snowflake]Attachment{
				777888: {ID: 777888, Filename: "test.pdf"},
			},
		},
	}

	u, ok := data.User("target")
	require.True(t, ok)
	assert.Equal(t, "Alice", u.Username)

	m, ok := data.Member("target")
	require.True(t, ok)
	assert.Equal(t, "Ali", m.Nickname)

	r, ok := data.Role("role")
	require.True(t, ok)
	assert.Equal(t, "Admin", r.Name)

	ch, ok := data.Channel("channel")
	require.True(t, ok)
	assert.Equal(t, "general", ch.Name)

	att, ok := data.Attachment("doc")
	require.True(t, ok)
	assert.Equal(t, "test.pdf", att.Filename)

	str, ok := data.String("title")
	require.True(t, ok)
	assert.Equal(t, "MyTitle", str)

	cnt, ok := data.Int("count")
	require.True(t, ok)
	assert.Equal(t, 42, cnt)

	rating, ok := data.Float("rating")
	require.True(t, ok)
	assert.Equal(t, 4.5, rating)

	b, ok := data.Bool("active")
	require.True(t, ok)
	assert.True(t, b)

	// non-existent option
	_, ok = data.User("unknown")
	assert.False(t, ok)
}

func TestChatInputCommandInteractionData_Subcommands(t *testing.T) {
	// 1. direct subcommand: /tag get name:foo
	directSub := ChatInputCommandInteractionData{
		Options: []ChatInputInteractionCommandOption{
			{
				Name: "get",
				Type: ApplicationCommandOptionTypeSubCommand,
				Options: []ChatInputInteractionCommandOption{
					{
						Name:  "name",
						Type:  ApplicationCommandOptionTypeString,
						Value: []byte(`"foo"`),
					},
				},
			},
		},
	}
	sub, ok := directSub.Subcommand()
	assert.True(t, ok)
	assert.Equal(t, "get", sub)
	group, ok := directSub.SubcommandGroup()
	assert.False(t, ok)
	assert.Empty(t, group)

	// 2. nested subcommand group: /settings logs set channel:#general
	groupSub := ChatInputCommandInteractionData{
		Options: []ChatInputInteractionCommandOption{
			{
				Name: "logs",
				Type: ApplicationCommandOptionTypeSubCommandGroup,
				Options: []ChatInputInteractionCommandOption{
					{
						Name: "set",
						Type: ApplicationCommandOptionTypeSubCommand,
						Options: []ChatInputInteractionCommandOption{
							{
								Name:  "channel",
								Type:  ApplicationCommandOptionTypeChannel,
								Value: []byte(`"123"`),
							},
						},
					},
				},
			},
		},
	}
	sub, ok = groupSub.Subcommand()
	assert.True(t, ok)
	assert.Equal(t, "set", sub)
	group, ok = groupSub.SubcommandGroup()
	assert.True(t, ok)
	assert.Equal(t, "logs", group)

	// 3. no subcommands (top-level command): /ping
	topLevel := ChatInputCommandInteractionData{
		Options: []ChatInputInteractionCommandOption{
			{
				Name:  "verbose",
				Type:  ApplicationCommandOptionTypeBool,
				Value: []byte(`true`),
			},
		},
	}
	sub, ok = topLevel.Subcommand()
	assert.False(t, ok)
	assert.Empty(t, sub)
	group, ok = topLevel.SubcommandGroup()
	assert.False(t, ok)
	assert.Empty(t, group)
}

func TestPartialMember_Helpers(t *testing.T) {
	future := time.Now().Add(10 * time.Minute)
	past := time.Now().Add(-10 * time.Minute)
	boostTime := time.Now().Add(-24 * time.Hour)

	pm := PartialMember{
		ID:                         2,
		GuildID:                    1,
		Nickname:                   "CoolNick",
		Avatar:                     "avatarhash",
		Banner:                     "bannerhash",
		RoleIDs:                    []Snowflake{100, 200, 300},
		PremiumSince:               &boostTime,
		Permissions:                PermissionAdministrator | PermissionManageMessages,
		CommunicationDisabledUntil: &future,
		AvatarDecorationData: &AvatarDecorationData{
			Asset: "decohash",
		},
	}

	assert.Equal(t, "CoolNick", pm.DisplayName())
	assert.True(t, pm.HasRole(200))
	assert.False(t, pm.HasRole(999))
	assert.True(t, pm.IsBoosting())
	assert.True(t, pm.IsTimedOut())
	assert.True(t, pm.HasPermission(PermissionAdministrator))
	assert.True(t, pm.HasPermission(PermissionManageMessages))
	// administrator implicitly grants all permissions
	assert.True(t, pm.HasPermission(PermissionBanMembers))

	// without Administrator:
	pmNormal := PartialMember{ID: 2, GuildID: 1, Permissions: PermissionManageMessages}
	assert.True(t, pmNormal.HasPermission(PermissionManageMessages))
	assert.False(t, pmNormal.HasPermission(PermissionBanMembers))

	assert.Contains(t, pm.AvatarURL(), "avatarhash")
	assert.Contains(t, pm.BannerURL(), "bannerhash")
	assert.Contains(t, pm.AvatarDecorationURL(), "decohash")

	// expired timeout
	pmPast := PartialMember{CommunicationDisabledUntil: &past}
	assert.False(t, pmPast.IsTimedOut())
}

func TestUnmarshalInteraction_HydratesMemberGuildIDAndID(t *testing.T) {
	raw := `{
		"id": "1000",
		"type": 2,
		"guild_id": "5000",
		"data": {
			"id": "2000",
			"name": "test",
			"type": 1,
			"resolved": {
				"members": {
					"3000": {
						"nick": "TargetUser"
					}
				}
			}
		},
		"member": {
			"nick": "Invoker"
		}
	}`

	in, err := UnmarshalInteraction([]byte(raw))
	require.NoError(t, err)
	chatIn, ok := in.(*ChatInputCommandInteraction)
	require.True(t, ok)

	require.NotNil(t, chatIn.Member)
	assert.Equal(t, Snowflake(5000), chatIn.Member.GuildID)

	targetMember, ok := chatIn.Data.Resolved.Members[Snowflake(3000)]
	require.True(t, ok)
	assert.Equal(t, Snowflake(3000), targetMember.ID)
	assert.Equal(t, Snowflake(5000), targetMember.GuildID)
}

func TestMember_Helpers(t *testing.T) {
	future := time.Now().Add(10 * time.Minute)
	boostTime := time.Now().Add(-24 * time.Hour)

	m := Member{
		ID:                         2,
		GuildID:                    1,
		Nickname:                   "Nick",
		Avatar:                     "avhash",
		Banner:                     "banhash",
		RoleIDs:                    []Snowflake{100, 200},
		PremiumSince:               &boostTime,
		CommunicationDisabledUntil: &future,
		AvatarDecorationData: &AvatarDecorationData{
			Asset: "decohash",
		},
	}

	assert.True(t, m.HasRole(100))
	assert.False(t, m.HasRole(300))
	assert.True(t, m.IsBoosting())
	assert.True(t, m.IsTimedOut())
	assert.Contains(t, m.AvatarURL(), "avhash")
	assert.Contains(t, m.BannerURL(), "banhash")
	assert.Contains(t, m.AvatarDecorationURL(), "decohash")
}

func TestChatInputInteractionCommandOption_GettersAndTypeMismatch(t *testing.T) {
	strOpt := ChatInputInteractionCommandOption{Name: "str", Value: []byte(`"hello"`)}
	assert.Equal(t, "hello", strOpt.String())

	intOpt := ChatInputInteractionCommandOption{Name: "num", Value: []byte(`42`)}
	assert.Equal(t, 42, intOpt.Int())

	floatOpt := ChatInputInteractionCommandOption{Name: "flt", Value: []byte(`3.14`)}
	assert.Equal(t, 3.14, floatOpt.Float())

	boolOpt := ChatInputInteractionCommandOption{Name: "b", Value: []byte(`true`)}
	assert.True(t, boolOpt.Bool())

	snowOpt := ChatInputInteractionCommandOption{Name: "id", Value: []byte(`"123456789"`)}
	assert.Equal(t, Snowflake(123456789), snowOpt.Snowflake())

	// typeMismatchError panics
	assert.Panics(t, func() {
		invalidStr := ChatInputInteractionCommandOption{Name: "invalid", Value: []byte(`{"not": "a string"}`)}
		_ = invalidStr.String()
	})

	assert.Panics(t, func() {
		invalidInt := ChatInputInteractionCommandOption{Name: "invalid", Value: []byte(`"not an int"`)}
		_ = invalidInt.Int()
	})

	assert.Panics(t, func() {
		invalidFloat := ChatInputInteractionCommandOption{Name: "invalid", Value: []byte(`"not a float"`)}
		_ = invalidFloat.Float()
	})

	assert.Panics(t, func() {
		invalidBool := ChatInputInteractionCommandOption{Name: "invalid", Value: []byte(`"not a bool"`)}
		_ = invalidBool.Bool()
	})

	assert.Panics(t, func() {
		invalidSnowflake := ChatInputInteractionCommandOption{Name: "invalid", Value: []byte(`"not-a-snowflake"`)}
		_ = invalidSnowflake.Snowflake()
	})

	// verify that panic value is *TypeMismatchError
	defer func() {
		r := recover()
		require.NotNil(t, r)
		tmErr, ok := r.(*TypeMismatchError)
		require.True(t, ok)
		assert.Equal(t, "str", tmErr.Option)
		assert.Equal(t, "int", tmErr.Expected)
		assert.Contains(t, tmErr.Error(), "type mismatch")
	}()

	_ = strOpt.Int()
}
