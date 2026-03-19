package storage

import (
	"testing"
	"time"
)

func TestNewMemoryStore(t *testing.T) {
	var store = NewMemoryStore()
	if store == nil {
		t.Fatal("NewMemoryStore returned nil")
	}
	if store.guilds == nil {
		t.Error("guilds map not initialized")
	}
	if store.guildConfigs == nil {
		t.Error("guildConfigs map not initialized")
	}
	if store.proxyMessages == nil {
		t.Error("proxyMessages map not initialized")
	}
}

func TestMemoryStore_SaveGuild(t *testing.T) {
	var tests = []struct {
		name      string
		guildID   string
		guildName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "success - save new guild",
			guildID:   "guild-123",
			guildName: "Test Guild",
			wantErr:   false,
		},
		{
			name:      "success - upsert existing guild",
			guildID:   "guild-123",
			guildName: "Updated Guild Name",
			wantErr:   false,
		},
		{
			name:      "error - empty guildID",
			guildID:   "",
			guildName: "Test Guild",
			wantErr:   true,
			errMsg:    "guildID cannot be empty",
		},
	}

	var store = NewMemoryStore()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err = store.SaveGuild(tt.guildID, tt.guildName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveGuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("SaveGuild() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestMemoryStore_GetGuild(t *testing.T) {
	var store = NewMemoryStore()

	// Setup: save a guild
	var err = store.SaveGuild("guild-123", "Test Guild")
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	var tests = []struct {
		name      string
		guildID   string
		wantFound bool
		wantName  string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "success - get existing guild",
			guildID:   "guild-123",
			wantFound: true,
			wantName:  "Test Guild",
			wantErr:   false,
		},
		{
			name:      "success - get non-existent guild returns nil",
			guildID:   "guild-999",
			wantFound: false,
			wantErr:   false,
		},
		{
			name:    "error - empty guildID",
			guildID: "",
			wantErr: true,
			errMsg:  "guildID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var guild, err = store.GetGuild(tt.guildID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("GetGuild() error message = %v, want %v", err.Error(), tt.errMsg)
				return
			}
			if tt.wantFound && guild == nil {
				t.Error("GetGuild() returned nil for existing guild")
				return
			}
			if tt.wantFound && guild.Name != tt.wantName {
				t.Errorf("GetGuild() returned guild with name = %v, want %v", guild.Name, tt.wantName)
			}
		})
	}
}

func TestMemoryStore_DeleteGuild(t *testing.T) {
	var store = NewMemoryStore()

	// Setup: save a guild and its config
	var err = store.SaveGuild("guild-123", "Test Guild")
	if err != nil {
		t.Fatalf("Failed to setup guild: %v", err)
	}
	var err2 = store.SaveGuildConfig(GuildConfig{GuildID: "guild-123", DefaultChannel: "channel-1"})
	if err2 != nil {
		t.Fatalf("Failed to setup guild config: %v", err2)
	}

	var tests = []struct {
		name    string
		guildID string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "success - delete existing guild",
			guildID: "guild-123",
			wantErr: false,
		},
		{
			name:    "success - delete non-existent guild no error",
			guildID: "guild-999",
			wantErr: false,
		},
		{
			name:    "error - empty guildID",
			guildID: "",
			wantErr: true,
			errMsg:  "guildID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err = store.DeleteGuild(tt.guildID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteGuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("DeleteGuild() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}

	// Verify guild and config were deleted
	var guild, _ = store.GetGuild("guild-123")
	if guild != nil {
		t.Error("DeleteGuild() did not remove guild data")
	}
	var config, _ = store.GetGuildConfig("guild-123")
	if config != nil {
		t.Error("DeleteGuild() did not remove guild config")
	}
}

func TestMemoryStore_GuildIsolation(t *testing.T) {
	var store = NewMemoryStore()

	// Save guilds with different IDs
	var err1 = store.SaveGuild("guild-1", "Guild One")
	var err2 = store.SaveGuild("guild-2", "Guild Two")
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to setup test data")
	}

	// Verify each guild can be retrieved independently
	var g1, _ = store.GetGuild("guild-1")
	var g2, _ = store.GetGuild("guild-2")

	if g1 == nil || g1.Name != "Guild One" {
		t.Error("Guild One data incorrect or missing")
	}
	if g2 == nil || g2.Name != "Guild Two" {
		t.Error("Guild Two data incorrect or missing")
	}

	// Delete one guild and verify the other remains
	var _ = store.DeleteGuild("guild-1")

	var g1After, _ = store.GetGuild("guild-1")
	var g2After, _ = store.GetGuild("guild-2")

	if g1After != nil {
		t.Error("Guild One should have been deleted")
	}
	if g2After == nil || g2After.Name != "Guild Two" {
		t.Error("Guild Two should still exist")
	}
}

