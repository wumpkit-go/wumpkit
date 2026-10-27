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
	"compress/zlib"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/klauspost/compress/zstd"
	"github.com/wumpkit/wumpkit/internal/ratelimit"
)

/*******************************
 * Shards Identify Rate Limiter
 *******************************/

// ShardsIdentifyRateLimiter defines the interface for a rate limiter
// that controls the frequency of Identify payloads sent per shard.
type ShardsIdentifyRateLimiter interface {
	// Wait blocks until the shard is allowed to send an Identify payload or the context is cancelled.
	Wait(ctx context.Context) error
}

// DefaultShardsRateLimiter implements a concurrent token bucket rate limiter for Gateway Identify.
type DefaultShardsRateLimiter struct {
	maxConcurrency int
	interval       time.Duration
	buckets        []chan struct{}
	stopChan       chan struct{}
	closed         sync.Once
}

var _ ShardsIdentifyRateLimiter = (*DefaultShardsRateLimiter)(nil)

// NewDefaultShardsRateLimiter creates a new concurrent token bucket rate limiter.
func NewDefaultShardsRateLimiter(r int, interval time.Duration) *DefaultShardsRateLimiter {
	r = max(1, r)
	rl := &DefaultShardsRateLimiter{
		maxConcurrency: r,
		interval:       interval,
		buckets:        make([]chan struct{}, r),
		stopChan:       make(chan struct{}),
	}
	for i := range r {
		ch := make(chan struct{}, 1)
		ch <- struct{}{}
		rl.buckets[i] = ch

		go func(bucketChan chan struct{}) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					select {
					case bucketChan <- struct{}{}:
					default:
					}
				case <-rl.stopChan:
					return
				}
			}
		}(ch)
	}
	return rl
}

// close stops all background refill tickers
func (rl *DefaultShardsRateLimiter) Close() {
	rl.closed.Do(func() {
		close(rl.stopChan)
	})
}

// Wait blocks until a token in bucket 0 is available for sending Identify or the context is cancelled.
func (rl *DefaultShardsRateLimiter) Wait(ctx context.Context) error {
	return rl.WaitShard(ctx, 0)
}

// WaitShard blocks until a token in the shard's bucket (shardID % maxConcurrency) is available.
func (rl *DefaultShardsRateLimiter) WaitShard(ctx context.Context, shardID int) error {
	bucket := 0
	if rl.maxConcurrency > 1 {
		bucket = shardID % rl.maxConcurrency
		if bucket < 0 {
			bucket = -bucket
		}
	}

	ch := rl.buckets[bucket]
	if ctx == nil {
		<-ch
		return nil
	}
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

/*************************************
 * ShardManager: manages multiple shards
 *************************************/

// IdentifyProperties configures the "properties" field in the Identify payload.
type IdentifyProperties struct {
	OS      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

// ShardManagerConfig holds configuration parameters for ShardManager and Gateway WebSocket connections.
type ShardManagerConfig struct {
	// TotalShards specifies the total number of shards. If 0, uses Discord's recommended shard count.
	TotalShards int
	// ShardIDs specifies the specific shard IDs managed by this instance (useful for clustering).
	ShardIDs []int
	// Intents specifies Gateway intents for all shards.
	Intents GatewayIntent
	// Compression specifies the Gateway compression mode (default: GatewayCompressionZlibStream).
	Compression GatewayCompression
	// GatewayURL specifies a custom Gateway base URL (useful for testing or reverse proxies).
	GatewayURL string
	// IdentifyProps specifies custom client identity properties sent during Gateway Identify.
	IdentifyProps IdentifyProperties
	// IdentifyLimiter specifies the rate limiter controlling Identify payloads across shards.
	IdentifyLimiter ShardsIdentifyRateLimiter
	// Options specifies additional functional options passed to ShardManager.
	Options []ShardManagerOption
}

// DefaultShardManagerConfig returns standard ShardManager configuration defaults.
func DefaultShardManagerConfig() ShardManagerConfig {
	return ShardManagerConfig{
		Intents: GatewayIntentGuilds |
			GatewayIntentGuildMessages |
			GatewayIntentGuildMembers,
		Compression: GatewayCompressionZlibStream,
	}
}

// ShardManagerOption configures a ShardManager instance.
type ShardManagerOption func(*ShardManager)

// applyShardManagerConfig applies a ShardManagerConfig to the ShardManager.
func applyShardManagerConfig(cfg ShardManagerConfig) ShardManagerOption {
	return func(sm *ShardManager) {
		if cfg.TotalShards > 0 {
			sm.totalShards = cfg.TotalShards
		}
		if len(cfg.ShardIDs) > 0 {
			sm.shardIDs = cfg.ShardIDs
		}
		if cfg.Intents != 0 {
			sm.intents = cfg.Intents
		}
		if cfg.Compression != "" {
			sm.compression = cfg.Compression
		}
		if cfg.GatewayURL != "" {
			sm.gatewayURL = cfg.GatewayURL
		}
		if cfg.IdentifyProps.OS != "" || cfg.IdentifyProps.Browser != "" ||
			cfg.IdentifyProps.Device != "" {
			sm.identify = cfg.IdentifyProps
		}
		if cfg.IdentifyLimiter != nil {
			sm.identifyLimiter = cfg.IdentifyLimiter
		}
		for _, opt := range cfg.Options {
			opt(sm)
		}
	}
}

// WithShardManagerTotalShards sets the total number of shards.
func WithShardManagerTotalShards(totalShards int) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.totalShards = totalShards
	}
}

