---
title: "State & Caching"
weight: 30
---

# State & Caching

Wumpkit includes a concurrent, thread-safe in-memory cache system that keeps Discord state synchronized as Gateway events arrive.

* [How Caching Works](overview/): Understand the in-memory cache architecture and automatic lifecycle updates.
* [Controlling Cached Entities](cache-flags/): How to toggle specific entity caches (messages, emojis, voice states) to minimize memory consumption.
* [Cache-Aside Pattern](cache-aside/): Choosing between cache-aside getters (`client.User`), in-memory lookups (`client.GetUser`), and direct REST calls (`client.FetchUser`).
