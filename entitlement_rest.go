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
	"strings"
)

// ListEntitlementsOptions contains query parameters for listing entitlements.
//
// Reference: https://discord.com/developers/docs/resources/entitlement#list-entitlements
type ListEntitlementsOptions struct {
	// UserID is the user ID to look up entitlements for.
	UserID Snowflake `json:"user_id,omitempty"`

	// SKUIDs is the list of SKU IDs to check entitlements for.
	SKUIDs []Snowflake `json:"sku_ids,omitempty"`

	// Before retrieves entitlements before this entitlement ID.
	Before Snowflake `json:"before,omitempty"`

	// After retrieves entitlements after this entitlement ID.
	After Snowflake `json:"after,omitempty"`

	// Limit is the number of entitlements to return (1-100, default 100).
	Limit int `json:"limit,omitempty"`

	// GuildID is the guild ID to look up entitlements for.
	GuildID Snowflake `json:"guild_id,omitempty"`

	// ExcludeEnded specifies whether ended entitlements should be omitted.
	ExcludeEnded bool `json:"exclude_ended,omitempty"`

	// ExcludeDeleted specifies whether deleted entitlements should be omitted.
	ExcludeDeleted *bool `json:"exclude_deleted,omitempty"`
}

// CreateTestEntitlementOptions contains parameters for creating a test entitlement.
//
// Reference: https://discord.com/developers/docs/resources/entitlement#create-test-entitlement
type CreateTestEntitlementOptions struct {
	// SKUID is the ID of the SKU to grant the entitlement to.
	SKUID Snowflake `json:"sku_id"`

	// OwnerID is the ID of the guild or user to grant the entitlement to.
	OwnerID Snowflake `json:"owner_id"`

	// OwnerType is 1 for a guild subscription, 2 for a user subscription.
	OwnerType int `json:"owner_type"`
}

// ListSKUSubscriptionsOptions contains query parameters for listing subscriptions for a SKU.
//
// Reference: https://discord.com/developers/docs/resources/subscription#list-sku-subscriptions
type ListSKUSubscriptionsOptions struct {
	// Before lists subscriptions before this ID.
	Before Snowflake `json:"before,omitempty"`

	// After lists subscriptions after this ID.
	After Snowflake `json:"after,omitempty"`

	// Limit is the number of results to return (1-100, default 50).
	Limit int `json:"limit,omitempty"`

	// UserID is the user ID for which to return subscriptions.
	UserID Snowflake `json:"user_id,omitempty"`
}

// ListSKUs returns all SKUs for a given application.
//
// Reference: https://discord.com/developers/docs/resources/sku#list-skus
func (r *requester) ListSKUs(ctx context.Context, applicationID Snowflake) ([]SKU, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/applications/" + applicationID.String() + "/skus",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var skus []SKU
	if err := json.NewDecoder(body).Decode(&skus); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/applications/{id}/skus",
				"error", err.Error(),
			)
		return nil, err
	}
	return skus, nil
}

// ListEntitlements returns all entitlements for a given app, active and expired.
//
// Reference: https://discord.com/developers/docs/resources/entitlement#list-entitlements
func (r *requester) ListEntitlements(
	ctx context.Context,
	applicationID Snowflake,
	opts ...ListEntitlementsOptions,
) ([]Entitlement, error) {
	var first ListEntitlementsOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	params := url.Values{}
	if first.UserID != 0 {
		params.Set("user_id", first.UserID.String())
	}
	if len(first.SKUIDs) > 0 {
		var sb strings.Builder
		for i, id := range first.SKUIDs {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(id.String())
		}
		params.Set("sku_ids", sb.String())
	}
	if first.Before != 0 {
		params.Set("before", first.Before.String())
	}
	if first.After != 0 {
		params.Set("after", first.After.String())
	}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if first.GuildID != 0 {
		params.Set("guild_id", first.GuildID.String())
	}
	if first.ExcludeEnded {
		params.Set("exclude_ended", "true")
	}
	if first.ExcludeDeleted != nil {
		params.Set("exclude_deleted", strconv.FormatBool(*first.ExcludeDeleted))
	}

	endpoint := "/applications/" + applicationID.String() + "/entitlements"
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

	var entitlements []Entitlement
	if err := json.NewDecoder(body).Decode(&entitlements); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/applications/{id}/entitlements",
				"error", err.Error(),
			)
		return nil, err
	}
	return entitlements, nil
}

