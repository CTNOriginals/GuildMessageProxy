package commands

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// MockStore implements storage.Store interface for testing
type MockStore struct {
	guilds        map[string]*storage.Guild
	guildConfigs  map[string]*storage.GuildConfig
	proxyMessages map[string]*storage.ProxyMessage
	saveProxyMsgCalled   bool
	getProxyMsgCalled    bool
	updateProxyMsgCalled bool
}

func NewMockStore() *MockStore {
	return &MockStore{
		guilds:        make(map[string]*storage.Guild),
		guildConfigs:  make(map[string]*storage.GuildConfig),
		proxyMessages: make(map[string]*storage.ProxyMessage),
	}
}

func (m *MockStore) SaveGuild(guildID, name string) error {
	m.guilds[guildID] = &storage.Guild{ID: guildID, Name: name}
	return nil
}

func (m *MockStore) GetGuild(guildID string) (*storage.Guild, error) {
	return m.guilds[guildID], nil
}

func (m *MockStore) DeleteGuild(guildID string) error {
	delete(m.guilds, guildID)
	delete(m.guildConfigs, guildID)
	return nil
}

func (m *MockStore) SaveGuildConfig(config storage.GuildConfig) error {
	m.guildConfigs[config.GuildID] = &config
	return nil
}

func (m *MockStore) GetGuildConfig(guildID string) (*storage.GuildConfig, error) {
	return m.guildConfigs[guildID], nil
}

func (m *MockStore) SaveProxyMessage(msg storage.ProxyMessage) error {
	m.saveProxyMsgCalled = true
	var key = msg.GuildID + ":" + msg.MessageID
	m.proxyMessages[key] = &msg
	return nil
}

func (m *MockStore) GetProxyMessage(guildID, messageID string) (*storage.ProxyMessage, error) {
	m.getProxyMsgCalled = true
	var key = guildID + ":" + messageID
	return m.proxyMessages[key], nil
}

func (m *MockStore) UpdateProxyMessage(msg storage.ProxyMessage) error {
	m.updateProxyMsgCalled = true
	var key = msg.GuildID + ":" + msg.MessageID
	m.proxyMessages[key] = &msg
	return nil
}

func (m *MockStore) DeleteProxyMessage(guildID, messageID string) error {
	var key = guildID + ":" + messageID
	delete(m.proxyMessages, key)
	return nil
}

// Test getDraftKey generates correct keys
func TestGetDraftKey(t *testing.T) {
	var testCases = []struct {
		name     string
		userID   string
		guildID  string
		expected string
	}{
		{
			name:     "standard IDs",
			userID:   "123456789",
			guildID:  "987654321",
			expected: "123456789:987654321",
		},
		{
			name:     "empty user ID",
			userID:   "",
			guildID:  "987654321",
			expected: ":987654321",
		},
		{
			name:     "empty guild ID",
			userID:   "123456789",
			guildID:  "",
			expected: "123456789:",
		},
		{
			name:     "both empty",
			userID:   "",
			guildID:  "",
			expected: ":",
		},
		{
			name:     "long IDs",
			userID:   "123456789012345678901234567890",
			guildID:  "987654321098765432109876543210",
			expected: "123456789012345678901234567890:987654321098765432109876543210",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result = getDraftKey(tc.userID, tc.guildID)
			if result != tc.expected {
				t.Errorf("getDraftKey(%q, %q) = %q, expected %q",
					tc.userID, tc.guildID, result, tc.expected)
			}
		})
	}
}

