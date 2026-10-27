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
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

/*****************************
 *          Client
 *****************************/

// Client manages your Discord connection at a high level, grouping multiple shards together.
//
// It encapsulates:
//   - REST API interactions via requester.
//   - Caching via CacheManager.
//
// Client manages interactions with Discord's API, WebSocket Gateway, and HTTP Webhook Events.
// It coordinates:
//   - REST API requests via requester.
//   - In-memory/Redis entity caching via CacheManager.
//   - Event dispatching via dispatcher.
//   - Shard management for scalable Gateway connections.
//
// Create a Client using wumpkit.New() with desired options, then call StartGateway(), StartGatewayWithSignals(), StartWebhookEvents(), StartWebhookEventsWithSignals(), or StartBothWithSignals().
type Client struct {
	ctx                    context.Context
	cancel                 context.CancelFunc
	shutdownOnce           sync.Once
	isGatewayRunning       atomic.Bool       // indicates whether Gateway is active on this client
	isWebhookEventsRunning atomic.Bool       // indicates whether Webhook Events server is active on this client
	logger                 Logger            // logger used throughout the client
	token                  string            // bot token (without "Bot " prefix)
	publicKey              string            // application public key for HTTP interaction verification
	publicKeyBytes         ed25519.PublicKey // pre-parsed ed25519 public key
	httpServer             *http.Server      // dedicated HTTP interactions server
	shardManager           *ShardManager     // manages Gateway shard lifecycle
	*requester                               // REST API client
	CacheManager                             // CacheManager for caching discord entities
	*dispatcher                              // event dispatcher
}

// clientConfig holds temporary configuration used only during Client initialization in New().
type clientConfig struct {
	token               string
	publicKey           string
	logger              Logger
	cacheManager        CacheManager
	shardManager        *ShardManager
	requesterConfig     RequesterConfig
	shardManagerConfig  ShardManagerConfig
	dispatcherConfig    DispatcherConfig
	cacheConfig         CacheConfig
	webhookEventsConfig WebhookEventsConfig
}

// clientOption defines a function used to configure Client initialization during creation.
type clientOption func(*clientConfig)

/*****************************
 *       Options
 *****************************/

// WithToken sets the bot token for your client.
//
// Usage:
//
//	y := wumpkit.New(ctx, wumpkit.WithToken("your_bot_token"))
//
// Notes:
//   - Logs fatal and exits if token is empty or obviously invalid (< 50 chars).
//   - Removes "Bot " prefix automatically if provided.
//
// Warning: Never share your bot token publicly.
func WithToken(token string) clientOption {
	if token == "" {
		log.Fatal("WithToken: token must not be empty")
	}
	token = cleanGatewayToken(token)
	return func(cfg *clientConfig) {
		cfg.token = token
	}
}

// WithLogger sets a custom Logger implementation for your client.
//
// Usage:
//
//	y := wumpkit.New(ctx, wumpkit.WithLogger(myLogger))
//
// If nil is provided or option omitted, slog is used by default.
func WithLogger(logger Logger) clientOption {
	return func(cfg *clientConfig) {
		if logger != nil {
			cfg.logger = logger
		}
	}
}

// WithCacheManager sets a custom CacheManager implementation for your client.
//
// Usage:
//
//	y := wumpkit.New(wumpkit.WithCacheManager(myCacheManager))
//
// Logs fatal and exits if cacheManager is nil.
func WithCacheManager(cacheManager CacheManager) clientOption {
	if cacheManager == nil {
		log.Fatal("WithCacheManager: cacheManager must not be nil")
	}
	return func(cfg *clientConfig) {
		cfg.cacheManager = cacheManager
	}
}

// WithShardManager sets a pre-configured or custom ShardManager instance on the client.
func WithShardManager(sm *ShardManager) clientOption {
	if sm == nil {
		log.Fatal("WithShardManager: shardManager must not be nil")
	}
	return func(cfg *clientConfig) {
		cfg.shardManager = sm
	}
}

// WithRequesterConfig sets the configuration for the HTTP requester.
func WithRequesterConfig(config RequesterConfig) clientOption {
	return func(cfg *clientConfig) {
		cfg.requesterConfig = config
	}
}

// WithRESTConfig is an alias for WithRequesterConfig.
func WithRESTConfig(config RequesterConfig) clientOption {
	return WithRequesterConfig(config)
}

