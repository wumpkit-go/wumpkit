---
title: "Sharding & Redis"
weight: 40
---

# Sharding & Redis

When your bot grows beyond 2,500 servers, Discord requires sharding to split the Gateway connection across multiple WebSocket sockets.

* [Manual & Auto Sharding](/clustering/sharding/): Shard configuration, shard routing formulas, and monitoring shard latency.
* [Distributed Redis Cache](/clustering/redis/): Sharing rate limits, Gateway identify limits, and caches across multiple bot processes using Redis.
