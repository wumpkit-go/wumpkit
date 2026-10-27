---
title: "Installation"
weight: 1
---

# Installation

## Requirements

Wumpkit requires Go 1.27 or newer. You can check your Go version in your terminal:

```bash
go version
```

If you are on an older version of Go, download the latest release from the official Go website (golang.org).

## Add Wumpkit to Your Project

In your Go project folder, initialize a module if you have not done so already:

```bash
mkdir mybot
cd mybot
go mod init mybot
```

Then download Wumpkit:

```bash
go get github.com/wumpkit/wumpkit
```

## Optional Redis Support

If you want to run multiple bot processes that share rate limits and cache over Redis, you can also install the official Redis client:

```bash
go get github.com/redis/go-redis/v9
```

Wumpkit does not require Redis for single-process bots. In-memory caching and in-memory rate limiting are included by default.
