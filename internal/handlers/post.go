// Package handlers provides message posting and URL extraction utilities for the GuildMessageProxy bot.
// It contains functions for posting proxied messages via webhooks, formatting content with attribution,
// and extracting Discord IDs from message URLs.
package handlers

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// PostResult contains the result of posting a proxied message.
type PostResult struct {
	// Success indicates whether the message was posted successfully.
	Success bool
	// MessageID is the Discord ID of the posted message (only valid when Success is true).
	MessageID string
	// Error contains a human-readable error message when Success is false.
	Error string
}

// PostProxiedMessage posts content to a Discord channel via webhook, creating the webhook if necessary.
// It triggers a typing indicator, retrieves or creates a webhook for the channel, executes the webhook
// to post the message, and stores proxy metadata in the provided store.
//
// Parameters:
//   - s: Discord session interface for API interactions
//   - guildID: Discord guild (server) ID where the message will be posted
//   - channelID: Discord channel ID where the message will be posted
//   - content: Message content to post
//   - ownerID: Discord user ID of the message owner (for attribution)
//   - store: Storage interface for persisting proxy message metadata
//
// Returns a PostResult indicating success or failure, with MessageID on success or Error on failure.
func PostProxiedMessage(s DiscordSession, guildID, channelID, content, ownerID string, store storage.Store) PostResult {
	// Trigger typing indicator before webhook operations
	var typingErr error = s.ChannelTyping(channelID)
	if typingErr != nil {
		// Non-fatal: log but continue
		logging.Debug("Failed to trigger typing indicator",
			logging.String("channel_id", channelID),
			logging.Err("error", typingErr),
		)
	}

	// Get or create webhook for channel
	webhook, err := getOrCreateWebhook(s, channelID)
	if err != nil {
		logging.Error("Failed to get or create webhook",
			logging.String("channel_id", channelID),
			logging.Err("error", err),
		)
		return PostResult{
			Success: false,
			Error:   "Unable to post message. The bot needs 'Manage Webhooks' permission in this channel. Ask a server admin to check bot permissions.",
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
			Error:   "Failed to post message. The webhook may have been deleted or the channel permissions changed. Try again, or ask a server admin to check the bot's access.",
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
func getOrCreateWebhook(s DiscordSession, channelID string) (*discordgo.Webhook, error) {
	// List existing webhooks in the channel
	webhooks, err := s.ChannelWebhooks(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	// Look for an existing webhook created by this bot
	var botUser *discordgo.User = s.BotUser()
	if botUser == nil {
		return nil, fmt.Errorf("bot user not available in session state")
	}
	for _, webhook := range webhooks {
		if webhook.User != nil && webhook.User.ID == botUser.ID {
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

// FormatProxiedContent adds attribution to proxied content, indicating who requested the message.
// The attribution is prepended as italicized text followed by two newlines.
func FormatProxiedContent(content string, requesterName string) string {
	// MVP: Simple text attribution
	// Future: Could use embeds or custom usernames via webhook
	var attribution string = fmt.Sprintf("_Requested by %s_\n\n", requesterName)
	return attribution + content
}

// ExtractMessageIDFromURL extracts the message ID from a Discord message URL.
// It validates that the URL is from a Discord domain and follows the expected path format.
//
// Supported URL formats:
//   - https://discord.com/channels/{guild_id}/{channel_id}/{message_id}
//   - https://www.discord.com/channels/{guild_id}/{channel_id}/{message_id}
//   - https://discordapp.com/channels/{guild_id}/{channel_id}/{message_id}
//   - https://www.discordapp.com/channels/{guild_id}/{channel_id}/{message_id}
//
// Parameter:
//   - urlStr: The Discord message URL to parse
//
// Returns the message ID string, or empty string if the URL is invalid or not a Discord URL.
func ExtractMessageIDFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Validate Discord domain
	host := strings.ToLower(parsedURL.Host)
	if host != "discord.com" && host != "www.discord.com" &&
		host != "discordapp.com" && host != "www.discordapp.com" {
		return ""
	}

	// Validate path pattern: /channels/{guild}/{channel}/{message}
	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "channels" {
		return ""
	}

	return parts[3]
}

// ExtractIDsFromURL extracts the guild ID, channel ID, and message ID from a Discord message URL.
// It validates that the URL is from a Discord domain and follows the /channels/{guild}/{channel}/{message} path format.
//
// Parameter:
//   - urlStr: The Discord message URL to parse
//
// Returns three strings in order: guildID, channelID, and messageID.
// Returns empty strings for all values if the URL is invalid or not a recognized Discord URL.
func ExtractIDsFromURL(urlStr string) (guildID, channelID, messageID string) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ""
	}

	// Validate Discord domain
	host := strings.ToLower(parsedURL.Host)
	if host != "discord.com" && host != "www.discord.com" &&
		host != "discordapp.com" && host != "www.discordapp.com" {
		return "", "", ""
	}

	// Validate path pattern: /channels/{guild}/{channel}/{message}
	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "channels" {
		return "", "", ""
	}

	return parts[1], parts[2], parts[3]
}