func TestMemoryStore_SaveGuildConfig(t *testing.T) {
	var tests = []struct {
		name   string
		config GuildConfig
		wantErr bool
		errMsg string
	}{
		{
			name: "success - save new config",
			config: GuildConfig{
				GuildID:        "guild-123",
				AllowedRoles:   []string{"role-1", "role-2"},
				DefaultChannel: "channel-1",
				LogChannel:     "log-1",
			},
			wantErr: false,
		},
		{
			name: "success - upsert existing config",
			config: GuildConfig{
				GuildID:        "guild-123",
				AllowedRoles:   []string{"role-3"},
				DefaultChannel: "channel-2",
				LogChannel:     "log-2",
			},
			wantErr: false,
		},
		{
			name: "error - empty GuildID",
			config: GuildConfig{
				GuildID:        "",
				AllowedRoles:   []string{"role-1"},
				DefaultChannel: "channel-1",
			},
			wantErr: true,
			errMsg:  "GuildID cannot be empty",
		},
	}

	var store = NewMemoryStore()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err = store.SaveGuildConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveGuildConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("SaveGuildConfig() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestMemoryStore_GetGuildConfig(t *testing.T) {
	var store = NewMemoryStore()

	// Setup: save a config
	var setupConfig = GuildConfig{
		GuildID:        "guild-123",
		AllowedRoles:   []string{"role-1", "role-2"},
		DefaultChannel: "channel-1",
		LogChannel:     "log-1",
	}
	var err = store.SaveGuildConfig(setupConfig)
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	var tests = []struct {
		name      string
		guildID   string
		wantFound bool
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "success - get existing config",
			guildID:   "guild-123",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "success - get non-existent config returns nil",
			guildID:   "guild-999",
			wantFound: false,
			wantErr:   false,
		},
		{
			name:    "error - empty guildID",
			guildID: "",
			wantErr: true,
			errMsg:  "guildID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config, err = store.GetGuildConfig(tt.guildID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGuildConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("GetGuildConfig() error message = %v, want %v", err.Error(), tt.errMsg)
				return
			}
			if tt.wantFound && config == nil {
				t.Error("GetGuildConfig() returned nil for existing config")
				return
			}
			if tt.wantFound && config.GuildID != "guild-123" {
				t.Errorf("GetGuildConfig() returned config with wrong GuildID = %v", config.GuildID)
			}
		})
	}
}

