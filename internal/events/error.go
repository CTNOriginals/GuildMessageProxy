package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// LogError logs an error with context information.
func LogError(err error, context string) {
	log.Printf("[Error] %s: %v", context, err)
}

// RespondToUser sends an ephemeral message to the user who triggered an interaction.
// Use this to provide feedback when something goes wrong.
func RespondToUser(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Printf("Failed to respond to user: %v", err)
	}
}

// RespondWithError sends a formatted error response to the user.
// Includes both a user-friendly message and logs the actual error.
func RespondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) {
	if err != nil {
		LogError(err, userMsg)
	}
	RespondToUser(s, i, userMsg)
}