// WithShardManagerShardIDs sets the specific shard IDs managed by this instance.
func WithShardManagerShardIDs(shardIDs ...int) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.shardIDs = shardIDs
	}
}

// WithShardManagerIdentify sets the Identify properties.
func WithShardManagerIdentify(identify IdentifyProperties) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.identify = identify
	}
}

// WithShardManagerCompression sets the Gateway compression mode.
func WithShardManagerCompression(compression GatewayCompression) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.compression = compression
	}
}

// WithShardManagerIntents sets the Gateway intents for all shards.
func WithShardManagerIntents(intents GatewayIntent) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.intents = intents
	}
}

// WithShardManagerIdentifyRateLimiter sets the rate limiter controlling Identify payloads.
func WithShardManagerIdentifyRateLimiter(limiter ShardsIdentifyRateLimiter) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.identifyLimiter = limiter
	}
}

// WithShardManagerGatewayURL sets a custom Gateway base URL (useful for testing, reverse proxies, or mocks).
func WithShardManagerGatewayURL(gatewayURL string) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.gatewayURL = gatewayURL
	}
}

// WithShardManagerHTTPClient sets a custom HTTP client for WebSocket dialing.
func WithShardManagerHTTPClient(httpClient *http.Client) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.httpClient = httpClient
	}
}

// WithShardManagerDialOptions sets custom WebSocket DialOptions.
func WithShardManagerDialOptions(opts *websocket.DialOptions) ShardManagerOption {
	return func(sm *ShardManager) {
		sm.dialOptions = opts
	}
}

// ShardManager manages the lifecycle of multiple Gateway shards.
type ShardManager struct {
	token           string
	totalShards     int
	shardIDs        []int
	identify        IdentifyProperties
	compression     GatewayCompression
	intents         GatewayIntent
	logger          Logger
	dispatcher      *dispatcher
	identifyLimiter ShardsIdentifyRateLimiter
	gatewayURL      string
	httpClient      *http.Client
	dialOptions     *websocket.DialOptions
	shards          []*Shard
	mu              sync.RWMutex
	shutdownOnce    sync.Once
}

// cleanGatewayToken strips leading "Bot " or "Bearer " prefixes (case-insensitive) from a Discord token.
func cleanGatewayToken(token string) string {
	token = strings.TrimSpace(token)
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bot ") {
		token = strings.TrimSpace(token[4:])
	} else if strings.HasPrefix(lower, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	return token
}

// NewShardManager creates a new ShardManager with the given token, logger, dispatcher, and options.
func NewShardManager(
	token string,
	logger Logger,
	dispatcher *dispatcher,
	options ...ShardManagerOption,
) *ShardManager {
	sm := &ShardManager{
		token:       cleanGatewayToken(token),
		logger:      logger,
		dispatcher:  dispatcher,
		compression: GatewayCompressionZlibStream,
		identify: IdentifyProperties{
			OS:      "linux",
			Browser: "wumpkit",
			Device:  "wumpkit",
		},
	}

	for _, opt := range options {
		opt(sm)
	}

	return sm
}

// Start connects all configured shards to Discord Gateway.
func (sm *ShardManager) Start(ctx context.Context, totalShards int) error {
	if totalShards <= 0 {
		totalShards = sm.totalShards
	}
	if totalShards <= 0 {
		totalShards = 1
	}

	var shardIDs []int
	if len(sm.shardIDs) > 0 {
		shardIDs = sm.shardIDs
	} else {
		shardIDs = make([]int, totalShards)
		for i := range totalShards {
			shardIDs[i] = i
		}
	}

	sm.logger.Info(
		"starting shard manager",
		"total_shards", totalShards,
		"managed_shards", shardIDs,
		"compression", sm.compression,
	)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	dialOpts := sm.dialOptions
	if dialOpts == nil && sm.httpClient != nil {
		dialOpts = &websocket.DialOptions{
			HTTPClient: sm.httpClient,
		}
	}

	for _, shardID := range shardIDs {
		shard := newShard(
			shardID, totalShards, sm.token, sm.intents,
			sm.logger, sm.dispatcher, sm.identifyLimiter,
			sm.compression, sm.identify, sm.gatewayURL, dialOpts,
		)
		if err := shard.connect(ctx); err != nil {
			return err
		}
		sm.shards = append(sm.shards, shard)
	}

	return nil
}

