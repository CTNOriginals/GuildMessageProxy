// Package handlers provides Discord interaction handlers and permission checking utilities.
// It contains functions for validating user permissions for compose commands and
// managing access control based on guild configuration.
package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// PermissionResult holds permission check outcome.
type PermissionResult struct {
	// Allowed indicates whether the permission check passed.
	Allowed bool
	// Error contains a human-readable error message if Allowed is false.
	Error string
}

// hasAnyRole checks if user has at least one of the allowed roles.
func hasAnyRole(memberRoles []string, allowedRoles []string) bool {
	for _, allowed := range allowedRoles {
		for _, memberRole := range memberRoles {
			if memberRole == allowed {
				return true
			}
		}
	}
	return false
}

// isChannelRestricted checks if channel is in the restricted list.
func isChannelRestricted(channelID string, config *storage.GuildConfig) bool {
	if config == nil {
		return false
	}
	for _, restricted := range config.RestrictedChannels {
		if restricted == channelID {
			return true
		}
	}
	return false
}

// isChannelAllowed checks if channel is in the allowed list.
// Returns true if allowedChannels is empty (whitelist not configured).
func isChannelAllowed(channelID string, config *storage.GuildConfig) bool {
	if config == nil || len(config.AllowedChannels) == 0 {
		return true
	}
	for _, allowed := range config.AllowedChannels {
		if allowed == channelID {
			return true
		}
	}
	return false
}

// CanUseCompose checks if a user has permission to use compose commands in a given channel.
//
// Parameters:
//   - s: Discord session for API calls
//   - guildID: The guild (server) ID to check permissions in
//   - channelID: The channel ID where the command will be used
//   - userID: The Discord user ID to check permissions for
//   - store: Storage interface for retrieving guild configuration
//   - memberRoles: List of role IDs assigned to the user
//
// Returns a PermissionResult indicating whether the user is allowed to use compose commands.
//
// Permission checks performed in order:
//   1. Channel existence and bot access
//   2. User's channel permissions retrieval
//   3. SendMessages permission check
//   4. Allowed roles check (if configured in guild)
//   5. Channel restrictions check
//   6. Channel whitelist check (if configured)
func CanUseCompose(s DiscordSession, guildID, channelID, userID string, store storage.Store, memberRoles []string) PermissionResult {
	// 1. Verify the channel exists and bot can access it
	_, err := s.Channel(channelID)
	if err != nil {
		return PermissionResult{
			Allowed: false,
			Error:   "Unable to verify channel permissions.",
		}
	}

	// 2. Get user's permissions in the channel
	var perms int64
	perms, err = s.UserChannelPermissions(userID, channelID)
	if err != nil {
		return PermissionResult{
			Allowed: false,
			Error:   "Unable to verify user permissions.",
		}
	}

	// 3. Check for SendMessages permission
	if perms&discordgo.PermissionSendMessages == 0 {
		return PermissionResult{
			Allowed: false,
			Error:   "You need 'Send Messages' permission in this channel to use this command.",
		}
	}

	// 4. Get guild config for role and channel checks
	var config *storage.GuildConfig
	config, err = store.GetGuildConfig(guildID)
	if err != nil {
		// Log error but don't block - continue with base permission only
		logging.Warn("Failed to get guild config for permission check",
			logging.String("guild_id", guildID),
			logging.Err("error", err),
		)
		config = nil
	}

	// 5. Check allowed roles if configured
	if config != nil && len(config.AllowedRoles) > 0 {
		if !hasAnyRole(memberRoles, config.AllowedRoles) {
			return PermissionResult{
				Allowed: false,
				Error:   "You need a specific role to use this command. Contact a server admin to check allowed roles.",
			}
		}
	}

	// 6. Check channel restrictions
	if isChannelRestricted(channelID, config) {
		return PermissionResult{
			Allowed: false,
			Error:   "This channel is restricted. Compose commands cannot be used here. Contact server admins to remove this channel from the restriction list.",
		}
	}

	// 7. Check channel whitelist if configured
	if !isChannelAllowed(channelID, config) {
		return PermissionResult{
			Allowed: false,
			Error:   "This channel is not on the allowed list for compose commands. Contact server admins to add this channel, or use a permitted channel.",
		}
	}

	return PermissionResult{
		Allowed: true,
		Error:   "",
	}
}

// IsMessageOwner checks if a user is the original owner of a proxied message.
//
// Parameters:
//   - proxyMsg: The proxied message to check ownership of
//   - userID: The Discord user ID to verify
//
// Returns true if the userID matches the message's OwnerID, false otherwise.
// Returns false if proxyMsg is nil.
func IsMessageOwner(proxyMsg *storage.ProxyMessage, userID string) bool {
	if proxyMsg == nil {
		return false
	}
	return proxyMsg.OwnerID == userID
}