// FetchEntitlement returns an entitlement for a given app.
//
// Reference: https://discord.com/developers/docs/resources/entitlement#get-entitlement
func (r *requester) FetchEntitlement(
	ctx context.Context,
	applicationID, entitlementID Snowflake,
) (Entitlement, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/applications/" + applicationID.String() + "/entitlements/" + entitlementID.String(),
	})
	if err != nil {
		return Entitlement{}, err
	}
	defer body.Close()

	var ent Entitlement
	if err := json.NewDecoder(body).Decode(&ent); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/applications/{id}/entitlements/{id}",
				"error", err.Error(),
			)
		return Entitlement{}, err
	}
	return ent, nil
}

// ConsumeEntitlement marks a given entitlement for the user as consumed.
//
// Reference: https://discord.com/developers/docs/resources/entitlement#consume-an-entitlement
func (r *requester) ConsumeEntitlement(
	ctx context.Context,
	applicationID, entitlementID Snowflake,
) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/applications/" + applicationID.String() + "/entitlements/" + entitlementID.String() + "/consume",
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// CreateTestEntitlement creates a test entitlement to a given SKU for a given guild or user.
//
// Reference: https://discord.com/developers/docs/resources/entitlement#create-test-entitlement
func (r *requester) CreateTestEntitlement(
	ctx context.Context,
	applicationID Snowflake,
	opts CreateTestEntitlementOptions,
) (Entitlement, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/applications/" + applicationID.String() + "/entitlements",
		Body:   reqBody,
	})
	if err != nil {
		return Entitlement{}, err
	}
	defer body.Close()

	var ent Entitlement
	if err := json.NewDecoder(body).Decode(&ent); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "POST",
				"url", "/applications/{id}/entitlements",
				"error", err.Error(),
			)
		return Entitlement{}, err
	}
	return ent, nil
}

// DeleteTestEntitlement deletes a currently-active test entitlement.
//
// Reference: https://discord.com/developers/docs/resources/entitlement#delete-test-entitlement
func (r *requester) DeleteTestEntitlement(
	ctx context.Context,
	applicationID, entitlementID Snowflake,
) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/applications/" + applicationID.String() + "/entitlements/" + entitlementID.String(),
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// ListSKUSubscriptions returns all subscriptions containing the SKU.
//
// Reference: https://discord.com/developers/docs/resources/subscription#list-sku-subscriptions
func (r *requester) ListSKUSubscriptions(
	ctx context.Context,
	skuID Snowflake,
	opts ...ListSKUSubscriptionsOptions,
) ([]Subscription, error) {
	var first ListSKUSubscriptionsOptions
	if len(opts) > 0 {
		first = opts[0]
	}

	params := url.Values{}
	if first.Before != 0 {
		params.Set("before", first.Before.String())
	}
	if first.After != 0 {
		params.Set("after", first.After.String())
	}
	if first.Limit > 0 {
		params.Set("limit", strconv.Itoa(first.Limit))
	}
	if first.UserID != 0 {
		params.Set("user_id", first.UserID.String())
	}

	endpoint := "/skus/" + skuID.String() + "/subscriptions"
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

	var subs []Subscription
	if err := json.NewDecoder(body).Decode(&subs); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/skus/{id}/subscriptions",
				"error", err.Error(),
			)
		return nil, err
	}
	return subs, nil
}

// FetchSKUSubscription gets a subscription by its ID.
//
// Reference: https://discord.com/developers/docs/resources/subscription#get-sku-subscription
func (r *requester) FetchSKUSubscription(
	ctx context.Context,
	skuID, subscriptionID Snowflake,
) (Subscription, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/skus/" + skuID.String() + "/subscriptions/" + subscriptionID.String(),
	})
	if err != nil {
		return Subscription{}, err
	}
	defer body.Close()

	var sub Subscription
	if err := json.NewDecoder(body).Decode(&sub); err != nil {
		r.logger.
			Error(
				"failed parsing response", "method", "GET",
				"url", "/skus/{id}/subscriptions/{id}",
				"error", err.Error(),
			)
		return Subscription{}, err
	}
	return sub, nil
}
