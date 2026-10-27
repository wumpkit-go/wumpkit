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
)

// GatewayInfo represents the basic gateway connection information.
type GatewayInfo struct {
	URL string `json:"url"`
}

// FetchGateway retrieves the standard Gateway WebSocket URL.
//
// Reference: https://discord.com/developers/docs/topics/gateway#get-gateway
func (r *requester) FetchGateway(ctx context.Context) (GatewayInfo, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/gateway",
	})
	if err != nil {
		return GatewayInfo{}, err
	}
	defer body.Close()

	var obj GatewayInfo
	if err := json.NewDecoder(body).Decode(&obj); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/gateway",
			"error", err.Error(),
		)
		return GatewayInfo{}, err
	}
	return obj, nil
}

// FetchGatewayBot retrieves bot gateway information including recommended shard count and session limits.
//
// Reference: https://discord.com/developers/docs/topics/gateway#get-gateway-bot
func (r *requester) FetchGatewayBot(ctx context.Context) (GatewayBot, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/gateway/bot",
	})
	if err != nil {
		return GatewayBot{}, err
	}
	defer body.Close()

	var obj GatewayBot
	if err := json.NewDecoder(body).Decode(&obj); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/gateway/bot",
			"error", err.Error(),
		)
		return GatewayBot{}, err
	}
	return obj, nil
}
