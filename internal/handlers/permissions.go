package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// PermissionResult holds permission check outcome
type PermissionResult struct {
	Allowed bool
	Error   string
}

// CanUseCompose checks if a user has permission to use compose commands.
// MVP: Check if user has SendMessages permission in the channel.
// Future: Check allowed roles from guild config.
func CanUseCompose(s *discordgo.Session, guildID, channelID, userID string, store storage.Store) PermissionResult {
	// Verify the channel exists and bot can access it
	_, err := s.Channel(channelID)
	if err != nil {
		return PermissionResult{
			Allowed: false,
			Error:   "Unable to verify channel permissions.",
		}
	}

	// Get user's permissions in the channel
	perms, err := s.UserChannelPermissions(userID, channelID)
	if err != nil {
		return PermissionResult{
			Allowed: false,
			Error:   "Unable to verify user permissions.",
		}
	}

	// Check for SendMessages permission
	if perms&discordgo.PermissionSendMessages == 0 {
		return PermissionResult{
			Allowed: false,
			Error:   "You need permission to send messages in this channel.",
		}
	}

	// Future: Check allowed roles from guild config
	// For now, allow all users with SendMessages permission
	_ = store // Will be used when role-based permissions are implemented

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
