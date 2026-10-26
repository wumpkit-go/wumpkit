# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-10-26

### Added
- Structured `DiscordRestError` wrapping Discord REST API status codes, error codes, and field validation messages, supporting `errors.As`.
- Comprehensive mock HTTP test suite for Client REST methods across user, guild, channel, message, webhook, and command endpoints.
- JDA-style per-channel bounded message ring buffer in `InMemoryCacheManager` with configurable `MaxMessagesPerChannel` (default 100) preventing memory bloat.
- Dedicated Gateway presence rate limiter enforcing Discord's 5 updates per 20 seconds limit to avoid close code 4008.
- Thread-safe session synchronization on `Shard` for `SessionID()` and `ResumeURL()`.
- Discord API v10 permission `PermissionSetVoiceChannelStatus` (`1 << 48`).
- Shard pagination option in `FetchCurrentUserGuildsOptions` for bots in over 10,000 guilds.
- `FlagsNew` field in `Application` struct ensuring 64-bit flag safety.
- Complete `GroupDMChannel` fields including `Recipients`, `OwnerID`, and `ApplicationID`.
- Automated GitHub Actions workflows for changelog extraction, version bumping, and releases.

### Changed
- Replaced fixed stack buffer with pre-allocated dynamic `strings.Builder` in rate-limit route normalizer, preventing out-of-bounds panics on deep routes.
- Webhook interaction event handler now returns valid deferred callback payloads (`type: 5`, `type: 6`, `type: 8`) to satisfy Discord's 3-second acknowledgement window.
- `READY` event handler preserves populated guild cache data instead of overwriting with bare unavailable stubs.
- HTTP 429 handler parses JSON response body for `retry_after` and `global` when headers are absent.
- Redis rate limit store now uses atomic optimistic locking (`Watch` and `TxPipelined`) with automatic conflict retry.
- Updated default REST User-Agent to `DiscordBot (wumpkit, v1.0.0)`.

### Fixed
- Fixed dispatcher crash caused by `send on closed channel` during concurrent shutdown.
- Fixed infinite reconnect loops on Gateway Opcode 9 invalid session by immediately invalidating session state when non-resumable.
- Fixed unhandled handshake error on Gateway Opcode 10 Hello by triggering reconnect on failure.
- Fixed background ticker goroutine leak in `DefaultShardsRateLimiter` by adding idempotent `Close()`.
- Fixed memory bloat in multipart request `bufferPool` by capping pooled buffer capacity to 512KB.

## [0.18.1] - 2026-10-20

### Added
- Native Zstandard (`zstd-stream`) Gateway frame decompression using Klauspost zstd with streaming decoder reuse.
- Outbound Gateway command rate limiter enforcing Discord's 120 commands per 60 seconds limit while prioritizing heartbeats.
- Per-shard FIFO sequential worker event queues preventing race conditions during message bursts and gateway replay.

### Fixed
- Decoupled gateway socket disconnect handling and handshake routines from the reading loop to eliminate reader freezes.
- Fixed invalid session resume attempts over dropped sockets.

## [0.18.0] - 2026-10-20

### Added
- Interaction option value getters and subcommand inspection helpers.
- Partial member helpers and resolved channel models.
- Runnable example bots: `pingpong`, `slash_commands`, `components`, `jointocreate`, `sharded_bot`, and `redis_cache`.

### Changed
- Made REST request options optional via variadic arguments.

## [0.17.0] - 2026-08-22

### Added
- Autocomplete interaction support and `RespondAutocomplete` helper.
- Subcommand and subcommand group builders.
- `ParseAppIDFromToken` helper in snowflake package.

## [0.16.0] - 2026-08-22

### Added
- Rebranded library to Wumpkit.
- High-performance `shardedmap` concurrent map implementation.
- Distributed Redis cache manager, Redis rate limit store, and distributed identify rate limiter.
- Ed25519-verified HTTP webhook server for Discord Webhook Events.
- Modular client subsystem configuration architecture.
- Cache-aside query helpers on Client.

## [0.15.0] - 2025-12-31

