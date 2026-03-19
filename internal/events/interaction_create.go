// Package events provides Discord event handlers for the GuildMessageProxy bot.
// It handles incoming Discord interactions including slash commands, message components
// (buttons and select menus), and modal submits.
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

// recoverPanic recovers from panics in handlers and logs them with stack traces.
// It should be used with defer at the start of each handler function to ensure
// panics are caught and logged rather than crashing the application.
func recoverPanic(context string) {
	if r := recover(); r != nil {
		logging.Error("panic recovered in handler",
			logging.String("context", context),
			logging.String("panic", fmt.Sprintf("%v", r)),
			logging.String("stack_trace", string(debug.Stack())),
		)
	}
}

// getInteractionContext returns standard Discord context fields for structured logging.
// It extracts interaction ID, type, guild ID, channel ID, and user ID for inclusion
// in log entries to aid debugging and traceability.
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

// HandleInteractionCreate routes all Discord interaction types to their appropriate handlers.
// It supports slash commands, message components (buttons and select menus), and modal submits.
// Unknown interaction types are logged as warnings.
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

// handleSlashCommand routes slash command interactions by their command name.
// It looks up the command in CommandDefinitions and executes it if found.
// Unknown commands are logged as warnings and the user is notified.
func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer recoverPanic("handleSlashCommand")

	if i.Member == nil || i.Member.User == nil {
		logging.Error("nil member or user in interaction",
			logging.String("interaction_id", i.ID),
		)
		return
	}

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

// handleMessageComponent routes message component interactions by their CustomID prefix.
// It dispatches to handleButton for CustomIDs starting with "button_" prefix,
// and to handleSelectMenu for CustomIDs starting with "select_" prefix.
// This routing dispatch logic allows consistent naming conventions for component IDs.
func handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer recoverPanic("handleMessageComponent")

	if i.Member == nil || i.Member.User == nil {
		logging.Error("nil member or user in interaction",
			logging.String("interaction_id", i.ID),
		)
		return
	}

	var data discordgo.MessageComponentInteractionData = i.MessageComponentData()
	var customID string = data.CustomID

	// Route based on CustomID prefix:
	// - "button_" prefix routes to handleButton
	// - "select_" prefix routes to handleSelectMenu
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
// It looks up the button in ButtonDefinitions and executes it if found.
// Unknown buttons are logged as warnings and the user is notified.
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

// handleSelectMenu routes select menu interactions by their CustomID.
// Currently a placeholder for future implementation as no select menu
// definitions are registered yet in the MVP.
func handleSelectMenu(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	defer recoverPanic("handleSelectMenu")

	if i.Member == nil || i.Member.User == nil {
		logging.Error("nil member or user in interaction",
			logging.String("interaction_id", i.ID),
		)
		return
	}

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
// Currently a placeholder for future implementation as no modal submit
// definitions are registered yet in the MVP.
func handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer recoverPanic("handleModalSubmit")

	if i.Member == nil || i.Member.User == nil {
		logging.Error("nil member or user in interaction",
			logging.String("interaction_id", i.ID),
		)
		return
	}

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
