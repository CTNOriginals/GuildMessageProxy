package commands

import "github.com/bwmarrin/discordgo"

// SCommandDef defines a slash, message, or user context command.
type SCommandDef struct {
	Definition   *discordgo.ApplicationCommand
	Execute      func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Autocomplete func(s *discordgo.Session, i *discordgo.InteractionCreate) // optional
}

// MCommandDefinitions maps slash command types to their definitions.
type MCommandDefinitions map[TSlashCommand]SCommandDef

// SSelectMenuDef defines a select menu interaction handler.
type SSelectMenuDef struct {
	Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// MSelectMenuDefinitions maps select menu types to their definitions.
type MSelectMenuDefinitions map[TSelectMenu]SSelectMenuDef

// SModalSubmitDef defines a modal submit interaction handler.
type SModalSubmitDef struct {
	Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// MModalSubmitDefinitions maps modal submit types to their definitions.
type MModalSubmitDefinitions map[TModalSubmit]SModalSubmitDef

// MMessageCommandDefinitions maps message context command types to their definitions.
// Empty for MVP - placeholder for future implementation.
type MMessageCommandDefinitions map[TMessageCommand]SCommandDef

// MUserCommandDefinitions maps user context command types to their definitions.
// Empty for MVP - placeholder for future implementation.
type MUserCommandDefinitions map[TUserCommand]SCommandDef
