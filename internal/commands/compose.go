package commands

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/handlers"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// Store is the global storage instance for command handlers.
// Must be initialized from main.go before commands are used.
var Store storage.Store

// DraftTTL is the duration after which drafts expire and are eligible for cleanup.
var DraftTTL = 24 * time.Hour

// Draft stores temporary compose state before posting.
type Draft struct {
	UserID        string
	GuildID       string
	ChannelID     string
	Content       string
	CreatedAt     time.Time
	ExpiresAt     time.Time // TTL expiration time for draft cleanup
	IsEdit        bool      // true if this is an edit proposal
	OriginalMsgID string    // for edit proposals: the original message ID
}

// draftStore holds pending drafts (key: userID:guildID).
var draftStore map[string]*Draft = make(map[string]*Draft)

// CleanupExpiredDrafts removes drafts older than their ExpiresAt time.
// Returns the count of cleaned drafts.
func CleanupExpiredDrafts() int {
	var now = time.Now()
	var cleaned int

	for key, draft := range draftStore {
		if now.After(draft.ExpiresAt) {
			delete(draftStore, key)
			cleaned++
		}
	}

	return cleaned
}

// getDraftKey generates a unique key for user's draft in a guild.
func getDraftKey(userID, guildID string) string {
	return userID + ":" + guildID
}

// ComposeDraftDefinition with content and optional channel option.
var ComposeDraftDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ComposeDraft),
	Description: "Create a new proxied message draft",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "content",
			Description: "The message content to post",
			Required:    true,
			MaxLength:   2000,
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Target channel (defaults to current channel)",
			Required:    false,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
		},
	},
}

// ComposeCreateDefinition is an alias for backward compatibility.
var ComposeCreateDefinition *discordgo.ApplicationCommand = ComposeDraftDefinition

// ComposeDraftExecute handles the compose-draft command.
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

	// Verify target channel permissions
	var targetPermResult handlers.PermissionResult = handlers.CanUseCompose(session, guildID, targetChannelID, userID, Store, i.Member.Roles)
	if !targetPermResult.Allowed {
		respondWithError(s, i, "You don't have permission to post in this channel. You need Send Messages permission, or an allowed role set by server admins.", nil)
		return
	}

	// Store draft
	var draftKey string = getDraftKey(userID, guildID)
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
	draftStore[draftKey] = &draft

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
		ConfirmButtonID: string(ButtonComposePreviewPost),
		CancelButtonID:  string(ButtonComposePreviewCancel),
	}

	var response *discordgo.InteractionResponse = handlers.RenderPreviewResponse(previewData)
	var err error = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		respondWithError(s, i, "Failed to show preview. Please try again.", err)
		return
	}
}

// ComposeCreateExecute is an alias for backward compatibility.
var ComposeCreateExecute func(s *discordgo.Session, i *discordgo.InteractionCreate) = ComposeDraftExecute

// ComposeSendDefinition for direct posting without preview.
var ComposeSendDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ComposeSend),
	Description: "Post a message directly without preview",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "content",
			Description: "The message content to post",
			Required:    true,
			MaxLength:   2000,
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Target channel (defaults to current channel)",
			Required:    false,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
		},
	},
}

// ComposePostDefinition is an alias for backward compatibility.
var ComposePostDefinition *discordgo.ApplicationCommand = ComposeSendDefinition

// ComposeSendExecute posts directly without preview.
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
		respondWithError(s, i, "You don't have permission to post in this channel. You need Send Messages permission, or an allowed role set by server admins.", nil)
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

	// Build jump URL and components for success message
	var jumpURL string = "https://discord.com/channels/" + guildID + "/" + targetChannelID + "/" + postResult.MessageID
	var components []discordgo.MessageComponent = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: "Jump to message",
					Style: discordgo.LinkButton,
					URL:   jumpURL,
				},
			},
		},
	}

	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "Message posted successfully!",
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})

	if err != nil {
		logging.Error("Failed to send success response",
			logging.Err("error", err),
			logging.String("user_id", userID),
		)
	}
}

// ComposePostExecute is an alias for backward compatibility.
var ComposePostExecute func(s *discordgo.Session, i *discordgo.InteractionCreate) = ComposeSendExecute

// ComposeEditDefinition for proposing edits to existing messages.
var ComposeEditDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ComposeEdit),
	Description: "Propose an edit to an existing proxied message",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "Message ID or link to the proxied message",
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

