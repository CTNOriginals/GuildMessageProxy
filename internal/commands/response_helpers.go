// Package commands provides response builder helpers to reduce code duplication.
package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
)

// BuildRetryButtons creates retry/cancel button components.
func BuildRetryButtons(retryButtonID string, cancelButtonID string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Retry",
					Style:    discordgo.PrimaryButton,
					CustomID: retryButtonID,
				},
				discordgo.Button{
					Label:    "Cancel",
					Style:    discordgo.SecondaryButton,
					CustomID: cancelButtonID,
				},
			},
		},
	}
}

// RespondWithRetry sends an error message with retry/cancel button options.
func RespondWithRetry(s *discordgo.Session, i *discordgo.InteractionCreate, errorMsg string, retryButtonID string, cancelButtonID string) error {
	components := BuildRetryButtons(retryButtonID, cancelButtonID)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    errorMsg,
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})

	if err != nil {
		var userID string
		if i.Member != nil && i.Member.User != nil {
			userID = i.Member.User.ID
		}
		logging.Error("Failed to send error response with retry/cancel buttons",
			logging.Err("error", err),
			logging.String("user_id", userID),
		)
	}

	return err
}

// RespondWithSuccessAndJumpURL sends a success message with a jump link button.
func RespondWithSuccessAndJumpURL(s *discordgo.Session, i *discordgo.InteractionCreate, content string, jumpURL string, buttonLabel string) error {
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: buttonLabel,
					Style: discordgo.LinkButton,
					URL:   jumpURL,
				},
			},
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})

	if err != nil {
		var userID string
		if i.Member != nil && i.Member.User != nil {
			userID = i.Member.User.ID
		}
		logging.Error("Failed to send success response",
			logging.Err("error", err),
			logging.String("user_id", userID),
		)
	}

	return err
}
