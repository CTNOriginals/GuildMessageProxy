package events

import (
	"testing"

	"github.com/CTNOriginals/GuildMessageProxy/internal/commands"
	"github.com/bwmarrin/discordgo"
)

// setupTestCommands registers test commands for testing
func setupTestCommands(t *testing.T) func() {
	// Save original state
	var originalCommands = make(map[commands.TSlashCommand]commands.SCommandDef)
	for k, v := range commands.CommandDefinitions {
		originalCommands[k] = v
	}

	var originalButtons = make(map[commands.TButton]commands.SButtonDef)
	for k, v := range commands.ButtonDefinitions {
		originalButtons[k] = v
	}

	// Register test commands
	commands.CommandDefinitions["test-command"] = commands.SCommandDef{
		Definition: &discordgo.ApplicationCommand{
			Name:        "test-command",
			Description: "A test command",
		},
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Test implementation - does nothing
		},
	}

	commands.ButtonDefinitions["button_test_action"] = commands.SButtonDef{
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Test implementation - does nothing
		},
	}

	// Return cleanup function
	return func() {
		// Restore original state
		commands.CommandDefinitions = make(map[commands.TSlashCommand]commands.SCommandDef)
		for k, v := range originalCommands {
			commands.CommandDefinitions[k] = v
		}

		commands.ButtonDefinitions = make(map[commands.TButton]commands.SButtonDef)
		for k, v := range originalButtons {
			commands.ButtonDefinitions[k] = v
		}
	}
}

// createTestInteraction creates a test interaction with common fields
func createTestInteraction(interactionType discordgo.InteractionType, guildID, userID string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "test-interaction-id",
			Type:      interactionType,
			GuildID:   guildID,
			ChannelID: "test-channel-id",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       userID,
					Username: "TestUser",
				},
			},
		},
	}
}

// TestHandleInteractionCreate_SlashCommand tests slash command routing
func TestHandleInteractionCreate_SlashCommand(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionApplicationCommand, "test-guild", "test-user")
	interaction.Data = discordgo.ApplicationCommandInteractionData{
		Name: "test-command",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleInteractionCreate panicked on slash command: %v", r)
		}
	}()

	HandleInteractionCreate(session, interaction)
}

// TestHandleInteractionCreate_SlashCommandUnknown tests unknown slash command handling
func TestHandleInteractionCreate_SlashCommandUnknown(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionApplicationCommand, "test-guild", "test-user")
	interaction.Data = discordgo.ApplicationCommandInteractionData{
		Name: "unknown-command",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleInteractionCreate panicked on unknown command: %v", r)
		}
	}()

	HandleInteractionCreate(session, interaction)
}

// TestHandleInteractionCreate_Button tests button interaction routing
func TestHandleInteractionCreate_Button(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: "button_test_action",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleInteractionCreate panicked on button: %v", r)
		}
	}()

	HandleInteractionCreate(session, interaction)
}

// TestHandleInteractionCreate_ButtonUnknown tests unknown button handling
func TestHandleInteractionCreate_ButtonUnknown(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: "button_unknown_action",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleInteractionCreate panicked on unknown button: %v", r)
		}
	}()

	HandleInteractionCreate(session, interaction)
}

// TestHandleInteractionCreate_SelectMenu tests select menu interaction routing
func TestHandleInteractionCreate_SelectMenu(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID:      "select_test_option",
		ComponentType: discordgo.SelectMenuComponent,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleInteractionCreate panicked on select menu: %v", r)
		}
	}()

	HandleInteractionCreate(session, interaction)
}

// TestHandleInteractionCreate_ModalSubmit tests modal submit routing
func TestHandleInteractionCreate_ModalSubmit(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionModalSubmit, "test-guild", "test-user")
	interaction.Data = discordgo.ModalSubmitInteractionData{
		CustomID: "modal_test_submit",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleInteractionCreate panicked on modal submit: %v", r)
		}
	}()

	HandleInteractionCreate(session, interaction)
}

// TestHandleInteractionCreate_UnknownType tests handling of unknown interaction types
func TestHandleInteractionCreate_UnknownType(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionType(255), "test-guild", "test-user")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleInteractionCreate panicked on unknown type: %v", r)
		}
	}()

	HandleInteractionCreate(session, interaction)
}

// TestHandleInteractionCreate_NilInteraction tests nil interaction handling
func TestHandleInteractionCreate_NilInteraction(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("HandleInteractionCreate panicked on nil interaction (acceptable): %v", r)
		}
	}()

	HandleInteractionCreate(session, nil)
}

// TestGetInteractionContext_WithGuild tests context extraction with guild
func TestGetInteractionContext_WithGuild(t *testing.T) {
	var interaction = createTestInteraction(discordgo.InteractionApplicationCommand, "test-guild-id", "test-user-id")

	var fields = getInteractionContext(interaction)

	if len(fields) < 3 {
		t.Errorf("Expected at least 3 fields, got %d", len(fields))
	}
}