// AddShard connects and adds a new shard at runtime.
func (sm *ShardManager) AddShard(ctx context.Context, shardID, totalShards int) (*Shard, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, s := range sm.shards {
		if s.shardID == shardID {
			return s, nil
		}
	}

	dialOpts := sm.dialOptions
	if dialOpts == nil && sm.httpClient != nil {
		dialOpts = &websocket.DialOptions{
			HTTPClient: sm.httpClient,
		}
	}

	shard := newShard(
		shardID, totalShards, sm.token, sm.intents,
		sm.logger, sm.dispatcher, sm.identifyLimiter,
		sm.compression, sm.identify, sm.gatewayURL, dialOpts,
	)
	if err := shard.connect(ctx); err != nil {
		return nil, err
	}
	sm.shards = append(sm.shards, shard)
	return shard, nil
}

// RemoveShard disconnects and removes a shard at runtime.
func (sm *ShardManager) RemoveShard(shardID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for i, s := range sm.shards {
		if s.shardID == shardID {
			_ = s.Shutdown()
			sm.shards = append(sm.shards[:i], sm.shards[i+1:]...)
			return nil
		}
	}
	return errors.New("shard not found")
}

// RestartShard closes and reconnects a specific shard.
func (sm *ShardManager) RestartShard(ctx context.Context, shardID int) error {
	shard := sm.ShardByID(shardID)
	if shard == nil {
		return errors.New("shard not found")
	}
	_ = shard.Shutdown()
	return shard.connect(ctx)
}

// RestartAll closes and reconnects all managed shards.
func (sm *ShardManager) RestartAll(ctx context.Context) error {
	for _, shard := range sm.Shards() {
		_ = shard.Shutdown()
		if err := shard.connect(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown gracefully closes all managed shards.
func (sm *ShardManager) Shutdown() {
	sm.shutdownOnce.Do(func() {
		sm.logger.Info("shard manager shutting down")
		sm.mu.Lock()
		shards := sm.shards
		sm.shards = nil
		sm.mu.Unlock()

		var wg sync.WaitGroup
		for _, shard := range shards {
			wg.Add(1)
			go func(s *Shard) {
				defer wg.Done()
				_ = s.Shutdown()
			}(shard)
		}
		wg.Wait()
		if closer, ok := sm.identifyLimiter.(interface{ Close() }); ok {
			closer.Close()
		}
	})
}

// Shards returns a copy of the list of managed shards.
func (sm *ShardManager) Shards() []*Shard {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	out := make([]*Shard, len(sm.shards))
	copy(out, sm.shards)
	return out
}

// ShardIDs returns the list of shard IDs managed by this instance.
func (sm *ShardManager) ShardIDs() []int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if len(sm.shards) > 0 {
		ids := make([]int, len(sm.shards))
		for i, s := range sm.shards {
			ids[i] = s.shardID
		}
		return ids
	}
	ids := make([]int, len(sm.shardIDs))
	copy(ids, sm.shardIDs)
	return ids
}

// ShardCount returns the number of shards currently managed.
func (sm *ShardManager) ShardCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.shards)
}

// ShardByID returns the Shard with the given ID, or nil if not managed.
func (sm *ShardManager) ShardByID(shardID int) *Shard {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, s := range sm.shards {
		if s.shardID == shardID {
			return s
		}
	}
	return nil
}

// TotalShards returns the total shard count.
func (sm *ShardManager) TotalShards() int {
	return sm.totalShards
}

// Token returns the bot token configured on this shard manager.
func (sm *ShardManager) Token() string {
	return sm.token
}

// Intents returns the Gateway intents configured on this shard manager.
func (sm *ShardManager) Intents() GatewayIntent {
	return sm.intents
}

// Compression returns the Gateway compression configured on this shard manager.
func (sm *ShardManager) Compression() GatewayCompression {
	return sm.compression
}

// Identify returns the Identify properties configured on this shard manager.
func (sm *ShardManager) Identify() IdentifyProperties {
	return sm.identify
}

// Dispatcher returns the event dispatcher configured on this shard manager.
func (sm *ShardManager) Dispatcher() *dispatcher {
	return sm.dispatcher
}

// IdentifyRateLimiter returns the rate limiter controlling Identify payloads.
func (sm *ShardManager) IdentifyRateLimiter() ShardsIdentifyRateLimiter {
	return sm.identifyLimiter
}

// AverageLatency returns the average heartbeat latency across all active shards.
func (sm *ShardManager) AverageLatency() time.Duration {
	shards := sm.Shards()
	if len(shards) == 0 {
		return 0
	}
	var total time.Duration
	for _, s := range shards {
		total += s.Latency()
	}
	return total / time.Duration(len(shards))
}

// UpdatePresence sends an Opcode 3 Presence Update across all active shards.
func (sm *ShardManager) UpdatePresence(presence GatewayPresenceUpdate) error {
	for _, s := range sm.Shards() {
		if err := s.UpdatePresence(presence); err != nil {
			return err
		}
	}
	return nil
}

