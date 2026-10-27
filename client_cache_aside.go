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

import "context"

/*********************************************************************
 *   Cache-Aside Getters (Check cache first, fallback to REST API)   *
 *********************************************************************/

// User returns the User from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) User(ctx context.Context, userID Snowflake) (User, error) {
	if u, ok := c.GetUser(ctx, userID); ok {
		return u, nil
	}
	user, err := c.FetchUser(ctx, userID)
	if err != nil {
		return User{}, err
	}
	c.AddUser(ctx, user)
	return user, nil
}

// Guild returns the Guild from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) Guild(ctx context.Context, guildID Snowflake) (Guild, error) {
	if g, ok := c.GetGuild(ctx, guildID); ok {
		return g, nil
	}
	restGuild, err := c.FetchGuild(ctx, guildID, FetchGuildOptions{})
	if err != nil {
		return Guild{}, err
	}
	c.AddGuild(ctx, restGuild.Guild)
	if len(restGuild.Roles) > 0 {
		c.AddRoles(ctx, restGuild.Roles...)
	}
	if len(restGuild.Emojis) > 0 {
		c.AddEmojis(ctx, restGuild.Emojis...)
	}
	if len(restGuild.Stickers) > 0 {
		c.AddStickers(ctx, restGuild.Stickers...)
	}
	return restGuild.Guild, nil
}

// Member returns the Member from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) Member(ctx context.Context, guildID, userID Snowflake) (Member, error) {
	if m, ok := c.GetMember(ctx, guildID, userID); ok {
		return m, nil
	}
	fullMember, err := c.FetchMember(ctx, guildID, userID)
	if err != nil {
		return Member{}, err
	}
	c.AddMember(ctx, fullMember.Member)
	if fullMember.User.ID != 0 {
		c.AddUser(ctx, fullMember.User)
	}
	return fullMember.Member, nil
}

// Channel returns the Channel from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) Channel(ctx context.Context, channelID Snowflake) (Channel, error) {
	if ch, ok := c.GetChannel(ctx, channelID); ok {
		return ch, nil
	}
	channel, err := c.FetchChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}
	c.AddChannel(ctx, channel)
	return channel, nil
}

// Message returns the Message from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) Message(ctx context.Context, channelID, messageID Snowflake) (Message, error) {
	if msg, ok := c.GetMessage(ctx, messageID); ok {
		return msg, nil
	}
	msg, err := c.FetchChannelMessage(ctx, channelID, messageID)
	if err != nil {
		return Message{}, err
	}
	c.AddMessage(ctx, msg)
	return msg, nil
}

// Role returns the Role from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) Role(ctx context.Context, guildID, roleID Snowflake) (Role, error) {
	if r, ok := c.GetRole(ctx, roleID); ok {
		return r, nil
	}
	role, err := c.FetchRole(ctx, guildID, roleID)
	if err != nil {
		return Role{}, err
	}
	c.AddRole(ctx, role)
	return role, nil
}

// Emoji returns the Emoji from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) Emoji(ctx context.Context, guildID, emojiID Snowflake) (Emoji, error) {
	if e, ok := c.GetEmoji(ctx, emojiID); ok {
		return e, nil
	}
	emoji, err := c.FetchGuildEmoji(ctx, guildID, emojiID)
	if err != nil {
		return Emoji{}, err
	}
	c.AddEmoji(ctx, emoji)
	return emoji, nil
}

// Sticker returns the Sticker from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) Sticker(ctx context.Context, stickerID Snowflake) (Sticker, error) {
	if s, ok := c.GetSticker(ctx, stickerID); ok {
		return s, nil
	}
	sticker, err := c.FetchSticker(ctx, stickerID)
	if err != nil {
		return Sticker{}, err
	}
	c.AddSticker(ctx, sticker)
	return sticker, nil
}

// GuildSticker returns the Guild Sticker from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) GuildSticker(ctx context.Context, guildID, stickerID Snowflake) (Sticker, error) {
	if s, ok := c.GetSticker(ctx, stickerID); ok {
		return s, nil
	}
	sticker, err := c.FetchGuildSticker(ctx, guildID, stickerID)
	if err != nil {
		return Sticker{}, err
	}
	c.AddSticker(ctx, sticker)
	return sticker, nil
}

// StageInstance returns the StageInstance from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) StageInstance(ctx context.Context, channelID Snowflake) (StageInstance, error) {
	if si, ok := c.GetStageInstance(ctx, channelID); ok {
		return si, nil
	}
	stage, err := c.FetchStageInstance(ctx, channelID)
	if err != nil {
		return StageInstance{}, err
	}
	c.AddStageInstance(ctx, stage)
	return stage, nil
}

