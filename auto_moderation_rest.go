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

// CreateAutoModerationRuleOptions contains parameters for creating a new Auto Moderation rule.
type CreateAutoModerationRuleOptions struct {
	// the rule name
	Name string `json:"name"`

	// the rule event type
	EventType AutoModerationEventType `json:"event_type"`

	// the rule trigger type
	TriggerType AutoModerationTriggerType `json:"trigger_type"`

	// the rule trigger metadata
	TriggerMetadata *AutoModerationTriggerMetadata `json:"trigger_metadata,omitempty"`

	// the actions which will execute when the rule is triggered
	Actions []AutoModerationAction `json:"actions"`

	// whether the rule is enabled (False by default)
	Enabled bool `json:"enabled,omitempty"`

	// the role ids that should not be affected by the rule (Maximum of 20)
	ExemptRoles []Snowflake `json:"exempt_roles,omitempty"`

	// the channel ids that should not be affected by the rule (Maximum of 50)
	ExemptChannels []Snowflake `json:"exempt_channels,omitempty"`

	// reason specifies the audit log reason for this action
	Reason string `json:"-"`
}

// ModifyAutoModerationRuleOptions contains parameters for modifying an existing Auto Moderation rule.
type ModifyAutoModerationRuleOptions struct {
	// the rule name. Zero value means not provided.
	Name string `json:"name,omitempty"`

	// the rule event type
	EventType *AutoModerationEventType `json:"event_type,omitempty"`

	// the rule trigger metadata
	TriggerMetadata *AutoModerationTriggerMetadata `json:"trigger_metadata,omitempty"`

	// the actions which will execute when the rule is triggered
	Actions []AutoModerationAction `json:"actions,omitempty"`

	// whether the rule is enabled
	Enabled *bool `json:"enabled,omitempty"`

	// the role ids that should not be affected by the rule (Maximum of 20)
	ExemptRoles []Snowflake `json:"exempt_roles,omitempty"`

	// the channel ids that should not be affected by the rule (Maximum of 50)
	ExemptChannels []Snowflake `json:"exempt_channels,omitempty"`

	// reason specifies the audit log reason for this action
	Reason string `json:"-"`
}

// ListAutoModerationRules retrieves a list of all rules currently configured for the guild.
//
// Requires the MANAGE_GUILD permission.
func (r *requester) ListAutoModerationRules(
	ctx context.Context,
	guildID Snowflake,
) ([]AutoModerationRule, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/auto-moderation/rules",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var rules []AutoModerationRule
	if err := json.NewDecoder(body).Decode(&rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// FetchAutoModerationRule retrieves a single Auto Moderation rule.
//
// Requires the MANAGE_GUILD permission.
func (r *requester) FetchAutoModerationRule(
	ctx context.Context,
	guildID, ruleID Snowflake,
) (AutoModerationRule, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/guilds/" + guildID.String() + "/auto-moderation/rules/" + ruleID.String(),
	})
	if err != nil {
		return AutoModerationRule{}, err
	}
	defer body.Close()

	var rule AutoModerationRule
	if err := json.NewDecoder(body).Decode(&rule); err != nil {
		return AutoModerationRule{}, err
	}
	return rule, nil
}

// CreateAutoModerationRule creates a new Auto Moderation rule.
//
// Requires the MANAGE_GUILD permission.
func (r *requester) CreateAutoModerationRule(
	ctx context.Context,
	guildID Snowflake,
	opts CreateAutoModerationRuleOptions,
) (AutoModerationRule, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "POST",
		URL:    "/guilds/" + guildID.String() + "/auto-moderation/rules",
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return AutoModerationRule{}, err
	}
	defer body.Close()

	var rule AutoModerationRule
	if err := json.NewDecoder(body).Decode(&rule); err != nil {
		return AutoModerationRule{}, err
	}
	return rule, nil
}

// ModifyAutoModerationRule modifies an existing Auto Moderation rule.
//
// Requires the MANAGE_GUILD permission.
func (r *requester) ModifyAutoModerationRule(
	ctx context.Context,
	guildID, ruleID Snowflake,
	opts ModifyAutoModerationRuleOptions,
) (AutoModerationRule, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/guilds/" + guildID.String() + "/auto-moderation/rules/" + ruleID.String(),
		Body:   reqBody,
		Reason: opts.Reason,
	})
	if err != nil {
		return AutoModerationRule{}, err
	}
	defer body.Close()

	var rule AutoModerationRule
	if err := json.NewDecoder(body).Decode(&rule); err != nil {
		return AutoModerationRule{}, err
	}
	return rule, nil
}

// DeleteAutoModerationRule deletes an Auto Moderation rule.
//
// Requires the MANAGE_GUILD permission.
func (r *requester) DeleteAutoModerationRule(
	ctx context.Context,
	guildID, ruleID Snowflake,
	reason string,
) error {
	body, err := r.DoRequest(ctx, Request{
		Method: "DELETE",
		URL:    "/guilds/" + guildID.String() + "/auto-moderation/rules/" + ruleID.String(),
		Reason: reason,
	})
	if err != nil {
		return err
	}
	body.Close()
	return nil
}
