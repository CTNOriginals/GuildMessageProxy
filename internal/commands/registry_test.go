package commands

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

// Test desiredCommands returns correct list of commands
func TestDesiredCommands(t *testing.T) {
	var commands = desiredCommands()

	// Should return same number of commands as in CommandDefinitions
	if len(commands) != len(CommandDefinitions) {
		t.Errorf("Expected %d commands, got %d", len(CommandDefinitions), len(commands))
	}

	// Verify each command from CommandDefinitions is in the list
	for cmdName, def := range CommandDefinitions {
		var found = false
		for _, cmd := range commands {
			if cmd.Name == string(cmdName) {
				found = true
				if cmd.Description != def.Definition.Description {
					t.Errorf("Command %q description mismatch", cmdName)
				}
				break
			}
		}
		if !found {
			t.Errorf("Command %q not found in desiredCommands result", cmdName)
		}
	}
}

// Test commandsEqual with identical commands
func TestCommandsEqual_Identical(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options:     []*discordgo.ApplicationCommandOption{},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options:     []*discordgo.ApplicationCommandOption{},
	}

	if !commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return true for identical commands")
	}
}

// Test commandsEqual with different names
func TestCommandsEqual_DifferentNames(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "command-one",
		Description: "A test command",
		Options:     []*discordgo.ApplicationCommandOption{},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "command-two",
		Description: "A test command",
		Options:     []*discordgo.ApplicationCommandOption{},
	}

	if commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return false for commands with different names")
	}
}

// Test commandsEqual with different descriptions
func TestCommandsEqual_DifferentDescriptions(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "First description",
		Options:     []*discordgo.ApplicationCommandOption{},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "Second description",
		Options:     []*discordgo.ApplicationCommandOption{},
	}

	if commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return false for commands with different descriptions")
	}
}

// Test commandsEqual with different option counts
func TestCommandsEqual_DifferentOptionCounts(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "First option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options:     []*discordgo.ApplicationCommandOption{},
	}

	if commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return false for commands with different option counts")
	}
}

// Test commandsEqual with same options
func TestCommandsEqual_SameOptions(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "First option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "option2",
				Description: "Second option",
				Type:        discordgo.ApplicationCommandOptionChannel,
				Required:    false,
			},
		},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "First option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "option2",
				Description: "Second option",
				Type:        discordgo.ApplicationCommandOptionChannel,
				Required:    false,
			},
		},
	}

	if !commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return true for commands with same options")
	}
}

// Test commandsEqual with different option names
func TestCommandsEqual_DifferentOptionNames(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "First option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "different-option",
				Description: "First option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}

	if commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return false for commands with different option names")
	}
}

// Test commandsEqual with different option descriptions
func TestCommandsEqual_DifferentOptionDescriptions(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "First description",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "Second description",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}

	if commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return false for commands with different option descriptions")
	}
}

// Test commandsEqual with different option types
func TestCommandsEqual_DifferentOptionTypes(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "An option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "An option",
				Type:        discordgo.ApplicationCommandOptionChannel,
				Required:    true,
			},
		},
	}

	if commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return false for commands with different option types")
	}
}

// Test commandsEqual with different option required status
func TestCommandsEqual_DifferentOptionRequired(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "An option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "option1",
				Description: "An option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    false,
			},
		},
	}

	if commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return false for commands with different required status")
	}
}

// Test commandsEqual with nested options
func TestCommandsEqual_NestedOptions(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "group1",
				Description: "An option group",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "sub1",
						Description: "A sub option",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
		},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "group1",
				Description: "An option group",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "sub1",
						Description: "A sub option",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
		},
	}

	if !commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return true for commands with identical nested options")
	}
}

// Test commandsEqual with different nested options
func TestCommandsEqual_DifferentNestedOptions(t *testing.T) {
	var cmd1 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "group1",
				Description: "An option group",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "sub1",
						Description: "A sub option",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
		},
	}
	var cmd2 = &discordgo.ApplicationCommand{
		Name:        "test-command",
		Description: "A test command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "group1",
				Description: "An option group",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "sub2", // Different name
						Description: "A sub option",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
		},
	}

	if commandsEqual(cmd1, cmd2) {
		t.Error("commandsEqual should return false for commands with different nested options")
	}
}

