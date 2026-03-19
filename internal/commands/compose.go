// Package commands provides the Discord slash command implementations for the GuildMessageProxy bot.
// This file implements the compose command suite for drafting, sending, and editing proxied messages.
package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/handlers"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// Store is the global storage instance that must be initialized from main.go before commands are used.
var Store storage.Store

// DraftTTL is the duration after which drafts expire.
var DraftTTL = 24 * time.Hour

// Draft stores temporary compose state before posting.
type Draft struct {
	UserID        string    // ID of the user who created the draft
	GuildID       string    // ID of the guild where the draft was created
	ChannelID     string    // ID of the target channel for the message
	Content       string    // Message content to be posted
	CreatedAt     time.Time // Timestamp when the draft was created
	ExpiresAt     time.Time // TTL expiration time
	IsEdit        bool      // true for edit proposals
	OriginalMsgID string    // for edit proposals: the original message ID
}

// DraftSvc is the global DraftService instance for managing draft storage.
var DraftSvc *DraftService

// init initializes the DraftService.
func init() {
	DraftSvc = NewDraftService()
}

// CleanupExpiredDrafts removes drafts older than their ExpiresAt time.
// It returns the count of cleaned drafts.
func CleanupExpiredDrafts() int {
	if DraftSvc == nil {
		return 0
	}
	return DraftSvc.CleanupExpired()
}

// ComposeDraftDefinition is the command definition for compose-draft.
var ComposeDraftDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ComposeDraft),
	Description: "Create a message preview before posting",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "content",
			Description: "The message to send (e.g., 'Hello everyone!')",
			Required:    true,
			MaxLength:   2000,
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Where to post the message (defaults to this channel)",
			Required:    false,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
		},
	},
}

// ComposeCreateDefinition is an alias for backward compatibility.
var ComposeCreateDefinition *discordgo.ApplicationCommand = ComposeDraftDefinition

// ComposeDraftExecute is the handler function for compose-draft command.
func ComposeDraftExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	var content string = ""
	var targetChannelID string = channelID

	for _, option := range data.Options {
		switch option.Name {
		case "content":
			content = option.StringValue()
		case "channel":
			if option.ChannelValue(s) != nil {
				targetChannelID = option.ChannelValue(s).ID
			}
		}
	}

	// Validate content
	var validationResult handlers.ValidationResult = handlers.ValidateContent(content)
	if !validationResult.Valid {
		respondWithError(s, i, validationResult.Error, nil)
		return
	}

	// Store draft FIRST (before checking target channel permissions)
	// This ensures user's work is preserved even if target channel permission fails
	var now = time.Now()
	var draft Draft = Draft{
		UserID:        userID,
		GuildID:       guildID,
		ChannelID:     targetChannelID,
		Content:       content,
		CreatedAt:     now,
		ExpiresAt:     now.Add(DraftTTL),
		IsEdit:        false,
		OriginalMsgID: "",
	}
	DraftSvc.Save(&draft)

	// Verify target channel permissions AFTER storing draft
	// Draft is already saved, so user can retry with /compose-draft if needed
	var targetPermResult handlers.PermissionResult = handlers.CanUseCompose(session, guildID, targetChannelID, userID, Store, i.Member.Roles)
	if !targetPermResult.Allowed {
		respondWithError(s, i, "You cannot post in this channel. You need 'Send Messages' permission, or a role allowed by server admins.\n\nYour draft has been saved. Use `/compose-draft` with a different channel to retry.", nil)
		return
	}

	logging.Info("Draft created",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("channel_id", targetChannelID),
	)

	// Render and send preview
	var previewData handlers.PreviewData = handlers.PreviewData{
		Content:         content,
		TargetChannel:   targetChannelID,
		IsEdit:          false,
		OriginalMsgID:   "",
		GuildID:         guildID,
		ConfirmButtonID: string(ButtonComposePreviewPost),
		CancelButtonID:  string(ButtonComposePreviewCancel),
		ExpiresAt:       draft.ExpiresAt,
	}

	var response *discordgo.InteractionResponse = handlers.RenderPreviewResponse(previewData)
	var err error = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		respondWithError(s, i, "Failed to show preview. The bot may be experiencing issues. Wait a moment and try `/compose-draft` again.", err)
		return
	}
}

// ComposeCreateExecute is an alias for backward compatibility.
var ComposeCreateExecute func(s *discordgo.Session, i *discordgo.InteractionCreate) = ComposeDraftExecute