// Test Draft struct creation and fields
func TestDraftStruct(t *testing.T) {
	var now = time.Now()
	var draft = Draft{
		UserID:        "user123",
		GuildID:       "guild456",
		ChannelID:     "channel789",
		Content:       "Test message content",
		CreatedAt:     now,
		IsEdit:        false,
		OriginalMsgID: "",
	}

	if draft.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got %q", draft.UserID)
	}
	if draft.GuildID != "guild456" {
		t.Errorf("Expected GuildID 'guild456', got %q", draft.GuildID)
	}
	if draft.ChannelID != "channel789" {
		t.Errorf("Expected ChannelID 'channel789', got %q", draft.ChannelID)
	}
	if draft.Content != "Test message content" {
		t.Errorf("Expected Content 'Test message content', got %q", draft.Content)
	}
	if draft.IsEdit != false {
		t.Errorf("Expected IsEdit false, got %v", draft.IsEdit)
	}
	if draft.OriginalMsgID != "" {
		t.Errorf("Expected OriginalMsgID empty, got %q", draft.OriginalMsgID)
	}
}

// Test Draft struct as edit proposal
func TestDraftStructAsEdit(t *testing.T) {
	var draft = Draft{
		UserID:        "user123",
		GuildID:       "guild456",
		ChannelID:     "channel789",
		Content:       "Updated message content",
		CreatedAt:     time.Now(),
		IsEdit:        true,
		OriginalMsgID: "msg456",
	}

	if draft.IsEdit != true {
		t.Errorf("Expected IsEdit true, got %v", draft.IsEdit)
	}
	if draft.OriginalMsgID != "msg456" {
		t.Errorf("Expected OriginalMsgID 'msg456', got %q", draft.OriginalMsgID)
	}
}

// Test draft store operations
func TestDraftStoreOperations(t *testing.T) {
	// Clear the draft store before test
	draftStore = make(map[string]*Draft)

	var userID = "user123"
	var guildID = "guild456"
	var draftKey = getDraftKey(userID, guildID)

	// Test storing a draft
	var draft = &Draft{
		UserID:    userID,
		GuildID:   guildID,
		ChannelID: "channel789",
		Content:   "Test content",
		CreatedAt: time.Now(),
	}
	draftStore[draftKey] = draft

	// Test retrieving the draft
	var retrievedDraft *Draft
	var exists bool
	retrievedDraft, exists = draftStore[draftKey]
	if !exists {
		t.Fatal("Expected draft to exist in store")
	}
	if retrievedDraft.Content != "Test content" {
		t.Errorf("Expected content 'Test content', got %q", retrievedDraft.Content)
	}

	// Test draft doesn't exist for different user
	var differentKey = getDraftKey("different_user", guildID)
	var differentDraft *Draft
	differentDraft, exists = draftStore[differentKey]
	if exists {
		t.Error("Expected draft to not exist for different user")
	}
	if differentDraft != nil {
		t.Error("Expected nil draft for non-existent key")
	}

	// Test draft doesn't exist for different guild
	var differentGuildKey = getDraftKey(userID, "different_guild")
	var differentGuildDraft *Draft
	differentGuildDraft, exists = draftStore[differentGuildKey]
	if exists {
		t.Error("Expected draft to not exist for different guild")
	}
	if differentGuildDraft != nil {
		t.Error("Expected nil draft for non-existent guild key")
	}

	// Test deleting draft
	delete(draftStore, draftKey)
	var deletedDraft *Draft
	deletedDraft, exists = draftStore[draftKey]
	if exists {
		t.Error("Expected draft to be deleted")
	}
	if deletedDraft != nil {
		t.Error("Expected nil after deletion")
	}
}