// WithShardManagerConfig sets the configuration for the ShardManager and Gateway connections.
func WithShardManagerConfig(config ShardManagerConfig) clientOption {
	return func(cfg *clientConfig) {
		cfg.shardManagerConfig = config
	}
}

// WithGatewayConfig is an alias for WithShardManagerConfig.
func WithGatewayConfig(config ShardManagerConfig) clientOption {
	return WithShardManagerConfig(config)
}

// WithDispatcherConfig sets the configuration for the event dispatcher.
func WithDispatcherConfig(config DispatcherConfig) clientOption {
	return func(cfg *clientConfig) {
		cfg.dispatcherConfig = config
	}
}

// WithCacheConfig sets the configuration for the cache manager.
func WithCacheConfig(config CacheConfig) clientOption {
	return func(cfg *clientConfig) {
		cfg.cacheConfig = config
	}
}

// WithWebhookEventsConfig sets the configuration for the Webhook Events HTTP server.
func WithWebhookEventsConfig(config WebhookEventsConfig) clientOption {
	return func(cfg *clientConfig) {
		cfg.webhookEventsConfig = config
	}
}

/*****************************
 *       Constructor
 *****************************/

// New creates a new Client instance with provided options.
//
// Example:
//
//	bot := wumpkit.New(
//	    ctx,
//	    wumpkit.WithToken("my_bot_token"),
//	    wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
//	        Intents: wumpkit.GatewayIntentGuilds | wumpkit.GatewayIntentGuildMessages,
//	    }),
//	    wumpkit.WithLogger(myLogger),
//	)
//
// Defaults:
//   - Logger: stdout logger at Info level.
//   - Gateway: DefaultShardManagerConfig() (Intents: Guilds | GuildMessages | GuildMembers, Compression: zlib-stream)
//   - Cache: Default InMemoryCacheManager with all flags enabled
//   - REST: DefaultRequesterConfig() (BaseURL: "https://discord.com/api/v10", MaxRetries: 3)
func New(ctx context.Context, options ...clientOption) *Client {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	cfg := clientConfig{
		logger:              slog.Default(),
		requesterConfig:     DefaultRequesterConfig(),
		shardManagerConfig:  DefaultShardManagerConfig(),
		dispatcherConfig:    DefaultDispatcherConfig(),
		cacheConfig:         DefaultCacheConfig(),
		webhookEventsConfig: DefaultWebhookEventsConfig(),
	}

	for _, option := range options {
		option(&cfg)
	}

	if cfg.logger == nil {
		cfg.logger = slog.Default()
	}

	cacheManager := cfg.cacheManager
	if cacheManager == nil {
		cacheManager = NewInMemoryCacheManager(
			CacheFlagGuilds|CacheFlagMembers|CacheFlagChannels|CacheFlagRoles|CacheFlagUsers|CacheFlagVoiceStates,
			cfg.cacheConfig,
		)
	}

	disp := newDispatcher(ctx, cacheManager, cfg.logger, cfg.dispatcherConfig.ExecutionMode)
	req := newRequester(cfg.token, cfg.requesterConfig, cfg.logger)

	var pubKeyBytes ed25519.PublicKey
	if cfg.publicKey != "" {
		if b, err := hex.DecodeString(cfg.publicKey); err == nil && len(b) == ed25519.PublicKeySize {
			pubKeyBytes = ed25519.PublicKey(b)
		}
	}

	sm := cfg.shardManager
	if sm == nil {
		opts := []ShardManagerOption{
			applyShardManagerConfig(cfg.shardManagerConfig),
		}
		sm = NewShardManager(cfg.token, cfg.logger, disp, opts...)
	} else {
		applyShardManagerConfig(cfg.shardManagerConfig)(sm)
		if sm.dispatcher == nil {
			sm.dispatcher = disp
		}
		if sm.logger == nil {
			sm.logger = cfg.logger
		}
	}

	client := &Client{
		ctx:            ctx,
		cancel:         cancel,
		logger:         cfg.logger,
		token:          cfg.token,
		publicKey:      cfg.publicKey,
		publicKeyBytes: pubKeyBytes,
		shardManager:   sm,
		requester:      req,
		CacheManager:   cacheManager,
		dispatcher:     disp,
	}

	return client
}

/*****************************
 *       Start
 *****************************/

// GatewayStartOption configures Gateway parameters when starting a Gateway connection.
type GatewayStartOption func(*ShardManager)