// ComposeSendDefinition is the command definition for compose-send (direct posting).
var ComposeSendDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ComposeSend),
	Description: "Send a message immediately (skips preview)",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "content",
			Description: "The message to send (e.g., 'Hello everyone!')",
			Required:    true,
			MaxLength:   2000,
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Where to post the message (defaults to this channel)",
			Required:    false,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
		},
	},
}

// ComposePostDefinition is an alias for backward compatibility.
var ComposePostDefinition *discordgo.ApplicationCommand = ComposeSendDefinition

// ComposeSendExecute is the handler function for compose-send command.
func ComposeSendExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	var content string = ""
	var targetChannelID string = channelID

	for _, option := range data.Options {
		switch option.Name {
		case "content":
			content = option.StringValue()
		case "channel":
			if option.ChannelValue(s) != nil {
				targetChannelID = option.ChannelValue(s).ID
			}
		}
	}

	// Validate content
	var validationResult handlers.ValidationResult = handlers.ValidateContent(content)
	if !validationResult.Valid {
		respondWithError(s, i, validationResult.Error, nil)
		return
	}

	// Verify target channel permissions
	var targetPermResult handlers.PermissionResult = handlers.CanUseCompose(session, guildID, targetChannelID, userID, Store, i.Member.Roles)
	if !targetPermResult.Allowed {
		// Store draft so user can retry with /compose-draft
		var now = time.Now()
		var draft Draft = Draft{
			UserID:        userID,
			GuildID:       guildID,
			ChannelID:     targetChannelID,
			Content:       content,
			CreatedAt:     now,
			ExpiresAt:     now.Add(DraftTTL),
			IsEdit:        false,
			OriginalMsgID: "",
		}
		DraftSvc.Save(&draft)

		logging.Info("Draft created on permission failure",
			logging.String("user_id", userID),
			logging.String("guild_id", guildID),
			logging.String("channel_id", targetChannelID),
		)

		respondWithError(s, i, "You cannot post in this channel. You need 'Send Messages' permission, or a role allowed by server admins.\n\nYour message has been saved as a draft. Use `/compose-draft` to review and post to a different channel.", nil)
		return
	}

	// Post directly
	var postResult handlers.PostResult = handlers.PostProxiedMessage(session, guildID, targetChannelID, content, userID, Store)
	if !postResult.Success {
		respondWithError(s, i, postResult.Error, nil)
		return
	}

	logging.Info("Direct post completed",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("message_id", postResult.MessageID),
	)

	// Build jump URL and send success message
	var jumpURL string = "https://discord.com/channels/" + guildID + "/" + targetChannelID + "/" + postResult.MessageID
	RespondWithSuccessAndJumpURL(s, i,
		fmt.Sprintf(":white_check_mark: **Message posted!** Here's what was sent:\n> %s", truncateContent(content, 200)),
		jumpURL,
		"Jump to message")
}

// ComposePostExecute is an alias for backward compatibility.
var ComposePostExecute func(s *discordgo.Session, i *discordgo.InteractionCreate) = ComposeSendExecute

// ComposeEditDefinition is the command definition for compose-edit.
var ComposeEditDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ComposeEdit),
	Description: "Edit a message you posted via /compose",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "Message to edit - paste a message link or ID",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "content",
			Description: "New message content",
			Required:    true,
			MaxLength:   2000,
		},
	},
}

// ComposeProposeDefinition is an alias for backward compatibility.
var ComposeProposeDefinition *discordgo.ApplicationCommand = ComposeEditDefinition