// Test multiple drafts for different user:guild combinations
func TestDraftStoreMultipleDrafts(t *testing.T) {
	// Clear the draft store before test
	draftStore = make(map[string]*Draft)

	var draft1 = &Draft{
		UserID:    "user1",
		GuildID:   "guild1",
		ChannelID: "channel1",
		Content:   "Draft 1 content",
		CreatedAt: time.Now(),
	}
	var draft2 = &Draft{
		UserID:    "user2",
		GuildID:   "guild1",
		ChannelID: "channel1",
		Content:   "Draft 2 content",
		CreatedAt: time.Now(),
	}
	var draft3 = &Draft{
		UserID:    "user1",
		GuildID:   "guild2",
		ChannelID: "channel2",
		Content:   "Draft 3 content",
		CreatedAt: time.Now(),
	}

	// Store all drafts
	draftStore[getDraftKey(draft1.UserID, draft1.GuildID)] = draft1
	draftStore[getDraftKey(draft2.UserID, draft2.GuildID)] = draft2
	draftStore[getDraftKey(draft3.UserID, draft3.GuildID)] = draft3

	// Verify all drafts exist
	if len(draftStore) != 3 {
		t.Errorf("Expected 3 drafts in store, got %d", len(draftStore))
	}

	// Verify each draft can be retrieved correctly
	var retrieved1 *Draft
	var retrieved2 *Draft
	var retrieved3 *Draft
	var exists bool

	retrieved1, exists = draftStore[getDraftKey("user1", "guild1")]
	if !exists || retrieved1.Content != "Draft 1 content" {
		t.Error("Failed to retrieve draft 1 correctly")
	}

	retrieved2, exists = draftStore[getDraftKey("user2", "guild1")]
	if !exists || retrieved2.Content != "Draft 2 content" {
		t.Error("Failed to retrieve draft 2 correctly")
	}

	retrieved3, exists = draftStore[getDraftKey("user1", "guild2")]
	if !exists || retrieved3.Content != "Draft 3 content" {
		t.Error("Failed to retrieve draft 3 correctly")
	}
}

// Test ComposeCreateDefinition structure
func TestComposeCreateDefinition(t *testing.T) {
	var def = ComposeCreateDefinition

	if def == nil {
		t.Fatal("ComposeCreateDefinition should not be nil")
	}

	if def.Name != string(ComposeCreate) {
		t.Errorf("Expected name %q, got %q", ComposeCreate, def.Name)
	}

	if def.Description != "Create a new proxied message draft" {
		t.Errorf("Expected description 'Create a new proxied message draft', got %q", def.Description)
	}

	if len(def.Options) != 2 {
		t.Fatalf("Expected 2 options, got %d", len(def.Options))
	}

	// Check content option
	var contentOption = def.Options[0]
	if contentOption.Name != "content" {
		t.Errorf("Expected first option name 'content', got %q", contentOption.Name)
	}
	if contentOption.Type != discordgo.ApplicationCommandOptionString {
		t.Errorf("Expected content option type String, got %v", contentOption.Type)
	}
	if !contentOption.Required {
		t.Error("Expected content option to be required")
	}
	if contentOption.MaxLength != 2000 {
		t.Errorf("Expected content option MaxLength 2000, got %d", contentOption.MaxLength)
	}

	// Check channel option
	var channelOption = def.Options[1]
	if channelOption.Name != "channel" {
		t.Errorf("Expected second option name 'channel', got %q", channelOption.Name)
	}
	if channelOption.Type != discordgo.ApplicationCommandOptionChannel {
		t.Errorf("Expected channel option type Channel, got %v", channelOption.Type)
	}
	if channelOption.Required {
		t.Error("Expected channel option to be optional")
	}
}

// Test ComposePostDefinition structure
func TestComposePostDefinition(t *testing.T) {
	var def = ComposePostDefinition

	if def == nil {
		t.Fatal("ComposePostDefinition should not be nil")
	}

	if def.Name != string(ComposePost) {
		t.Errorf("Expected name %q, got %q", ComposePost, def.Name)
	}

	if def.Description != "Post a message directly without preview" {
		t.Errorf("Expected description 'Post a message directly without preview', got %q", def.Description)
	}

	if len(def.Options) != 2 {
		t.Fatalf("Expected 2 options, got %d", len(def.Options))
	}
}