// Send broadcasts a raw Gateway opcode and data across all active shards.
func (sm *ShardManager) Send(op int, data any) error {
	payload, err := json.Marshal(map[string]any{
		"op": op,
		"d":  data,
	})
	if err != nil {
		return err
	}
	for _, s := range sm.Shards() {
		if err := s.writeMessage(payload); err != nil {
			return err
		}
	}
	return nil
}

/*************************************
 * Shard: a single Gateway connection
 *************************************/

const (
	gatewayVersion  = "10"
	gatewayBaseURL  = "wss://gateway.discord.gg"
	zlibFlushSuffix = "\x00\x00\xff\xff"
)

// Shard manages a single WebSocket connection to Discord Gateway,
// including session state, event handling, heartbeats, and reconnects.
type Shard struct {
	shardID     int           // shard number (zero-based)
	totalShards int           // total number of shards in the bot
	token       string        // Discord bot token
	intents     GatewayIntent // Gateway intents bitmask

	logger          Logger                    // logger interface for informational and error messages
	dispatcher      *dispatcher               // event dispatcher for received Gateway events
	identifyLimiter ShardsIdentifyRateLimiter // rate limiter controlling Identify payloads

	writeMu sync.Mutex      // protects concurrent writes on conn
	conn    *websocket.Conn // websocket connection

	seq       atomic.Int64 // last received sequence number from Gateway
	sessionMu sync.RWMutex
	sessionID string // current session id for resuming
	resumeURL string // Gateway URL to resume session on

	ctx    context.Context
	cancel context.CancelFunc

	latency           atomic.Int64  // heartbeat latency in milliseconds
	lastHeartbeatSent atomic.Int64  // timestamp (unix nano) of last heartbeat sent
	lastHeartbeatACK  atomic.Bool   // true if last heartbeat was acknowledged
	heartbeatStop     chan struct{} // signal to stop heartbeat goroutine
	stopChan          chan struct{} // signal to stop all shard background workers immediately
	shutdownOnce      sync.Once     // ensures shutdown logic runs only once

	reconnecting atomic.Bool // prevents overlapping reconnection loops
	closed       atomic.Bool // indicates graceful shutdown

	// Compression support
	compression GatewayCompression
	properties  IdentifyProperties
	gatewayURL  string
	dialOptions *websocket.DialOptions

	// zlib buffer and reader for decompression
	zlibBuf    bytes.Buffer
	zlibReader io.ReadCloser
	zlibMu     sync.Mutex

	// zstd decompression state
	zstdDecoder *zstd.Decoder
	zstdMu      sync.Mutex

	// command rate limiter for outbound gateway commands
	commandLimiter ratelimit.Bucket
	// presence rate limiter (5 updates per 20s per shard)
	presenceLimiter ratelimit.Bucket
}

func (s *Shard) getSession() (sessionID, resumeURL string) {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()
	return s.sessionID, s.resumeURL
}

func (s *Shard) setSession(sessionID, resumeURL string) {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	s.sessionID = sessionID
	s.resumeURL = resumeURL
}

func (s *Shard) clearSession() {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	s.sessionID = ""
	s.resumeURL = ""
}

// newShard constructs a new Shard instance with the specified parameters
func newShard(
	shardID, totalShards int, token string, intents GatewayIntent,
	logger Logger, dispatcher *dispatcher, limiter ShardsIdentifyRateLimiter,
	compression GatewayCompression, properties IdentifyProperties,
	gatewayURL string, dialOptions *websocket.DialOptions,
) *Shard {
	ctx, cancel := context.WithCancel(context.Background())
	return &Shard{
		ctx:             ctx,
		cancel:          cancel,
		shardID:         shardID,
		totalShards:     totalShards,
		token:           cleanGatewayToken(token),
		intents:         intents,
		logger:          logger,
		dispatcher:      dispatcher,
		identifyLimiter: limiter,
		compression:     compression,
		properties:      properties,
		gatewayURL:      gatewayURL,
		dialOptions:     dialOptions,
		commandLimiter:  ratelimit.NewTokenBucket(120, 60*time.Second),
		presenceLimiter: ratelimit.NewTokenBucket(5, 20*time.Second),
		heartbeatStop:   make(chan struct{}),
		stopChan:        make(chan struct{}),
	}
}

