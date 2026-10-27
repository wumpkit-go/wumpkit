---
title: "REST API & Rate Limits"
weight: 40
---

# REST API & Rate Limits

Wumpkit includes a high-performance HTTP client that communicates with Discord REST API v10, handles multipart file uploads, and manages rate limits pre-emptively.

* [Making REST Requests](client/): Performing HTTP calls, context timeouts, retry backoff, and error handling.
* [Rate Limiting & Buckets](rate-limits/): How Wumpkit tracks Discord route buckets, avoids 429 errors, and shares state via in-memory or Redis stores.