// Test ComposeProposeDefinition structure
func TestComposeProposeDefinition(t *testing.T) {
	var def = ComposeProposeDefinition

	if def == nil {
		t.Fatal("ComposeProposeDefinition should not be nil")
	}

	if def.Name != string(ComposePropose) {
		t.Errorf("Expected name %q, got %q", ComposePropose, def.Name)
	}

	if def.Description != "Propose an edit to an existing proxied message" {
		t.Errorf("Expected description 'Propose an edit to an existing proxied message', got %q", def.Description)
	}

	if len(def.Options) != 2 {
		t.Fatalf("Expected 2 options, got %d", len(def.Options))
	}

	// Check message option
	var messageOption = def.Options[0]
	if messageOption.Name != "message" {
		t.Errorf("Expected first option name 'message', got %q", messageOption.Name)
	}
	if messageOption.Type != discordgo.ApplicationCommandOptionString {
		t.Errorf("Expected message option type String, got %v", messageOption.Type)
	}
	if !messageOption.Required {
		t.Error("Expected message option to be required")
	}

	// Check content option
	var contentOption = def.Options[1]
	if contentOption.Name != "content" {
		t.Errorf("Expected second option name 'content', got %q", contentOption.Name)
	}
	if !contentOption.Required {
		t.Error("Expected content option to be required")
	}
}

// Test that button definitions are registered
func TestButtonDefinitionsRegistered(t *testing.T) {
	// Check that all expected buttons are registered
	var expectedButtons = []TButton{
		ButtonComposePreviewPost,
		ButtonComposePreviewCancel,
		ButtonEditPreviewApply,
		ButtonEditPreviewCancel,
	}

	for _, buttonID := range expectedButtons {
		var def, exists = ButtonDefinitions[buttonID]
		if !exists {
			t.Errorf("Button %q should be registered in ButtonDefinitions", buttonID)
			continue
		}
		if def.Execute == nil {
			t.Errorf("Button %q should have a non-nil Execute function", buttonID)
		}
	}
}

// Test that command definitions are registered
func TestCommandDefinitionsRegistered(t *testing.T) {
	// Check that all expected commands are registered
	var expectedCommands = []TSlashCommand{
		ComposeCreate,
		ComposePost,
		ComposePropose,
	}

	for _, cmdName := range expectedCommands {
		var def, exists = CommandDefinitions[cmdName]
		if !exists {
			t.Errorf("Command %q should be registered in CommandDefinitions", cmdName)
			continue
		}
		if def.Definition == nil {
			t.Errorf("Command %q should have a non-nil Definition", cmdName)
		}
		if def.Execute == nil {
			t.Errorf("Command %q should have a non-nil Execute function", cmdName)
		}
	}
}

// Test CommandDefinitions map structure
func TestCommandDefinitionsStructure(t *testing.T) {
	if CommandDefinitions == nil {
		t.Fatal("CommandDefinitions should not be nil")
	}

	// Verify each command has correct type
	for cmdName, def := range CommandDefinitions {
		if def.Definition == nil {
			t.Errorf("Command %q has nil Definition", cmdName)
		}
		if def.Execute == nil {
			t.Errorf("Command %q has nil Execute", cmdName)
		}
		// Autocomplete can be nil
	}
}

// Test ButtonDefinitions map structure
func TestButtonDefinitionsStructure(t *testing.T) {
	if ButtonDefinitions == nil {
		t.Fatal("ButtonDefinitions should not be nil")
	}

	// Verify each button has correct type
	for buttonID, def := range ButtonDefinitions {
		if def.Execute == nil {
			t.Errorf("Button %q has nil Execute function", buttonID)
		}
	}
}

// Test that Store variable can be set
func TestStoreVariable(t *testing.T) {
	var mockStore = NewMockStore()

	// Save original store
	var originalStore = Store

	// Set mock store
	Store = mockStore

	// Verify it was set
	if Store != mockStore {
		t.Error("Store should be settable to mock implementation")
	}

	// Restore original store
	Store = originalStore
}
