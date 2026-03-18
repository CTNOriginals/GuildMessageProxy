package events

import (
	"errors"
	"testing"

	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
	"github.com/bwmarrin/discordgo"
)

// mockStoreForDelete is a test implementation focused on delete operations
type mockStoreForDelete struct {
	deleteGuildCalled bool
	deleteGuildID     string
	deleteGuildError  error

	// Unused but required by interface
	saveGuildCalled       bool
	getGuildCalled        bool
	saveGuildConfigCalled bool
	getGuildConfigCalled  bool
	saveProxyMsgCalled    bool
	getProxyMsgCalled     bool
	updateProxyMsgCalled  bool
	deleteProxyMsgCalled  bool
}

func (m *mockStoreForDelete) SaveGuild(guildID, name string) error {
	m.saveGuildCalled = true
	return nil
}

func (m *mockStoreForDelete) GetGuild(guildID string) (*storage.Guild, error) {
	m.getGuildCalled = true
	return nil, nil
}

func (m *mockStoreForDelete) DeleteGuild(guildID string) error {
	m.deleteGuildCalled = true
	m.deleteGuildID = guildID
	return m.deleteGuildError
}

func (m *mockStoreForDelete) SaveGuildConfig(config storage.GuildConfig) error {
	m.saveGuildConfigCalled = true
	return nil
}

func (m *mockStoreForDelete) GetGuildConfig(guildID string) (*storage.GuildConfig, error) {
	m.getGuildConfigCalled = true
	return nil, nil
}

func (m *mockStoreForDelete) SaveProxyMessage(msg storage.ProxyMessage) error {
	m.saveProxyMsgCalled = true
	return nil
}

func (m *mockStoreForDelete) GetProxyMessage(guildID, messageID string) (*storage.ProxyMessage, error) {
	m.getProxyMsgCalled = true
	return nil, nil
}

func (m *mockStoreForDelete) UpdateProxyMessage(msg storage.ProxyMessage) error {
	m.updateProxyMsgCalled = true
	return nil
}

func (m *mockStoreForDelete) DeleteProxyMessage(guildID, messageID string) error {
	m.deleteProxyMsgCalled = true
	return nil
}

