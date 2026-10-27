package ratelimit

import (
	"fmt"
	"testing"
)

func TestPaths(t *testing.T) {
	tests := []struct {
		path, method, want string
	}{
		// guild Major
		{"/api/v9/guilds/103039963636301824", "GET", "/guilds/103039963636301824"},
		// channel major
		{"/api/v8/channels/203039963636301824", "GET", "/channels/203039963636301824"},
		{"/api/v7/channels/203039963636301824/pins", "GET", "/channels/203039963636301824/pins"},
		{"/api/v6/channels/872712139712913438/messages/872712150509047809/reactions/%F0%9F%98%8B", "GET", "/channels/872712139712913438/messages/!/reactions/!/!"},
		{"/api/v10/channels/872712139712913438/messages/872712150509047809/reactions/PandaOhShit:863985751205085195", "GET", "/channels/872712139712913438/messages/!/reactions/!/!"},
		{"/api/v9/channels/872712139712913438/messages/872712150509047809/reactions/PandaOhShit:863985751205085195", "PUT", "/channels/872712139712913438/messages/!/reactions/!modify"},
		{"/api/v9/channels/872712139712913438/messages/872712150509047809/reactions/PandaOhShit:863985751205085195", "DELETE", "/channels/872712139712913438/messages/!/reactions/!modify"},
		{"/api/v9/channels/872712139712913438/messages/872712150509047809/reactions/PandaOhShit:863985751205085195/@me", "DELETE", "/channels/872712139712913438/messages/!/reactions/!modify"},
		{"/api/v9/channels/872712139712913438/messages/872712150509047809/reactions/PandaOhShit:863985751205085195/203039963636301824", "DELETE", "/channels/872712139712913438/messages/!/reactions/!modify"},
		// hooks major
		{"/api/v9/webhooks/203039963636301824", "GET", "/webhooks/203039963636301824"},
		{"/api/v9/webhooks/203039963636301824/VSOzAqY1OZFF5WJVtbIzFtmjGupk-84Hn0A_ZzToF_CHsPIeCk0Q9Uok_mjxR0dNtApI", "POST", "/webhooks/203039963636301824/!"},
		// invites major
		{"/api/v9/invites/dyno", "GET", "/invites/!"},
		// interactions major
		{"/api/v9/interactions/203039963636301824/aW50ZXJhY3Rpb246ODg3NTU5MDA01AY4NTUxNDU0OnZwS3QycDhvREk2aVF3U1BqN2prcXBkRmNqNlp4VEhGRjZvSVlXSGh4WG4yb3l6Z3B6NTBPNVc3OHphV05OULLMOHBMa2RTZmVKd3lzVDA2b2h3OTUxaFJ4QlN0dkxXallPcmhnSHNJb0tSV0M5ZzY1NkN4VGRvemFOSHY4b05c/callback", "GET", "/interactions/203039963636301824/!/callback"},
		// interaction followup webhooks
		{"/api/v10/webhooks/203039963636301824/aW50ZXJhY3Rpb246MTEwMzA0OTQyMDkzMDU2ODMyMjpOZUllWHdNU2J4RXBFMHVYRjBpU0pHMDdEb3BhM3ZlYklBODlMUmtlUXlRbzlpZzYyTnpLU0dqdWlyVlBvZnBSUlJHbUJHYlJ0N29MbE9KQUJVTFk4bTR4UzFtZEpEeXJyY0hBUERmTEhKVE9wRkNzU1FFWUkwTnlpWFY2WHdrRg/messages/@original", "POST", "/webhooks/203039963636301824/1103049420930568322/messages/@original"},
		// no known major
		{"/api/v9/invalid/203039963636301824", "GET", "/invalid/203039963636301824"},
		{"/api/v9/invalid/203039963636301824/route/203039963636301824", "GET", "/invalid/203039963636301824/route/!"},
		// special case for /guilds/:id/channels
		{"/api/v9/guilds/203039963636301824/channels", "GET", "/guilds/203039963636301824/channels"},
		// wierd routes
		{"/api/v9/guilds/templates/203039963636301824", "GET", "/guilds/templates/!"},
		// unversioned routes
		{"/api/webhooks/203039963636301824/VSOzAqY1OZFF5WJVtbIzFtmjGupk-84Hn0A_ZzToF_CHsPIeCk0Q9Uok_mjxR0dNtApI", "POST", "/webhooks/203039963636301824/!"},
		{"/api/interactions/203039963636301824/aW50ZXJhY3Rpb246ODg3NTU5MDA01AY4NTUxNDU0OnZwS3QycDhvREk2aVF3U1BqN2prcXBkRmNqNlp4VEhGRjZvSVlXSGh4WG4yb3l6Z3B6NTBPNVc3OHphV05OULLMOHBMa2RTZmVKd3lzVDA2b2h3OTUxaFJ4QlN0dkxXallPcmhnSHNJb0tSV0M5ZzY1NkN4VGRvemFOSHY4b05c/callback", "GET", "/interactions/203039963636301824/!/callback"},
		{"/api/channels/872712139712913438/messages/872712150509047809/reactions/PandaOhShit:863985751205085195", "GET", "/channels/872712139712913438/messages/!/reactions/!/!"},
		{"/api/invites/dyno", "GET", "/invites/!"},
		// application commands
		{"/api/v9/applications/203039963636301824/commands", "GET", "/applications/203039963636301824/commands"},
		{"/api/v9/applications/203039963636301824/commands/203039963636301824", "GET", "/applications/203039963636301824/commands/!"},
		// deep nested route exceeding 160 characters
		{"/api/v10/applications/103039963636301824/guilds/203039963636301824/commands/303039963636301824/permissions/extra/nested/subresource/item/identifier/field/configuration/detail", "GET", "/applications/103039963636301824/guilds/!/commands/!/permissions/extra/nested/subresource/item/identifier/field/configuration/detail"},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s-%s", tt.method, tt.path)
		t.Run(testname, func(t *testing.T) {
			bucket := NormalizeRoute(tt.path, tt.method)
			if bucket != tt.want {
				t.Errorf("Expected %s but got %s", tt.want, bucket)
			}
		})
	}
}

func TestIsInteractionRoute(t *testing.T) {
	tests := []struct {
		name           string
		optimisticPath string
		want           bool
	}{
		// interaction callback routes — MUST bypass global rate limit
		{"interaction callback", "/interactions/203039963636301824/!/callback", true},
		{"interaction base", "/interactions/203039963636301824/!", true},

		// webhook interaction followup routes — MUST bypass global rate limit
		// these have a decoded interaction snowflake ID (not "!") as the token segment
		{"webhook interaction followup", "/webhooks/203039963636301824/1103049420930568322/messages/@original", true},

		// regular webhook routes — MUST NOT bypass global rate limit
		{"regular webhook execute", "/webhooks/203039963636301824/!", false},
		{"regular webhook base", "/webhooks/203039963636301824", false},

		// non-interaction routes — MUST NOT bypass global rate limit
		{"channel messages", "/channels/203039963636301824/messages", false},
		{"guild route", "/guilds/103039963636301824", false},
		{"invites", "/invites/!", false},
		{"application commands", "/applications/103039963636301824/commands", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInteractionRoute(tt.optimisticPath)
			if got != tt.want {
				t.Errorf("IsInteractionRoute(%q) = %v, want %v", tt.optimisticPath, got, tt.want)
			}
		})
	}
}
