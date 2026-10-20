/************************************************************************************
 *
 * wumpkit, A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD-3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package ratelimit

import (
	"cmp"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

// Major parameter constants used for Discord rate limit bucket routing.
// These identify the top-level resource in a Discord API path that influences
// per-route rate limit isolation.
const (
	MajorUnknown      = "unk"
	MajorChannels     = "channels"
	MajorGuilds       = "guilds"
	MajorWebhooks     = "webhooks"
	MajorInvites      = "invites"
	MajorInteractions = "interactions"
)

// epochDiscord is the Discord epoch used for Snowflake ID timestamp extraction (2015-01-01T00:00:00Z).
const epochDiscord = 1420070400000

// interactionTokenPrefix is the base64 encoding of "interaction:" which prefixes interaction tokens.
const interactionTokenPrefix = "aW50ZXJhY3Rpb246"

// SnowflakeCreatedAt extracts the creation timestamp from a Discord Snowflake ID string.
func SnowflakeCreatedAt(snowflake string) (time.Time, error) {
	parsedID, err := strconv.ParseUint(snowflake, 10, 64)
	if err != nil {
		return time.Now(), err
	}
	epoch := (parsedID >> 22) + epochDiscord
	return time.UnixMilli(int64(epoch)), nil
}

// IsSnowflake reports whether str looks like a valid Discord Snowflake ID (17-20 digit numeric string).
func IsSnowflake(str string) bool {
	l := len(str)
	if l < 17 || l > 20 {
		return false
	}
	for i := range str {
		c := str[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// IsInteractionRoute reports whether the given normalized bucket path corresponds to a
// Discord interaction endpoint. Per Discord documentation, interaction endpoints are NOT
// bound to the bot's global rate limit (50 req/s), so they must bypass global throttling.
//
// Matched routes:
//   - /interactions/{id}/{token}/callback
//   - /webhooks/{id}/{interaction_id}/... (webhook followups with decoded interaction tokens)
func IsInteractionRoute(bucketPath string) bool {
	return strings.HasPrefix(bucketPath, "/interactions/") ||
		isWebhookInteractionFollowup(bucketPath)
}

// isWebhookInteractionFollowup checks if a webhook path is an interaction followup.
// These are paths like /webhooks/{app_id}/{interaction_id}/messages/...
// where the interaction ID was decoded from a base64 interaction token.
func isWebhookInteractionFollowup(path string) bool {
	if !strings.HasPrefix(path, "/webhooks/") {
		return false
	}

	// path format: /webhooks/{app_id}/{decoded_interaction_id}/...
	// the decoded interaction ID is a snowflake inserted by NormalizeRoute
	// when it decodes a base64 interaction token
	rest := path[len("/webhooks/"):]

	// skip webhook_id (first snowflake)
	slashIdx := strings.IndexByte(rest, '/')
	if slashIdx == -1 {
		return false
	}
	rest = rest[slashIdx+1:]

	// the next segment should be a decoded interaction snowflake ID (not "!" which is a regular token)
	nextSlash := strings.IndexByte(rest, '/')
	var segment string
	if nextSlash == -1 {
		segment = rest
	} else {
		segment = rest[:nextSlash]
	}

	// regular webhook tokens are replaced with "!" by NormalizeRoute
	// interaction tokens are decoded to their interaction snowflake ID
	// so if this segment is a snowflake, it's an interaction followup
	return IsSnowflake(segment)
}

// NormalizeRoute resolves the canonical route bucket path for a given Discord URL and method.
// It uses stack-allocated arrays for zero-allocation path tokenization and produces a canonical
// representation that groups routes into the same rate limit bucket per Discord's rules.
//
// Major parameters (channel_id, guild_id, webhook_id) are preserved in the path since they
// represent independent rate limit buckets. Minor parameters (message_id, role_id, etc.) are
// replaced with "!" since they share buckets within their major parameter scope.
func NormalizeRoute(url string, method string) string {
	// 1. strip Query String
	if qIdx := strings.IndexByte(url, '?'); qIdx != -1 {
		url = url[:qIdx]
	}

	// 2. strip /api/v* or /api/ prefix
	var cleanPath string
	if strings.HasPrefix(url, "/api/v") {
		rest := url[6:]
		if slashIdx := strings.IndexByte(rest, '/'); slashIdx != -1 {
			cleanPath = rest[slashIdx+1:]
		} else {
			cleanPath = rest
		}
	} else if strings.HasPrefix(url, "/api/") {
		cleanPath = url[5:]
	} else if strings.HasPrefix(url, "/") {
		cleanPath = url[1:]
	} else {
		cleanPath = url
	}

	// 3. tokenize path segments into stack array (max depth 16)
	var parts [16]string
	numParts := 0
	rem := cleanPath
	for len(rem) > 0 && numParts < 16 {
		if slashIdx := strings.IndexByte(rem, '/'); slashIdx != -1 {
			parts[numParts] = rem[:slashIdx]
			rem = rem[slashIdx+1:]
		} else {
			parts[numParts] = rem
			rem = ""
		}
		numParts++
	}

	if numParts <= 1 {
		if strings.HasPrefix(cleanPath, "/") {
			return cleanPath
		}
		return "/" + cleanPath
	}

	// 4. build normalized bucket into stack buffer
	var out [160]byte
	w := 0
	out[w] = '/'
	w++

	writeStr := func(s string) {
		w += copy(out[w:], s)
	}

	currMajor := MajorUnknown

	switch parts[0] {
	case MajorChannels:
		writeStr("channels/")
		writeStr(parts[1])
		if numParts == 2 {
			return string(out[:w])
		}
		currMajor = MajorChannels
	case MajorInvites:
		writeStr("invites/!")
		currMajor = MajorInvites
	case MajorGuilds:
		if numParts == 3 && parts[2] == "channels" {
			writeStr("guilds/")
			writeStr(parts[1])
			writeStr("/channels")
			return string(out[:w])
		}
		writeStr(parts[0])
		out[w] = '/'
		w++
		writeStr(parts[1])
		currMajor = parts[0]
	case MajorInteractions:
		if numParts == 4 && parts[3] == "callback" {
			writeStr("interactions/")
			writeStr(parts[1])
			writeStr("/!/callback")
			return string(out[:w])
		}
		writeStr(parts[0])
		out[w] = '/'
		w++
		writeStr(parts[1])
		currMajor = parts[0]
	default:
		writeStr(parts[0])
		out[w] = '/'
		w++
		writeStr(parts[1])
		currMajor = parts[0]
	}

	if numParts == 2 {
		return string(out[:w])
	}

	// 5. process sub-resources
	for _, part := range parts[2:numParts] {
		if IsSnowflake(part) {
			// discord 14-day message delete rate limit rule
			if currMajor == MajorChannels && numParts == 4 && parts[2] == "messages" && method == "DELETE" {
				createdAt, _ := SnowflakeCreatedAt(part)
				now := time.Now()
				if createdAt.Before(now.Add(-14 * 24 * time.Hour)) {
					writeStr("/!14dmsg")
				} else if createdAt.After(now.Add(-10 * time.Second)) {
					writeStr("/!10smsg")
				} else {
					writeStr("/!")
				}
				continue
			}
			writeStr("/!")
		} else {
			if currMajor == MajorChannels && part == "reactions" {
				if method == "PUT" || method == "DELETE" {
					writeStr("/reactions/!modify")
					break
				}
				writeStr("/reactions/!/!")
				break
			}

			// webhook tokens and base64 interaction tokens (always >= 64 chars)
			if len(part) >= 64 {
				if !strings.HasPrefix(part, interactionTokenPrefix) {
					// regular webhook token — replace with placeholder
					writeStr("/!")
					continue
				}

				// interaction token: decode to extract the interaction snowflake ID
				// so interaction followup webhooks resolve to a per-interaction bucket
				padded := part
				if remLen := len(part) % 4; remLen != 0 {
					padded += strings.Repeat("=", 4-remLen)
				}

				var interactionID string
				if decoded, err := base64.StdEncoding.DecodeString(padded); err == nil {
					split := strings.Split(string(decoded), ":")
					if len(split) > 1 {
						interactionID = split[1]
					}
				}
				interactionID = cmp.Or(interactionID, "Unknown")

				out[w] = '/'
				w++
				writeStr(interactionID)
				continue
			}

			out[w] = '/'
			w++
			writeStr(part)
		}
	}

	return string(out[:w])
}