func TestMemoryStore_SaveProxyMessage(t *testing.T) {
	var now = time.Now()
	var tests = []struct {
		name    string
		msg     ProxyMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "success - save new message",
			msg: ProxyMessage{
				GuildID:   "guild-123",
				ChannelID: "channel-1",
				MessageID: "msg-1",
				OwnerID:   "user-1",
				Content:   "Hello World",
				CreatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "success - upsert existing message",
			msg: ProxyMessage{
				GuildID:   "guild-123",
				ChannelID: "channel-1",
				MessageID: "msg-1",
				OwnerID:   "user-1",
				Content:   "Updated Content",
				CreatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "error - empty guildID",
			msg: ProxyMessage{
				GuildID:   "",
				ChannelID: "channel-1",
				MessageID: "msg-1",
				Content:   "Hello",
				CreatedAt: now,
			},
			wantErr: true,
			errMsg:  "guildID and messageID cannot be empty",
		},
		{
			name: "error - empty messageID",
			msg: ProxyMessage{
				GuildID:   "guild-123",
				ChannelID: "channel-1",
				MessageID: "",
				Content:   "Hello",
				CreatedAt: now,
			},
			wantErr: true,
			errMsg:  "guildID and messageID cannot be empty",
		},
		{
			name: "error - both empty",
			msg: ProxyMessage{
				GuildID:   "",
				ChannelID: "channel-1",
				MessageID: "",
				Content:   "Hello",
				CreatedAt: now,
			},
			wantErr: true,
			errMsg:  "guildID and messageID cannot be empty",
		},
	}

	var store = NewMemoryStore()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err = store.SaveProxyMessage(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveProxyMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("SaveProxyMessage() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestMemoryStore_GetProxyMessage(t *testing.T) {
	var store = NewMemoryStore()
	var now = time.Now()

	// Setup: save a message
	var setupMsg = ProxyMessage{
		GuildID:      "guild-123",
		ChannelID:    "channel-1",
		MessageID:    "msg-1",
		OwnerID:      "user-1",
		Content:      "Test Content",
		CreatedAt:    now,
		WebhookID:    "webhook-1",
		WebhookToken: "token-1",
	}
	var err = store.SaveProxyMessage(setupMsg)
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	var tests = []struct {
		name      string
		guildID   string
		messageID string
		wantFound bool
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "success - get existing message",
			guildID:   "guild-123",
			messageID: "msg-1",
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "success - get non-existent message returns nil",
			guildID:   "guild-123",
			messageID: "msg-999",
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "error - empty guildID",
			guildID:   "",
			messageID: "msg-1",
			wantErr:   true,
			errMsg:    "guildID and messageID cannot be empty",
		},
		{
			name:      "error - empty messageID",
			guildID:   "guild-123",
			messageID: "",
			wantErr:   true,
			errMsg:    "guildID and messageID cannot be empty",
		},
		{
			name:      "error - both empty",
			guildID:   "",
			messageID: "",
			wantErr:   true,
			errMsg:    "guildID and messageID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg, err = store.GetProxyMessage(tt.guildID, tt.messageID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProxyMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("GetProxyMessage() error message = %v, want %v", err.Error(), tt.errMsg)
				return
			}
			if tt.wantFound && msg == nil {
				t.Error("GetProxyMessage() returned nil for existing message")
				return
			}
			if tt.wantFound && msg.Content != "Test Content" {
				t.Errorf("GetProxyMessage() returned message with wrong Content = %v", msg.Content)
			}
		})
	}
}

func TestMemoryStore_UpdateProxyMessage(t *testing.T) {
	var store = NewMemoryStore()
	var now = time.Now()
	var editedTime = now.Add(time.Hour)

	// Setup: save a message
	var setupMsg = ProxyMessage{
		GuildID:   "guild-123",
		ChannelID: "channel-1",
		MessageID: "msg-1",
		OwnerID:   "user-1",
		Content:   "Original Content",
		CreatedAt: now,
	}
	var err = store.SaveProxyMessage(setupMsg)
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	var tests = []struct {
		name    string
		msg     ProxyMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "success - update existing message",
			msg: ProxyMessage{
				GuildID:      "guild-123",
				ChannelID:    "channel-1",
				MessageID:    "msg-1",
				OwnerID:      "user-1",
				Content:      "Updated Content",
				CreatedAt:    now,
				LastEditedAt: &editedTime,
				LastEditedBy: "user-2",
			},
			wantErr: false,
		},
		{
			name: "error - update non-existent message",
			msg: ProxyMessage{
				GuildID:   "guild-123",
				ChannelID: "channel-1",
				MessageID: "msg-999",
				OwnerID:   "user-1",
				Content:   "Content",
				CreatedAt: now,
			},
			wantErr: true,
			errMsg:  "proxy message not found: guild-123:msg-999",
		},
		{
			name: "error - empty guildID",
			msg: ProxyMessage{
				GuildID:   "",
				ChannelID: "channel-1",
				MessageID: "msg-1",
				Content:   "Content",
				CreatedAt: now,
			},
			wantErr: true,
			errMsg:  "guildID and messageID cannot be empty",
		},
		{
			name: "error - empty messageID",
			msg: ProxyMessage{
				GuildID:   "guild-123",
				ChannelID: "channel-1",
				MessageID: "",
				Content:   "Content",
				CreatedAt: now,
			},
			wantErr: true,
			errMsg:  "guildID and messageID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err = store.UpdateProxyMessage(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateProxyMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("UpdateProxyMessage() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}

	// Verify the update actually persisted
	var updated, _ = store.GetProxyMessage("guild-123", "msg-1")
	if updated == nil {
		t.Fatal("Updated message not found")
	}
	if updated.Content != "Updated Content" {
		t.Errorf("UpdateProxyMessage() did not persist Content, got = %v", updated.Content)
	}
	if updated.LastEditedBy != "user-2" {
		t.Errorf("UpdateProxyMessage() did not persist LastEditedBy, got = %v", updated.LastEditedBy)
	}
}

