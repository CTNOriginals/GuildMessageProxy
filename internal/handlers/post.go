package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// PostResult contains the result of posting a proxied message
type PostResult struct {
	Success   bool
	MessageID string
	Error     string
}

// PostProxiedMessage posts content to a channel via webhook.
// Creates webhook if needed, stores metadata, returns result.
func PostProxiedMessage(s *discordgo.Session, guildID, channelID, content, ownerID string, store storage.Store) PostResult {
	// Get or create webhook for channel
	webhook, err := getOrCreateWebhook(s, channelID)
	if err != nil {
		logging.Error("Failed to get or create webhook",
			logging.String("channel_id", channelID),
			logging.Err("error", err),
		)
		return PostResult{
			Success: false,
			Error:   "Failed to create webhook for posting. Please check bot permissions.",
		}
	}

	// Execute webhook with content
	// MVP: Use bot's own avatar/username, add attribution in content if desired
	params := &discordgo.WebhookParams{
		Content: content,
	}

	// Use wait=true to get the created message back (needed for message ID)
	message, err := s.WebhookExecute(webhook.ID, webhook.Token, true, params)
	if err != nil {
		logging.Error("Failed to execute webhook",
			logging.String("webhook_id", webhook.ID),
			logging.String("channel_id", channelID),
			logging.Err("error", err),
		)
		return PostResult{
			Success: false,
			Error:   "Failed to post message via webhook.",
		}
	}

	// Store proxy metadata
	proxyMsg := storage.ProxyMessage{
		GuildID:      guildID,
		ChannelID:    channelID,
		MessageID:    message.ID,
		OwnerID:      ownerID,
		Content:      content,
		CreatedAt:    time.Now(),
		LastEditedAt: nil,
		LastEditedBy: "",
		WebhookID:    webhook.ID,
		WebhookToken: webhook.Token,
	}

	err = store.SaveProxyMessage(proxyMsg)
	if err != nil {
		logging.Error("Failed to save proxy message metadata",
			logging.String("message_id", message.ID),
			logging.String("guild_id", guildID),
			logging.Err("error", err),
		)
		// Don't fail the post if storage fails - the message is already sent
		// But log it for investigation
	}

	logging.Info("Posted proxied message",
		logging.String("message_id", message.ID),
		logging.String("channel_id", channelID),
		logging.String("guild_id", guildID),
		logging.String("owner_id", ownerID),
	)

	return PostResult{
		Success:   true,
		MessageID: message.ID,
		Error:     "",
	}
}

// getOrCreateWebhook finds existing webhook or creates new one
func getOrCreateWebhook(s *discordgo.Session, channelID string) (*discordgo.Webhook, error) {
	// List existing webhooks in the channel
	webhooks, err := s.ChannelWebhooks(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	// Look for an existing webhook created by this bot
	for _, webhook := range webhooks {
		if webhook.User != nil && webhook.User.ID == s.State.User.ID {
			// Found existing webhook created by bot
			return webhook, nil
		}
	}

	// No existing webhook found, create new one
	// Use bot's name for the webhook
	createdWebhook, err := s.WebhookCreate(channelID, "GuildMessageProxy", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	logging.Info("Created new webhook",
		logging.String("webhook_id", createdWebhook.ID),
		logging.String("channel_id", channelID),
	)

	return createdWebhook, nil
}

// FormatProxiedContent adds attribution to proxied content
func FormatProxiedContent(content string, requesterName string) string {
	// MVP: Simple text attribution
	// Future: Could use embeds or custom usernames via webhook
	var attribution string = fmt.Sprintf("_Requested by %s_\n\n", requesterName)
	return attribution + content
}

// ExtractMessageIDFromURL extracts message ID from a Discord message URL
func ExtractMessageIDFromURL(url string) string {
	// Discord message URLs: https://discord.com/channels/{guildID}/{channelID}/{messageID}
	parts := strings.Split(url, "/")
	if len(parts) >= 6 {
		return parts[len(parts)-1]
	}
	return ""
}

// ExtractIDsFromURL extracts guild, channel, and message IDs from a Discord message URL
func ExtractIDsFromURL(url string) (guildID, channelID, messageID string) {
	parts := strings.Split(url, "/")
	if len(parts) >= 6 {
		return parts[len(parts)-3], parts[len(parts)-2], parts[len(parts)-1]
	}
	return "", "", ""
}
