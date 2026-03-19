package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
)

// CommandDefinitions maps slash command types to their full definitions.
// This is the source of truth for command registration.
// Commands are registered by individual command files in their init() functions.
var CommandDefinitions MCommandDefinitions = make(MCommandDefinitions)

// desiredCommands returns the list of ApplicationCommand definitions to register.
func desiredCommands() []*discordgo.ApplicationCommand {
	var commands []*discordgo.ApplicationCommand
	for _, def := range CommandDefinitions {
		commands = append(commands, def.Definition)
	}
	return commands
}

// commandsEqual compares two ApplicationCommand definitions for equality.
// Compares name, description, and options. Ignores Discord-managed fields (ID, version, etc.)
func commandsEqual(a, b *discordgo.ApplicationCommand) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Description != b.Description {
		return false
	}
	if len(a.Options) != len(b.Options) {
		return false
	}
	for i := range a.Options {
		if !optionsEqual(a.Options[i], b.Options[i]) {
			return false
		}
	}
	return true
}

// optionsEqual compares two ApplicationCommandOption definitions.
func optionsEqual(a, b *discordgo.ApplicationCommandOption) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Description != b.Description {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	if a.Required != b.Required {
		return false
	}
	if len(a.Options) != len(b.Options) {
		return false
	}
	for i := range a.Options {
		if !optionsEqual(a.Options[i], b.Options[i]) {
			return false
		}
	}
	return true
}

// needsSync compares desired commands with existing commands.
// Returns true if sync is needed (commands differ or counts differ).
func needsSync(desired, existing []*discordgo.ApplicationCommand) bool {
	if len(desired) != len(existing) {
		return true
	}

	// Build map of existing commands by name for comparison
	existingMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range existing {
		existingMap[cmd.Name] = cmd
	}

	// Check each desired command exists and matches
	for _, desiredCmd := range desired {
		existingCmd, ok := existingMap[desiredCmd.Name]
		if !ok {
			return true
		}
		if !commandsEqual(desiredCmd, existingCmd) {
			return true
		}
	}

	return false
}

// SyncCommands synchronizes application commands with Discord.
// If guildID is empty, syncs globally. If set, syncs to that specific guild.
// Uses bulk overwrite (PUT) only when definitions differ.
func SyncCommands(session *discordgo.Session, guildID string) error {
	var desired []*discordgo.ApplicationCommand = desiredCommands()

	// Need application ID (bot user ID) for API calls
	var appID string = session.State.User.ID

	// Fetch existing commands
	var existing []*discordgo.ApplicationCommand
	var err error

	existing, err = session.ApplicationCommands(appID, guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing commands: %w", err)
	}

	// Check if sync is needed
	if !needsSync(desired, existing) {
		logging.Info("Commands are up to date, skipping sync")
		return nil
	}

	// Perform bulk overwrite
	logging.Info("Syncing commands", logging.Int("count", len(desired)))

	var synced []*discordgo.ApplicationCommand
	synced, err = session.ApplicationCommandBulkOverwrite(appID, guildID, desired)

	if err != nil {
		return fmt.Errorf("failed to sync commands: %w", err)
	}

	logging.Info("Successfully synced commands", logging.Int("count", len(synced)))
	return nil
}
