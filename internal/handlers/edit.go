package handlers

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// EditResult contains the result of editing a proxied message
type EditResult struct {
	Success bool
	Error   string
}

// EditProxiedMessage updates an existing proxied message via webhook.
// Uses stored webhook ID and token from proxy metadata.
func EditProxiedMessage(s DiscordSession, proxyMsg *storage.ProxyMessage, newContent, editedBy string, store storage.Store) EditResult {
	// Check if we have webhook credentials
	if proxyMsg.WebhookID == "" || proxyMsg.WebhookToken == "" {
		return EditResult{
			Success: false,
			Error:   "Cannot edit: webhook credentials not found for this message.",
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
			Error:   "Failed to edit message. It may have been deleted or is no longer editable.",
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

// GetProxiedMessage retrieves proxy message metadata by message ID.
// Convenience wrapper around store.GetProxyMessage.
func GetProxiedMessage(store storage.Store, guildID, messageID string) (*storage.ProxyMessage, error) {
	return store.GetProxyMessage(guildID, messageID)
}