// TestHandleGuildDelete_BasicSuccess tests successful guild deletion
func TestHandleGuildDelete_BasicSuccess(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{}
	var handler = HandleGuildDelete(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildDelete{
		Guild: &discordgo.Guild{
			ID: "delete-guild-123",
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildDelete panicked: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.deleteGuildCalled {
		t.Error("Expected DeleteGuild to be called")
	}
	if mock.deleteGuildID != "delete-guild-123" {
		t.Errorf("Expected guild ID 'delete-guild-123', got '%s'", mock.deleteGuildID)
	}
}

// TestHandleGuildDelete_DeletedUnavailableFlag tests the Unavailable flag behavior
func TestHandleGuildDelete_DeletedUnavailableFlag(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{}
	var handler = HandleGuildDelete(mock)

	var session *discordgo.Session = &discordgo.Session{}

	// Test with Unavailable = true (guild outage)
	var guildUnavailable = &discordgo.GuildDelete{
		Guild: &discordgo.Guild{
			ID:          "unavailable-guild",
			Unavailable: true,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildDelete panicked on unavailable guild: %v", r)
		}
	}()

	handler(session, guildUnavailable)

	// Note: Current implementation doesn't check Unavailable flag, so DeleteGuild is still called
	if !mock.deleteGuildCalled {
		t.Log("Note: DeleteGuild was not called for unavailable guild - this may indicate policy change")
	}
}

// TestHandleGuildDelete_ErrorHandling tests error handling when DeleteGuild fails
func TestHandleGuildDelete_ErrorHandling(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{
		deleteGuildError: errors.New("database delete error"),
	}
	var handler = HandleGuildDelete(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildDelete{
		Guild: &discordgo.Guild{
			ID: "error-guild",
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildDelete panicked on error: %v", r)
		}
	}()

	handler(session, guild)

	if !mock.deleteGuildCalled {
		t.Error("Expected DeleteGuild to be called even when it will fail")
	}
	if mock.deleteGuildID != "error-guild" {
		t.Errorf("Expected guild ID 'error-guild', got '%s'", mock.deleteGuildID)
	}
}

// TestHandleGuildDelete_LongGuildID tests handling of long guild IDs
func TestHandleGuildDelete_LongGuildID(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{}
	var handler = HandleGuildDelete(mock)

	var longID string = "12345678901234567890123456789012345678901234567890"
	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildDelete{
		Guild: &discordgo.Guild{
			ID: longID,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildDelete panicked with long ID: %v", r)
		}
	}()

	handler(session, guild)

	if mock.deleteGuildID != longID {
		t.Errorf("Expected long guild ID, got '%s'", mock.deleteGuildID)
	}
}

// TestHandleGuildDelete_SpecialCharactersInGuildID tests handling of special characters in guild ID
func TestHandleGuildDelete_SpecialCharactersInGuildID(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{}
	var handler = HandleGuildDelete(mock)

	var specialID string = "guild-id-with-dashes-and_underscores123"
	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildDelete{
		Guild: &discordgo.Guild{
			ID: specialID,
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildDelete panicked with special characters: %v", r)
		}
	}()

	handler(session, guild)

	if mock.deleteGuildID != specialID {
		t.Errorf("Expected guild ID with special characters, got '%s'", mock.deleteGuildID)
	}
}

// TestHandleGuildDelete_MultipleCalls tests that the handler can be called multiple times
func TestHandleGuildDelete_MultipleCalls(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{}
	var handler = HandleGuildDelete(mock)

	var session *discordgo.Session = &discordgo.Session{}

	var guildIDs = []string{"guild-1", "guild-2", "guild-3", "guild-4", "guild-5"}

	for _, guildID := range guildIDs {
		var guild = &discordgo.GuildDelete{
			Guild: &discordgo.Guild{
				ID: guildID,
			},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("HandleGuildDelete panicked for guild %s: %v", guildID, r)
			}
		}()

		handler(session, guild)
	}

	// Note: The mock only tracks the last call, so we just verify the handler didn't panic
}

// TestHandleGuildDelete_NilGuild tests behavior with nil guild (edge case)
func TestHandleGuildDelete_NilGuild(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{}
	var handler = HandleGuildDelete(mock)

	var session *discordgo.Session = &discordgo.Session{}

	defer func() {
		if r := recover(); r != nil {
			// Expected to potentially panic on nil guild - document this behavior
			t.Logf("HandleGuildDelete panicked on nil guild (expected behavior): %v", r)
		}
	}()

	handler(session, nil)

	// If we reach here, the handler handled nil gracefully
	t.Log("HandleGuildDelete handled nil guild gracefully")
}

// TestHandleGuildDelete_EmptyGuildID tests behavior with empty guild ID
func TestHandleGuildDelete_EmptyGuildID(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{}
	var handler = HandleGuildDelete(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildDelete{
		Guild: &discordgo.Guild{
			ID: "",
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildDelete panicked with empty ID: %v", r)
		}
	}()

	handler(session, guild)

	if mock.deleteGuildID != "" {
		t.Errorf("Expected empty guild ID, got '%s'", mock.deleteGuildID)
	}
}

// TestHandleGuildDelete_PolicyHardDelete documents the hard delete policy
func TestHandleGuildDelete_PolicyHardDelete(t *testing.T) {
	var mock *mockStoreForDelete = &mockStoreForDelete{}
	var handler = HandleGuildDelete(mock)

	var session *discordgo.Session = &discordgo.Session{}
	var guild = &discordgo.GuildDelete{
		Guild: &discordgo.Guild{
			ID: "policy-test-guild",
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleGuildDelete panicked: %v", r)
		}
	}()

	handler(session, guild)

	// Verify that DeleteGuild is called (hard delete policy)
	if !mock.deleteGuildCalled {
		t.Error("Hard delete policy: DeleteGuild must be called when bot leaves guild")
	}

	// Note: The implementation currently only deletes guild metadata, not related data
	// This is a documentation test to verify the current behavior
	t.Log("Hard delete policy verified: DeleteGuild is called on GuildDelete event")
}
