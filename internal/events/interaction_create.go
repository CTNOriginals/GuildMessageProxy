package events

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/commands"
)

// HandleInteractionCreate routes all interaction types to their appropriate handlers.
// Supports: slash commands, message components (buttons, select menus), and modal submits.
func HandleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		handleSlashCommand(s, i)
	case discordgo.InteractionMessageComponent:
		handleMessageComponent(s, i)
	case discordgo.InteractionModalSubmit:
		handleModalSubmit(s, i)
	default:
		log.Printf("Unknown interaction type: %d", i.Type)
	}
}

// handleSlashCommand routes slash commands by their name.
func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var data discordgo.ApplicationCommandInteractionData = i.ApplicationCommandData()
	var cmdType commands.TSlashCommand = commands.TSlashCommand(data.Name)

	if def, ok := commands.CommandDefinitions[cmdType]; ok {
		def.Execute(s, i)
	} else {
		log.Printf("Unknown slash command: %s", data.Name)
		RespondToUser(s, i, "Unknown command: "+data.Name)
	}
}

// handleMessageComponent routes buttons and select menus by their CustomID.
func handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var data discordgo.MessageComponentInteractionData = i.MessageComponentData()
	var customID string = data.CustomID

	// Route based on CustomID prefix
	switch {
	case strings.HasPrefix(customID, "button_"):
		handleButton(s, i, customID)
	case strings.HasPrefix(customID, "select_"):
		handleSelectMenu(s, i, customID)
	default:
		log.Printf("Unknown message component CustomID: %s", customID)
		RespondToUser(s, i, "Unknown component: "+customID)
	}
}

// handleButton routes button interactions.
func handleButton(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	_ = commands.TButton(customID) // Will be used when button definitions are registered

	// Placeholder: No button definitions registered yet in MVP
	// This will be populated as features are added
	log.Printf("Button interaction received: %s (no handler registered)", customID)
	RespondToUser(s, i, "Button action not yet implemented: "+customID)
}

// handleSelectMenu routes select menu interactions.
func handleSelectMenu(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	_ = commands.TSelectMenu(customID) // Will be used when select menu definitions are registered

	// Placeholder: No select menu definitions registered yet in MVP
	// This will be populated as features are added
	log.Printf("Select menu interaction received: %s (no handler registered)", customID)
	RespondToUser(s, i, "Select menu action not yet implemented: "+customID)
}

// handleModalSubmit routes modal submit interactions by their CustomID.
func handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var data discordgo.ModalSubmitInteractionData = i.ModalSubmitData()
	var customID string = data.CustomID
	var modalType commands.TModalSubmit = commands.TModalSubmit(customID)

	// Placeholder: No modal definitions registered yet in MVP
	// This will be populated as features are added
	log.Printf("Modal submit received: %s (no handler registered)", customID)
	_ = modalType // Will be used when modal definitions are registered

	RespondToUser(s, i, "Modal submit not yet implemented: "+customID)
}
