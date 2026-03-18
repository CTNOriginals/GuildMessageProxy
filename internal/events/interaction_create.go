package events

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/CTNOriginals/GuildMessageProxy/internal/commands"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/bwmarrin/discordgo"
)

// recoverPanic recovers from panics in handlers and logs them
func recoverPanic(context string) {
	if r := recover(); r != nil {
		logging.Error("panic recovered in handler",
			logging.String("context", context),
			logging.String("panic", fmt.Sprintf("%v", r)),
			logging.String("stack_trace", string(debug.Stack())),
		)
	}
}

// getInteractionContext returns standard Discord context fields for an interaction
func getInteractionContext(i *discordgo.InteractionCreate) []logging.Field {
	var fields = []logging.Field{
		logging.String("interaction_id", i.ID),
		logging.String("type", i.Type.String()),
	}
	if i.GuildID != "" {
		fields = append(fields, logging.String("guild_id", i.GuildID))
	}
	if i.ChannelID != "" {
		fields = append(fields, logging.String("channel_id", i.ChannelID))
	}
	if i.Member != nil && i.Member.User != nil {
		fields = append(fields, logging.String("user_id", i.Member.User.ID))
	}
	return fields
}

// HandleInteractionCreate routes all interaction types to their appropriate handlers.
// Supports: slash commands, message components (buttons, select menus), and modal submits.
func HandleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer recoverPanic("HandleInteractionCreate")
	logging.Debug("interaction received", getInteractionContext(i)...)

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		handleSlashCommand(s, i)
	case discordgo.InteractionMessageComponent:
		handleMessageComponent(s, i)
	case discordgo.InteractionModalSubmit:
		handleModalSubmit(s, i)
	default:
		logging.Warn("unknown interaction type",
			logging.String("interaction_id", i.ID),
			logging.Int("type", int(i.Type)),
		)
	}
}

// handleSlashCommand routes slash commands by their name.
func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer recoverPanic("handleSlashCommand")
	var data discordgo.ApplicationCommandInteractionData = i.ApplicationCommandData()
	var cmdType commands.TSlashCommand = commands.TSlashCommand(data.Name)

	var startTime = time.Now()

	if def, ok := commands.CommandDefinitions[cmdType]; ok {
		logging.Info("command execution started",
			logging.String("command", data.Name),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
		)

		def.Execute(s, i)

		logging.Info("command execution completed",
			logging.String("command", data.Name),
			logging.Duration("duration", time.Since(startTime)),
		)
	} else {
		logging.Warn("unknown slash command",
			logging.String("command", data.Name),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
		)
		RespondToUser(s, i, "Unknown command: "+data.Name)
	}
}

// handleMessageComponent routes buttons and select menus by their CustomID.
func handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer recoverPanic("handleMessageComponent")
	var data discordgo.MessageComponentInteractionData = i.MessageComponentData()
	var customID string = data.CustomID

	// Route based on CustomID prefix
	switch {
	case strings.HasPrefix(customID, "button_"):
		handleButton(s, i, customID)
	case strings.HasPrefix(customID, "select_"):
		handleSelectMenu(s, i, customID)
	default:
		logging.Warn("unknown message component",
			logging.String("custom_id", customID),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
		)
		RespondToUser(s, i, "Unknown component: "+customID)
	}
}

// handleButton routes button interactions by their CustomID.
func handleButton(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	defer recoverPanic("handleButton")
	var buttonType commands.TButton = commands.TButton(customID)

	if def, ok := commands.ButtonDefinitions[buttonType]; ok {
		logging.Info("button execution started",
			logging.String("button_id", customID),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
		)

		def.Execute(s, i)

		logging.Info("button execution completed",
			logging.String("button_id", customID),
		)
	} else {
		logging.Warn("unknown button clicked",
			logging.String("button_id", customID),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
		)
		RespondToUser(s, i, "Unknown button action: "+customID)
	}
}

// handleSelectMenu routes select menu interactions.
func handleSelectMenu(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	defer recoverPanic("handleSelectMenu")
	_ = commands.TSelectMenu(customID) // Will be used when select menu definitions are registered

	// Placeholder: No select menu definitions registered yet in MVP
	// This will be populated as features are added
	logging.Info("select menu selected",
		logging.String("select_id", customID),
		logging.String("user_id", i.Member.User.ID),
		logging.String("guild_id", i.GuildID),
	)
	RespondToUser(s, i, "Select menu action not yet implemented: "+customID)
}

// handleModalSubmit routes modal submit interactions by their CustomID.
func handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer recoverPanic("handleModalSubmit")
	var data discordgo.ModalSubmitInteractionData = i.ModalSubmitData()
	var customID string = data.CustomID
	var modalType commands.TModalSubmit = commands.TModalSubmit(customID)

	// Placeholder: No modal definitions registered yet in MVP
	// This will be populated as features are added
	logging.Info("modal submitted",
		logging.String("modal_id", customID),
		logging.String("user_id", i.Member.User.ID),
		logging.String("guild_id", i.GuildID),
	)
	_ = modalType // Will be used when modal definitions are registered

	RespondToUser(s, i, "Modal submit not yet implemented: "+customID)
}
