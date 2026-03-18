package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// PreviewData contains all information needed to render a preview
type PreviewData struct {
	Content       string
	TargetChannel string
	IsEdit        bool
	OriginalMsgID string // only for edits
	// Button CustomIDs - caller provides these to avoid import cycles
	ConfirmButtonID string // "Post" for compose, "Apply" for edit
	CancelButtonID  string
}

// RenderPreviewResponse creates an ephemeral preview response with Post/Cancel or Apply/Cancel buttons.
// Returns InteractionResponse ready to send.
func RenderPreviewResponse(data PreviewData) *discordgo.InteractionResponse {
	var content string = buildPreviewContent(data)

	// Build action row with appropriate buttons
	var components []discordgo.MessageComponent

	// Determine button label based on edit/compose flow
	var confirmLabel string = "Post"
	if data.IsEdit {
		confirmLabel = "Apply"
	}

	components = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    confirmLabel,
					Style:    discordgo.PrimaryButton,
					CustomID: data.ConfirmButtonID,
				},
				discordgo.Button{
					Label:    "Cancel",
					Style:    discordgo.SecondaryButton,
					CustomID: data.CancelButtonID,
				},
			},
		},
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	}
}

// buildPreviewContent formats the preview message content
func buildPreviewContent(data PreviewData) string {
	var header string
	if data.IsEdit {
		header = "**Edit Preview**\nPreview of your edited message:"
	} else {
		header = "**Compose Preview**\nPreview of your message:"
	}

	// Format the message content in a quote block for clarity
	var messagePreview string = fmt.Sprintf("> %s", data.Content)

	// Add metadata
	var metadata string = fmt.Sprintf(
		"\n\n**Target Channel:** <#%s>\n**Posted via:** GuildMessageProxy",
		data.TargetChannel,
	)

	if data.IsEdit && data.OriginalMsgID != "" {
		metadata += fmt.Sprintf("\n**Original Message ID:** %s", data.OriginalMsgID)
	}

	return header + "\n\n" + messagePreview + metadata + "\n\nClick **Post** to confirm or **Cancel** to discard."
}