func TestMemoryStore_DeleteProxyMessage(t *testing.T) {
	var store = NewMemoryStore()
	var now = time.Now()

	// Setup: save a message
	var setupMsg = ProxyMessage{
		GuildID:   "guild-123",
		ChannelID: "channel-1",
		MessageID: "msg-1",
		OwnerID:   "user-1",
		Content:   "Test Content",
		CreatedAt: now,
	}
	var err = store.SaveProxyMessage(setupMsg)
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	var tests = []struct {
		name      string
		guildID   string
		messageID string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "success - delete existing message",
			guildID:   "guild-123",
			messageID: "msg-1",
			wantErr:   false,
		},
		{
			name:      "success - delete non-existent message no error",
			guildID:   "guild-123",
			messageID: "msg-999",
			wantErr:   false,
		},
		{
			name:      "error - empty guildID",
			guildID:   "",
			messageID: "msg-1",
			wantErr:   true,
			errMsg:    "guildID and messageID cannot be empty",
		},
		{
			name:      "error - empty messageID",
			guildID:   "guild-123",
			messageID: "",
			wantErr:   true,
			errMsg:    "guildID and messageID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err = store.DeleteProxyMessage(tt.guildID, tt.messageID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteProxyMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("DeleteProxyMessage() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}

	// Verify the message was deleted
	var msg, _ = store.GetProxyMessage("guild-123", "msg-1")
	if msg != nil {
		t.Error("DeleteProxyMessage() did not remove the message")
	}
}

func TestMemoryStore_ProxyMessageIsolation(t *testing.T) {
	var store = NewMemoryStore()
	var now = time.Now()

	// Save messages in different guilds with the same message ID
	var msg1 = ProxyMessage{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		MessageID: "msg-same",
		OwnerID:   "user-1",
		Content:   "Guild 1 Message",
		CreatedAt: now,
	}
	var msg2 = ProxyMessage{
		GuildID:   "guild-2",
		ChannelID: "channel-2",
		MessageID: "msg-same",
		OwnerID:   "user-2",
		Content:   "Guild 2 Message",
		CreatedAt: now,
	}

	var err1 = store.SaveProxyMessage(msg1)
	var err2 = store.SaveProxyMessage(msg2)
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to setup test data")
	}

	// Verify each message can be retrieved independently with correct key format
	var retrieved1, _ = store.GetProxyMessage("guild-1", "msg-same")
	var retrieved2, _ = store.GetProxyMessage("guild-2", "msg-same")

	if retrieved1 == nil || retrieved1.Content != "Guild 1 Message" {
		t.Error("Guild 1 message incorrect or missing")
	}
	if retrieved2 == nil || retrieved2.Content != "Guild 2 Message" {
		t.Error("Guild 2 message incorrect or missing")
	}

	// Delete one message and verify the other remains
	var _ = store.DeleteProxyMessage("guild-1", "msg-same")

	var after1, _ = store.GetProxyMessage("guild-1", "msg-same")
	var after2, _ = store.GetProxyMessage("guild-2", "msg-same")

	if after1 != nil {
		t.Error("Guild 1 message should have been deleted")
	}
	if after2 == nil || after2.Content != "Guild 2 Message" {
		t.Error("Guild 2 message should still exist")
	}
}

func TestMemoryStore_ProxyMessageKeyFormat(t *testing.T) {
	var store = NewMemoryStore()
	var now = time.Now()

	// Save a message
	var msg = ProxyMessage{
		GuildID:   "guild-123",
		ChannelID: "channel-1",
		MessageID: "msg-456",
		OwnerID:   "user-1",
		Content:   "Test",
		CreatedAt: now,
	}
	var err = store.SaveProxyMessage(msg)
	if err != nil {
		t.Fatalf("Failed to save message: %v", err)
	}

	// Verify the key is stored as "guildID:messageID"
	var key = "guild-123:msg-456"
	var stored, exists = store.proxyMessages[key]
	if !exists {
		t.Errorf("Message not stored with expected key format: %s", key)
	}
	if stored != nil && stored.Content != "Test" {
		t.Error("Stored message has wrong content")
	}

	// Verify we can retrieve using the separate ID parameters (not the key directly)
	var retrieved, _ = store.GetProxyMessage("guild-123", "msg-456")
	if retrieved == nil {
		t.Error("Could not retrieve message using separate IDs")
	}
}