// ComposeEditExecute is the handler function for compose-edit command.
func ComposeEditExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	var newContent string = ""

	for _, option := range data.Options {
		switch option.Name {
		case "message":
			messageRef = option.StringValue()
		case "content":
			newContent = option.StringValue()
		}
	}

	// Validate message reference length
	if len(messageRef) > 200 {
		respondWithError(s, i, "Invalid input. Paste a message link (right-click message → Copy Link) or the message ID number.", nil)
		return
	}

	// Extract message ID from URL or use directly
	var messageID string = messageRef
	if strings.Contains(messageRef, "/") {
		messageID = handlers.ExtractMessageIDFromURL(messageRef)
	}

	if messageID == "" {
		respondWithError(s, i, "Invalid message ID or URL. Provide a message ID (e.g., `1234567890123456789`) or a full Discord message link (e.g., `https://discord.com/channels/...`).", nil)
		return
	}

	// Look up proxy message
	var proxyMsg *storage.ProxyMessage
	var lookupErr error
	proxyMsg, lookupErr = handlers.GetProxiedMessage(Store, guildID, messageID)
	if lookupErr != nil {
		respondWithError(s, i, "Failed to find the specified message. Make sure you copied the message ID correctly, or check that the message hasn't been deleted.", lookupErr)
		return
	}

	if proxyMsg == nil {
		respondWithError(s, i, "Message not found. Only messages posted via `/compose` can be edited. Check the message ID, or verify this message was created by the bot.", nil)
		return
	}

	// Validate edit permission (must be owner for MVP)
	var editPermResult handlers.ValidationResult = handlers.ValidateEditPermission(proxyMsg, userID)
	if !editPermResult.Valid {
		respondWithError(s, i, editPermResult.Error, nil)
		return
	}

	// Validate new content
	var validationResult handlers.ValidationResult = handlers.ValidateContent(newContent)
	if !validationResult.Valid {
		respondWithError(s, i, validationResult.Error, nil)
		return
	}

	// Store edit proposal as draft
	var now = time.Now()
	var draft Draft = Draft{
		UserID:        userID,
		GuildID:       guildID,
		ChannelID:     proxyMsg.ChannelID,
		Content:       newContent,
		CreatedAt:     now,
		ExpiresAt:     now.Add(DraftTTL),
		IsEdit:        true,
		OriginalMsgID: messageID,
	}
	DraftSvc.Save(&draft)

	logging.Info("Edit draft created",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("original_message_id", messageID),
	)

	// Render and send preview
	var previewData handlers.PreviewData = handlers.PreviewData{
		Content:         newContent,
		TargetChannel:   proxyMsg.ChannelID,
		IsEdit:          true,
		OriginalMsgID:   messageID,
		GuildID:         guildID,
		ConfirmButtonID: string(ButtonEditPreviewApply),
		CancelButtonID:  string(ButtonEditPreviewCancel),
		ExpiresAt:       draft.ExpiresAt,
	}

	var response *discordgo.InteractionResponse = handlers.RenderPreviewResponse(previewData)
	var err error = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		respondWithError(s, i, "Failed to show preview. The bot may be experiencing issues. Wait a moment and try `/compose-edit` again.", err)
		return
	}
}

// ComposeProposeExecute is an alias for backward compatibility.
var ComposeProposeExecute func(s *discordgo.Session, i *discordgo.InteractionCreate) = ComposeEditExecute

// handleComposePreviewPost posts the draft when Post button is clicked.
func handleComposePreviewPost(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var session handlers.DiscordSession = handlers.NewDiscordSession(s)
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Retrieve draft
	var draft *Draft
	var exists bool
	draft, exists = DraftSvc.Get(userID, guildID)
	if !exists || draft == nil {
		respondWithError(s, i, "No pending draft found. Use `/compose-draft` to create a new message, or check if your draft expired.", nil)
		return
	}

	// Check if draft has expired
	if time.Now().After(draft.ExpiresAt) {
		DraftSvc.Delete(userID, guildID)
		respondWithError(s, i, fmt.Sprintf("Draft expired %s ago. Please create a new draft.", formatDurationPast(draft.ExpiresAt)), nil)
		return
	}

	// Check if user owns the draft
	if draft.UserID != userID {
		respondWithError(s, i, "You can only post your own drafts. This draft was created by someone else.", nil)
		return
	}

	// Verify this is not an edit draft
	if draft.IsEdit {
		respondWithError(s, i, "This draft is for editing an existing message. Use 'Apply Edit' instead.", nil)
		return
	}

	// Post the message
	var postResult handlers.PostResult = handlers.PostProxiedMessage(session, draft.GuildID, draft.ChannelID, draft.Content, draft.UserID, Store)
	if !postResult.Success {
		// Do NOT delete draft on failure - keep it for retry
		RespondWithRetry(s, i,
			fmt.Sprintf(":x: **Failed to post message**\n\n%s\n\nYour draft is preserved. Try again?", postResult.Error),
			string(ButtonComposeRetryPost),
			string(ButtonComposeCancelAfterError))
		return
	}

	// Delete draft on success
	DraftSvc.Delete(userID, guildID)

	logging.Info("Draft posted",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("message_id", postResult.MessageID),
	)

	// Build jump URL and send success message
	var jumpURL string = "https://discord.com/channels/" + draft.GuildID + "/" + draft.ChannelID + "/" + postResult.MessageID
	RespondWithSuccessAndJumpURL(s, i,
		fmt.Sprintf(":white_check_mark: **Message posted!** Here's what was sent:\n> %s", truncateContent(draft.Content, 200)),
		jumpURL,
		"Jump to message")
}

