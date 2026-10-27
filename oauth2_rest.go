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
	"net/http"
	"net/url"
	"strings"
)

// ExchangeOAuth2Code exchanges an authorization code for an access token and refresh token.
//
// Reference: https://discord.com/developers/docs/topics/oauth2#authorization-code-grant-access-token-exchange
func (r *requester) ExchangeOAuth2Code(
	ctx context.Context,
	clientID Snowflake,
	clientSecret, code, redirectURI string,
) (OAuth2TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", clientID.String())
	form.Set("client_secret", clientSecret)

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPost,
		URL:    "/oauth2/token",
		Body:   []byte(form.Encode()),
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		NoAuth: true,
	})
	if err != nil {
		return OAuth2TokenResponse{}, err
	}
	defer body.Close()

	var resp OAuth2TokenResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "POST",
			"url", "/oauth2/token",
			"error", err.Error(),
		)
		return OAuth2TokenResponse{}, err
	}
	return resp, nil
}

// RefreshOAuth2Token exchanges a refresh token for a newly refreshed access token and refresh token.
//
// Reference: https://discord.com/developers/docs/topics/oauth2#authorization-code-grant-refresh-token-exchange
func (r *requester) RefreshOAuth2Token(
	ctx context.Context,
	clientID Snowflake,
	clientSecret, refreshToken string,
) (OAuth2TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", clientID.String())
	form.Set("client_secret", clientSecret)

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPost,
		URL:    "/oauth2/token",
		Body:   []byte(form.Encode()),
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		NoAuth: true,
	})
	if err != nil {
		return OAuth2TokenResponse{}, err
	}
	defer body.Close()

	var resp OAuth2TokenResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "POST",
			"url", "/oauth2/token",
			"error", err.Error(),
		)
		return OAuth2TokenResponse{}, err
	}
	return resp, nil
}

// ClientCredentialsGrant requests an access token using client credentials grant.
//
// Reference: https://discord.com/developers/docs/topics/oauth2#client-credentials-grant
func (r *requester) ClientCredentialsGrant(
	ctx context.Context,
	clientID Snowflake,
	clientSecret string,
	scopes ...OAuth2Scope,
) (OAuth2TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID.String())
	form.Set("client_secret", clientSecret)

	if len(scopes) > 0 {
		scopeStrs := make([]string, len(scopes))
		for i, s := range scopes {
			scopeStrs[i] = string(s)
		}
		form.Set("scope", strings.Join(scopeStrs, " "))
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPost,
		URL:    "/oauth2/token",
		Body:   []byte(form.Encode()),
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		NoAuth: true,
	})
	if err != nil {
		return OAuth2TokenResponse{}, err
	}
	defer body.Close()

	var resp OAuth2TokenResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "POST",
			"url", "/oauth2/token",
			"error", err.Error(),
		)
		return OAuth2TokenResponse{}, err
	}
	return resp, nil
}

// RevokeOAuth2Token revokes an access token or refresh token.
//
// Reference: https://discord.com/developers/docs/topics/oauth2#authorization-code-grant-token-revocation
func (r *requester) RevokeOAuth2Token(
	ctx context.Context,
	clientID Snowflake,
	clientSecret, token string,
	tokenTypeHint ...string,
) error {
	form := url.Values{}
	form.Set("token", token)
	form.Set("client_id", clientID.String())
	form.Set("client_secret", clientSecret)

	if len(tokenTypeHint) > 0 && tokenTypeHint[0] != "" {
		form.Set("token_type_hint", tokenTypeHint[0])
	}

	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodPost,
		URL:    "/oauth2/token/revoke",
		Body:   []byte(form.Encode()),
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		NoAuth: true,
	})
	if err != nil {
		return err
	}
	defer body.Close()
	return nil
}

// FetchCurrentAuthorization returns current authorization information for the given user Bearer access token.
//
// Reference: https://discord.com/developers/docs/topics/oauth2#get-current-authorization-information
func (r *requester) FetchCurrentAuthorization(
	ctx context.Context,
	accessToken string,
) (CurrentAuthorizationInformation, error) {
	body, err := r.DoRequest(ctx, Request{
		Method: http.MethodGet,
		URL:    "/oauth2/@me",
		Headers: map[string]string{
			"Authorization": "Bearer " + accessToken,
		},
		NoAuth: true,
	})
	if err != nil {
		return CurrentAuthorizationInformation{}, err
	}
	defer body.Close()

	var info CurrentAuthorizationInformation
	if err := json.NewDecoder(body).Decode(&info); err != nil {
		r.logger.Error(
			"failed parsing response",
			"method", "GET",
			"url", "/oauth2/@me",
			"error", err.Error(),
		)
		return CurrentAuthorizationInformation{}, err
	}
	return info, nil
}
