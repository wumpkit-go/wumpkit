package wumpkit

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardManager_SingleShardCreation(t *testing.T) {
	client := New(context.Background(), WithToken("test_token"))
	d := client.dispatcher
	limiter := NewDefaultShardsRateLimiter(1, 5*time.Second)
	sm := NewShardManager(
		"test_token",
		slog.Default(),
		d,
		WithShardManagerTotalShards(1),
		WithShardManagerIdentify(IdentifyProperties{
			OS:      "linux",
			Browser: "wumpkit",
			Device:  "wumpkit",
		}),
		WithShardManagerIntents(GatewayIntentGuilds),
		WithShardManagerCompression(GatewayCompressionZlibStream),
		WithShardManagerIdentifyRateLimiter(limiter),
	)

	require.NotNil(t, sm)
	assert.Equal(t, 0, sm.ShardCount())
}

func TestShardManager_GuildShardingRouting(t *testing.T) {
	// guild ID: 41771983444115456
	// formula: (41771983444115456 >> 22) % 1 = 0
	guildID := Snowflake(41771983444115456)
	assert.Equal(t, 0, GuildShardID(guildID, 1))

	// sharding with 5 shards
	shardID := GuildShardID(guildID, 5)
	expected := int((uint64(guildID) >> 22) % 5)
	assert.Equal(t, expected, shardID)
}

func TestShardManager_IdentifyRateLimiter(t *testing.T) {
	limiter := NewDefaultShardsRateLimiter(2, 50*time.Millisecond)

	start := time.Now()
	_ = limiter.WaitShard(context.Background(), 0) // consumes bucket 0
	_ = limiter.WaitShard(context.Background(), 1) // consumes bucket 1
	elapsed := time.Since(start)
	assert.Less(t, elapsed, 30*time.Millisecond)
}

func TestShardManager_OptionsAndGetters(t *testing.T) {
	limiter := NewDefaultShardsRateLimiter(1, 5*time.Second)
	sm := NewShardManager(
		"test_token_123",
		slog.Default(),
		nil,
		WithShardManagerTotalShards(8),
		WithShardManagerShardIDs(0, 1, 2, 3),
		WithShardManagerIdentify(IdentifyProperties{
			OS:      "darwin",
			Browser: "my-custom-browser",
			Device:  "my-custom-device",
		}),
		WithShardManagerIntents(GatewayIntentGuilds|GatewayIntentGuildMessages),
		WithShardManagerCompression(GatewayCompressionZlibStream),
		WithShardManagerGatewayURL("wss://custom.gateway.discord.gg"),
		WithShardManagerIdentifyRateLimiter(limiter),
	)

	assert.Equal(t, "test_token_123", sm.Token())
	assert.Equal(t, 8, sm.TotalShards())
	assert.Equal(t, []int{0, 1, 2, 3}, sm.ShardIDs())
	assert.Equal(t, GatewayIntentGuilds|GatewayIntentGuildMessages, sm.Intents())
	assert.Equal(t, GatewayCompressionZlibStream, sm.Compression())
	assert.Equal(t, "darwin", sm.Identify().OS)
	assert.Equal(t, "my-custom-browser", sm.Identify().Browser)
	assert.Equal(t, "my-custom-device", sm.Identify().Device)
	assert.Equal(t, limiter, sm.IdentifyRateLimiter())
	assert.Equal(t, 0, sm.ShardCount())
	assert.Equal(t, time.Duration(0), sm.AverageLatency())
}

func TestClient_CustomizationOptionsAndGetters(t *testing.T) {
	customCache := NewInMemoryCacheManager(CacheFlagsNone, DefaultCacheConfig())
	sm := NewShardManager("custom_token_456", slog.Default(), nil)

	client := New(
		context.Background(),
		WithToken("custom_token_456"),
		WithShardManagerConfig(ShardManagerConfig{
			TotalShards: 16,
			ShardIDs:    []int{0, 1, 2},
			Intents:     GatewayIntentGuilds | GatewayIntentMessageContent,
			Compression: GatewayCompressionNone,
			GatewayURL:  "wss://proxy.gateway.local",
		}),
		WithRequesterConfig(RequesterConfig{
			BaseURL:   "https://proxy.api.local",
			UserAgent: "MyBot/1.0",
		}),
		WithCacheManager(customCache),
		WithShardManager(sm),
	)

	assert.Equal(t, "custom_token_456", client.Token())
	assert.Equal(t, 16, client.TotalShards())
	assert.Equal(t, []int{0, 1, 2}, client.ShardIDs())
	assert.Equal(t, GatewayIntentGuilds|GatewayIntentMessageContent, client.Intents())
	assert.Equal(t, GatewayCompressionNone, client.Compression())
	assert.Equal(t, customCache, client.Cache())
	assert.Equal(t, sm, client.ShardManager())
	assert.NotNil(t, client.Dispatcher())
	assert.NotNil(t, client.Requester())
	assert.NotNil(t, client.Context())
	assert.NotNil(t, client.Logger())
}

func TestClient_GracefulShutdown_Idempotent(t *testing.T) {
	client := New(
		context.Background(),
		WithToken("test_token_shutdown"),
	)

	// calling Shutdown multiple times should be safe and idempotent
	client.Shutdown()
	client.Shutdown()

	err := client.Close()
	assert.NoError(t, err)

	// internal context should be cancelled
	select {
	case <-client.Context().Done():
		// success: context was cancelled on shutdown
	default:
		t.Fatal("expected client context to be cancelled after shutdown")
	}
}

