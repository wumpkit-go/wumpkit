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
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CacheAside_User(t *testing.T) {
	var apiCalls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"12345","username":"wumpkit_bot","discriminator":"0001"}`))
	}))
	defer server.Close()

	client := New(
		context.Background(),
		WithToken("test_token"),
		WithRESTConfig(RequesterConfig{
			BaseURL: server.URL,
		}),
	)

	// 1. initial state: not in cache -> fetches from API & caches
	user, err := client.User(context.Background(), 12345)
	require.NoError(t, err)
	assert.Equal(t, "wumpkit_bot", user.Username)
	assert.Equal(t, int64(1), apiCalls.Load())

	// 2. second call: found in cache -> 0 additional API calls
	cachedUser, err := client.User(context.Background(), 12345)
	require.NoError(t, err)
	assert.Equal(t, "wumpkit_bot", cachedUser.Username)
	assert.Equal(t, int64(1), apiCalls.Load())
}

func TestClient_CacheAside_Guild(t *testing.T) {
	var apiCalls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"55555","name":"Wumpkit Guild","roles":[],"emojis":[],"stickers":[]}`))
	}))
	defer server.Close()

	client := New(
		context.Background(),
		WithToken("test_token"),
		WithRESTConfig(RequesterConfig{
			BaseURL: server.URL,
		}),
	)

	guild, err := client.Guild(context.Background(), 55555)
	require.NoError(t, err)
	assert.Equal(t, "Wumpkit Guild", guild.Name)
	assert.Equal(t, int64(1), apiCalls.Load())

	// cache check
	cachedGuild, err := client.Guild(context.Background(), 55555)
	require.NoError(t, err)
	assert.Equal(t, "Wumpkit Guild", cachedGuild.Name)
	assert.Equal(t, int64(1), apiCalls.Load())
}

func TestClient_CacheAside_Member(t *testing.T) {
	var apiCalls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"user":{"id":"2001","username":"zwafriya"},"nick":"Zwafr","roles":[]}`))
	}))
	defer server.Close()

	client := New(
		context.Background(),
		WithToken("test_token"),
		WithRESTConfig(RequesterConfig{
			BaseURL: server.URL,
		}),
	)

	member, err := client.Member(context.Background(), 1001, 2001)
	require.NoError(t, err)
	assert.Equal(t, "Zwafr", member.Nickname)
	assert.Equal(t, int64(1), apiCalls.Load())

	// cache check
	cachedMember, err := client.Member(context.Background(), 1001, 2001)
	require.NoError(t, err)
	assert.Equal(t, "Zwafr", cachedMember.Nickname)
	assert.Equal(t, int64(1), apiCalls.Load())
}

func TestClient_CacheAside_Channel(t *testing.T) {
	var apiCalls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"3001","type":0,"name":"general","guild_id":"1001"}`))
	}))
	defer server.Close()

	client := New(
		context.Background(),
		WithToken("test_token"),
		WithRESTConfig(RequesterConfig{
			BaseURL: server.URL,
		}),
	)

	ch, err := client.Channel(context.Background(), 3001)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(3001), ch.GetID())
	assert.Equal(t, int64(1), apiCalls.Load())

	// cache check
	cachedCh, err := client.Channel(context.Background(), 3001)
	require.NoError(t, err)
	assert.Equal(t, Snowflake(3001), cachedCh.GetID())
	assert.Equal(t, int64(1), apiCalls.Load())
}
