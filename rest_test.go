package wumpkit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClientWithServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	cfg := DefaultRequesterConfig()
	cfg.BaseURL = ts.URL
	c := New(context.Background(),
		WithToken("test_token_secret"),
		WithRequesterConfig(cfg),
	)
	return c, ts
}

func TestDiscordRestError_Parsing(t *testing.T) {
	// simulate 403 Forbidden with Discord JSON error code 50013
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{
			"code": 50013,
			"message": "Missing Permissions"
		}`))
	}))
	defer ts.Close()

	cfg := DefaultRequesterConfig()
	cfg.BaseURL = ts.URL
	c := New(context.Background(),
		WithToken("test_token"),
		WithRequesterConfig(cfg),
	)

	_, err := c.FetchChannel(context.Background(), 123456789)
	require.Error(t, err)

	var discordErr *DiscordRestError
	require.True(t, errors.As(err, &discordErr), "error must wrap DiscordRestError")
	assert.Equal(t, http.StatusForbidden, discordErr.StatusCode)
	assert.Equal(t, 50013, discordErr.Code)
	assert.Equal(t, "Missing Permissions", discordErr.Message)
	assert.Contains(t, discordErr.Error(), "code 50013")
}

func TestClient_UserREST(t *testing.T) {
	c, ts := newTestClientWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bot test_token_secret", r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/users/@me":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "10001",
				"username": "bot_user",
				"discriminator": "0001"
			}`))
		case "/users/10002":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "10002",
				"username": "target_user"
			}`))
		case "/users/@me/guilds":
			assert.Equal(t, "1", r.URL.Query().Get("shard"))
			assert.Equal(t, "true", r.URL.Query().Get("with_counts"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"id": "20001",
					"name": "Guild One",
					"approximate_member_count": 42
				}
			]`))
		default:
			http.NotFound(w, r)
		}
	})
	defer ts.Close()

	ctx := context.Background()

	// fetch current user
	me, err := c.FetchCurrentUser(ctx)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(10001), me.ID)
	assert.Equal(t, "bot_user", me.Username)

	// fetch user by ID
	user, err := c.FetchUser(ctx, 10002)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(10002), user.ID)
	assert.Equal(t, "target_user", user.Username)

	// fetch current user guilds with shard pagination
	shardID := 1
	guilds, err := c.FetchCurrentUserGuilds(ctx, FetchCurrentUserGuildsOptions{
		Shard:      &shardID,
		WithCounts: true,
	})
	require.NoError(t, err)
	require.Len(t, guilds, 1)
	assert.Equal(t, Snowflake(20001), guilds[0].ID)
	assert.Equal(t, "Guild One", guilds[0].Name)
}

func TestClient_ChannelAndMessageREST(t *testing.T) {
	c, ts := newTestClientWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/channels/30001":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "30001",
				"type": 0,
				"name": "general",
				"guild_id": "20001"
			}`))
		case r.Method == http.MethodPost && r.URL.Path == "/channels/30001/messages":
			var payload CreateMessageOptions
			err := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, err)
			assert.Equal(t, "hello wumpkit", payload.Content)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "40001",
				"channel_id": "30001",
				"content": "hello wumpkit"
			}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/30001/messages/40001":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/reactions/"):
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	})
	defer ts.Close()

	ctx := context.Background()

	// fetch channel
	ch, err := c.FetchChannel(ctx, 30001)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(30001), ch.GetID())

	// create message
	msg, err := c.CreateMessage(ctx, 30001, CreateMessageOptions{
		Content: "hello wumpkit",
	})
	require.NoError(t, err)
	assert.Equal(t, Snowflake(40001), msg.ID)
	assert.Equal(t, "hello wumpkit", msg.Content)

	// add reaction
	err = c.CreateReaction(ctx, 30001, 40001, "👍")
	require.NoError(t, err)

	// delete message
	err = c.DeleteMessage(ctx, 30001, 40001)
	require.NoError(t, err)
}

func TestClient_GuildREST(t *testing.T) {
	c, ts := newTestClientWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/20001":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "20001",
				"name": "Super Guild",
				"owner_id": "10001"
			}`))
		case r.Method == http.MethodPost && r.URL.Path == "/guilds/20001/channels":
			var opts CreateChannelOptions
			err := json.NewDecoder(r.Body).Decode(&opts)
			assert.NoError(t, err)
			assert.Equal(t, "announcements", opts.Name)

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{
				"id": "30002",
				"guild_id": "20001",
				"name": "announcements",
				"type": 0
			}`))
		default:
			http.NotFound(w, r)
		}
	})
	defer ts.Close()

	ctx := context.Background()

	guild, err := c.FetchGuild(ctx, 20001)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(20001), guild.ID)
	assert.Equal(t, "Super Guild", guild.Name)

	newChan, err := c.CreateChannel(ctx, 20001, CreateChannelOptions{
		Name: "announcements",
	})
	require.NoError(t, err)
	assert.Equal(t, Snowflake(30002), newChan.GetID())
}

func TestClient_WebhookREST(t *testing.T) {
	c, ts := newTestClientWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/channels/30001/webhooks":
			var opts CreateWebhookOptions
			err := json.NewDecoder(r.Body).Decode(&opts)
			assert.NoError(t, err)
			assert.Equal(t, "Captain Hook", opts.Name)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "50001",
				"channel_id": "30001",
				"name": "Captain Hook",
				"token": "hook_token_123"
			}`))
		case r.Method == http.MethodPost && r.URL.Path == "/webhooks/50001/hook_token_123":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	})
	defer ts.Close()

	ctx := context.Background()

	wh, err := c.CreateWebhook(ctx, 30001, CreateWebhookOptions{
		Name: "Captain Hook",
	})
	require.NoError(t, err)
	assert.Equal(t, Snowflake(50001), wh.ID)
	assert.Equal(t, "Captain Hook", wh.Name)

	_, err = c.ExecuteWebhook(ctx, 50001, "hook_token_123", ExecuteWebhookOptions{
		Content: "Ahoy from webhook",
	})
	require.NoError(t, err)
}

func TestClient_ApplicationCommandREST(t *testing.T) {
	c, ts := newTestClientWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/applications/88755900418551454/commands":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"id": "60001",
					"application_id": "88755900418551454",
					"name": "ping",
					"description": "ping pong command",
					"type": 1
				}
			]`))
		case r.Method == http.MethodPost && r.URL.Path == "/applications/88755900418551454/commands":
			var cmd ChatInputCommand
			err := json.NewDecoder(r.Body).Decode(&cmd)
			assert.NoError(t, err)
			assert.Equal(t, "ping", cmd.Name)

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{
				"id": "60001",
				"application_id": "88755900418551454",
				"name": "ping",
				"description": "ping pong command",
				"type": 1
			}`))
		default:
			http.NotFound(w, r)
		}
	})
	defer ts.Close()

	ctx := context.Background()
	appID := Snowflake(88755900418551454)

	cmds, err := c.FetchGlobalApplicationCommands(ctx, appID)
	require.NoError(t, err)
	require.Len(t, cmds, 1)
	assert.Equal(t, Snowflake(60001), cmds[0].GetID())
	assert.Equal(t, "ping", cmds[0].GetName())

	cmdToCreate := &ChatInputCommand{
		ApplicationCommandBase: ApplicationCommandBase{
			Name: "ping",
			Type: ApplicationCommandTypeChatInput,
		},
		DescriptionConstraints: DescriptionConstraints{
			Description: "ping pong command",
		},
	}
	newCmd, err := c.CreateGlobalApplicationCommand(ctx, appID, cmdToCreate)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(60001), newCmd.GetID())
}