// ComposeEditExecute handles edit proposals.
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

	// Extract message ID from URL or use directly
	var messageID string = messageRef
	if strings.Contains(messageRef, "/") {
		messageID = handlers.ExtractMessageIDFromURL(messageRef)
	}

	if messageID == "" {
		respondWithError(s, i, "Invalid message ID or URL provided.", nil)
		return
	}

	// Look up proxy message
	var proxyMsg *storage.ProxyMessage
	var lookupErr error
	proxyMsg, lookupErr = handlers.GetProxiedMessage(Store, guildID, messageID)
	if lookupErr != nil {
		respondWithError(s, i, "Failed to find the specified message. Please verify the message ID.", lookupErr)
		return
	}

	if proxyMsg == nil {
		respondWithError(s, i, "Message not found. Only proxied messages can be edited.", nil)
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
	var draftKey string = getDraftKey(userID, guildID)
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
	draftStore[draftKey] = &draft

	logging.Info("Edit proposal created",
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
		ConfirmButtonID: string(ButtonEditPreviewApply),
		CancelButtonID:  string(ButtonEditPreviewCancel),
	}

	var response *discordgo.InteractionResponse = handlers.RenderPreviewResponse(previewData)
	var err error = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		respondWithError(s, i, "Failed to show preview. Please try again.", err)
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
	var draftKey string = getDraftKey(userID, guildID)
	var draft *Draft
	var exists bool
	draft, exists = draftStore[draftKey]
	if !exists || draft == nil {
		respondWithError(s, i, "No pending draft found. Create one with `/compose draft`.", nil)
		return
	}

	// Check if user owns the draft
	if draft.UserID != userID {
		respondWithError(s, i, "You can only post your own drafts.", nil)
		return
	}

	// Verify this is not an edit draft
	if draft.IsEdit {
		respondWithError(s, i, "This is an edit proposal. Please use the apply button for edits.", nil)
		return
	}

	// Post the message
	var postResult handlers.PostResult = handlers.PostProxiedMessage(session, draft.GuildID, draft.ChannelID, draft.Content, draft.UserID, Store)
	if !postResult.Success {
		respondWithError(s, i, postResult.Error, nil)
		return
	}

	// Delete draft on success
	delete(draftStore, draftKey)

	logging.Info("Draft posted",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("message_id", postResult.MessageID),
	)

	// Build jump URL and components for success message
	var jumpURL string = "https://discord.com/channels/" + draft.GuildID + "/" + draft.ChannelID + "/" + postResult.MessageID
	var components []discordgo.MessageComponent = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: "Jump to message",
					Style: discordgo.LinkButton,
					URL:   jumpURL,
				},
			},
		},
	}

	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "Message posted successfully!",
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})
	if err != nil {
		logging.Error("Failed to send success response",
			logging.Err("error", err),
			logging.String("user_id", userID),
		)
	}
}

// handleComposePreviewCancel cancels the draft.
func handleComposePreviewCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Delete draft
	var draftKey string = getDraftKey(userID, guildID)
	delete(draftStore, draftKey)

	logging.Info("Draft cancelled",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
	)

	// Update interaction to show cancellation
	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Draft cancelled. Your message was not posted.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logging.Error("Failed to send cancel response",
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
	var draftKey string = getDraftKey(userID, guildID)
	var draft *Draft
	var exists bool
	draft, exists = draftStore[draftKey]
	if !exists || draft == nil {
		respondWithError(s, i, "No pending edit proposal found. Create one with `/compose edit`.", nil)
		return
	}

	// Check if user owns the draft
	if draft.UserID != userID {
		respondWithError(s, i, "You can only apply your own edit proposals.", nil)
		return
	}

	// Verify this is an edit draft
	if !draft.IsEdit {
		respondWithError(s, i, "This is not an edit proposal. Please use the post button for new messages.", nil)
		return
	}

	// Look up original proxy message
	var proxyMsg *storage.ProxyMessage
	var lookupErr error
	proxyMsg, lookupErr = handlers.GetProxiedMessage(Store, guildID, draft.OriginalMsgID)
	if lookupErr != nil || proxyMsg == nil {
		respondWithError(s, i, "Message not found. Only proxied messages can be edited. Check the message ID or link.", lookupErr)
		return
	}

	// Apply the edit
	var editResult handlers.EditResult = handlers.EditProxiedMessage(session, proxyMsg, draft.Content, userID, Store)
	if !editResult.Success {
		respondWithError(s, i, editResult.Error, nil)
		return
	}

	// Delete draft on success
	delete(draftStore, draftKey)

	logging.Info("Edit applied",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
		logging.String("message_id", draft.OriginalMsgID),
	)

	// Build jump URL and components for success message
	var jumpURL string = "https://discord.com/channels/" + draft.GuildID + "/" + draft.ChannelID + "/" + draft.OriginalMsgID
	var components []discordgo.MessageComponent = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: "View edited message",
					Style: discordgo.LinkButton,
					URL:   jumpURL,
				},
			},
		},
	}

	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "Message edited successfully!",
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})
	if err != nil {
		logging.Error("Failed to send edit success response",
			logging.Err("error", err),
			logging.String("user_id", userID),
		)
	}
}

// handleEditPreviewCancel cancels the edit proposal.
func handleEditPreviewCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string = i.Member.User.ID
	var guildID string = i.GuildID

	// Delete draft
	var draftKey string = getDraftKey(userID, guildID)
	delete(draftStore, draftKey)

	logging.Info("Edit proposal cancelled",
		logging.String("user_id", userID),
		logging.String("guild_id", guildID),
	)

	// Update interaction to show cancellation
	respondToUser(s, i, "Edit proposal cancelled. The message was not modified.")
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
		logging.Error("Error in command execution",
			logging.Err("error", err),
			logging.String("context", userMsg),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
			logging.String("interaction_id", i.ID),
		)
	}
	respondToUser(s, i, userMsg)
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
}