### Added
- Gateway event models for all Discord dispatch events.
- Shard manager with multi-shard coordination and automatic sharding.
- Guild channel permission overwrite calculation helpers.
- Administrator permission override resolution.

## [0.14.5] - 2025-09-10

### Added
- Component message builders.
- Cache manager abstraction with `InMemoryCacheManager`.
- Gateway v10 default endpoint URL update.

## [0.14.0] - 2025-09-05

### Added
- Discord message component models: buttons, action rows, and select menus.
- Channel custom JSON serialization.
- Renamed project from `yada` to `dwaz`.

## [0.13.0] - 2025-08-28

### Added
- Channel thread support.
- Stage instance models and privacy level constants.
- Soundboard models and file helpers.
- Application command option types and models.
- Discord Poll models and options.
- Entitlement models for Discord monetization.

## [0.12.0] - 2025-08-12

### Added
- Guild models, guild discovery splash helpers, and guild tag badge URLs.
- Member banner and avatar decoration URL helpers.
- Voice state and voice region models.

## [0.11.0] - 2025-08-08

### Added
- Dynamic worker pool for asynchronous event execution.
- Base64 image file loader helper.

## [0.10.0] - 2025-08-03

### Added
- Role models, role colors, and role tags.
- Custom emoji models and mention formatters.
- Embed models and `EmbedBuilder`.
- Sticker and sticker pack models.

## [0.9.0] - 2025-07-25

### Added
- Channel type models and helpers.
- `X-Audit-Log-Reason` request header support.
- Color type and standard Discord color palette constants.

## [0.8.0] - 2025-07-23

### Added
- Permissions bitmask type with permission flags and bitwise operations.
- User display name, avatar, and mention formatting helpers.

## [0.7.0] - 2025-07-22

### Added
- Per-resource rate limiting with major parameter bucket hashing.
- Image URL builders for all Discord CDN resources.

## [0.6.0] - 2025-07-21

### Added
- Cluster manager for multi-shard bot coordination.
- Discord User models and REST endpoints.

## [0.5.0] - 2025-07-21

### Added
- WebSocket shard management with connection loop, heartbeats, and identify rate limiter.

## [0.4.0] - 2025-07-21

### Added
- Gateway intent constants, gateway opcodes, and close event codes.

## [0.3.0] - 2025-07-19

### Added
- `MESSAGE_CREATE` and `MESSAGE_DELETE` gateway event structures.
- Initial event dispatcher and handler manager.

## [0.2.0] - 2025-07-17

### Added
- Gateway bot structure and REST API call wrapper.

## [0.1.0] - 2025-07-17

### Added
- Initial library foundation: Snowflake type, logger interface, and HTTP requester.

[Unreleased]: https://github.com/wumpkit-go/wumpkit/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.18.1...v1.0.0
[0.18.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.18.0...v0.18.1
[0.18.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.16.0...v0.17.0
[0.16.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.14.5...v0.15.0
[0.14.5]: https://github.com/wumpkit-go/wumpkit/compare/v0.14.4...v0.14.5
[0.14.4]: https://github.com/wumpkit-go/wumpkit/compare/v0.14.3...v0.14.4
[0.14.3]: https://github.com/wumpkit-go/wumpkit/compare/v0.14.2...v0.14.3
[0.14.2]: https://github.com/wumpkit-go/wumpkit/compare/v0.14.1...v0.14.2
[0.14.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.14.0...v0.14.1
[0.14.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.13.2...v0.14.0
[0.13.2]: https://github.com/wumpkit-go/wumpkit/compare/v0.13.1...v0.13.2
[0.13.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.13.0...v0.13.1
[0.13.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.12.1...v0.13.0
[0.12.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.12.0...v0.12.1
[0.12.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.10.1...v0.11.0
[0.10.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.10.0...v0.10.1
[0.10.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.9.2...v0.10.0
[0.9.2]: https://github.com/wumpkit-go/wumpkit/compare/v0.9.1...v0.9.2
[0.9.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/wumpkit-go/wumpkit/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/wumpkit-go/wumpkit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/wumpkit-go/wumpkit/releases/tag/v0.1.0
