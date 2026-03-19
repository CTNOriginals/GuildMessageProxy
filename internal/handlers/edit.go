package handlers

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// EditResult contains the result of editing a proxied message.
type EditResult struct {
	// Success indicates whether the edit operation completed successfully.
	Success bool
	// Error contains a human-readable error message if Success is false.
	Error string
}

// EditProxiedMessage updates an existing proxied message via webhook and persists metadata changes.
//
// Parameters:
//   - s: Discord session for API operations
//   - proxyMsg: The proxy message to edit, containing webhook credentials and message ID
//   - newContent: The new message content to apply
//   - editedBy: Discord user ID of the editor for audit tracking
//   - store: Storage interface for persisting metadata updates
//
// Returns an EditResult indicating success or failure with an error message.
//
// The function edits the message via webhook using stored webhook ID and token.
// Webhook messages must be edited through the webhook endpoint rather than
// standard message edit APIs. After successful webhook edit, metadata including
// content, LastEditedAt timestamp, and LastEditedBy are updated in storage.
func EditProxiedMessage(s DiscordSession, proxyMsg *storage.ProxyMessage, newContent, editedBy string, store storage.Store) EditResult {
	// Check if we have webhook credentials
	if proxyMsg.WebhookID == "" || proxyMsg.WebhookToken == "" {
		return EditResult{
			Success: false,
			Error:   "Cannot edit this message. The webhook used to post it is no longer available. This can happen if the webhook was deleted or the bot lost access. Contact a server admin if you need help.",
		}
	}

	// Trigger typing indicator before webhook edit
	var typingErr error = s.ChannelTyping(proxyMsg.ChannelID)
	if typingErr != nil {
		// Non-fatal: log but continue
		logging.Debug("Failed to trigger typing indicator",
			logging.String("channel_id", proxyMsg.ChannelID),
			logging.Err("error", typingErr),
		)
	}

	// Edit the webhook message using WebhookMessageEdit
	// Webhook messages must be edited via the webhook endpoint, not ChannelMessageEdit
	editParams := &discordgo.WebhookEdit{
		Content: &newContent,
	}

	_, err := s.WebhookMessageEdit(proxyMsg.WebhookID, proxyMsg.WebhookToken, proxyMsg.MessageID, editParams)
	if err != nil {
		logging.Error("Failed to edit webhook message",
			logging.String("webhook_id", proxyMsg.WebhookID),
			logging.String("message_id", proxyMsg.MessageID),
			logging.Err("error", err),
		)
		return EditResult{
			Success: false,
			Error:   "Failed to edit message. It may have been deleted, or the bot no longer has access. Try again, or use `/compose edit` to start a fresh edit proposal.",
		}
	}

	// Update proxy metadata
	var now time.Time = time.Now()
	proxyMsg.Content = newContent
	proxyMsg.LastEditedAt = &now
	proxyMsg.LastEditedBy = editedBy

	err = store.UpdateProxyMessage(*proxyMsg)
	if err != nil {
		logging.Error("Failed to update proxy message metadata",
			logging.String("message_id", proxyMsg.MessageID),
			logging.Err("error", err),
		)
		// Don't fail the edit if storage update fails - the message is already edited
		// But log it for investigation
	}

	logging.Info("Edited proxied message",
		logging.String("message_id", proxyMsg.MessageID),
		logging.String("channel_id", proxyMsg.ChannelID),
		logging.String("edited_by", editedBy),
	)

	return EditResult{
		Success: true,
		Error:   "",
	}
}

// GetProxiedMessage retrieves proxy message metadata by guild and message ID.
//
// Parameters:
//   - store: Storage interface for database operations
//   - guildID: Discord guild ID where the message was posted
//   - messageID: Discord message ID of the proxied message
//
// Returns the ProxyMessage metadata or an error if not found.
//
// This is a convenience wrapper around store.GetProxyMessage that provides
// a handler-level abstraction for retrieving proxied message records.
func GetProxiedMessage(store storage.Store, guildID, messageID string) (*storage.ProxyMessage, error) {
	return store.GetProxyMessage(guildID, messageID)
}
