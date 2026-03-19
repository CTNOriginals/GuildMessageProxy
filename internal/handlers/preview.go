package handlers

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/colors"
)

// PreviewData contains all information needed to render a preview
type PreviewData struct {
	// Content is the message content to be previewed and posted
	Content string
	// TargetChannel is the Discord channel ID where the message will be sent
	TargetChannel string
	// IsEdit indicates whether this is an edit preview (true) or compose preview (false)
	IsEdit bool
	// OriginalMsgID is the ID of the original message being edited (only used when IsEdit is true)
	OriginalMsgID string
	// GuildID is the Discord guild ID where the message will be posted (needed for jump URLs)
	GuildID string
	// ConfirmButtonID is the CustomID for the confirm button ("Post" for compose, "Apply" for edit)
	// Caller provides this to avoid import cycles
	ConfirmButtonID string
	// CancelButtonID is the CustomID for the cancel button
	// Caller provides this to avoid import cycles
	CancelButtonID string
	// ExpiresAt is the time when the draft expires
	ExpiresAt time.Time
}

// RenderPreviewResponse creates an ephemeral preview response with confirmation buttons.
// It returns a complete InteractionResponse ready to be sent to Discord.
//
// The data parameter contains all information needed to render the preview including
// the message content, target channel, and button CustomIDs.
//
// Button logic: For compose operations (IsEdit=false), the confirm button is labeled
// "Post" and will send a new message. For edit operations (IsEdit=true), the confirm
// button is labeled "Apply" and will update an existing message. The cancel button
// discards the operation in both cases.
//
// Returns an InteractionResponse containing an ephemeral message with an embed
// showing the preview and an action row with the confirmation and cancel buttons.
func RenderPreviewResponse(data PreviewData) *discordgo.InteractionResponse {
	var embed *discordgo.MessageEmbed = buildPreviewEmbed(data)
	var content string = buildPreviewContent(data)

	// Build action row with appropriate buttons
	var components []discordgo.MessageComponent

	// Determine button label based on edit/compose flow
	var confirmLabel string = "Post Message"
	if data.IsEdit {
		confirmLabel = "Apply Edit"
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

// formatDurationUntil returns a human-readable string representing the duration until the given time.
// If the time is zero, it returns "Unknown". If the time has passed, it returns "Expired".
// For future times, it returns simple integer durations like "2 days", "23 hours", or "45 minutes".
func formatDurationUntil(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}

	duration := time.Until(t)
	if duration <= 0 {
		return "Expired"
	}

	hours := int(duration.Hours())
	if hours >= 24 {
		return fmt.Sprintf("%d days", hours/24)
	}
	if hours >= 1 {
		return fmt.Sprintf("%d hours", hours)
	}

	minutes := int(duration.Minutes())
	return fmt.Sprintf("%d minutes", minutes)
}

// buildPreviewEmbed creates a Discord embed for preview display.
// It generates an embed with appropriate title, color, and footer text based on
// whether this is an edit or compose operation. The embed includes the message
// content formatted in a code block, the target channel information, and for edits,
// the original message ID.
func buildPreviewEmbed(data PreviewData) *discordgo.MessageEmbed {
	var title string
	var color int
	var footerText string

	if data.IsEdit {
		title = "Edit Preview"
		color = colors.Edit
		footerText = "Review your changes above. Click 'Apply Edit' to save, or 'Cancel' to discard."
	} else {
		title = "Compose Preview"
		color = colors.Primary
		footerText = "Review your message above. Click 'Post Message' to send, or 'Cancel' to discard."
	}

	// Check if draft expires in under 1 hour and apply warning styling
	var expiresSoon bool
	if !data.ExpiresAt.IsZero() {
		timeUntil := time.Until(data.ExpiresAt)
		expiresSoon = timeUntil > 0 && timeUntil < time.Hour
	}

	if expiresSoon {
		color = colors.Warning
		title = ":warning: " + title
	}

	// Build fields for metadata
	var fields []*discordgo.MessageEmbedField = []*discordgo.MessageEmbedField{
		{
			Name:   "Target Channel",
			Value:  fmt.Sprintf("<#%s>", data.TargetChannel),
			Inline: true,
		},
		{
			Name:   "Expires",
			Value:  fmt.Sprintf("Draft expires in %s", formatDurationUntil(data.ExpiresAt)),
			Inline: true,
		},
	}

	if data.IsEdit && data.OriginalMsgID != "" && data.GuildID != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Original Message",
			Value:  fmt.Sprintf("[Jump to message](https://discord.com/channels/%s/%s/%s)", data.GuildID, data.TargetChannel, data.OriginalMsgID),
			Inline: true,
		})
	}

	// Add expiration warning if draft expires soon
	if expiresSoon {
		footerText += " - This draft expires soon. Post or lose your work."
	}

	// Add storage warning
	footerText += " Note: Drafts are stored temporarily and will be lost if the bot restarts."

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
// needs to customize the response wrapper. This is useful when the caller wants
// to build a custom InteractionResponse but still use the standard preview embed
// format. The embed will have the same content and styling as the one used by
// RenderPreviewResponse.
func RenderPreviewEmbed(data PreviewData) *discordgo.MessageEmbed {
	return buildPreviewEmbed(data)
}

// buildPreviewContent creates a text representation of the preview for display.
// It generates a formatted text preview with the message title, quoted content,
// target channel information, and original message ID for edits. The content is
// used for simple text-based previews alongside the main embed. The text includes
// appropriate action verbs ("Apply" for edits, "Post" for compose) to match the
// button labels shown in the interactive response.
func buildPreviewContent(data PreviewData) string {
	var title string
	var actionVerb string

	if data.IsEdit {
		title = "**Edit Preview**"
		actionVerb = "Apply Edit"
	} else {
		title = "**Compose Preview**"
		actionVerb = "Post Message"
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
