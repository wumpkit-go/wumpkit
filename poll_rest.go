/************************************************************************************
 *
 * wumpkit, A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD 3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package wumpkit

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// FetchAnswerVotersOptions contains parameters for fetching users who voted for a specific poll answer.
//
// Reference: https://discord.com/developers/docs/resources/poll#get-answer-voters
type FetchAnswerVotersOptions struct {
	// After gets users after this user ID.
	After Snowflake `json:"after,omitempty"`

	// Limit is max number of users to return (1-100, default 25).
	Limit int `json:"limit,omitempty"`
}

// FetchAnswerVoters returns a list of users that voted for the specific answer.
//
// Reference: https://discord.com/developers/docs/resources/poll#get-answer-voters
func (r *requester) FetchAnswerVoters(ctx context.Context, channelID, messageID Snowflake, answerID int, opts ...FetchAnswerVotersOptions) ([]User, error) {
	var first FetchAnswerVotersOptions
	if len(opts) > 0 {
		first = opts[0]
	}
	params := url.Values{}
	if first.After != 0 {
		params.Set("after", first.After.String())
	}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}

	endpoint := "/channels/" + channelID.String() + "/polls/" + messageID.String() + "/answers/" + strconv.Itoa(answerID)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp PollAnswerVotersResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/channels/{id}/polls/{id}/answers/{id}",
				"error", err.Error(),
			)
		return nil, err
	}
	return resp.Users, nil
}

// EndPoll immediately ends a poll. You cannot end polls from other users.
//
// Returns the updated message object.
//
// Reference: https://discord.com/developers/docs/resources/poll#end-poll
func (r *requester) EndPoll(ctx context.Context, channelID, messageID Snowflake) (Message, error) {
	endpoint := "/channels/" + channelID.String() + "/polls/" + messageID.String() + "/expire"
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    endpoint,
	})
	if err != nil {
		return Message{}, err
	}
	defer body.Close()

	var msg Message
	if err := json.NewDecoder(body).Decode(&msg); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/channels/{id}/polls/{id}/expire",
				"error", err.Error(),
			)
		return Message{}, err
	}
	msg.ChannelID = channelID
	return msg, nil
}