func TestMemoryStore_CompleteCRUDWorkflow(t *testing.T) {
	var store = NewMemoryStore()
	var now = time.Now()

	// 1. Create
	var guildID = "guild-workflow"
	var messageID = "msg-workflow"

	var err = store.SaveGuild(guildID, "Workflow Test Guild")
	if err != nil {
		t.Fatalf("Create guild failed: %v", err)
	}

	var config = GuildConfig{
		GuildID:        guildID,
		AllowedRoles:   []string{"role-1"},
		DefaultChannel: "default",
		LogChannel:     "logs",
	}
	var err2 = store.SaveGuildConfig(config)
	if err2 != nil {
		t.Fatalf("Create config failed: %v", err2)
	}

	var msg = ProxyMessage{
		GuildID:      guildID,
		ChannelID:    "channel-1",
		MessageID:    messageID,
		OwnerID:      "user-1",
		Content:      "Original",
		CreatedAt:    now,
		WebhookID:    "hook-1",
		WebhookToken: "token-1",
	}
	var err3 = store.SaveProxyMessage(msg)
	if err3 != nil {
		t.Fatalf("Create message failed: %v", err3)
	}

	// 2. Read
	var g, _ = store.GetGuild(guildID)
	if g == nil || g.Name != "Workflow Test Guild" {
		t.Error("Read guild failed")
	}

	var c, _ = store.GetGuildConfig(guildID)
	if c == nil || c.DefaultChannel != "default" {
		t.Error("Read config failed")
	}

	var m, _ = store.GetProxyMessage(guildID, messageID)
	if m == nil || m.Content != "Original" {
		t.Error("Read message failed")
	}

	// 3. Update (upsert for guild/config, UpdateProxyMessage for messages)
	var err4 = store.SaveGuild(guildID, "Updated Guild Name")
	if err4 != nil {
		t.Fatalf("Update guild failed: %v", err4)
	}

	var updatedConfig = GuildConfig{
		GuildID:        guildID,
		AllowedRoles:   []string{"role-1", "role-2"},
		DefaultChannel: "new-default",
		LogChannel:     "new-logs",
	}
	var err5 = store.SaveGuildConfig(updatedConfig)
	if err5 != nil {
		t.Fatalf("Update config failed: %v", err5)
	}

	var editedTime = now.Add(time.Hour)
	var updatedMsg = ProxyMessage{
		GuildID:      guildID,
		ChannelID:    "channel-1",
		MessageID:    messageID,
		OwnerID:      "user-1",
		Content:      "Updated Content",
		CreatedAt:    now,
		LastEditedAt: &editedTime,
		LastEditedBy: "user-2",
		WebhookID:    "hook-1",
		WebhookToken: "token-1",
	}
	var err6 = store.UpdateProxyMessage(updatedMsg)
	if err6 != nil {
		t.Fatalf("Update message failed: %v", err6)
	}

	// Verify updates
	var g2, _ = store.GetGuild(guildID)
	if g2 == nil || g2.Name != "Updated Guild Name" {
		t.Error("Guild update not persisted")
	}

	var c2, _ = store.GetGuildConfig(guildID)
	if c2 == nil || c2.DefaultChannel != "new-default" {
		t.Error("Config update not persisted")
	}

	var m2, _ = store.GetProxyMessage(guildID, messageID)
	if m2 == nil || m2.Content != "Updated Content" {
		t.Error("Message update not persisted")
	}
	if m2.LastEditedBy != "user-2" {
		t.Error("Message LastEditedBy not persisted")
	}

	// 4. Delete
	var err7 = store.DeleteProxyMessage(guildID, messageID)
	if err7 != nil {
		t.Fatalf("Delete message failed: %v", err7)
	}

	var m3, _ = store.GetProxyMessage(guildID, messageID)
	if m3 != nil {
		t.Error("Message should be deleted")
	}

	// 5. Delete guild (also deletes config)
	var err8 = store.DeleteGuild(guildID)
	if err8 != nil {
		t.Fatalf("Delete guild failed: %v", err8)
	}

	var g3, _ = store.GetGuild(guildID)
	if g3 != nil {
		t.Error("Guild should be deleted")
	}

	var c3, _ = store.GetGuildConfig(guildID)
	if c3 != nil {
		t.Error("Config should be deleted with guild")
	}
}
