package events

import (
	"errors"
	"testing"

	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
	"github.com/bwmarrin/discordgo"
)

// mockStore is a test implementation of storage.Store
type mockStore struct {
	saveGuildCalled bool
	saveGuildID     string
	saveGuildName   string
	saveGuildError  error

	getGuildConfigCalled bool
	getGuildConfigID     string
	getGuildConfigResult *storage.GuildConfig
	getGuildConfigError  error

	saveGuildConfigCalled bool
	saveGuildConfigResult storage.GuildConfig
	saveGuildConfigError  error

	// Unused but required by interface
	getGuildCalled       bool
	deleteGuildCalled    bool
	saveProxyMsgCalled   bool
	getProxyMsgCalled    bool
	updateProxyMsgCalled bool
	deleteProxyMsgCalled bool
}

func (m *mockStore) SaveGuild(guildID, name string) error {
	m.saveGuildCalled = true
	m.saveGuildID = guildID
	m.saveGuildName = name
	return m.saveGuildError
}

func (m *mockStore) GetGuild(guildID string) (*storage.Guild, error) {
	m.getGuildCalled = true
	return nil, nil
}

func (m *mockStore) DeleteGuild(guildID string) error {
	m.deleteGuildCalled = true
	return nil
}

func (m *mockStore) SaveGuildConfig(config storage.GuildConfig) error {
	m.saveGuildConfigCalled = true
	m.saveGuildConfigResult = config
	return m.saveGuildConfigError
}

func (m *mockStore) GetGuildConfig(guildID string) (*storage.GuildConfig, error) {
	m.getGuildConfigCalled = true
	m.getGuildConfigID = guildID
	return m.getGuildConfigResult, m.getGuildConfigError
}

func (m *mockStore) SaveProxyMessage(msg storage.ProxyMessage) error {
	m.saveProxyMsgCalled = true
	return nil
}

func (m *mockStore) GetProxyMessage(guildID, messageID string) (*storage.ProxyMessage, error) {
	m.getProxyMsgCalled = true
	return nil, nil
}

func (m *mockStore) UpdateProxyMessage(msg storage.ProxyMessage) error {
	m.updateProxyMsgCalled = true
	return nil
}

func (m *mockStore) DeleteProxyMessage(guildID, messageID string) error {
	m.deleteProxyMsgCalled = true
	return nil
}

// TestHandleGuildCreate_BasicSuccess tests successful guild creation
func TestHandleGuildCreate_BasicSuccess(t *testing.T) {
	var mock *mockStore = &mockStore{}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "123456789",
			Name:        "Test Guild",
			MemberCount: 100,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.saveGuildCalled {
		t.Error("Expected SaveGuild to be called")
	}
	if mock.saveGuildID != "123456789" {
		t.Errorf("Expected guild ID '123456789', got '%s'", mock.saveGuildID)
	}
	if mock.saveGuildName != "Test Guild" {
		t.Errorf("Expected guild name 'Test Guild', got '%s'", mock.saveGuildName)
	}
	if !mock.getGuildConfigCalled {
		t.Error("Expected GetGuildConfig to be called")
	}
}

// TestHandleGuildCreate_CreatesDefaultConfig tests that default config is created when none exists
func TestHandleGuildCreate_CreatesDefaultConfig(t *testing.T) {
	var mock *mockStore = &mockStore{
		getGuildConfigResult: nil, // No config exists
	}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "987654321",
			Name:        "New Guild",
			MemberCount: 50,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.saveGuildConfigCalled {
		t.Error("Expected SaveGuildConfig to be called when no config exists")
	}
	if mock.saveGuildConfigResult.GuildID != "987654321" {
		t.Errorf("Expected config GuildID '987654321', got '%s'", mock.saveGuildConfigResult.GuildID)
	}
}

// TestHandleGuildCreate_SkipsConfigCreationWhenExists tests that config is not created when it already exists
func TestHandleGuildCreate_SkipsConfigCreationWhenExists(t *testing.T) {
	var existingConfig *storage.GuildConfig = &storage.GuildConfig{
		GuildID:        "existing-guild",
		AllowedRoles:   []string{"admin"},
		DefaultChannel: "general",
	}
	var mock *mockStore = &mockStore{
		getGuildConfigResult: existingConfig,
	}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "existing-guild",
			Name:        "Existing Guild",
			MemberCount: 200,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked: %v", r)
		}
	}()

	handler(session, guild)

	if mock.saveGuildConfigCalled {
		t.Error("Expected SaveGuildConfig NOT to be called when config already exists")
	}
}