// writeMessage safely writes a text message to the WebSocket connection
func (s *Shard) writeMessage(payload []byte) error {
	if s.closed.Load() {
		return errors.New("shard: connection is nil or closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.conn == nil {
		return errors.New("shard: connection is nil or closed")
	}
	return s.conn.Write(ctx, websocket.MessageText, payload)
}

// writeCommand writes an outbound gateway command with rate limiting
func (s *Shard) writeCommand(op gatewayOpcode, payload []byte) error {
	if op != gatewayOpcodeHeartbeat && s.commandLimiter != nil {
		for {
			if err := s.commandLimiter.Add(1); err == nil {
				break
			}
			wait := time.Until(s.commandLimiter.Reset())
			if wait <= 0 {
				continue
			}
			select {
			case <-time.After(wait):
			case <-s.stopChan:
				return errors.New("shard: connection stopped")
			case <-s.ctx.Done():
				return s.ctx.Err()
			}
		}
	}
	return s.writeMessage(payload)
}

// connect establishes or resumes a WebSocket connection to Discord Gateway.
func (s *Shard) connect(ctx context.Context) error {
	if s.heartbeatStop != nil {
		select {
		case <-s.heartbeatStop:
		default:
			close(s.heartbeatStop)
		}
	}
	s.heartbeatStop = make(chan struct{})

	s.writeMu.Lock()
	if s.conn != nil {
		_ = s.conn.CloseNow()
	}
	s.conn = nil
	s.writeMu.Unlock()

	s.zlibMu.Lock()
	if s.zlibReader != nil {
		_ = s.zlibReader.Close()
		s.zlibReader = nil
	}
	s.zlibBuf.Reset()
	s.zlibMu.Unlock()

	s.zstdMu.Lock()
	if s.zstdDecoder != nil {
		s.zstdDecoder.Close()
		s.zstdDecoder = nil
	}
	s.zstdMu.Unlock()

	connURL := s.ResumeURL()
	if connURL == "" {
		if s.gatewayURL != "" {
			connURL = s.gatewayURL
		} else {
			connURL = gatewayBaseURL
		}
	}
	connURL = s.buildGatewayURL(connURL)

	opts := s.dialOptions
	if opts == nil {
		opts = &websocket.DialOptions{
			HTTPClient: &http.Client{
				Timeout: 20 * time.Second,
			},
		}
	}

	conn, _, err := websocket.Dial(ctx, connURL, opts)
	if err != nil {
		return err
	}

	// disable read limit to allow large ready / guild payloads (up to 64MB)
	conn.SetReadLimit(64 * 1024 * 1024)

	s.logger.Info("connected to gateway", "url", connURL)
	s.writeMu.Lock()
	s.conn = conn
	s.writeMu.Unlock()

	s.lastHeartbeatACK.Store(true)
	s.latency.Store(0)

	go s.readLoop(conn)
	return nil
}

// buildGatewayURL appends the required query params to the gateway endpoint.
func (s *Shard) buildGatewayURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := parsed.Query()
	if q.Get("v") == "" {
		q.Set("v", gatewayVersion)
	}
	if q.Get("encoding") == "" {
		q.Set("encoding", "json")
	}
	if s.compression != GatewayCompressionNone && s.compression != "" && q.Get("compress") == "" {
		q.Set("compress", string(s.compression))
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

// readLoop continuously reads messages from the Gateway WebSocket.
func (s *Shard) readLoop(conn *websocket.Conn) {
	defer func() {
		s.writeMu.Lock()
		if s.conn == conn {
			_ = conn.CloseNow()
			s.conn = nil
		}
		s.writeMu.Unlock()
	}()

	for {
		if s.closed.Load() {
			return
		}

		typ, msg, err := conn.Read(context.Background())
		if err != nil {
			if s.closed.Load() {
				return
			}
			go s.handleDisconnectError(err)
			return
		}

		var payload gatewayPayload

		if typ == websocket.MessageBinary {
			var data []byte
			var err error
			if s.compression == GatewayCompressionZstdStream {
				data, err = s.decompressZstd(msg)
			} else {
				data, err = s.decompressZlib(msg)
			}
			if err != nil {
				s.logger.Error(
					"payload decompression failed",
					"compression",
					s.compression,
					"error",
					err,
				)
				s.reconnect(true)
				return
			}
			if data == nil {
				// chunk was buffered, waiting for complete flush
				continue
			}
			if err := json.Unmarshal(data, &payload); err != nil {
				s.logger.Error("unmarshal compressed json error", "error", err)
				continue
			}
		} else if typ == websocket.MessageText {
			if err := json.Unmarshal(msg, &payload); err != nil {
				s.logger.Error("unmarshal text json error", "error", err)
				continue
			}
		} else {
			continue
		}

		s.handleGatewayPayload(payload)
	}
}

// decompressZlib buffers binary chunks and decompresses when flush suffix is reached
func (s *Shard) decompressZlib(chunk []byte) ([]byte, error) {
	s.zlibMu.Lock()
	defer s.zlibMu.Unlock()

	s.zlibBuf.Write(chunk)

	// in discord zlib-stream, each full packet ends with flush suffix
	if len(chunk) < 4 || string(chunk[len(chunk)-4:]) != zlibFlushSuffix {
		return nil, nil
	}

	var err error
	if s.zlibReader == nil {
		s.zlibReader, err = zlib.NewReader(&s.zlibBuf)
		if err != nil {
			s.zlibBuf.Reset()
			return nil, err
		}
	}

	var out bytes.Buffer
	_, err = io.Copy(&out, s.zlibReader)
	s.zlibBuf.Reset()

	if err != nil {
		_ = s.zlibReader.Close()
		s.zlibReader = nil
		return nil, err
	}

	return out.Bytes(), nil
}

// decompressZstd decompresses a zstd-stream chunk
func (s *Shard) decompressZstd(chunk []byte) ([]byte, error) {
	s.zstdMu.Lock()
	defer s.zstdMu.Unlock()

	if s.zstdDecoder == nil {
		dec, err := zstd.NewReader(nil, zstd.WithDecoderLowmem(true))
		if err != nil {
			return nil, err
		}
		s.zstdDecoder = dec
	}

	if err := s.zstdDecoder.Reset(bytes.NewReader(chunk)); err != nil {
		return nil, err
	}

	return io.ReadAll(s.zstdDecoder)
}

// isFatalCloseCode returns true if the close code is unrecoverable without configuration change.
func isFatalCloseCode(code int) bool {
	switch GatewayCloseEventCode(code) {
	case GatewayCloseEventCodeAuthenticationFailed,
		GatewayCloseEventCodeInvalidShard,
		GatewayCloseEventCodeShardingRequired,
		GatewayCloseEventCodeInvalidAPIVersion,
		GatewayCloseEventCodeInvalidIntents,
		GatewayCloseEventCodeDisallowedIntents:
		return true
	default:
		return false
	}
}

// isReidentifyCloseCode returns true if the session must be re-identified rather than resumed.
func isReidentifyCloseCode(code int) bool {
	switch GatewayCloseEventCode(code) {
	case GatewayCloseEventCodeInvalidSeq,
		GatewayCloseEventCodeSessionTimedOut:
		return true
	default:
		return false
	}
}

func (s *Shard) handleDisconnectError(err error) {
	code := int(websocket.CloseStatus(err))
	var closeErr websocket.CloseError
	reason := ""
	if errors.As(err, &closeErr) {
		reason = closeErr.Reason
	}

	s.logger.Warn(
		"gateway connection closed",
		"code", code,
		"reason", reason,
		"error", err.Error(),
	)

	if isFatalCloseCode(code) {
		s.logger.Error("fatal gateway close code received, stopping reconnects", "code", code)
		return
	}

	if isReidentifyCloseCode(code) {
		s.logger.Info(
			"session invalidated by gateway, resetting session and re-identifying",
			"code",
			code,
		)
		s.reconnect(false)
		return
	}

	// reconnect and attempt resume
	s.reconnect(true)
}

func (s *Shard) handleGatewayPayload(payload gatewayPayload) {
	if payload.S > 0 {
		s.seq.Store(payload.S)
	}

	s.dispatcher.Dispatch(ShardID(s.shardID), payload.T, payload.D)

	switch payload.Op {
	case gatewayOpcodeDispatch:
		switch payload.T {
		case string(GatewayEventReady):
			var ready struct {
				SessionID        string `json:"session_id"`
				ResumeGatewayURL string `json:"resume_gateway_url"`
			}
			_ = json.Unmarshal(payload.D, &ready)
			s.setSession(ready.SessionID, ready.ResumeGatewayURL)
			s.logger.Info("READY received from gateway")
		case string(GatewayEventResumed):
			s.logger.Info("RESUMED received, session replayed with 0 lost events")
		}

	case gatewayOpcodeReconnect:
		s.logger.Info("RECONNECT opcode 7 received, reconnecting immediately to resume")
		go s.reconnect(true)

	case gatewayOpcodeInvalidSession:
		var resumable bool
		_ = json.Unmarshal(payload.D, &resumable)
		jitter := time.Duration(1000+rand.IntN(4000)) * time.Millisecond
		s.logger.Info("invalid session received", "resumable", resumable, "jitter", jitter.String())

		if !resumable {
			s.clearSession()
			s.seq.Store(0)
		}

		// wait jitter then reconnect to resume or identify
		go func() {
			time.Sleep(jitter)
			s.reconnect(resumable)
		}()

	case gatewayOpcodeHello:
		var hello struct {
			HeartbeatInterval float64 `json:"heartbeat_interval"`
		}
		_ = json.Unmarshal(payload.D, &hello)
		interval := time.Duration(hello.HeartbeatInterval) * time.Millisecond
		s.logger.Debug("HELLO received", "heartbeat_interval", interval.String())
		go s.startHeartbeat(interval)

		// send handshake asynchronously so rate limit wait does not block read loop
		go func() {
			sessionID, _ := s.getSession()
			if sessionID != "" && s.seq.Load() > 0 {
				s.logger.Info("resuming previous session")
				if err := s.sendResume(); err != nil {
					s.logger.Error("failed to send resume handshake", "error", err)
					go s.reconnect(false)
				}
			} else {
				s.logger.Debug("identifying new session")
				if err := s.sendIdentify(); err != nil {
					s.logger.Error("failed to send identify handshake", "error", err)
					go s.reconnect(false)
				}
			}
		}()

	case gatewayOpcodeHeartbeatACK:
		s.lastHeartbeatACK.Store(true)
		sent := s.lastHeartbeatSent.Load()
		if sent > 0 {
			rtt := time.Since(time.Unix(0, sent)).Milliseconds()
			s.latency.Store(rtt)
			s.logger.Debug("heartbeat ACK received", "rtt_ms", rtt)
		}

	case gatewayOpcodeHeartbeat:
		// discord requested immediate heartbeat
		_ = s.sendHeartbeat()
	}
}

// sendIdentify sends an Identify payload to Discord Gateway.
func (s *Shard) sendIdentify() error {
	payload, _ := json.Marshal(map[string]any{
		"op": gatewayOpcodeIdentify,
		"d": map[string]any{
			"token": s.token,
			"properties": map[string]string{
				"os":      s.properties.OS,
				"browser": s.properties.Browser,
				"device":  s.properties.Device,
			},
			"shard":   [2]int{s.shardID, s.totalShards},
			"intents": s.intents,
		},
	})
	if s.identifyLimiter != nil {
		if bl, ok := s.identifyLimiter.(interface {
			WaitShard(ctx context.Context, shardID int) error
		}); ok {
			if err := bl.WaitShard(s.ctx, s.shardID); err != nil {
				return err
			}
		} else if err := s.identifyLimiter.Wait(s.ctx); err != nil {
			return err
		}
	}
	return s.writeCommand(gatewayOpcodeIdentify, payload)
}

// sendResume sends a Resume payload to Discord Gateway
func (s *Shard) sendResume() error {
	sessionID, _ := s.getSession()
	payload, _ := json.Marshal(map[string]any{
		"op": gatewayOpcodeResume,
		"d": map[string]any{
			"token":      s.token,
			"session_id": sessionID,
			"seq":        s.seq.Load(),
		},
	})
	return s.writeCommand(gatewayOpcodeResume, payload)
}

// sendHeartbeat sends a Heartbeat payload to Discord Gateway.
func (s *Shard) sendHeartbeat() error {
	seq := s.seq.Load()
	var d any
	if seq > 0 {
		d = seq
	}
	payload, _ := json.Marshal(map[string]any{
		"op": gatewayOpcodeHeartbeat,
		"d":  d,
	})
	return s.writeMessage(payload)
}

// startHeartbeat begins sending heartbeats at the given interval.
func (s *Shard) startHeartbeat(interval time.Duration) {
	jitter := time.Duration(rand.Float64() * float64(interval))
	select {
	case <-time.After(jitter):
	case <-s.heartbeatStop:
		return
	}

	if err := s.sendHeartbeat(); err != nil {
		s.logger.Error("first heartbeat send failed", "error", err)
		return
	}
	s.lastHeartbeatACK.Store(false)
	s.lastHeartbeatSent.Store(time.Now().UnixNano())

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.heartbeatStop:
			return
		case <-ticker.C:
			if s.closed.Load() {
				return
			}

			if !s.lastHeartbeatACK.Load() {
				s.logger.Warn(
					"heartbeat ACK not received before next interval (zombie connection), reconnecting to resume",
				)
				go s.reconnect(true)
				return
			}

			s.lastHeartbeatACK.Store(false)
			s.lastHeartbeatSent.Store(time.Now().UnixNano())

			if err := s.sendHeartbeat(); err != nil {
				s.logger.Error("heartbeat send failed, reconnecting", "error", err)
				go s.reconnect(true)
				return
			}
		}
	}
}

