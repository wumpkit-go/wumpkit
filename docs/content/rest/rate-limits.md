---
title: "Rate Limiting & Buckets"
weight: 2
---

# Rate Limiting & Buckets

Discord uses bucket-based rate limits to prevent API abuse. Each route belongs to a bucket that permits a specific number of requests within a time window.

Wumpkit includes a **pre-emptive rate limiter**. Instead of waiting for Discord to return an HTTP `429 Too Many Requests` error and then sleeping, Wumpkit tracks remaining requests locally. When a bucket is exhausted, subsequent requests pause automatically until the bucket resets.

## How Buckets Work

Discord routes are grouped by their route pattern and major parameters:
- Route paths like `/channels/{channel_id}/messages` share limits per `channel_id`.
- Guild management routes share limits per `guild_id`.
- Webhook routes share limits per `webhook_id`.

When Wumpkit makes an HTTP call, it reads response headers:
- `X-RateLimit-Bucket`: Unique hash identifying the bucket
- `X-RateLimit-Limit`: Maximum requests allowed in the current window
- `X-RateLimit-Remaining`: Remaining requests before rate limiting kicks in
- `X-RateLimit-Reset-After`: Seconds until the bucket resets

Wumpkit saves this metadata to avoid sending requests that are guaranteed to fail.

## Pluggable Rate Limit Stores

The rate limiter abstracts storage using the `RateLimitStore` interface.

### In-Memory Store (Default)
For standard, single-process bots, Wumpkit uses `NewMemoryRateLimitStore()`. It uses atomic timers and memory maps with zero configuration.

### Redis Store (Multi-Cluster Setup)
If you run multiple bot instances or separate Gateway and REST worker processes, memory stores cannot communicate with each other. You can plug in Redis so all processes share a single, synchronized rate limit bucket queue:

```go
import (
	"github.com/redis/go-redis/v9"
	wumpredis "github.com/wumpkit/wumpkit/extra/redis"
)

rdb := redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

redisStore := wumpredis.NewRateLimitStore(rdb, wumpredis.RateLimitStoreConfig{
	KeyPrefix: "wumpkit:ratelimit:",
})

bot := wumpkit.New(
	ctx,
	wumpkit.WithToken(token),
	wumpkit.WithRateLimitStore(redisStore),
)
```