// handleComposePreviewCancel shows a confirmation dialog before discarding the draft.
func handleComposePreviewCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Look up draft
	var draft *Draft
	var exists bool
	draft, exists = DraftSvc.Get(userID, guildID)
	if !exists || draft == nil {
		respondWithError(s, i, "No pending draft found. Use `/compose-draft` to create a new message, or check if your draft expired.", nil)
		return
	}

	logging.Info("Draft discard confirmation shown",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
	)

	// Show confirmation dialog with draft preview
	var truncatedContent string = truncateContent(draft.Content, 100)
	var content string = fmt.Sprintf(":warning: **Discard this draft?**\n\n> %s\n\nThis cannot be undone.", truncatedContent)

	var components []discordgo.MessageComponent = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Yes, discard draft",
					Style:    discordgo.DangerButton,
					CustomID: string(ButtonComposeConfirmDiscard),
				},
				discordgo.Button{
					Label:    "No, keep it",
					Style:    discordgo.SecondaryButton,
					CustomID: string(ButtonComposeKeepDraft),
				},
			},
		},
	}

	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})
	if err != nil {
		logging.Error("Failed to send discard confirmation",
			logging.Err("error", err),
			logging.String("user_id", userID),
		)
	}
}

// handleEditPreviewApply applies the edit.
func handleEditPreviewApply(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var session handlers.DiscordSession = handlers.NewDiscordSession(s)
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Retrieve draft
	var draft *Draft
	var exists bool
	draft, exists = DraftSvc.Get(userID, guildID)
	if !exists || draft == nil {
		respondWithError(s, i, "No pending edit draft found. Create an edit draft with `/compose-edit <message> <content>`.", nil)
		return
	}

	// Check if draft has expired
	if time.Now().After(draft.ExpiresAt) {
		DraftSvc.Delete(userID, guildID)
		respondWithError(s, i, fmt.Sprintf("Draft expired %s ago. Please create a new draft.", formatDurationPast(draft.ExpiresAt)), nil)
		return
	}

	// Check if user owns the draft
	if draft.UserID != userID {
		respondWithError(s, i, "You can only apply your own edit drafts. This draft was created by someone else.", nil)
		return
	}

	// Verify this is an edit draft
	if !draft.IsEdit {
		respondWithError(s, i, "This draft is for a new message, not an edit. Use 'Post Message' instead.", nil)
		return
	}

	// Look up original proxy message
	var proxyMsg *storage.ProxyMessage
	var lookupErr error
	proxyMsg, lookupErr = handlers.GetProxiedMessage(Store, guildID, draft.OriginalMsgID)
	if lookupErr != nil || proxyMsg == nil {
		respondWithError(s, i, "Message not found. Only messages you created with /compose can be edited. Check the message ID or verify this message was posted by this bot.", lookupErr)
		return
	}

	// Get original message content before deleting draft
	var originalContent string
	if proxyMsg != nil {
		originalContent = proxyMsg.Content
	}

	// Apply the edit
	var editResult handlers.EditResult = handlers.EditProxiedMessage(session, proxyMsg, draft.Content, userID, Store)
	if !editResult.Success {
		// Keep draft on failure and show error with retry/cancel buttons
		RespondWithRetry(s, i,
			fmt.Sprintf(":x: **Failed to apply edit**\n\n%s\n\nYour edit draft is preserved. Try again?", editResult.Error),
			string(ButtonEditRetryApply),
			string(ButtonEditCancelAfterError))
		return
	}

	// Delete draft on success
	DraftSvc.Delete(userID, guildID)

	logging.Info("Edit applied",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("message_id", draft.OriginalMsgID),
	)

	// Build jump URL and success message content
	var jumpURL string = "https://discord.com/channels/" + draft.GuildID + "/" + draft.ChannelID + "/" + draft.OriginalMsgID
	var successContent string = ":white_check_mark: **Message edited successfully!**\n\n"
	if originalContent != "" {
		successContent += fmt.Sprintf("**Before:**\n> %s\n\n", truncateContent(originalContent, 100))
	}
	successContent += fmt.Sprintf("**After:**\n> %s", truncateContent(draft.Content, 100))

	RespondWithSuccessAndJumpURL(s, i, successContent, jumpURL, "View edited message")
}