// StartGateway initializes internal components and connects shards to Discord Gateway.
// Blocks until client context is cancelled or Shutdown() is called.
func (c *Client) StartGateway(options ...GatewayStartOption) error {
	for _, opt := range options {
		opt(c.shardManager)
	}

	if !c.isGatewayRunning.CompareAndSwap(false, true) {
		return errors.New("cannot start Gateway client: Gateway is already running on this client")
	}
	defer c.isGatewayRunning.Store(false)

	// discover Gateway configuration & recommended shard count if not configured
	totalShards := c.shardManager.totalShards
	if totalShards <= 0 || c.shardManager.gatewayURL == "" || c.shardManager.identifyLimiter == nil {
		gatewayBotData, err := c.FetchGatewayBot(c.ctx)
		if err != nil {
			return fmt.Errorf("failed fetching gateway bot info: %w", err)
		}

		if totalShards <= 0 {
			totalShards = gatewayBotData.Shards
			c.shardManager.totalShards = totalShards
		}
		if c.shardManager.gatewayURL == "" {
			c.shardManager.gatewayURL = gatewayBotData.URL
		}
		if c.shardManager.identifyLimiter == nil {
			c.shardManager.identifyLimiter = NewDefaultShardsRateLimiter(
				gatewayBotData.SessionStartLimit.MaxConcurrency,
				5*time.Second,
			)
		}
	}

	// start all configured shards
	if err := c.shardManager.Start(c.ctx, totalShards); err != nil {
		return err
	}

	<-c.ctx.Done()
	c.Shutdown()
	return nil
}

// StartGatewayWithSignals starts the Gateway client and blocks until an interrupt or termination signal
// (SIGINT, SIGTERM) is received, then performs a graceful shutdown.
func (c *Client) StartGatewayWithSignals(signals ...os.Signal) error {
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}
	sigCtx, stop := signal.NotifyContext(c.ctx, signals...)
	defer stop()

	errChan := make(chan error, 1)
	go func() {
		errChan <- c.StartGateway()
	}()

	select {
	case err := <-errChan:
		return err
	case <-sigCtx.Done():
		if c.logger != nil {
			c.logger.Info("OS signal received, initiating graceful shutdown")
		}
		c.Shutdown()
		return <-errChan
	}
}

