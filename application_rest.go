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

// EditCurrentApplicationOptions contains parameters for editing the current application.
//
// Reference: https://discord.com/developers/docs/resources/application#edit-current-application
type EditCurrentApplicationOptions struct {
	// CustomInstallURL is the default custom authorization URL for the app, if enabled.
	CustomInstallURL *string `json:"custom_install_url,omitempty"`

	// Description is the description of the app.
	Description *string `json:"description,omitempty"`

	// RoleConnectionsVerificationURL is the role connection verification URL for the app.
	RoleConnectionsVerificationURL *string `json:"role_connections_verification_url,omitempty"`

	// InstallParams is settings for the app's default in-app authorization link, if enabled.
	InstallParams *ApplicationInstallParams `json:"install_params,omitempty"`

	// IntegrationTypesConfig is default scopes and permissions for each supported installation context.
	IntegrationTypesConfig map[ApplicationIntegrationType]ApplicationIntegrationTypeConfiguration `json:"integration_types_config,omitempty"`

	// Flags is the app's public flags.
	Flags *ApplicationFlags `json:"flags,omitempty"`

	// Icon is the icon for the app (base64 image data). Pointer allows setting nil to clear.
	Icon *Base64Image `json:"icon,omitempty"`

	// CoverImage is the default rich presence invite cover image for the app (base64 image data). Pointer allows setting nil to clear.
	CoverImage *Base64Image `json:"cover_image,omitempty"`

	// InteractionsEndpointURL is the interactions endpoint URL for the app. Pointer allows setting nil to clear.
	InteractionsEndpointURL *string `json:"interactions_endpoint_url,omitempty"`

	// Tags is list of tags describing the content and functionality of the app (max 5 tags, max 20 chars each).
	Tags []string `json:"tags,omitempty"`

	// EventWebhooksURL is the event webhooks URL for the app to receive webhook events.
	EventWebhooksURL *string `json:"event_webhooks_url,omitempty"`

	// EventWebhooksStatus indicates whether webhook events are enabled or disabled.
	EventWebhooksStatus *ApplicationEventWebhookStatus `json:"event_webhooks_status,omitempty"`

	// EventWebhooksTypes is the list of Webhook event types to subscribe to.
	EventWebhooksTypes []WebhookEventTypes `json:"event_webhooks_types,omitempty"`
}

// FetchCurrentApplication returns the application object associated with the requesting bot user.
//
// Reference: https://discord.com/developers/docs/resources/application#get-current-application
func (r *requester) FetchCurrentApplication(ctx context.Context) (Application, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/applications/@me",
	})
	if err != nil {
		return Application{}, err
	}
	defer func() { _ = body.Close() }()

	var app Application
	if err := json.NewDecoder(body).Decode(&app); err != nil {
		r.logger.Error("failed parsing response",
			"method", "GET",
			"url", "/applications/@me",
			"error", err.Error(),
		)
		return Application{}, err
	}
	return app, nil
}

// EditCurrentApplication edits properties of the app associated with the requesting bot user.
//
// Reference: https://discord.com/developers/docs/resources/application#edit-current-application
func (r *requester) EditCurrentApplication(ctx context.Context, opts EditCurrentApplicationOptions) (Application, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.DoRequest(ctx, Request{
		Method: "PATCH",
		URL:    "/applications/@me",
		Body:   reqBody,
	})
	if err != nil {
		return Application{}, err
	}
	defer func() { _ = body.Close() }()

	var app Application
	if err := json.NewDecoder(body).Decode(&app); err != nil {
		r.logger.Error("failed parsing response",
			"method", "PATCH",
			"url", "/applications/@me",
			"error", err.Error(),
		)
		return Application{}, err
	}
	return app, nil
}

// FetchApplicationActivityInstance returns a serialized activity instance, if it exists.
//
// Reference: https://discord.com/developers/docs/resources/application#get-application-activity-instance
func (r *requester) FetchApplicationActivityInstance(ctx context.Context, applicationID Snowflake, instanceID string) (ApplicationActivityInstance, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/applications/" + applicationID.String() + "/activity-instances/" + instanceID,
	})
	if err != nil {
		return ApplicationActivityInstance{}, err
	}
	defer func() { _ = body.Close() }()

	var instance ApplicationActivityInstance
	if err := json.NewDecoder(body).Decode(&instance); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/applications/{id}/activity-instances/{instance_id}",
				"error", err.Error(),
			)
		return ApplicationActivityInstance{}, err
	}
	return instance, nil
}

// FetchApplicationRoleConnectionMetadata returns a list of application role connection metadata objects.
//
// Reference: https://discord.com/developers/docs/resources/application-role-connection-metadata#get-application-role-connection-metadata-records
func (r *requester) FetchApplicationRoleConnectionMetadata(ctx context.Context, applicationID Snowflake) ([]ApplicationRoleConnectionMetadata, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: "GET",
		URL:    "/applications/" + applicationID.String() + "/role-connections/metadata",
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	var records []ApplicationRoleConnectionMetadata
	if err := json.NewDecoder(body).Decode(&records); err != nil {
		r.logger.
			Error("failed parsing response", "method", "GET",
				"url", "/applications/{id}/role-connections/metadata",
				"error", err.Error(),
			)
		return nil, err
	}
	return records, nil
}

// UpdateApplicationRoleConnectionMetadata updates and returns a list of role connection metadata objects (max 5).
//
// Reference: https://discord.com/developers/docs/resources/application-role-connection-metadata#update-application-role-connection-metadata-records
func (r *requester) UpdateApplicationRoleConnectionMetadata(ctx context.Context, applicationID Snowflake, records []ApplicationRoleConnectionMetadata) ([]ApplicationRoleConnectionMetadata, error) {
	reqBody, _ := json.Marshal(records)
	body, err := r.DoRequest(ctx, Request{
		Method: "PUT",
		URL:    "/applications/" + applicationID.String() + "/role-connections/metadata",
		Body:   reqBody,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	var updated []ApplicationRoleConnectionMetadata
	if err := json.NewDecoder(body).Decode(&updated); err != nil {
		r.logger.
			Error("failed parsing response", "method", "PUT",
				"url", "/applications/{id}/role-connections/metadata",
				"error", err.Error(),
			)
		return nil, err
	}
	return updated, nil
}