// handleEditPreviewCancel shows a confirmation dialog before discarding the edit proposal.
func handleEditPreviewCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Look up draft
	var draft *Draft
	var exists bool
	draft, exists = DraftSvc.Get(userID, guildID)
	if !exists || draft == nil {
		respondWithError(s, i, "No pending edit draft found. Create one with `/compose-edit`.", nil)
		return
	}

	logging.Info("Edit draft discard confirmation shown",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
	)

	// Show confirmation dialog with draft preview
	var truncatedContent string = truncateContent(draft.Content, 100)
	var content string = fmt.Sprintf(":warning: **Discard this edit draft?**\n\n> %s\n\nThis cannot be undone.", truncatedContent)

	var components []discordgo.MessageComponent = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Yes, discard draft",
					Style:    discordgo.DangerButton,
					CustomID: string(ButtonComposeConfirmDiscard),
				},
				discordgo.Button{
					Label:    "No, keep it",
					Style:    discordgo.SecondaryButton,
					CustomID: string(ButtonComposeKeepDraft),
				},
			},
		},
	}

	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})
	if err != nil {
		logging.Error("Failed to send discard confirmation",
			logging.Err("error", err),
			logging.String("user_id", userID),
		)
	}
}

// handleConfirmDiscard confirms and deletes the draft.
func handleConfirmDiscard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Look up draft
	var draft *Draft
	var exists bool
	draft, exists = DraftSvc.Get(userID, guildID)
	if !exists || draft == nil {
		respondWithError(s, i, "No pending draft found. Use `/compose-draft` to create a new message, or check if your draft expired.", nil)
		return
	}

	// Delete draft
	DraftSvc.Delete(userID, guildID)

	logging.Info("Draft discarded",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
	)

	respondToUser(s, i, ":wastebasket: **Draft discarded.** Use `/compose-draft` to create a new draft.")
}

// handleKeepDraft preserves the draft and re-shows the preview.
func handleKeepDraft(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Look up draft
	var draft *Draft
	var exists bool
	draft, exists = DraftSvc.Get(userID, guildID)
	if !exists || draft == nil {
		respondWithError(s, i, "No pending draft found. Use `/compose-draft` to create a new message, or check if your draft expired.", nil)
		return
	}

	logging.Info("Draft kept",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
	)

	// Re-show preview using RenderPreviewResponse
	var previewData handlers.PreviewData = handlers.PreviewData{
		Content:         draft.Content,
		TargetChannel:   draft.ChannelID,
		IsEdit:          draft.IsEdit,
		OriginalMsgID:   draft.OriginalMsgID,
		GuildID:         draft.GuildID,
		ConfirmButtonID: func() string {
			if draft.IsEdit {
				return string(ButtonEditPreviewApply)
			}
			return string(ButtonComposePreviewPost)
		}(),
		CancelButtonID: func() string {
			if draft.IsEdit {
				return string(ButtonEditPreviewCancel)
			}
			return string(ButtonComposePreviewCancel)
		}(),
		ExpiresAt: draft.ExpiresAt,
	}

	var response *discordgo.InteractionResponse = handlers.RenderPreviewResponse(previewData)
	var err error = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		logging.Error("Failed to re-show preview",
			logging.Err("error", err),
			logging.String("user_id", userID),
		)
	}
}

// truncateContent truncates content if it exceeds maxLen, adding "..." suffix.
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}

// formatDurationPast returns a human-readable duration string for a past time.
// Returns empty string if the time is zero.
func formatDurationPast(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	var d time.Duration = time.Since(t)
	var hours int = int(d.Hours())
	var minutes int = int(d.Minutes())
	var days int = hours / 24

	if days > 0 {
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
	if hours > 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if minutes > 1 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	return "less than a minute"
}

// respondToUser sends an ephemeral message to the user.
func respondToUser(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		logging.Error("Failed to respond to user",
			logging.Err("error", err),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
		)
	}
}