// Test optionsEqual directly
func TestOptionsEqual(t *testing.T) {
	var opt1 = &discordgo.ApplicationCommandOption{
		Name:        "test",
		Description: "A test option",
		Type:        discordgo.ApplicationCommandOptionString,
		Required:    true,
	}
	var opt2 = &discordgo.ApplicationCommandOption{
		Name:        "test",
		Description: "A test option",
		Type:        discordgo.ApplicationCommandOptionString,
		Required:    true,
	}

	if !optionsEqual(opt1, opt2) {
		t.Error("optionsEqual should return true for identical options")
	}
}

// Test optionsEqual with different names
func TestOptionsEqual_DifferentNames(t *testing.T) {
	var opt1 = &discordgo.ApplicationCommandOption{
		Name:        "option1",
		Description: "A test option",
		Type:        discordgo.ApplicationCommandOptionString,
		Required:    true,
	}
	var opt2 = &discordgo.ApplicationCommandOption{
		Name:        "option2",
		Description: "A test option",
		Type:        discordgo.ApplicationCommandOptionString,
		Required:    true,
	}

	if optionsEqual(opt1, opt2) {
		t.Error("optionsEqual should return false for options with different names")
	}
}

// Test optionsEqual with nested suboptions
func TestOptionsEqual_NestedSuboptions(t *testing.T) {
	var opt1 = &discordgo.ApplicationCommandOption{
		Name:        "group",
		Description: "A group option",
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "sub1",
				Description: "A sub option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
	var opt2 = &discordgo.ApplicationCommandOption{
		Name:        "group",
		Description: "A group option",
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "sub1",
				Description: "A sub option",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}

	if !optionsEqual(opt1, opt2) {
		t.Error("optionsEqual should return true for identical nested options")
	}
}

// Test needsSync with empty lists (no sync needed)
func TestNeedsSync_EmptyLists(t *testing.T) {
	var desired []*discordgo.ApplicationCommand = []*discordgo.ApplicationCommand{}
	var existing []*discordgo.ApplicationCommand = []*discordgo.ApplicationCommand{}

	if needsSync(desired, existing) {
		t.Error("needsSync should return false for empty lists")
	}
}

// Test needsSync with different counts
func TestNeedsSync_DifferentCounts(t *testing.T) {
	var desired = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Command 1"},
	}
	var existing = []*discordgo.ApplicationCommand{}

	if !needsSync(desired, existing) {
		t.Error("needsSync should return true when counts differ")
	}
}

// Test needsSync with same commands
func TestNeedsSync_SameCommands(t *testing.T) {
	var desired = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
	}
	var existing = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
	}

	if needsSync(desired, existing) {
		t.Error("needsSync should return false for identical command sets")
	}
}

// Test needsSync with missing command
func TestNeedsSync_MissingCommand(t *testing.T) {
	var desired = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
	}
	var existing = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Command 1"},
	}

	if !needsSync(desired, existing) {
		t.Error("needsSync should return true when a command is missing")
	}
}

// Test needsSync with different command description
func TestNeedsSync_DifferentDescription(t *testing.T) {
	var desired = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "New description"},
	}
	var existing = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Old description"},
	}

	if !needsSync(desired, existing) {
		t.Error("needsSync should return true when description differs")
	}
}

// Test needsSync with extra existing commands (should still sync)
func TestNeedsSync_ExtraExistingCommands(t *testing.T) {
	var desired = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Command 1"},
	}
	var existing = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
	}

	if !needsSync(desired, existing) {
		t.Error("needsSync should return true when existing has extra commands")
	}
}