func TestShard_GracefulShutdown_Idempotent(t *testing.T) {
	s := newShard(
		0, 1, "test_token", GatewayIntentGuilds,
		slog.Default(), nil, nil,
		GatewayCompressionNone, IdentifyProperties{},
		"wss://gateway.discord.gg", nil,
	)

	// shard shutdown should be safe to call multiple times without panicking
	err1 := s.Shutdown()
	err2 := s.Shutdown()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.True(t, s.closed.Load())
}

func TestDispatcher_GracefulShutdown(t *testing.T) {
	d := newDispatcher(context.Background(), nil, slog.Default(), HandlersExecutionSequential)

	d.OnMessageCreate(func(ctx context.Context, e MessageCreateEvent) {
		// handled event
	})

	// dispatch dummy payload
	d.Dispatch(0, "MESSAGE_CREATE", []byte(`{"id":"123","channel_id":"456","content":"hi"}`))

	// graceful shutdown should drain
	d.Shutdown(1 * time.Second)
	assert.True(t, d.closed.Load())
}

func TestToken_PrefixSanitization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my_token_raw", "my_token_raw"},
		{"Bot my_bot_token", "my_bot_token"},
		{"bot my_bot_token_lower", "my_bot_token_lower"},
		{"BOT my_bot_token_upper", "my_bot_token_upper"},
		{"Bearer my_bearer_token", "my_bearer_token"},
		{"bearer my_bearer_token_lower", "my_bearer_token_lower"},
		{"BEARER my_bearer_token_upper", "my_bearer_token_upper"},
		{"  Bot   my_spaced_token  ", "my_spaced_token"},
	}

	for _, tt := range tests {
		sm := NewShardManager(tt.input, slog.Default(), nil)
		assert.Equal(t, tt.expected, sm.Token())

		s := newShard(
			0, 1, tt.input, GatewayIntentGuilds,
			slog.Default(), nil, nil,
			GatewayCompressionNone, IdentifyProperties{},
			"wss://gateway.discord.gg", nil,
		)
		assert.Equal(t, tt.expected, s.token)

		c := New(context.Background(), WithToken(tt.input))
		assert.Equal(t, tt.expected, c.Token())
	}
}

func TestShard_ZstdDecompression(t *testing.T) {
	s := newShard(
		0, 1, "test_token", GatewayIntentGuilds,
		slog.Default(), nil, nil,
		GatewayCompressionZstdStream, IdentifyProperties{},
		"wss://gateway.discord.gg", nil,
	)
	defer func() {
		_ = s.Shutdown()
	}()

	original := []byte(
		`{"op":0,"s":1,"t":"READY","d":{"session_id":"abc","resume_gateway_url":"wss://resume.discord.gg"}}`,
	)

	var compressedBuf bytes.Buffer
	enc, err := zstd.NewWriter(&compressedBuf)
	require.NoError(t, err)
	_, err = enc.Write(original)
	require.NoError(t, err)
	require.NoError(t, enc.Close())

	decompressed, err := s.decompressZstd(compressedBuf.Bytes())
	require.NoError(t, err)
	assert.Equal(t, string(original), string(decompressed))
}

func TestShard_CommandRateLimiter(t *testing.T) {
	s := newShard(
		0, 1, "test_token", GatewayIntentGuilds,
		slog.Default(), nil, nil,
		GatewayCompressionNone, IdentifyProperties{},
		"wss://gateway.discord.gg", nil,
	)
	defer func() {
		_ = s.Shutdown()
	}()

	require.NotNil(t, s.commandLimiter)
	assert.Equal(t, uint(120), s.commandLimiter.Capacity())
	assert.Equal(t, uint(120), s.commandLimiter.Remaining())

	err := s.commandLimiter.Add(1)
	assert.NoError(t, err)
	assert.Equal(t, uint(119), s.commandLimiter.Remaining())
}

func TestShard_PresenceRateLimiter(t *testing.T) {
	s := newShard(
		0, 1, "test_token", GatewayIntentGuilds,
		slog.Default(), nil, nil,
		GatewayCompressionNone, IdentifyProperties{},
		"wss://gateway.discord.gg", nil,
	)
	defer func() {
		_ = s.Shutdown()
	}()

	require.NotNil(t, s.presenceLimiter)
	assert.Equal(t, uint(5), s.presenceLimiter.Capacity())
	assert.Equal(t, uint(5), s.presenceLimiter.Remaining())

	err := s.presenceLimiter.Add(1)
	assert.NoError(t, err)
	assert.Equal(t, uint(4), s.presenceLimiter.Remaining())
}

func TestShard_SessionSynchronization(t *testing.T) {
	s := newShard(
		0, 1, "test_token", GatewayIntentGuilds,
		slog.Default(), nil, nil,
		GatewayCompressionNone, IdentifyProperties{},
		"wss://gateway.discord.gg", nil,
	)
	defer func() {
		_ = s.Shutdown()
	}()

	s.setSession("sess123", "wss://resume.discord.gg")
	assert.Equal(t, "sess123", s.SessionID())
	assert.Equal(t, "wss://resume.discord.gg", s.ResumeURL())

	sess, url := s.getSession()
	assert.Equal(t, "sess123", sess)
	assert.Equal(t, "wss://resume.discord.gg", url)

	s.clearSession()
	assert.Empty(t, s.SessionID())
	assert.Empty(t, s.ResumeURL())
}

func TestDefaultShardsRateLimiter_Close(t *testing.T) {
	limiter := NewDefaultShardsRateLimiter(2, 50*time.Millisecond)
	require.NotNil(t, limiter)
	// closing must terminate goroutines without panic
	limiter.Close()
	// multiple close calls must be safe and idempotent
	limiter.Close()
}
