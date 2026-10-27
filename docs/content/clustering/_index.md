---
title: "Sharding & Redis"
weight: 80
---

# Sharding & Redis

When your bot grows beyond 2,500 servers, Discord requires sharding to split Gateway WebSocket traffic across multiple connections.

wumpkit supports automatic sharding for single-process bots, manual shard routing for multi-process container clusters, and official Redis extensions to coordinate state, rate limits, and caches.

* [Manual & Auto Sharding](sharding/): Shard configuration, shard routing formulas, and monitoring shard latency.
* [Distributed Redis Cache](redis/): Sharing rate limits, Gateway identify limits, and caches across multiple bot processes using Redis.
