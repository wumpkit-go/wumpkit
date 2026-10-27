---
title: "Markdown & Mentions"
weight: 1
---

# Discord Markdown & Mentions

Discord supports a customized version of Markdown for chat messages, embeds, and modal dialogs. Wumpkit provides type-safe helpers to format text and generate mention strings.

## Text Styling Helpers

```go
package main

import (
	"fmt"

	"github.com/wumpkit/wumpkit"
)

func main() {
	bold := wumpkit.Bold("Important notice")           // **Important notice**
	italic := wumpkit.Italic("curved text")             // *curved text*
	underline := wumpkit.Underline("underlined")        // __underlined__
	strike := wumpkit.Strikethrough("deleted")          // ~~deleted~~
	spoiler := wumpkit.Spoiler("secret plot twist")     // ||secret plot twist||
	code := wumpkit.InlineCode("err != nil")            // `err != nil`
	subtext := wumpkit.Subtext("small footer note")     // -# small footer note

	// multi-line block quote
	quote := wumpkit.BlockQuote("First line\nSecond line")

	// syntax-highlighted code block
	block := wumpkit.CodeBlock("go", "func main() {\n\tprintln(\"Hello\")\n}")

	fmt.Println(bold)
	fmt.Println(spoiler)
	fmt.Println(quote)
	fmt.Println(block)
}
```

## Dynamic Timestamps

Discord dynamically formats timestamps in chat according to each user's local timezone and device language settings. Use `FormatTimestamp`:

```go
package main

import (
	"fmt"
	"time"

	"github.com/wumpkit/wumpkit"
)

func main() {
	eventTime := time.Now().Add(2 * time.Hour)

	// relative time (e.g. "in 2 hours")
	relative := wumpkit.FormatTimestamp(eventTime, wumpkit.TimestampStyleRelative)

	// short time (e.g. "16:20")
	shortTime := wumpkit.FormatTimestamp(eventTime, wumpkit.TimestampStyleShortTime)

	// long date and time (e.g. "Tuesday, 20 April 2026 16:20")
	longDateTime := wumpkit.FormatTimestamp(eventTime, wumpkit.TimestampStyleLongDateTime)

	// default style (omitting style parameter defaults to ShortDateTime)
	defaultTime := wumpkit.FormatTimestamp(eventTime)

	fmt.Printf("Event starts %s (%s)\n", relative, defaultTime)
}
```

Available timestamp styles:

| Style Constant | Format Example | Discord Code |
| :--- | :--- | :--- |
| `TimestampStyleShortTime` | 16:20 | `<t:1620000000:t>` |
| `TimestampStyleLongTime` | 16:20:30 | `<t:1620000000:T>` |
| `TimestampStyleShortDate` | 20/04/2026 | `<t:1620000000:d>` |
| `TimestampStyleLongDate` | 20 April 2026 | `<t:1620000000:D>` |
| `TimestampStyleShortDateTime` | 20 April 2026 16:20 (Default) | `<t:1620000000:f>` |
| `TimestampStyleLongDateTime` | Tuesday, 20 April 2026 16:20 | `<t:1620000000:F>` |
| `TimestampStyleRelative` | in 2 hours / 5 minutes ago | `<t:1620000000:R>` |

## Mention Formatting

Format snowflakes into clickable Discord mention strings:

```go
userID := wumpkit.Snowflake(123456789012345678)
channelID := wumpkit.Snowflake(234567890123456789)
roleID := wumpkit.Snowflake(345678901234567890)

userMention := wumpkit.FormatUserMention(userID)       // <@123456789012345678>
channelMention := wumpkit.FormatChannelMention(channelID) // <#234567890123456789>
roleMention := wumpkit.FormatRoleMention(roleID)       // <@&345678901234567890>
```