// respondWithError sends a formatted error response to the user.
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) {
	if err != nil {
		var userID string
		if i.Member != nil && i.Member.User != nil {
			userID = i.Member.User.ID
		}
		logging.Error("Error in command execution",
			logging.Err("error", err),
			logging.String("context", userMsg),
			logging.String("user_id", userID),
			logging.String("guild_id", i.GuildID),
			logging.String("interaction_id", i.ID),
		)
	}
	respondToUser(s, i, userMsg)
}

// handleEditRetryApply retries applying the edit after a failure.
func handleEditRetryApply(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var session handlers.DiscordSession = handlers.NewDiscordSession(s)
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Retrieve draft
	var draft *Draft
	var exists bool
	draft, exists = DraftSvc.Get(userID, guildID)
	if !exists || draft == nil {
		respondWithError(s, i, "No pending edit draft found. Create an edit draft with `/compose-edit <message> <content>`.", nil)
		return
	}

	// Check if draft has expired
	if time.Now().After(draft.ExpiresAt) {
		DraftSvc.Delete(userID, guildID)
		respondWithError(s, i, fmt.Sprintf("Draft expired %s ago. Please create a new draft.", formatDurationPast(draft.ExpiresAt)), nil)
		return
	}

	// Check if user owns the draft
	if draft.UserID != userID {
		respondWithError(s, i, "You can only edit your own drafts.", nil)
		return
	}

	// Verify this is an edit draft
	if !draft.IsEdit {
		respondWithError(s, i, "This is not an edit draft. Use the Post button to send new messages.", nil)
		return
	}

	// Look up original proxy message
	var proxyMsg *storage.ProxyMessage
	var lookupErr error
	proxyMsg, lookupErr = handlers.GetProxiedMessage(Store, guildID, draft.OriginalMsgID)
	if lookupErr != nil || proxyMsg == nil {
		respondWithError(s, i, "Message not found. Only messages posted via /compose can be edited. Check the message ID or link.", lookupErr)
		return
	}

	// Get original message content before editing
	var originalContent string
	if proxyMsg != nil {
		originalContent = proxyMsg.Content
	}

	// Retry applying the edit
	var editResult handlers.EditResult = handlers.EditProxiedMessage(session, proxyMsg, draft.Content, userID, Store)
	if !editResult.Success {
		// Keep draft on failure and show error with retry/cancel buttons again
		RespondWithRetry(s, i,
			fmt.Sprintf(":x: **Failed to apply edit**\n\n%s\n\nYour edit draft is preserved. Try again?", editResult.Error),
			string(ButtonEditRetryApply),
			string(ButtonEditCancelAfterError))
		return
	}

	// Delete draft on success
	DraftSvc.Delete(userID, guildID)

	logging.Info("Edit applied on retry",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("message_id", draft.OriginalMsgID),
	)

	// Build jump URL and success message content
	var jumpURL string = "https://discord.com/channels/" + draft.GuildID + "/" + draft.ChannelID + "/" + draft.OriginalMsgID
	var successContent string = ":white_check_mark: **Message edited successfully!**\n\n"
	if originalContent != "" {
		successContent += fmt.Sprintf("**Before:**\n> %s\n\n", truncateContent(originalContent, 100))
	}
	successContent += fmt.Sprintf("**After:**\n> %s", truncateContent(draft.Content, 100))

	RespondWithSuccessAndJumpURL(s, i, successContent, jumpURL, "View edited message")
}

// handleEditCancelAfterError deletes the edit proposal after an edit failure.
func handleEditCancelAfterError(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Delete draft if it exists
	DraftSvc.Delete(userID, guildID)

	logging.Info("Edit draft cancelled after error",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
	)

	respondToUser(s, i, ":wastebasket: **Edit draft cancelled.** Use `/compose-edit` to create a new edit draft.")
}