// Test needsSync with different options
func TestNeedsSync_DifferentOptions(t *testing.T) {
	var desired = []*discordgo.ApplicationCommand{
		{
			Name:        "cmd1",
			Description: "Command 1",
			Options: []*discordgo.ApplicationCommandOption{
				{Name: "opt1", Description: "Option 1", Type: discordgo.ApplicationCommandOptionString, Required: true},
			},
		},
	}
	var existing = []*discordgo.ApplicationCommand{
		{
			Name:        "cmd1",
			Description: "Command 1",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
	}

	if !needsSync(desired, existing) {
		t.Error("needsSync should return true when options differ")
	}
}

// Test needsSync ignores Discord-managed fields
func TestNeedsSync_IgnoresDiscordFields(t *testing.T) {
	var desired = &discordgo.ApplicationCommand{
		Name:        "cmd1",
		Description: "Command 1",
	}
	var existing = &discordgo.ApplicationCommand{
		ID:          "123456789",
		Name:        "cmd1",
		Description: "Command 1",
		Version:     "987654321",
	}

	var desiredList = []*discordgo.ApplicationCommand{desired}
	var existingList = []*discordgo.ApplicationCommand{existing}

	if needsSync(desiredList, existingList) {
		t.Error("needsSync should ignore Discord-managed fields (ID, Version)")
	}
}

// Test needsSync with same commands in different order
func TestNeedsSync_DifferentOrder(t *testing.T) {
	var desired = []*discordgo.ApplicationCommand{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
	}
	var existing = []*discordgo.ApplicationCommand{
		{Name: "cmd2", Description: "Command 2"},
		{Name: "cmd1", Description: "Command 1"},
	}

	if needsSync(desired, existing) {
		t.Error("needsSync should return false regardless of command order")
	}
}

// Test CommandDefinitions is properly populated
func TestCommandDefinitionsPopulation(t *testing.T) {
	// Ensure CommandDefinitions is initialized
	if CommandDefinitions == nil {
		t.Fatal("CommandDefinitions should be initialized")
	}

	// Check that all compose commands are present
	var expectedCommands = []TSlashCommand{
		ComposeCreate,
		ComposePost,
		ComposePropose,
	}

	if len(CommandDefinitions) < len(expectedCommands) {
		t.Errorf("CommandDefinitions should have at least %d commands, has %d",
			len(expectedCommands), len(CommandDefinitions))
	}

	for _, cmdName := range expectedCommands {
		var def, exists = CommandDefinitions[cmdName]
		if !exists {
			t.Errorf("CommandDefinitions should contain %q", cmdName)
			continue
		}
		if def.Definition == nil {
			t.Errorf("Command %q should have a Definition", cmdName)
		}
		if def.Execute == nil {
			t.Errorf("Command %q should have an Execute function", cmdName)
		}
	}
}

// Test that command type constants have correct values
func TestCommandTypeConstants(t *testing.T) {
	var testCases = []struct {
		name     string
		value    string
		expected string
	}{
		{"ComposeCreate", string(ComposeCreate), "compose-draft"},
		{"ComposePost", string(ComposePost), "compose-send"},
		{"ComposePropose", string(ComposePropose), "compose-edit"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, tc.value)
			}
		})
	}
}

// Test that button type constants have correct values
func TestButtonTypeConstants(t *testing.T) {
	var testCases = []struct {
		name     string
		value    string
		expected string
	}{
		{"ButtonComposePreviewPost", string(ButtonComposePreviewPost), "button_compose_preview_post"},
		{"ButtonComposePreviewCancel", string(ButtonComposePreviewCancel), "button_compose_preview_cancel"},
		{"ButtonEditPreviewApply", string(ButtonEditPreviewApply), "button_edit_preview_apply"},
		{"ButtonEditPreviewCancel", string(ButtonEditPreviewCancel), "button_edit_preview_cancel"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, tc.value)
			}
		})
	}
}

// Test MCommandDefinitions type
func TestMCommandDefinitionsType(t *testing.T) {
	var defs MCommandDefinitions = make(MCommandDefinitions)

	var cmdDef = SCommandDef{
		Definition: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Test command",
		},
		Execute: nil,
	}

	defs["test"] = cmdDef

	var retrieved SCommandDef
	var exists bool
	retrieved, exists = defs["test"]
	if !exists {
		t.Error("Expected to retrieve command from MCommandDefinitions")
	}
	if retrieved.Definition.Name != "test" {
		t.Errorf("Expected command name 'test', got %q", retrieved.Definition.Name)
	}
}

// Test MButtonDefinitions type
func TestMButtonDefinitionsType(t *testing.T) {
	var defs MButtonDefinitions = make(MButtonDefinitions)

	var buttonDef = SButtonDef{
		Execute: nil,
	}

	defs["test_button"] = buttonDef

	var retrieved SButtonDef
	var exists bool
	retrieved, exists = defs["test_button"]
	if !exists {
		t.Error("Expected to retrieve button from MButtonDefinitions")
	}
	// Execute can be nil in this test
	if retrieved.Execute != nil {
		t.Error("Expected nil Execute function")
	}
}

// Test CommandDefinitions integrity
func TestCommandDefinitionsIntegrity(t *testing.T) {
	for cmdName, def := range CommandDefinitions {
		if def.Definition == nil {
			t.Errorf("Command %q has nil Definition", cmdName)
			continue
		}
		if def.Definition.Name != string(cmdName) {
			t.Errorf("Command %q Definition.Name (%q) does not match key",
				cmdName, def.Definition.Name)
		}
		if def.Definition.Description == "" {
			t.Errorf("Command %q has empty Description", cmdName)
		}
	}
}