// StartBoth starts both the Gateway WebSocket connection and the Webhook Events HTTP server simultaneously on the same client.
// Blocks until client context is cancelled or Shutdown() is called.
func (c *Client) StartBoth(addr string, gatewayOpts []GatewayStartOption, webhookOpts ...WebhookEventsStartOption) error {
	errChan := make(chan error, 2)

	go func() {
		if err := c.StartGateway(gatewayOpts...); err != nil {
			errChan <- fmt.Errorf("gateway error: %w", err)
		}
	}()

	go func() {
		if err := c.StartWebhookEvents(addr, webhookOpts...); err != nil {
			errChan <- fmt.Errorf("webhook events error: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		c.Shutdown()
		return err
	case <-c.ctx.Done():
		c.Shutdown()
		return nil
	}
}

// StartBothWithSignals starts both Gateway and Webhook Events simultaneously, blocking until an OS signal is received.
func (c *Client) StartBothWithSignals(addr string, gatewayOpts []GatewayStartOption, webhookOpts ...WebhookEventsStartOption) error {
	sigCtx, stop := signal.NotifyContext(c.ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 1)
	go func() {
		errChan <- c.StartBoth(addr, gatewayOpts, webhookOpts...)
	}()

	select {
	case err := <-errChan:
		return err
	case <-sigCtx.Done():
		if c.logger != nil {
			c.logger.Info("OS signal received, initiating graceful shutdown")
		}
		c.Shutdown()
		return <-errChan
	}
}

// StartCluster initializes and starts the client for a specific cluster (subset of shards).
func (c *Client) StartCluster(ctx context.Context, totalShards int, shardIDs ...int) error {
	if ctx != nil {
		c.ctx = ctx
	}
	if c.shardManager != nil {
		c.shardManager.totalShards = totalShards
		c.shardManager.shardIDs = shardIDs
	}
	return c.StartGateway()
}

// StartShards initializes and starts specific shards on this client instance.
func (c *Client) StartShards(ctx context.Context, shardIDs ...int) error {
	if ctx != nil {
		c.ctx = ctx
	}
	if c.shardManager != nil {
		c.shardManager.shardIDs = shardIDs
	}
	return c.StartGateway()
}

/*****************************
 *       Shutdown
 *****************************/

// Shutdown cleanly shuts down the Client.
//
// It:
//   - Cancels the client's internal context so Start() unblocks cleanly.
//   - Shuts down all managed shards via ShardManager (closing WebSocket connections with StatusNormalClosure).
//   - Gracefully drains in-flight event handlers via dispatcher.
//   - Shuts down the REST API requester (closes idle connections).
func (c *Client) Shutdown() {
	c.shutdownOnce.Do(func() {
		if c.logger != nil {
			c.logger.Info("Client gracefully shutting down")
		}
		if c.cancel != nil {
			c.cancel()
		}
		if c.shardManager != nil {
			c.shardManager.Shutdown()
		}
		if c.dispatcher != nil {
			c.dispatcher.Shutdown(5 * time.Second)
		}
		if c.requester != nil {
			c.requester.Shutdown()
		}
		if c.httpServer != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = c.httpServer.Shutdown(shutdownCtx)
		}
	})
}

// Close is an alias for Shutdown, implementing io.Closer.
func (c *Client) Close() error {
	c.Shutdown()
	return nil
}

/*****************************
 *       Helpers
 *****************************/

// Shards returns the slice of active Gateway shards.
func (c *Client) Shards() []*Shard {
	if c.shardManager == nil {
		return nil
	}
	return c.shardManager.Shards()
}

// ShardByID retrieves a specific shard instance by its ID.
func (c *Client) ShardByID(shardID int) *Shard {
	if c.shardManager == nil {
		return nil
	}
	return c.shardManager.ShardByID(shardID)
}

// ShardCount returns the number of active Gateway shards managed by this instance.
func (c *Client) ShardCount() int {
	if c.shardManager == nil {
		return 0
	}
	return c.shardManager.ShardCount()
}

// Latency returns the average heartbeat latency across all active shards.
func (c *Client) Latency() time.Duration {
	if c.shardManager == nil {
		return 0
	}
	return c.shardManager.AverageLatency()
}

// ShardManager returns the underlying ShardManager instance.
func (c *Client) ShardManager() *ShardManager {
	return c.shardManager
}

// Requester returns the underlying REST API requester.
func (c *Client) Requester() *requester {
	return c.requester
}

// RateLimitStore returns the configured RateLimitStore instance.
func (c *Client) RateLimitStore() RateLimitStore {
	if c.requester != nil {
		return c.requester.rateLimitStore
	}
	return nil
}

// Cache returns the active CacheManager instance.
func (c *Client) Cache() CacheManager {
	return c.CacheManager
}

// Dispatcher returns the underlying event dispatcher.
func (c *Client) Dispatcher() *dispatcher {
	return c.dispatcher
}

// Logger returns the Logger configured on the client.
func (c *Client) Logger() Logger {
	return c.logger
}

// Context returns the client root context.
func (c *Client) Context() context.Context {
	return c.ctx
}

// Token returns the bot token configured on the client.
func (c *Client) Token() string {
	return c.token
}

// Intents returns the Gateway intents configured on the client.
func (c *Client) Intents() GatewayIntent {
	if c.shardManager != nil {
		return c.shardManager.Intents()
	}
	return 0
}

// Compression returns the Gateway compression mode configured on the client.
func (c *Client) Compression() GatewayCompression {
	if c.shardManager != nil {
		return c.shardManager.Compression()
	}
	return GatewayCompressionNone
}

// TotalShards returns the configured total shard count (or 0 if using Discord recommendation).
func (c *Client) TotalShards() int {
	if c.shardManager != nil {
		return c.shardManager.TotalShards()
	}
	return 0
}

// ShardIDs returns the specific shard IDs configured for this client.
func (c *Client) ShardIDs() []int {
	if c.shardManager != nil {
		return c.shardManager.ShardIDs()
	}
	return nil
}

// IdentifyProperties returns the Identify properties configured on the client.
func (c *Client) IdentifyProperties() IdentifyProperties {
	if c.shardManager != nil {
		return c.shardManager.Identify()
	}
	return IdentifyProperties{}
}

// AverageLatency returns the average heartbeat latency across all active shards.
func (c *Client) AverageLatency() time.Duration {
	return c.Latency()
}

// UpdatePresence updates the presence for all managed shards.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#update-presence
func (c *Client) UpdatePresence(presence GatewayPresenceUpdate) error {
	if c.shardManager == nil {
		return errors.New("client: shard manager is not started")
	}
	for _, s := range c.shardManager.Shards() {
		if err := s.UpdatePresence(presence); err != nil {
			return err
		}
	}
	return nil
}

// UpdateVoiceState joins, moves, or disconnects the client from a voice channel in a guild.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#update-voice-state
func (c *Client) UpdateVoiceState(guildID Snowflake, channelID *Snowflake, selfMute, selfDeaf bool) error {
	if c.shardManager == nil {
		return errors.New("client: shard manager is not started")
	}
	totalShards := c.shardManager.ShardCount()
	shardID := GuildShardID(guildID, totalShards)
	shard := c.shardManager.ShardByID(shardID)
	if shard == nil {
		// fallback to first shard if exact shard not in this cluster
		shards := c.shardManager.Shards()
		if len(shards) == 0 {
			return errors.New("client: no active shards")
		}
		shard = shards[0]
	}
	return shard.UpdateVoiceState(guildID, channelID, selfMute, selfDeaf)
}

// RequestGuildMembers requests offline guild members for a guild.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#request-guild-members
func (c *Client) RequestGuildMembers(opts RequestGuildMembersOptions) error {
	if c.shardManager == nil {
		return errors.New("client: shard manager is not started")
	}
	totalShards := c.shardManager.ShardCount()
	shardID := GuildShardID(opts.GuildID, totalShards)
	shard := c.shardManager.ShardByID(shardID)
	if shard == nil {
		shards := c.shardManager.Shards()
		if len(shards) == 0 {
			return errors.New("client: no active shards")
		}
		shard = shards[0]
	}
	return shard.RequestGuildMembers(opts)
}

// RequestSoundboardSounds requests soundboard sounds for a list of guilds.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#request-soundboard-sounds
func (c *Client) RequestSoundboardSounds(guildIDs []Snowflake) error {
	if c.shardManager == nil {
		return errors.New("client: shard manager is not started")
	}
	shards := c.shardManager.Shards()
	if len(shards) == 0 {
		return errors.New("client: no active shards")
	}
	return shards[0].RequestSoundboardSounds(guildIDs)
}

// RequestChannelInfo requests ephemeral channel data for channels in a guild.
//
// Reference: https://discord.com/developers/docs/events/gateway-events#request-channel-info
func (c *Client) RequestChannelInfo(opts RequestChannelInfoOptions) error {
	if c.shardManager == nil {
		return errors.New("client: shard manager is not started")
	}
	totalShards := c.shardManager.ShardCount()
	shardID := GuildShardID(opts.GuildID, totalShards)
	shard := c.shardManager.ShardByID(shardID)
	if shard == nil {
		shards := c.shardManager.Shards()
		if len(shards) == 0 {
			return errors.New("client: no active shards")
		}
		shard = shards[0]
	}
	return shard.RequestChannelInfo(opts)
}

// ComputeMemberPermissions calculates a member's effective base permissions in a guild using cached roles and @everyone role.
// If the member is the owner of the guild, it returns PermissionAll.
func (c *Client) ComputeMemberPermissions(ctx context.Context, member Member) Permissions {
	if ctx == nil {
		ctx = c.ctx
	}
	guild, ok := c.GetGuild(ctx, member.GuildID)
	if !ok {
		return 0
	}
	if guild.OwnerID == member.ID {
		return PermissionAll
	}
	roles := c.GetRoles(ctx, member.RoleIDs...)
	if everyone, ok := c.GetRole(ctx, member.GuildID); ok {
		roles[member.GuildID] = everyone
	}
	var perms Permissions
	for _, r := range roles {
		perms.Add(r.Permissions)
	}
	return perms
}

// ComputeChannelPermissions calculates a member's effective permissions in a guild channel,
// factoring in base permissions and channel-specific permission overwrites.
func (c *Client) ComputeChannelPermissions(ctx context.Context, member Member, channel GuildChannel) Permissions {
	basePerms := c.ComputeMemberPermissions(ctx, member)
	return channel.PermissionsFor(basePerms, member)
}

// MemberPermissions is an alias for ComputeMemberPermissions.
func (c *Client) MemberPermissions(ctx context.Context, member Member) Permissions {
	return c.ComputeMemberPermissions(ctx, member)
}

// ChannelPermissions is an alias for ComputeChannelPermissions.
func (c *Client) ChannelPermissions(ctx context.Context, member Member, channel GuildChannel) Permissions {
	return c.ComputeChannelPermissions(ctx, member, channel)
}
