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
	var embed *discordgo.MessageEmbed = buildPreviewEmbed(data)
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
			Embeds:     []*discordgo.MessageEmbed{embed},
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	}
}

// buildPreviewEmbed creates a Discord embed for preview display
func buildPreviewEmbed(data PreviewData) *discordgo.MessageEmbed {
	var title string
	var color int
	var footerText string

	if data.IsEdit {
		title = "Edit Preview"
		color = 0xe67e22 // Orange for edit
		footerText = "Click Apply to confirm the edit, or Cancel to discard."
	} else {
		title = "Compose Preview"
		color = 0x3498db // Blue for compose
		footerText = "Click Post to send the message, or Cancel to discard."
	}

	// Build fields for metadata
	var fields []*discordgo.MessageEmbedField = []*discordgo.MessageEmbedField{
		{
			Name:   "Target Channel",
			Value:  fmt.Sprintf("<#%s>", data.TargetChannel),
			Inline: true,
		},
	}

	if data.IsEdit && data.OriginalMsgID != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Original Message",
			Value:  fmt.Sprintf("`%s`", data.OriginalMsgID),
			Inline: true,
		})
	}

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: fmt.Sprintf("```\n%s\n```", data.Content),
		Color:       color,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
}

// RenderPreviewEmbed returns just the embed for cases where the caller
// needs to customize the response wrapper.
func RenderPreviewEmbed(data PreviewData) *discordgo.MessageEmbed {
	return buildPreviewEmbed(data)
}

// buildPreviewContent creates a text representation of the preview for display.
// Used for simple text-based previews without embeds.
func buildPreviewContent(data PreviewData) string {
	var title string
	var actionVerb string

	if data.IsEdit {
		title = "**Edit Preview**"
		actionVerb = "Apply"
	} else {
		title = "**Compose Preview**"
		actionVerb = "Post"
	}

	var content string = data.Content
	if content != "" {
		// Quote the content
		content = "> " + content
	}

	var result string = fmt.Sprintf("%s\n\n%s\n\n", title, content)
	result += fmt.Sprintf("**Target Channel:** <#%s>\n", data.TargetChannel)

	if data.IsEdit && data.OriginalMsgID != "" {
		result += fmt.Sprintf("**Original Message:** `%s`\n", data.OriginalMsgID)
	}

	result += "\n**Posted via:** GuildMessageProxy\n"
	result += fmt.Sprintf("Click **%s** to confirm", actionVerb)

	return result
}