// GuildScheduledEvent returns the GuildScheduledEvent from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) GuildScheduledEvent(
	ctx context.Context,
	guildID, eventID Snowflake,
	withUserCount ...bool,
) (GuildScheduledEvent, error) {
	if e, ok := c.GetGuildScheduledEvent(ctx, eventID); ok {
		return e, nil
	}
	userCount := false
	if len(withUserCount) > 0 {
		userCount = withUserCount[0]
	}
	event, err := c.FetchGuildScheduledEvent(ctx, guildID, eventID, userCount)
	if err != nil {
		return GuildScheduledEvent{}, err
	}
	c.AddGuildScheduledEvent(ctx, event)
	return event, nil
}

// AutoModerationRule returns the AutoModerationRule from cache if present; otherwise fetches it from the Discord API,
// populates the cache with the response, and returns it.
func (c *Client) AutoModerationRule(
	ctx context.Context,
	guildID, ruleID Snowflake,
) (AutoModerationRule, error) {
	if r, ok := c.GetAutoModerationRule(ctx, ruleID); ok {
		return r, nil
	}
	rule, err := c.FetchAutoModerationRule(ctx, guildID, ruleID)
	if err != nil {
		return AutoModerationRule{}, err
	}
	c.AddAutoModerationRule(ctx, rule)
	return rule, nil
}

/*****************************
 *   Group Relational Helpers
 *****************************/

// GuildChannels returns all guild channels from cache if available; otherwise fetches them from the Discord API,
// populates the cache, and returns them as a map keyed by ChannelID.
func (c *Client) GuildChannels(
	ctx context.Context,
	guildID Snowflake,
) (map[Snowflake]GuildChannel, error) {
	if chs, ok := c.GetGuildChannels(ctx, guildID); ok {
		return chs, nil
	}
	channels, err := c.FetchGuildChannels(ctx, guildID)
	if err != nil {
		return nil, err
	}
	out := make(map[Snowflake]GuildChannel, len(channels))
	for _, ch := range channels {
		c.AddChannel(ctx, ch)
		out[ch.GetID()] = ch
	}
	return out, nil
}

// GuildRoles returns all guild roles from cache if available; otherwise fetches them from the Discord API,
// populates the cache, and returns them as a map keyed by RoleID.
func (c *Client) GuildRoles(ctx context.Context, guildID Snowflake) (map[Snowflake]Role, error) {
	if roles, ok := c.GetGuildRoles(ctx, guildID); ok {
		return roles, nil
	}
	rolesList, err := c.FetchRoles(ctx, guildID)
	if err != nil {
		return nil, err
	}
	c.AddRoles(ctx, rolesList...)
	out := make(map[Snowflake]Role, len(rolesList))
	for _, r := range rolesList {
		out[r.ID] = r
	}
	return out, nil
}

// GuildEmojis returns all guild emojis from cache if available; otherwise fetches them from the Discord API,
// populates the cache, and returns them as a map keyed by EmojiID.
func (c *Client) GuildEmojis(ctx context.Context, guildID Snowflake) (map[Snowflake]Emoji, error) {
	if emojis, ok := c.GetGuildEmojis(ctx, guildID); ok {
		return emojis, nil
	}
	emojisList, err := c.ListGuildEmojis(ctx, guildID)
	if err != nil {
		return nil, err
	}
	c.AddEmojis(ctx, emojisList...)
	out := make(map[Snowflake]Emoji, len(emojisList))
	for _, e := range emojisList {
		out[e.ID] = e
	}
	return out, nil
}

// GuildStickers returns all guild stickers from cache if available; otherwise fetches them from the Discord API,
// populates the cache, and returns them as a map keyed by StickerID.
func (c *Client) GuildStickers(
	ctx context.Context,
	guildID Snowflake,
) (map[Snowflake]Sticker, error) {
	if stickers, ok := c.GetGuildStickers(ctx, guildID); ok {
		return stickers, nil
	}
	stickersList, err := c.ListGuildStickers(ctx, guildID)
	if err != nil {
		return nil, err
	}
	c.AddStickers(ctx, stickersList...)
	out := make(map[Snowflake]Sticker, len(stickersList))
	for _, s := range stickersList {
		out[s.ID] = s
	}
	return out, nil
}

// GuildMembers returns guild members from cache if available; otherwise fetches them from the Discord API,
// populates the cache, and returns them as a map keyed by UserID.
func (c *Client) GuildMembers(
	ctx context.Context,
	guildID Snowflake,
) (map[Snowflake]Member, error) {
	if members, ok := c.GetGuildMembers(ctx, guildID); ok {
		return members, nil
	}
	membersList, err := c.ListMembers(ctx, guildID)
	if err != nil {
		return nil, err
	}
	out := make(map[Snowflake]Member, len(membersList))
	for _, m := range membersList {
		c.AddMember(ctx, m.Member)
		if m.User.ID != 0 {
			c.AddUser(ctx, m.User)
		}
		out[m.ID] = m.Member
	}
	return out, nil
}