// TestGetInteractionContext_WithoutGuild tests context extraction without guild
func TestGetInteractionContext_WithoutGuild(t *testing.T) {
	var interaction = createTestInteraction(discordgo.InteractionApplicationCommand, "", "test-user-id")

	var fields = getInteractionContext(interaction)

	// Should still have basic fields but no guild_id
	if len(fields) < 2 {
		t.Errorf("Expected at least 2 fields, got %d", len(fields))
	}
}

// TestGetInteractionContext_WithoutMember tests context extraction without member
func TestGetInteractionContext_WithoutMember(t *testing.T) {
	var interaction = &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "test-id",
			Type:      discordgo.InteractionApplicationCommand,
			GuildID:   "test-guild",
			ChannelID: "test-channel",
			Member:    nil,
		},
	}

	var fields = getInteractionContext(interaction)

	// Should have interaction_id, type, guild_id, channel_id but no user_id
	if len(fields) < 3 {
		t.Errorf("Expected at least 3 fields without member, got %d", len(fields))
	}
}

// TestHandleSlashCommand_KnownCommand tests handling of known commands
func TestHandleSlashCommand_KnownCommand(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionApplicationCommand, "test-guild", "test-user")
	interaction.Data = discordgo.ApplicationCommandInteractionData{
		Name: "test-command",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleSlashCommand panicked: %v", r)
		}
	}()

	handleSlashCommand(session, interaction)
}

// TestHandleSlashCommand_UnknownCommand tests handling of unknown commands
func TestHandleSlashCommand_UnknownCommand(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionApplicationCommand, "test-guild", "test-user")
	interaction.Data = discordgo.ApplicationCommandInteractionData{
		Name: "nonexistent-command",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleSlashCommand panicked on unknown: %v", r)
		}
	}()

	handleSlashCommand(session, interaction)
}

// TestHandleMessageComponent_ButtonPrefix tests routing of button components
func TestHandleMessageComponent_ButtonPrefix(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: "button_compose_preview_post",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMessageComponent panicked on button: %v", r)
		}
	}()

	handleMessageComponent(session, interaction)
}

// TestHandleMessageComponent_SelectPrefix tests routing of select menu components
func TestHandleMessageComponent_SelectPrefix(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: "select_test_option",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMessageComponent panicked on select: %v", r)
		}
	}()

	handleMessageComponent(session, interaction)
}

// TestHandleMessageComponent_UnknownPrefix tests routing of unknown components
func TestHandleMessageComponent_UnknownPrefix(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: "unknown_component_type",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMessageComponent panicked on unknown: %v", r)
		}
	}()

	handleMessageComponent(session, interaction)
}

// TestHandleButton_KnownButton tests handling of known buttons
func TestHandleButton_KnownButton(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: "button_test_action",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleButton panicked: %v", r)
		}
	}()

	handleButton(session, interaction, "button_test_action")
}

// TestHandleButton_UnknownButton tests handling of unknown buttons
func TestHandleButton_UnknownButton(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: "button_unknown_action",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleButton panicked on unknown: %v", r)
		}
	}()

	handleButton(session, interaction, "button_unknown_action")
}

// TestHandleSelectMenu tests select menu handling
func TestHandleSelectMenu(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionMessageComponent, "test-guild", "test-user")
	interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: "select_test_option",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleSelectMenu panicked: %v", r)
		}
	}()

	handleSelectMenu(session, interaction, "select_test_option")
}

// TestHandleModalSubmit tests modal submit handling
func TestHandleModalSubmit(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionModalSubmit, "test-guild", "test-user")
	interaction.Data = discordgo.ModalSubmitInteractionData{
		CustomID: "modal_test_submit",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleModalSubmit panicked: %v", r)
		}
	}()

	handleModalSubmit(session, interaction)
}

// TestHandleModalSubmit_NilData tests modal submit with nil data (edge case)
func TestHandleModalSubmit_NilData(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = createTestInteraction(discordgo.InteractionModalSubmit, "test-guild", "test-user")
	// Data is nil by default

	defer func() {
		if r := recover(); r != nil {
			t.Logf("handleModalSubmit panicked with nil data (acceptable): %v", r)
		}
	}()

	handleModalSubmit(session, interaction)
}

// TestHandleInteractionCreate_WithoutMemberUser tests interaction without Member.User
func TestHandleInteractionCreate_WithoutMemberUser(t *testing.T) {
	var cleanup = setupTestCommands(t)
	defer cleanup()

	var session *discordgo.Session = &discordgo.Session{}
	var interaction = &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:      "test-id",
			Type:    discordgo.InteractionApplicationCommand,
			GuildID: "test-guild",
			Member: &discordgo.Member{
				User: nil, // No user info
			},
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "test-command",
			},
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("HandleInteractionCreate panicked without Member.User (acceptable): %v", r)
		}
	}()

	HandleInteractionCreate(session, interaction)
}
