package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// PermissionResult holds permission check outcome
type PermissionResult struct {
	Allowed bool
	Error   string
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

// CanUseCompose checks if a user has permission to use compose commands.
// Checks: SendMessages permission, allowed roles from guild config, channel restrictions.
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
			Error:   "You need permission to send messages in this channel.",
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
				Error:   "You need one of the allowed roles to use this command. Contact server admins.",
			}
		}
	}

	// 6. Check channel restrictions
	if isChannelRestricted(channelID, config) {
		return PermissionResult{
			Allowed: false,
			Error:   "This channel is restricted from using compose commands.",
		}
	}

	// 7. Check channel whitelist if configured
	if !isChannelAllowed(channelID, config) {
		return PermissionResult{
			Allowed: false,
			Error:   "This channel is not allowed for compose commands.",
		}
	}

	return PermissionResult{
		Allowed: true,
		Error:   "",
	}
}

// IsMessageOwner checks if user is the original message owner.
func IsMessageOwner(proxyMsg *storage.ProxyMessage, userID string) bool {
	if proxyMsg == nil {
		return false
	}
	return proxyMsg.OwnerID == userID
}