// handleComposeRetryPost retries posting the draft after a failure.
func handleComposeRetryPost(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var session handlers.DiscordSession = handlers.NewDiscordSession(s)
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Retrieve draft
	var draft *Draft
	var exists bool
	draft, exists = DraftSvc.Get(userID, guildID)
	if !exists || draft == nil {
		respondWithError(s, i, "Draft no longer available. Create a new draft with `/compose-draft`.", nil)
		return
	}

	// Check if draft has expired
	if time.Now().After(draft.ExpiresAt) {
		DraftSvc.Delete(userID, guildID)
		respondWithError(s, i, fmt.Sprintf("Draft expired %s ago. Please create a new draft.", formatDurationPast(draft.ExpiresAt)), nil)
		return
	}

	// Check if user owns the draft
	if draft.UserID != userID {
		respondWithError(s, i, "You can only post your own drafts.", nil)
		return
	}

	// Verify this is not an edit draft
	if draft.IsEdit {
		respondWithError(s, i, "This is an edit draft. Use the Apply button to save edits.", nil)
		return
	}

	// Retry posting the message
	var postResult handlers.PostResult = handlers.PostProxiedMessage(session, draft.GuildID, draft.ChannelID, draft.Content, draft.UserID, Store)
	if !postResult.Success {
		// Keep draft on failure and show error with retry/cancel buttons again
		RespondWithRetry(s, i,
			fmt.Sprintf(":x: **Failed to post message**\n\n%s\n\nYour draft is preserved. Try again?", postResult.Error),
			string(ButtonComposeRetryPost),
			string(ButtonComposeCancelAfterError))
		return
	}

	// Delete draft on success
	DraftSvc.Delete(userID, guildID)

	logging.Info("Draft posted on retry",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("message_id", postResult.MessageID),
	)

	// Build jump URL and send success message
	var jumpURL string = "https://discord.com/channels/" + draft.GuildID + "/" + draft.ChannelID + "/" + postResult.MessageID
	RespondWithSuccessAndJumpURL(s, i,
		fmt.Sprintf(":white_check_mark: **Message posted!** Here's what was sent:\n> %s", truncateContent(draft.Content, 200)),
		jumpURL,
		"Jump to message")
}

// handleComposeCancelAfterError deletes the draft after a post failure.
func handleComposeCancelAfterError(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Delete draft
	DraftSvc.Delete(userID, guildID)

	logging.Info("Draft discarded after error",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
	)

	respondToUser(s, i, ":wastebasket: **Draft discarded.** Use `/compose-draft` to create a new draft.")
}

func init() {
	// Register new command definitions (primary names)
	CommandDefinitions[ComposeDraft] = SCommandDef{
		Definition: ComposeDraftDefinition,
		Execute:    ComposeDraftExecute,
		Autocomplete: nil,
	}
	CommandDefinitions[ComposeSend] = SCommandDef{
		Definition: ComposeSendDefinition,
		Execute:    ComposeSendExecute,
		Autocomplete: nil,
	}
	CommandDefinitions[ComposeEdit] = SCommandDef{
		Definition: ComposeEditDefinition,
		Execute:    ComposeEditExecute,
		Autocomplete: nil,
	}

	// Register backward compatibility aliases (old names)
	CommandDefinitions[ComposeCreate] = SCommandDef{
		Definition: ComposeDraftDefinition,
		Execute:    ComposeDraftExecute,
		Autocomplete: nil,
	}
	CommandDefinitions[ComposePost] = SCommandDef{
		Definition: ComposeSendDefinition,
		Execute:    ComposeSendExecute,
		Autocomplete: nil,
	}
	CommandDefinitions[ComposePropose] = SCommandDef{
		Definition: ComposeEditDefinition,
		Execute:    ComposeEditExecute,
		Autocomplete: nil,
	}

	// Register button handlers
	ButtonDefinitions[ButtonComposePreviewPost] = SButtonDef{
		Execute: handleComposePreviewPost,
	}
	ButtonDefinitions[ButtonComposePreviewCancel] = SButtonDef{
		Execute: handleComposePreviewCancel,
	}
	ButtonDefinitions[ButtonEditPreviewApply] = SButtonDef{
		Execute: handleEditPreviewApply,
	}
	ButtonDefinitions[ButtonEditPreviewCancel] = SButtonDef{
		Execute: handleEditPreviewCancel,
	}
	ButtonDefinitions[ButtonComposeConfirmDiscard] = SButtonDef{
		Execute: handleConfirmDiscard,
	}
	ButtonDefinitions[ButtonComposeKeepDraft] = SButtonDef{
		Execute: handleKeepDraft,
	}
	ButtonDefinitions[ButtonComposeRetryPost] = SButtonDef{
		Execute: handleComposeRetryPost,
	}
	ButtonDefinitions[ButtonComposeCancelAfterError] = SButtonDef{
		Execute: handleComposeCancelAfterError,
	}
	ButtonDefinitions[ButtonEditRetryApply] = SButtonDef{
		Execute: handleEditRetryApply,
	}
	ButtonDefinitions[ButtonEditCancelAfterError] = SButtonDef{
		Execute: handleEditCancelAfterError,
	}
}
