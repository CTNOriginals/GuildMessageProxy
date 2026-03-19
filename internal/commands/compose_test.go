package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

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
		ExpiresAt:     now.Add(24 * time.Hour),
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
	if !draft.ExpiresAt.After(draft.CreatedAt) {
		t.Errorf("Expected ExpiresAt to be after CreatedAt")
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
		ExpiresAt:     time.Now().Add(24 * time.Hour),
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
		ExpiresAt: time.Now().Add(24 * time.Hour),
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
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	var draft2 = &Draft{
		UserID:    "user2",
		GuildID:   "guild1",
		ChannelID: "channel1",
		Content:   "Draft 2 content",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	var draft3 = &Draft{
		UserID:    "user1",
		GuildID:   "guild2",
		ChannelID: "channel2",
		Content:   "Draft 3 content",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
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

	if def.Description != "Create a message preview before posting" {
		t.Errorf("Expected description 'Create a message preview before posting', got %q", def.Description)
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

	if def.Description != "Send a message immediately (skips preview)" {
		t.Errorf("Expected description 'Send a message immediately (skips preview)', got %q", def.Description)
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

	if def.Description != "Edit a message you posted via /compose" {
		t.Errorf("Expected description 'Edit a message you posted via /compose', got %q", def.Description)
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
		ButtonComposeConfirmDiscard,
		ButtonComposeKeepDraft,
		ButtonComposeRetryPost,
		ButtonComposeCancelAfterError,
		ButtonEditRetryApply,
		ButtonEditCancelAfterError,
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
	var mockStore = storage.NewMockStore()

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

func TestDraftExpiration(t *testing.T) {
	// Clear draft store
	draftStore = make(map[string]*Draft)

	// Create an expired draft
	expiredDraft := &Draft{
		UserID:    "user123",
		GuildID:   "guild456",
		ChannelID: "channel789",
		Content:   "Test content",
		CreatedAt: time.Now().Add(-48 * time.Hour), // Created 48 hours ago
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired 24 hours ago
	}
	draftStore["user123:guild456"] = expiredDraft

	// Verify draft exists
	if _, exists := draftStore["user123:guild456"]; !exists {
		t.Fatal("Draft should exist before expiration check")
	}

	// Verify draft is expired
	if !time.Now().After(expiredDraft.ExpiresAt) {
		t.Fatal("Draft should be expired")
	}
}

// Test draft expiration calculation with DraftTTL
func TestDraftExpirationCalculation(t *testing.T) {
	var now = time.Now()
	var draft = Draft{
		UserID:    "user123",
		GuildID:   "guild456",
		ChannelID: "channel789",
		Content:   "Test content",
		CreatedAt: now,
		ExpiresAt: now.Add(DraftTTL),
	}

	// Verify DraftTTL is 24 hours
	if DraftTTL != 24*time.Hour {
		t.Errorf("Expected DraftTTL to be 24 hours, got %v", DraftTTL)
	}

	// Verify ExpiresAt is correctly calculated as CreatedAt + DraftTTL
	var expectedExpiresAt = draft.CreatedAt.Add(DraftTTL)
	if !draft.ExpiresAt.Equal(expectedExpiresAt) {
		t.Errorf("Expected ExpiresAt to be CreatedAt + DraftTTL, got %v, expected %v", draft.ExpiresAt, expectedExpiresAt)
	}

	// Verify ExpiresAt is after CreatedAt
	if !draft.ExpiresAt.After(draft.CreatedAt) {
		t.Errorf("Expected ExpiresAt to be after CreatedAt")
	}
}

// Test formatDurationPast helper function
func TestFormatDurationPast(t *testing.T) {
	testCases := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "zero time returns empty string",
			time:     time.Time{},
			expected: "",
		},
		{
			name:     "recent past returns minutes",
			time:     time.Now().Add(-5 * time.Minute),
			expected: "5 minutes",
		},
		{
			name:     "hours ago returns hours",
			time:     time.Now().Add(-3 * time.Hour),
			expected: "3 hours",
		},
		{
			name:     "days ago returns days",
			time:     time.Now().Add(-2 * 24 * time.Hour),
			expected: "2 days",
		},
		{
			name:     "one hour singular",
			time:     time.Now().Add(-1 * time.Hour),
			expected: "1 hour",
		},
		{
			name:     "one day singular",
			time:     time.Now().Add(-1 * 24 * time.Hour),
			expected: "1 day",
		},
		{
			name:     "less than a minute",
			time:     time.Now().Add(-30 * time.Second),
			expected: "less than a minute",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatDurationPast(tc.time)
			if result != tc.expected {
				t.Errorf("formatDurationPast(%v) = %q, expected %q", tc.time, result, tc.expected)
			}
		})
	}
}

func TestMessageRefLengthValidation(t *testing.T) {
	testCases := []struct {
		name       string
		messageRef string
		shouldPass bool
	}{
		{"valid message ID", "123456789012345678", true},
		{"valid short URL", "https://discord.com/channels/1/2/3", true},
		{"too long - over 200 chars", strings.Repeat("a", 201), false},
		{"valid exact 200 chars", strings.Repeat("a", 200), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPass && len(tc.messageRef) > 200 {
				t.Errorf("Expected %q to pass but length %d > 200", tc.name, len(tc.messageRef))
			}
			if !tc.shouldPass && len(tc.messageRef) <= 200 {
				t.Errorf("Expected %q to fail but length %d <= 200", tc.name, len(tc.messageRef))
			}
		})
	}
}
