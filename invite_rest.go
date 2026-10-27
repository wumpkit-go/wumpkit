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
	"strconv"
)

// FetchInviteOptions contains parameters for fetching a invite.
type FetchInviteOptions struct {
	// WithCounts is whether the invite should contain approximate member counts.
	WithCounts bool `json:"with_counts"`

	// GuildScheduledEventID is the guild scheduled event to include with the invite.
	GuildScheduledEventID Snowflake `json:"guild_scheduled_event_id,omitempty"`
}

// DeleteInviteOptions contains parameters for deleting invite.
type DeleteInviteOptions struct {
	// Reason is the reason shown in the audit log for this action.
	Reason string `json:"-"`
}

// FetchInvite retrieves a invite by its code.
func (r *requester) FetchInvite(ctx context.Context, code string, opts FetchInviteOptions) (Invite, error) {
	endpoint := "/invites/" + code + "?with_counts=" + strconv.FormatBool(opts.WithCounts)
	if opts.GuildScheduledEventID > 0 {
		endpoint += "&guild_scheduled_event_id=" + opts.GuildScheduledEventID.String()
	}
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    endpoint,
	})
	if err != nil {
		return Invite{}, err
	}
	defer body.Close()

	var invite Invite
	if err := json.NewDecoder(body).Decode(&invite); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/invites/{code}",
				"error", err.Error(),
			)
		return Invite{}, err
	}
	return invite, nil
}

// DeleteInvite deletes an invite.
//
// Requires the PermissionManageChannels permission on the channel this invite belongs to, or PermissionManageGuild to remove any invite across the guild.
func (r *requester) DeleteInvite(ctx context.Context, code string, opts DeleteInviteOptions) (Invite, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/invites/" + code,
		Reason: opts.Reason,
	})
	if err != nil {
		return Invite{}, err
	}
	defer body.Close()

	var invite Invite
	if err := json.NewDecoder(body).Decode(&invite); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "DELETE",
				"url", "/invites/{code}",
				"error", err.Error(),
			)
		return Invite{}, err
	}
	return invite, nil
}