// TestHandleGuildCreate_SaveGuildError tests error handling when SaveGuild fails
func TestHandleGuildCreate_SaveGuildError(t *testing.T) {
	var mock *mockStore = &mockStore{
		saveGuildError: errors.New("database error"),
	}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "error-guild",
			Name:        "Error Guild",
			MemberCount: 10,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked on error: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.saveGuildCalled {
		t.Error("Expected SaveGuild to be called")
	}
	if mock.getGuildConfigCalled {
		t.Error("Expected GetGuildConfig NOT to be called when SaveGuild fails")
	}
}

// TestHandleGuildCreate_GetGuildConfigError tests error handling when GetGuildConfig fails
func TestHandleGuildCreate_GetGuildConfigError(t *testing.T) {
	var mock *mockStore = &mockStore{
		getGuildConfigError: errors.New("config read error"),
	}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "config-error-guild",
			Name:        "Config Error Guild",
			MemberCount: 75,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked on config error: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.saveGuildCalled {
		t.Error("Expected SaveGuild to be called")
	}
	if !mock.getGuildConfigCalled {
		t.Error("Expected GetGuildConfig to be called")
	}
	if mock.saveGuildConfigCalled {
		t.Error("Expected SaveGuildConfig NOT to be called when GetGuildConfig fails")
	}
}

// TestHandleGuildCreate_SaveGuildConfigError tests error handling when SaveGuildConfig fails
func TestHandleGuildCreate_SaveGuildConfigError(t *testing.T) {
	var mock *mockStore = &mockStore{
		getGuildConfigResult: nil, // No config exists
		saveGuildConfigError: errors.New("config save error"),
	}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "save-config-error-guild",
			Name:        "Save Config Error Guild",
			MemberCount: 30,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked on save config error: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.saveGuildConfigCalled {
		t.Error("Expected SaveGuildConfig to be called")
	}
}

// TestHandleGuildCreate_SpecialCharactersInGuildName tests handling of special characters in guild name
func TestHandleGuildCreate_SpecialCharactersInGuildName(t *testing.T) {
	var mock *mockStore = &mockStore{}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "special-guild",
			Name:        "Test Guild with Special!@#$%^&*() Characters",
			MemberCount: 25,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked with special characters: %v", r)
		}
	}()

	handler(session, guild)

	if mock.saveGuildName != "Test Guild with Special!@#$%^&*() Characters" {
		t.Errorf("Expected special characters in name, got '%s'", mock.saveGuildName)
	}
}

// TestHandleGuildCreate_LargeMemberCount tests handling of large guilds
func TestHandleGuildCreate_LargeMemberCount(t *testing.T) {
	var mock *mockStore = &mockStore{}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "large-guild",
			Name:        "Very Large Guild",
			MemberCount: 500000,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked with large guild: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.saveGuildCalled {
		t.Error("Expected SaveGuild to be called for large guild")
	}
}

// TestHandleGuildCreate_ZeroMemberCount tests handling of empty guilds
func TestHandleGuildCreate_ZeroMemberCount(t *testing.T) {
	var mock *mockStore = &mockStore{}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildCreate{
		Guild: &discordgo.Guild{
			ID:          "empty-guild",
			Name:        "Empty Guild",
			MemberCount: 0,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildCreate panicked with empty guild: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.saveGuildCalled {
		t.Error("Expected SaveGuild to be called for empty guild")
	}
}

// TestHandleGuildCreate_MultipleCalls tests that the handler can be called multiple times
func TestHandleGuildCreate_MultipleCalls(t *testing.T) {
	var mock *mockStore = &mockStore{}
	var handler = HandleGuildCreate(mock)

	var session *discordgo.Session = &discordgo.Session{}

	for i := 0; i < 5; i++ {
		var guildID string = string(rune('A' + i))
		var guild = &discordgo.GuildCreate{
			Guild: &discordgo.Guild{
				ID:          guildID,
				Name:        "Guild " + guildID,
				MemberCount: i * 10,
			},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("HandleGuildCreate panicked on call %d: %v", i, r)
			}
		}()

		handler(session, guild)
	}
}
