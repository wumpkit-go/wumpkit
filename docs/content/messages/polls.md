---
title: "Native Polls"
weight: 40
---

# Native Polls

Discord messages support interactive polls natively. Users can vote for answers directly inside the client without using reactions or buttons.

wumpkit lets you create polls, inspect votes, and close polls when voting ends.

## Creating a Poll

To add a poll to a message, pass a `PollCreateOptions` struct to the `Poll` field in `CreateMessageOptions`:

```go
package main

import (
	"context"

	"github.com/wumpkit/wumpkit"
)

func createServerPoll(ctx context.Context, client *wumpkit.Client, channelID wumpkit.Snowflake) {
	poll := &wumpkit.PollCreateOptions{
		Question: wumpkit.PollMedia{
			Text: "What is your favorite programming language?",
		},
		Answers: []wumpkit.PollAnswer{
			{
				PollMedia: wumpkit.PollMedia{
					Text: "Go",
				},
			},
			{
				PollMedia: wumpkit.PollMedia{
					Text: "Rust",
				},
			},
			{
				PollMedia: wumpkit.PollMedia{
					Text: "TypeScript",
				},
			},
			{
				PollMedia: wumpkit.PollMedia{
					Text: "Python",
				},
			},
		},
		Duration:         24, // poll stays open for 24 hours
		AllowMultiselect: false,
	}

	_, _ = client.CreateMessage(ctx, channelID, wumpkit.CreateMessageOptions{
		Content: "Community vote is now live!",
		Poll:    poll,
	})
}
```

## Adding Emojis to Answers

You can attach an emoji to each answer option:

```go
Answers: []wumpkit.PollAnswer{
	{
		PollMedia: wumpkit.PollMedia{
			Text: "Pizza",
			Emoji: &wumpkit.PartialEmoji{
				Name: "🍕",
			},
		},
	},
	{
		PollMedia: wumpkit.PollMedia{
			Text: "Burgers",
			Emoji: &wumpkit.PartialEmoji{
				Name: "🍔",
			},
		},
	},
}
```

## Ending a Poll Early

You can immediately expire and close a poll before its natural duration runs out using `EndPoll`:

```go
updatedMessage, err := client.EndPoll(ctx, channelID, messageID)
if err != nil {
	log.Printf("failed to end poll: %v", err)
	return
}

if updatedMessage.Poll != nil && updatedMessage.Poll.Results != nil {
	log.Printf("poll ended with finalized: %v", updatedMessage.Poll.Results.IsFinalized)
}
```

## Fetching Voters for an Answer

To see which users voted for a specific answer, call `FetchAnswerVoters` with the answer number (1 for the first option):

```go
voters, err := client.FetchAnswerVoters(ctx, channelID, messageID, 1, wumpkit.FetchAnswerVotersOptions{
	Limit: 50,
})
if err != nil {
	log.Printf("failed to fetch voters: %v", err)
	return
}

for _, voter := range voters {
	log.Printf("User %s voted for answer 1", voter.Username)
}
```