// reconnect closes the current connection and attempts to reconnect with fast retry & exponential backoff.
func (s *Shard) reconnect(canResume bool) {
	if s.closed.Load() {
		return
	}
	if !s.reconnecting.CompareAndSwap(false, true) {
		return
	}
	defer s.reconnecting.Store(false)

	if !canResume {
		s.clearSession()
		s.seq.Store(0)
	}

	if s.heartbeatStop != nil {
		select {
		case <-s.heartbeatStop:
		default:
			close(s.heartbeatStop)
		}
	}

	s.writeMu.Lock()
	if s.conn != nil {
		_ = s.conn.CloseNow()
		s.conn = nil
	}
	s.writeMu.Unlock()

	// fast initial reconnect attempt (almost zero delay)
	backoff := 250 * time.Millisecond
	maxBackoff := 60 * time.Second

	for {
		if s.closed.Load() {
			return
		}
		s.logger.Info("attempting to reconnect and resume", "backoff", backoff.String())
		select {
		case <-time.After(backoff):
		case <-s.stopChan:
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		err := s.connect(ctx)
		cancel()

		if err == nil {
			s.logger.Info("reconnected to gateway successfully")
			return
		}

		s.logger.Error("reconnect attempt failed", "error", err)
		backoff = min(backoff*2, maxBackoff)
	}
}

// UpdatePresence sends an Opcode 3 Presence Update to Discord Gateway.
func (s *Shard) UpdatePresence(presence GatewayPresenceUpdate) error {
	if s.presenceLimiter != nil {
		for {
			if err := s.presenceLimiter.Add(1); err == nil {
				break
			}
			wait := time.Until(s.presenceLimiter.Reset())
			if wait <= 0 {
				continue
			}
			select {
			case <-time.After(wait):
			case <-s.stopChan:
				return errors.New("shard: connection stopped")
			case <-s.ctx.Done():
				return s.ctx.Err()
			}
		}
	}
	payload, err := json.Marshal(map[string]any{
		"op": gatewayOpcodePresenceUpdate,
		"d":  presence,
	})
	if err != nil {
		return err
	}
	return s.writeCommand(gatewayOpcodePresenceUpdate, payload)
}

// UpdateVoiceState sends an Opcode 4 Voice State Update to Discord Gateway.
func (s *Shard) UpdateVoiceState(
	guildID Snowflake,
	channelID *Snowflake,
	selfMute, selfDeaf bool,
) error {
	payload, err := json.Marshal(map[string]any{
		"op": gatewayOpcodeVoiceStateUpdate,
		"d": GatewayVoiceStateUpdate{
			GuildID:   guildID,
			ChannelID: channelID,
			SelfMute:  selfMute,
			SelfDeaf:  selfDeaf,
		},
	})
	if err != nil {
		return err
	}
	return s.writeCommand(gatewayOpcodeVoiceStateUpdate, payload)
}

// RequestGuildMembers sends an Opcode 8 Request Guild Members payload to Discord Gateway.
func (s *Shard) RequestGuildMembers(opts RequestGuildMembersOptions) error {
	payload, err := json.Marshal(map[string]any{
		"op": gatewayOpcodeRequestGuildMembers,
		"d":  opts,
	})
	if err != nil {
		return err
	}
	return s.writeCommand(gatewayOpcodeRequestGuildMembers, payload)
}

// RequestSoundboardSounds sends an Opcode 31 Request Soundboard Sounds payload to Discord Gateway.
func (s *Shard) RequestSoundboardSounds(guildIDs []Snowflake) error {
	payload, err := json.Marshal(map[string]any{
		"op": gatewayOpcodeRequestSoundboardSounds,
		"d": RequestSoundboardSoundsOptions{
			GuildIDs: guildIDs,
		},
	})
	if err != nil {
		return err
	}
	return s.writeCommand(gatewayOpcodeRequestSoundboardSounds, payload)
}

// RequestChannelInfo sends an Opcode 43 Request Channel Info payload to Discord Gateway.
func (s *Shard) RequestChannelInfo(opts RequestChannelInfoOptions) error {
	payload, err := json.Marshal(map[string]any{
		"op": gatewayOpcodeRequestChannelInfo,
		"d":  opts,
	})
	if err != nil {
		return err
	}
	return s.writeCommand(gatewayOpcodeRequestChannelInfo, payload)
}

// ShardID returns the zero-based shard ID.
func (s *Shard) ShardID() int {
	return s.shardID
}

// ID returns the zero-based shard ID (alias for ShardID).
func (s *Shard) ID() int {
	return s.shardID
}

// TotalShards returns the total number of shards configured for this shard.
func (s *Shard) TotalShards() int {
	return s.totalShards
}

// SessionID returns the current Gateway session ID.
func (s *Shard) SessionID() string {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()
	return s.sessionID
}

// ResumeURL returns the gateway resume URL provided by Discord in the READY event.
func (s *Shard) ResumeURL() string {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()
	return s.resumeURL
}

// Sequence returns the last sequence number received from Discord Gateway.
func (s *Shard) Sequence() int64 {
	return s.seq.Load()
}

// IsConnected returns true if the WebSocket connection is active and open.
func (s *Shard) IsConnected() bool {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.conn != nil && !s.closed.Load()
}

// Latency returns the current heartbeat round-trip latency as a time.Duration.
func (s *Shard) Latency() time.Duration {
	return time.Duration(s.latency.Load()) * time.Millisecond
}

// LatencyMillis returns the current heartbeat latency in milliseconds.
func (s *Shard) LatencyMillis() int64 {
	return s.latency.Load()
}

// Send sends a custom Gateway payload with opcode and data on this shard.
func (s *Shard) Send(op int, data any) error {
	payload, err := json.Marshal(map[string]any{
		"op": op,
		"d":  data,
	})
	if err != nil {
		return err
	}
	return s.writeCommand(gatewayOpcode(op), payload)
}

// Context returns the shard's internal lifecycle context.
func (s *Shard) Context() context.Context {
	return s.ctx
}

// Shutdown cleanly closes the shard's websocket connection.
func (s *Shard) Shutdown() error {
	s.shutdownOnce.Do(func() {
		s.closed.Store(true)
		if s.cancel != nil {
			s.cancel()
		}
		if s.stopChan != nil {
			close(s.stopChan)
		}
		if s.heartbeatStop != nil {
			select {
			case <-s.heartbeatStop:
			default:
				close(s.heartbeatStop)
			}
		}
		s.zstdMu.Lock()
		if s.zstdDecoder != nil {
			s.zstdDecoder.Close()
			s.zstdDecoder = nil
		}
		s.zstdMu.Unlock()
		s.writeMu.Lock()
		defer s.writeMu.Unlock()
		if s.conn != nil {
			s.logger.Info("shutting down shard connection")
			_ = s.conn.Close(websocket.StatusNormalClosure, "shutting down")
			_ = s.conn.CloseNow()
			s.conn = nil
		}
	})
	return nil
}
