package commands

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/handlers"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// MessageDeleteDefinition for deleting a proxied message.
var MessageDeleteDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(MessageDelete),
	Description: "Delete a proxied message",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "Message ID or link to the proxied message",
			Required:    true,
		},
	},
}

// MessageDeleteExecute handles the message-delete command.
func MessageDeleteExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var session handlers.DiscordSession = handlers.NewDiscordSession(s)
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID
	var channelID string = i.ChannelID

	// Check permissions
	var permResult handlers.PermissionResult = handlers.CanUseCompose(session, guildID, channelID, userID, Store, i.Member.Roles)
	if !permResult.Allowed {
		respondWithError(s, i, permResult.Error, nil)
		return
	}

	// Extract options
	var data discordgo.ApplicationCommandInteractionData = i.ApplicationCommandData()
	var messageRef string = ""

	for _, option := range data.Options {
		if option.Name == "message" {
			messageRef = option.StringValue()
		}
	}

	// Extract message ID from URL or use directly
	var messageID string = messageRef
	if strings.Contains(messageRef, "/") {
		messageID = handlers.ExtractMessageIDFromURL(messageRef)
	}

	if messageID == "" {
		respondWithError(s, i, "Invalid message ID or URL provided.", nil)
		return
	}

	// Look up proxy message in storage
	var proxyMsg *storage.ProxyMessage
	var lookupErr error
	proxyMsg, lookupErr = handlers.GetProxiedMessage(Store, guildID, messageID)
	if lookupErr != nil {
		respondWithError(s, i, "Failed to find the specified message. Please verify the message ID.", lookupErr)
		return
	}

	if proxyMsg == nil {
		respondWithError(s, i, "Message not found.", nil)
		return
	}

	// Verify user is message owner
	if !handlers.IsMessageOwner(proxyMsg, userID) {
		respondWithError(s, i, "Only the message owner can delete this message.", nil)
		return
	}

	// Delete webhook message via Discord API
	var deleteErr error = deleteWebhookMessage(s, proxyMsg)
	if deleteErr != nil {
		respondWithError(s, i, "Failed to delete message. It may have already been deleted or is no longer accessible.", deleteErr)
		return
	}

	// Remove from storage
	var storageErr error = Store.DeleteProxyMessage(guildID, messageID)
	if storageErr != nil {
		logging.Error("Failed to delete proxy message from storage",
			logging.String("message_id", messageID),
			logging.String("guild_id", guildID),
			logging.Err("error", storageErr),
		)
		// Don't fail the deletion if storage removal fails - the message is already deleted from Discord
	}

	// Log the deletion
	logging.Info("Message deleted",
		logging.String("message_id", messageID),
		logging.String("channel_id", proxyMsg.ChannelID),
		logging.String("guild_id", guildID),
		logging.String("user_id", userID),
	)

	// Confirm success to user
	respondToUser(s, i, "Message deleted successfully.")
}

// deleteWebhookMessage deletes a message via webhook API.
func deleteWebhookMessage(s *discordgo.Session, proxyMsg *storage.ProxyMessage) error {
	// Check if we have webhook credentials
	if proxyMsg.WebhookID == "" || proxyMsg.WebhookToken == "" {
		return errors.New("webhook credentials not found for this message")
	}

	// Delete the webhook message using WebhookMessageDelete
	err := s.WebhookMessageDelete(proxyMsg.WebhookID, proxyMsg.WebhookToken, proxyMsg.MessageID)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	CommandDefinitions[MessageDelete] = SCommandDef{
		Definition:   MessageDeleteDefinition,
		Execute:      MessageDeleteExecute,
		Autocomplete: nil,
	}
}
