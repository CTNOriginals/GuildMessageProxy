package handlers

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// MockDiscordSession implements DiscordSession interface for testing
type MockDiscordSession struct {
	ChannelFunc                func(channelID string) (*discordgo.Channel, error)
	UserChannelPermissionsFunc func(userID, channelID string) (int64, error)
	ChannelWebhooksFunc        func(channelID string) ([]*discordgo.Webhook, error)
	WebhookCreateFunc          func(channelID, name, avatar string) (*discordgo.Webhook, error)
	WebhookExecuteFunc         func(webhookID, token string, wait bool, data *discordgo.WebhookParams) (*discordgo.Message, error)
	WebhookMessageEditFunc     func(webhookID, token, messageID string, data *discordgo.WebhookEdit) (*discordgo.Message, error)
	BotUserFunc                func() *discordgo.User
	ChannelTypingFunc          func(channelID string) error
}

func (m *MockDiscordSession) Channel(channelID string) (*discordgo.Channel, error) {
	if m.ChannelFunc != nil {
		return m.ChannelFunc(channelID)
	}
	return nil, errors.New("Channel not implemented")
}

func (m *MockDiscordSession) UserChannelPermissions(userID, channelID string) (int64, error) {
	if m.UserChannelPermissionsFunc != nil {
		return m.UserChannelPermissionsFunc(userID, channelID)
	}
	return 0, errors.New("UserChannelPermissions not implemented")
}

func (m *MockDiscordSession) ChannelWebhooks(channelID string) ([]*discordgo.Webhook, error) {
	if m.ChannelWebhooksFunc != nil {
		return m.ChannelWebhooksFunc(channelID)
	}
	return nil, errors.New("ChannelWebhooks not implemented")
}

func (m *MockDiscordSession) WebhookCreate(channelID, name, avatar string) (*discordgo.Webhook, error) {
	if m.WebhookCreateFunc != nil {
		return m.WebhookCreateFunc(channelID, name, avatar)
	}
	return nil, errors.New("WebhookCreate not implemented")
}

func (m *MockDiscordSession) WebhookExecute(webhookID, token string, wait bool, data *discordgo.WebhookParams) (*discordgo.Message, error) {
	if m.WebhookExecuteFunc != nil {
		return m.WebhookExecuteFunc(webhookID, token, wait, data)
	}
	return nil, errors.New("WebhookExecute not implemented")
}

func (m *MockDiscordSession) WebhookMessageEdit(webhookID, token, messageID string, data *discordgo.WebhookEdit) (*discordgo.Message, error) {
	if m.WebhookMessageEditFunc != nil {
		return m.WebhookMessageEditFunc(webhookID, token, messageID, data)
	}
	return nil, errors.New("WebhookMessageEdit not implemented")
}

func (m *MockDiscordSession) BotUser() *discordgo.User {
	if m.BotUserFunc != nil {
		return m.BotUserFunc()
	}
	return nil
}

func (m *MockDiscordSession) ChannelTyping(channelID string) error {
	if m.ChannelTypingFunc != nil {
		return m.ChannelTypingFunc(channelID)
	}
	return nil // Default to no error for tests that don't care about typing
}

// MockStore implements storage.Store interface for testing
type MockStore struct {
	guilds        map[string]*storage.Guild
	guildConfigs  map[string]*storage.GuildConfig
	proxyMessages map[string]*storage.ProxyMessage
	saveError     error
	getError      error
	updateError   error
	deleteError   error
}

func NewMockStore() *MockStore {
	return &MockStore{
		guilds:        make(map[string]*storage.Guild),
		guildConfigs:  make(map[string]*storage.GuildConfig),
		proxyMessages: make(map[string]*storage.ProxyMessage),
	}
}

func (m *MockStore) SaveGuild(guildID, name string) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.guilds[guildID] = &storage.Guild{ID: guildID, Name: name}
	return nil
}

func (m *MockStore) GetGuild(guildID string) (*storage.Guild, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.guilds[guildID], nil
}

func (m *MockStore) DeleteGuild(guildID string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.guilds, guildID)
	delete(m.guildConfigs, guildID)
	return nil
}

func (m *MockStore) SaveGuildConfig(config storage.GuildConfig) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.guildConfigs[config.GuildID] = &config
	return nil
}

func (m *MockStore) GetGuildConfig(guildID string) (*storage.GuildConfig, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.guildConfigs[guildID], nil
}

func (m *MockStore) SaveProxyMessage(msg storage.ProxyMessage) error {
	if m.saveError != nil {
		return m.saveError
	}
	var key = msg.GuildID + ":" + msg.MessageID
	m.proxyMessages[key] = &msg
	return nil
}

func (m *MockStore) GetProxyMessage(guildID, messageID string) (*storage.ProxyMessage, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	var key = guildID + ":" + messageID
	return m.proxyMessages[key], nil
}

func (m *MockStore) UpdateProxyMessage(msg storage.ProxyMessage) error {
	if m.updateError != nil {
		return m.updateError
	}
	var key = msg.GuildID + ":" + msg.MessageID
	if _, exists := m.proxyMessages[key]; !exists {
		return errors.New("proxy message not found")
	}
	m.proxyMessages[key] = &msg
	return nil
}

func (m *MockStore) DeleteProxyMessage(guildID, messageID string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	var key = guildID + ":" + messageID
	delete(m.proxyMessages, key)
	return nil
}

// ==================== VALIDATION TESTS ====================

func TestValidateContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantValid bool
		wantError string
	}{
		{
			name:      "valid content",
			content:   "Hello, World!",
			wantValid: true,
			wantError: "",
		},
		{
			name:      "empty content",
			content:   "",
			wantValid: false,
			wantError: "Message content cannot be empty.",
		},
		{
			name:      "content at max length",
			content:   strings.Repeat("a", MaxMessageLength),
			wantValid: true,
			wantError: "",
		},
		{
			name:      "content exceeds max length",
			content:   strings.Repeat("a", MaxMessageLength+1),
			wantValid: false,
			wantError: "Message exceeds maximum length of 2000 characters (current: 2001).",
		},
		{
			name:      "whitespace only content is valid",
			content:   "   ",
			wantValid: true,
			wantError: "",
		},
		{
			name:      "single character",
			content:   "x",
			wantValid: true,
			wantError: "",
		},
		{
			name:      "unicode content",
			content:   "Hello 世界! 🌍",
			wantValid: true,
			wantError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateContent(tt.content)
			if result.Valid != tt.wantValid {
				t.Errorf("ValidateContent() Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if result.Error != tt.wantError {
				t.Errorf("ValidateContent() Error = %q, want %q", result.Error, tt.wantError)
			}
		})
	}
}

func TestValidateEditPermission(t *testing.T) {
	var testTime time.Time = time.Now()
	var validProxyMsg *storage.ProxyMessage = &storage.ProxyMessage{
		GuildID:      "guild123",
		ChannelID:    "channel123",
		MessageID:    "msg123",
		OwnerID:      "user123",
		Content:      "Test content",
		CreatedAt:    testTime,
		LastEditedAt: nil,
		LastEditedBy: "",
		WebhookID:    "webhook123",
		WebhookToken: "token123",
	}

	tests := []struct {
		name      string
		proxyMsg  *storage.ProxyMessage
		userID    string
		wantValid bool
		wantError string
	}{
		{
			name:      "owner can edit",
			proxyMsg:  validProxyMsg,
			userID:    "user123",
			wantValid: true,
			wantError: "",
		},
		{
			name:      "other user cannot edit",
			proxyMsg:  validProxyMsg,
			userID:    "user456",
			wantValid: false,
			wantError: "Only the original message author can edit this message.",
		},
		{
			name:      "nil proxy message",
			proxyMsg:  nil,
			userID:    "user123",
			wantValid: false,
			wantError: "Message not found or no longer exists.",
		},
		{
			name: "empty owner id in proxy message",
			proxyMsg: &storage.ProxyMessage{
				GuildID:   "guild123",
				MessageID: "msg123",
				OwnerID:   "",
				Content:   "Test",
			},
			userID:    "user123",
			wantValid: false,
			wantError: "Only the original message author can edit this message.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEditPermission(tt.proxyMsg, tt.userID)
			if result.Valid != tt.wantValid {
				t.Errorf("ValidateEditPermission() Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if result.Error != tt.wantError {
				t.Errorf("ValidateEditPermission() Error = %q, want %q", result.Error, tt.wantError)
			}
		})
	}
}

// ==================== PERMISSIONS TESTS ====================

func TestCanUseCompose(t *testing.T) {
	tests := []struct {
		name       string
		channelErr error
		perms      int64
		permsErr   error
		wantAllowed bool
		wantError  string
	}{
		{
			name:       "user has send messages permission",
			channelErr: nil,
			perms:      discordgo.PermissionSendMessages,
			permsErr:   nil,
			wantAllowed: true,
			wantError:  "",
		},
		{
			name:       "user has multiple permissions including send",
			channelErr: nil,
			perms:      discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles,
			permsErr:   nil,
			wantAllowed: true,
			wantError:  "",
		},
		{
			name:       "channel not found",
			channelErr: errors.New("channel not found"),
			perms:      0,
			permsErr:   nil,
			wantAllowed: false,
			wantError:  "Unable to verify channel permissions.",
		},
		{
			name:       "permission check error",
			channelErr: nil,
			perms:      0,
			permsErr:   errors.New("permission check failed"),
			wantAllowed: false,
			wantError:  "Unable to verify user permissions.",
		},
		{
			name:       "user lacks send messages permission",
			channelErr: nil,
			perms:      discordgo.PermissionViewChannel, // only view, no send
			permsErr:   nil,
			wantAllowed: false,
			wantError:  "You need permission to send messages in this channel.",
		},
		{
			name:       "user has zero permissions",
			channelErr: nil,
			perms:      0,
			permsErr:   nil,
			wantAllowed: false,
			wantError:  "You need permission to send messages in this channel.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockStore *MockStore = NewMockStore()
			var mockSession *MockDiscordSession = &MockDiscordSession{
				ChannelFunc: func(channelID string) (*discordgo.Channel, error) {
					if tt.channelErr != nil {
						return nil, tt.channelErr
					}
					return &discordgo.Channel{ID: "channel123", Name: "test-channel"}, nil
				},
				UserChannelPermissionsFunc: func(userID, channelID string) (int64, error) {
					return tt.perms, tt.permsErr
				},
			}

			result := CanUseCompose(mockSession, "guild123", "channel123", "user123", mockStore, []string{"role123"})
			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanUseCompose() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if result.Error != tt.wantError {
				t.Errorf("CanUseCompose() Error = %q, want %q", result.Error, tt.wantError)
			}
		})
	}
}

func TestIsMessageOwner(t *testing.T) {
	tests := []struct {
		name     string
		proxyMsg *storage.ProxyMessage
		userID   string
		want     bool
	}{
		{
			name: "user is owner",
			proxyMsg: &storage.ProxyMessage{
				OwnerID: "user123",
			},
			userID: "user123",
			want:   true,
		},
		{
			name: "user is not owner",
			proxyMsg: &storage.ProxyMessage{
				OwnerID: "user123",
			},
			userID: "user456",
			want:   false,
		},
		{
			name:     "nil proxy message",
			proxyMsg: nil,
			userID:   "user123",
			want:     false,
		},
		{
			name: "empty owner id",
			proxyMsg: &storage.ProxyMessage{
				OwnerID: "",
			},
			userID: "user123",
			want:   false,
		},
		{
			name: "empty user id",
			proxyMsg: &storage.ProxyMessage{
				OwnerID: "user123",
			},
			userID: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMessageOwner(tt.proxyMsg, tt.userID)
			if got != tt.want {
				t.Errorf("IsMessageOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==================== POST TESTS ====================

func TestPostProxiedMessage(t *testing.T) {
	tests := []struct {
		name           string
		webhookList    []*discordgo.Webhook
		webhookListErr error
		webhookCreate  *discordgo.Webhook
		webhookCreateErr error
		webhookExecute   *discordgo.Message
		webhookExecuteErr error
		storeSaveErr     error
		wantSuccess      bool
		wantError        string
	}{
		{
			name:           "successful post with new webhook",
			webhookList:    []*discordgo.Webhook{}, // no existing webhooks
			webhookListErr: nil,
			webhookCreate: &discordgo.Webhook{
				ID:       "webhook123",
				Token:    "token123",
				ChannelID: "channel123",
			},
			webhookCreateErr: nil,
			webhookExecute: &discordgo.Message{
				ID: "msg123",
			},
			webhookExecuteErr: nil,
			storeSaveErr:      nil,
			wantSuccess:       true,
			wantError:         "",
		},
		{
			name: "successful post with existing webhook",
			webhookList: []*discordgo.Webhook{
				{
					ID:       "existing123",
					Token:    "existingToken",
					ChannelID: "channel123",
					User: &discordgo.User{
						ID: "bot123", // matches bot user ID
					},
				},
			},
			webhookListErr: nil,
			webhookCreate:  nil, // won't be called
			webhookCreateErr: nil,
			webhookExecute: &discordgo.Message{
				ID: "msg456",
			},
			webhookExecuteErr: nil,
			storeSaveErr:      nil,
			wantSuccess:       true,
			wantError:         "",
		},
		{
			name:           "webhook list fails",
			webhookList:    nil,
			webhookListErr: errors.New("failed to list webhooks"),
			webhookCreate:  nil,
			webhookCreateErr: nil,
			webhookExecute:   nil,
			webhookExecuteErr: nil,
			storeSaveErr:     nil,
			wantSuccess:      false,
			wantError:        "Failed to create webhook. Ensure the bot has Manage Webhooks permission in this channel.",
		},
		{
			name:           "webhook creation fails",
			webhookList:    []*discordgo.Webhook{},
			webhookListErr: nil,
			webhookCreate:  nil,
			webhookCreateErr: errors.New("insufficient permissions"),
			webhookExecute:   nil,
			webhookExecuteErr: nil,
			storeSaveErr:     nil,
			wantSuccess:      false,
			wantError:        "Failed to create webhook. Ensure the bot has Manage Webhooks permission in this channel.",
		},
		{
			name:           "webhook execution fails",
			webhookList:    []*discordgo.Webhook{},
			webhookListErr: nil,
			webhookCreate: &discordgo.Webhook{
				ID:    "webhook123",
				Token: "token123",
			},
			webhookCreateErr: nil,
			webhookExecute:   nil,
			webhookExecuteErr: errors.New("webhook execution failed"),
			storeSaveErr:     nil,
			wantSuccess:      false,
			wantError:        "Failed to post message. The webhook may have been deleted. Try again or contact an admin.",
		},
		{
			name:           "store save fails but post succeeds",
			webhookList:    []*discordgo.Webhook{},
			webhookListErr: nil,
			webhookCreate: &discordgo.Webhook{
				ID:    "webhook123",
				Token: "token123",
			},
			webhookCreateErr: nil,
			webhookExecute: &discordgo.Message{
				ID: "msg123",
			},
			webhookExecuteErr: nil,
			storeSaveErr:      errors.New("storage error"),
			wantSuccess:       true, // message was still posted
			wantError:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockStore *MockStore = NewMockStore()
			mockStore.saveError = tt.storeSaveErr

			var mockSession *MockDiscordSession = &MockDiscordSession{
				BotUserFunc: func() *discordgo.User {
					return &discordgo.User{ID: "bot123"}
				},
				ChannelWebhooksFunc: func(channelID string) ([]*discordgo.Webhook, error) {
					return tt.webhookList, tt.webhookListErr
				},
				WebhookCreateFunc: func(channelID, name, avatar string) (*discordgo.Webhook, error) {
					return tt.webhookCreate, tt.webhookCreateErr
				},
				WebhookExecuteFunc: func(webhookID, token string, wait bool, data *discordgo.WebhookParams) (*discordgo.Message, error) {
					return tt.webhookExecute, tt.webhookExecuteErr
				},
			}

			result := PostProxiedMessage(mockSession, "guild123", "channel123", "Test content", "user123", mockStore)
			if result.Success != tt.wantSuccess {
				t.Errorf("PostProxiedMessage() Success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if result.Error != tt.wantError {
				t.Errorf("PostProxiedMessage() Error = %q, want %q", result.Error, tt.wantError)
			}
			if tt.wantSuccess && result.MessageID == "" {
				t.Error("PostProxiedMessage() MessageID should not be empty on success")
			}
		})
	}
}

func TestGetOrCreateWebhook(t *testing.T) {
	tests := []struct {
		name           string
		webhookList    []*discordgo.Webhook
		webhookListErr error
		webhookCreate  *discordgo.Webhook
		webhookCreateErr error
		wantErr        bool
		wantWebhookID  string
	}{
		{
			name:           "returns existing bot webhook",
			webhookList: []*discordgo.Webhook{
				{
					ID:    "webhook1",
					Token: "token1",
					User:  &discordgo.User{ID: "bot123"},
				},
			},
			webhookListErr: nil,
			webhookCreate:  nil,
			webhookCreateErr: nil,
			wantErr:        false,
			wantWebhookID:  "webhook1",
		},
		{
			name: "skips webhooks created by other users",
			webhookList: []*discordgo.Webhook{
				{
					ID:    "otherwebhook",
					Token: "othertoken",
					User:  &discordgo.User{ID: "otheruser"},
				},
			},
			webhookListErr: nil,
			webhookCreate: &discordgo.Webhook{
				ID:    "newwebhook",
				Token: "newtoken",
			},
			webhookCreateErr: nil,
			wantErr:        false,
			wantWebhookID:  "newwebhook",
		},
		{
			name:           "creates new webhook when none exist",
			webhookList:    []*discordgo.Webhook{},
			webhookListErr: nil,
			webhookCreate: &discordgo.Webhook{
				ID:    "created123",
				Token: "createdToken",
			},
			webhookCreateErr: nil,
			wantErr:        false,
			wantWebhookID:  "created123",
		},
		{
			name:           "error listing webhooks",
			webhookList:    nil,
			webhookListErr: errors.New("permission denied"),
			webhookCreate:  nil,
			webhookCreateErr: nil,
			wantErr:        true,
			wantWebhookID:  "",
		},
		{
			name:           "error creating webhook",
			webhookList:    []*discordgo.Webhook{},
			webhookListErr: nil,
			webhookCreate:  nil,
			webhookCreateErr: errors.New("cannot create webhook"),
			wantErr:        true,
			wantWebhookID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockSession *MockDiscordSession = &MockDiscordSession{
				BotUserFunc: func() *discordgo.User {
					return &discordgo.User{ID: "bot123"}
				},
				ChannelWebhooksFunc: func(channelID string) ([]*discordgo.Webhook, error) {
					return tt.webhookList, tt.webhookListErr
				},
				WebhookCreateFunc: func(channelID, name, avatar string) (*discordgo.Webhook, error) {
					return tt.webhookCreate, tt.webhookCreateErr
				},
			}

			webhook, err := getOrCreateWebhook(mockSession, "channel123")
			if (err != nil) != tt.wantErr {
				t.Errorf("getOrCreateWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && webhook.ID != tt.wantWebhookID {
				t.Errorf("getOrCreateWebhook() webhook.ID = %q, want %q", webhook.ID, tt.wantWebhookID)
			}
		})
	}
}

func TestFormatProxiedContent(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		requesterName string
		want          string
	}{
		{
			name:          "simple content",
			content:       "Hello World",
			requesterName: "TestUser",
			want:          "_Requested by TestUser_\n\nHello World",
		},
		{
			name:          "empty content",
			content:       "",
			requesterName: "TestUser",
			want:          "_Requested by TestUser_\n\n",
		},
		{
			name:          "multi-line content",
			content:       "Line 1\nLine 2\nLine 3",
			requesterName: "AnotherUser",
			want:          "_Requested by AnotherUser_\n\nLine 1\nLine 2\nLine 3",
		},
		{
			name:          "content with special characters",
			content:       "Hello *world* with **markdown**",
			requesterName: "MarkdownUser",
			want:          "_Requested by MarkdownUser_\n\nHello *world* with **markdown**",
		},
		{
			name:          "unicode content",
			content:       "Hello 世界! 🌍",
			requesterName: "UnicodeUser",
			want:          "_Requested by UnicodeUser_\n\nHello 世界! 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatProxiedContent(tt.content, tt.requesterName)
			if got != tt.want {
				t.Errorf("FormatProxiedContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractMessageIDFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "valid discord message URL",
			url:  "https://discord.com/channels/123/456/789",
			want: "789",
		},
		{
			name: "discord.com URL with www",
			url:  "https://www.discord.com/channels/123/456/789",
			want: "789",
		},
		{
			name: "discordapp.com URL",
			url:  "https://discordapp.com/channels/123/456/789",
			want: "789",
		},
		{
			name: "empty URL",
			url:  "",
			want: "",
		},
		{
			name: "invalid URL - too few parts",
			url:  "https://discord.com/channels/123",
			want: "",
		},
		{
			name: "invalid URL - not discord",
			url:  "https://example.com/path/to/resource",
			want: "resource",
		},
		{
			name: "URL with extra path segments",
			url:  "https://discord.com/channels/123/456/789/extra",
			want: "extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractMessageIDFromURL(tt.url)
			if got != tt.want {
				t.Errorf("ExtractMessageIDFromURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractIDsFromURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantGuild string
		wantChan  string
		wantMsg   string
	}{
		{
			name:      "valid discord message URL",
			url:       "https://discord.com/channels/123/456/789",
			wantGuild: "123",
			wantChan:  "456",
			wantMsg:   "789",
		},
		{
			name:      "empty URL",
			url:       "",
			wantGuild: "",
			wantChan:  "",
			wantMsg:   "",
		},
		{
			name:      "invalid URL - too few parts",
			url:       "https://discord.com/channels/123",
			wantGuild: "",
			wantChan:  "",
			wantMsg:   "",
		},
		{
			name:      "URL with extra segments",
			url:       "https://discord.com/channels/123/456/789/extra",
			wantGuild: "456",
			wantChan:  "789",
			wantMsg:   "extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGuild, gotChan, gotMsg := ExtractIDsFromURL(tt.url)
			if gotGuild != tt.wantGuild {
				t.Errorf("ExtractIDsFromURL() guild = %q, want %q", gotGuild, tt.wantGuild)
			}
			if gotChan != tt.wantChan {
				t.Errorf("ExtractIDsFromURL() channel = %q, want %q", gotChan, tt.wantChan)
			}
			if gotMsg != tt.wantMsg {
				t.Errorf("ExtractIDsFromURL() message = %q, want %q", gotMsg, tt.wantMsg)
			}
		})
	}
}

// ==================== EDIT TESTS ====================

func TestEditProxiedMessage(t *testing.T) {
	var now time.Time = time.Now()
	var validProxyMsg storage.ProxyMessage = storage.ProxyMessage{
		GuildID:      "guild123",
		ChannelID:    "channel123",
		MessageID:    "msg123",
		OwnerID:      "user123",
		Content:      "Original content",
		CreatedAt:    now,
		LastEditedAt: nil,
		LastEditedBy: "",
		WebhookID:    "webhook123",
		WebhookToken: "token123",
	}

	tests := []struct {
		name             string
		proxyMsg         storage.ProxyMessage
		newContent       string
		editedBy         string
		webhookEditResult *discordgo.Message
		webhookEditErr   error
		storeUpdateErr   error
		wantSuccess      bool
		wantError        string
	}{
		{
			name:             "successful edit",
			proxyMsg:         validProxyMsg,
			newContent:       "Updated content",
			editedBy:         "user123",
			webhookEditResult: &discordgo.Message{ID: "msg123"},
			webhookEditErr:   nil,
			storeUpdateErr:   nil,
			wantSuccess:      true,
			wantError:        "",
		},
		{
			name: "missing webhook ID",
			proxyMsg: storage.ProxyMessage{
				GuildID:      "guild123",
				MessageID:    "msg123",
				WebhookID:    "",
				WebhookToken: "token123",
			},
			newContent:       "Updated content",
			editedBy:         "user123",
			webhookEditResult: nil,
			webhookEditErr:   nil,
			storeUpdateErr:   nil,
			wantSuccess:      false,
			wantError:        "Cannot edit: webhook credentials not found for this message.",
		},
		{
			name: "missing webhook token",
			proxyMsg: storage.ProxyMessage{
				GuildID:      "guild123",
				MessageID:    "msg123",
				WebhookID:    "webhook123",
				WebhookToken: "",
			},
			newContent:       "Updated content",
			editedBy:         "user123",
			webhookEditResult: nil,
			webhookEditErr:   nil,
			storeUpdateErr:   nil,
			wantSuccess:      false,
			wantError:        "Cannot edit: webhook credentials not found for this message.",
		},
		{
			name:             "webhook edit fails",
			proxyMsg:         validProxyMsg,
			newContent:       "Updated content",
			editedBy:         "user123",
			webhookEditResult: nil,
			webhookEditErr:   errors.New("message not found or no longer editable"),
			storeUpdateErr:   nil,
			wantSuccess:      false,
			wantError:        "Failed to edit message. It may have been deleted or is no longer editable.",
		},
		{
			name:             "store update fails but edit succeeds",
			proxyMsg:         validProxyMsg,
			newContent:       "Updated content",
			editedBy:         "user123",
			webhookEditResult: &discordgo.Message{ID: "msg123"},
			webhookEditErr:   nil,
			storeUpdateErr:   errors.New("storage error"),
			wantSuccess:      true, // edit succeeded even if storage failed
			wantError:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockStore *MockStore = NewMockStore()
			mockStore.updateError = tt.storeUpdateErr

			// Pre-populate store with the proxy message for update test
			if tt.proxyMsg.GuildID != "" && tt.proxyMsg.MessageID != "" {
				var key string = tt.proxyMsg.GuildID + ":" + tt.proxyMsg.MessageID
				var msgCopy storage.ProxyMessage = tt.proxyMsg
				mockStore.proxyMessages[key] = &msgCopy
			}

			var mockSession *MockDiscordSession = &MockDiscordSession{
				WebhookMessageEditFunc: func(webhookID, token, messageID string, data *discordgo.WebhookEdit) (*discordgo.Message, error) {
					return tt.webhookEditResult, tt.webhookEditErr
				},
			}

			result := EditProxiedMessage(mockSession, &tt.proxyMsg, tt.newContent, tt.editedBy, mockStore)
			if result.Success != tt.wantSuccess {
				t.Errorf("EditProxiedMessage() Success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if result.Error != tt.wantError {
				t.Errorf("EditProxiedMessage() Error = %q, want %q", result.Error, tt.wantError)
			}
		})
	}
}

func TestGetProxiedMessage(t *testing.T) {
	var mockStore *MockStore = NewMockStore()
	var testMsg *storage.ProxyMessage = &storage.ProxyMessage{
		GuildID:   "guild123",
		MessageID: "msg123",
		Content:   "Test content",
	}

	// Pre-populate store
	mockStore.proxyMessages["guild123:msg123"] = testMsg

	tests := []struct {
		name      string
		guildID   string
		messageID string
		wantMsg   bool // whether we expect a non-nil message
		wantErr   bool
	}{
		{
			name:      "existing message",
			guildID:   "guild123",
			messageID: "msg123",
			wantMsg:   true,
			wantErr:   false,
		},
		{
			name:      "non-existing message",
			guildID:   "guild123",
			messageID: "nonexistent",
			wantMsg:   false,
			wantErr:   false, // store returns nil, not error
		},
		{
			name:      "empty guild id",
			guildID:   "",
			messageID: "msg123",
			wantMsg:   false,
			wantErr:   false,
		},
		{
			name:      "empty message id",
			guildID:   "guild123",
			messageID: "",
			wantMsg:   false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := GetProxiedMessage(mockStore, tt.guildID, tt.messageID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProxiedMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantMsg && msg == nil {
				t.Error("GetProxiedMessage() expected non-nil message, got nil")
			}
			if !tt.wantMsg && msg != nil {
				t.Error("GetProxiedMessage() expected nil message, got non-nil")
			}
		})
	}
}

// ==================== PREVIEW TESTS ====================

func TestRenderPreviewResponse(t *testing.T) {
	tests := []struct {
		name             string
		data             PreviewData
		wantType         discordgo.InteractionResponseType
		wantEphemeral    bool
		wantConfirmLabel string
		wantTitle        string
		wantColor        int
		wantContent      string
		wantChannel      string
	}{
		{
			name: "compose preview",
			data: PreviewData{
				Content:         "Hello World",
				TargetChannel:   "channel123",
				IsEdit:          false,
				ConfirmButtonID: "confirm_123",
				CancelButtonID:  "cancel_123",
			},
			wantType:         discordgo.InteractionResponseChannelMessageWithSource,
			wantEphemeral:    true,
			wantConfirmLabel: "Post",
			wantTitle:        "Compose Preview",
			wantColor:        0x3498db,
			wantContent:      "```\nHello World\n```",
			wantChannel:      "<#channel123>",
		},
		{
			name: "edit preview",
			data: PreviewData{
				Content:         "Updated content",
				TargetChannel:   "channel456",
				IsEdit:          true,
				OriginalMsgID:   "original789",
				ConfirmButtonID: "apply_456",
				CancelButtonID:  "cancel_456",
			},
			wantType:         discordgo.InteractionResponseChannelMessageWithSource,
			wantEphemeral:    true,
			wantConfirmLabel: "Apply",
			wantTitle:        "Edit Preview",
			wantColor:        0xe67e22,
			wantContent:      "```\nUpdated content\n```",
			wantChannel:      "<#channel456>",
		},
		{
			name: "preview with multiline content",
			data: PreviewData{
				Content:         "Line 1\nLine 2\nLine 3",
				TargetChannel:   "channel789",
				IsEdit:          false,
				ConfirmButtonID: "confirm_789",
				CancelButtonID:  "cancel_789",
			},
			wantType:         discordgo.InteractionResponseChannelMessageWithSource,
			wantEphemeral:    true,
			wantConfirmLabel: "Post",
			wantTitle:        "Compose Preview",
			wantColor:        0x3498db,
			wantContent:      "```\nLine 1\nLine 2\nLine 3\n```",
			wantChannel:      "<#channel789>",
		},
		{
			name: "empty content preview",
			data: PreviewData{
				Content:         "",
				TargetChannel:   "channel000",
				IsEdit:          false,
				ConfirmButtonID: "confirm_000",
				CancelButtonID:  "cancel_000",
			},
			wantType:         discordgo.InteractionResponseChannelMessageWithSource,
			wantEphemeral:    true,
			wantConfirmLabel: "Post",
			wantTitle:        "Compose Preview",
			wantColor:        0x3498db,
			wantContent:      "```\n\n```",
			wantChannel:      "<#channel000>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPreviewResponse(tt.data)

			if result.Type != tt.wantType {
				t.Errorf("RenderPreviewResponse() Type = %v, want %v", result.Type, tt.wantType)
			}

			if result.Data == nil {
				t.Fatal("RenderPreviewResponse() Data is nil")
			}

			if result.Data.Flags != discordgo.MessageFlagsEphemeral {
				t.Errorf("RenderPreviewResponse() Flags = %v, want Ephemeral", result.Data.Flags)
			}

			// Check embed
			if len(result.Data.Embeds) == 0 {
				t.Fatal("RenderPreviewResponse() Embeds is empty")
			}

			embed := result.Data.Embeds[0]

			if embed.Title != tt.wantTitle {
				t.Errorf("RenderPreviewResponse() Embed Title = %q, want %q", embed.Title, tt.wantTitle)
			}

			if embed.Color != tt.wantColor {
				t.Errorf("RenderPreviewResponse() Embed Color = %d, want %d", embed.Color, tt.wantColor)
			}

			if embed.Description != tt.wantContent {
				t.Errorf("RenderPreviewResponse() Embed Description = %q, want %q", embed.Description, tt.wantContent)
			}

			// Check target channel field
			if len(embed.Fields) == 0 {
				t.Fatal("RenderPreviewResponse() Embed Fields is empty")
			}

			if embed.Fields[0].Value != tt.wantChannel {
				t.Errorf("RenderPreviewResponse() Embed Fields[0].Value = %q, want %q", embed.Fields[0].Value, tt.wantChannel)
			}

			// Check footer text
			if embed.Footer == nil {
				t.Fatal("RenderPreviewResponse() Embed Footer is nil")
			}

			var expectedFooter string
			if tt.data.IsEdit {
				expectedFooter = "Click Apply to confirm the edit, or Cancel to discard."
			} else {
				expectedFooter = "Click Post to send the message, or Cancel to discard."
			}
			if embed.Footer.Text != expectedFooter {
				t.Errorf("RenderPreviewResponse() Embed Footer.Text = %q, want %q", embed.Footer.Text, expectedFooter)
			}

			// Check components
			if len(result.Data.Components) == 0 {
				t.Fatal("RenderPreviewResponse() Components is empty")
			}

			actionsRow, ok := result.Data.Components[0].(discordgo.ActionsRow)
			if !ok {
				t.Fatal("RenderPreviewResponse() First component is not ActionsRow")
			}

			if len(actionsRow.Components) != 2 {
				t.Errorf("RenderPreviewResponse() Expected 2 buttons, got %d", len(actionsRow.Components))
			}

			// Check confirm button
			confirmBtn, ok := actionsRow.Components[0].(discordgo.Button)
			if !ok {
				t.Fatal("RenderPreviewResponse() First button is not Button")
			}
			if confirmBtn.Label != tt.wantConfirmLabel {
				t.Errorf("RenderPreviewResponse() Confirm button label = %q, want %q", confirmBtn.Label, tt.wantConfirmLabel)
			}
			if confirmBtn.Style != discordgo.PrimaryButton {
				t.Errorf("RenderPreviewResponse() Confirm button style = %v, want PrimaryButton", confirmBtn.Style)
			}
			if confirmBtn.CustomID != tt.data.ConfirmButtonID {
				t.Errorf("RenderPreviewResponse() Confirm button CustomID = %q, want %q", confirmBtn.CustomID, tt.data.ConfirmButtonID)
			}

			// Check cancel button
			cancelBtn, ok := actionsRow.Components[1].(discordgo.Button)
			if !ok {
				t.Fatal("RenderPreviewResponse() Second button is not Button")
			}
			if cancelBtn.Label != "Cancel" {
				t.Errorf("RenderPreviewResponse() Cancel button label = %q, want Cancel", cancelBtn.Label)
			}
			if cancelBtn.Style != discordgo.SecondaryButton {
				t.Errorf("RenderPreviewResponse() Cancel button style = %v, want SecondaryButton", cancelBtn.Style)
			}
			if cancelBtn.CustomID != tt.data.CancelButtonID {
				t.Errorf("RenderPreviewResponse() Cancel button CustomID = %q, want %q", cancelBtn.CustomID, tt.data.CancelButtonID)
			}
		})
	}
}

func TestBuildPreviewEmbed(t *testing.T) {
	tests := []struct {
		name           string
		data           PreviewData
		wantTitle      string
		wantContent    string
		wantColor      int
		wantChannel    string
		wantHasMsgID   bool
		wantMsgID      string
		wantFooterText string
	}{
		{
			name: "compose preview embed",
			data: PreviewData{
				Content:       "Test message",
				TargetChannel: "123456789",
				IsEdit:        false,
			},
			wantTitle:      "Compose Preview",
			wantContent:    "```\nTest message\n```",
			wantColor:      0x3498db,
			wantChannel:    "<#123456789>",
			wantHasMsgID:   false,
			wantFooterText: "Click Post to send the message, or Cancel to discard.",
		},
		{
			name: "edit preview embed",
			data: PreviewData{
				Content:       "Edited message",
				TargetChannel: "987654321",
				IsEdit:        true,
				OriginalMsgID: "msg123456",
			},
			wantTitle:      "Edit Preview",
			wantContent:    "```\nEdited message\n```",
			wantColor:      0xe67e22,
			wantChannel:    "<#987654321>",
			wantHasMsgID:   true,
			wantMsgID:      "msg123456",
			wantFooterText: "Click Apply to confirm the edit, or Cancel to discard.",
		},
		{
			name: "edit preview without original message id",
			data: PreviewData{
				Content:       "Edited message",
				TargetChannel: "987654321",
				IsEdit:        true,
				OriginalMsgID: "",
			},
			wantTitle:      "Edit Preview",
			wantContent:    "```\nEdited message\n```",
			wantColor:      0xe67e22,
			wantChannel:    "<#987654321>",
			wantHasMsgID:   false,
			wantFooterText: "Click Apply to confirm the edit, or Cancel to discard.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPreviewEmbed(tt.data)

			if got.Title != tt.wantTitle {
				t.Errorf("buildPreviewEmbed() Title = %q, want %q", got.Title, tt.wantTitle)
			}

			if got.Description != tt.wantContent {
				t.Errorf("buildPreviewEmbed() Description = %q, want %q", got.Description, tt.wantContent)
			}

			if got.Color != tt.wantColor {
				t.Errorf("buildPreviewEmbed() Color = %d, want %d", got.Color, tt.wantColor)
			}

			if got.Footer == nil || got.Footer.Text != tt.wantFooterText {
				t.Errorf("buildPreviewEmbed() Footer.Text = %q, want %q",
					func() string {
						if got.Footer == nil {
							return ""
						}
						return got.Footer.Text
					}(),
					tt.wantFooterText)
			}

			// Check target channel field
			if len(got.Fields) < 1 {
				t.Errorf("buildPreviewEmbed() Fields length = %d, want at least 1", len(got.Fields))
			} else {
				if got.Fields[0].Name != "Target Channel" {
					t.Errorf("buildPreviewEmbed() Fields[0].Name = %q, want Target Channel", got.Fields[0].Name)
				}
				if got.Fields[0].Value != tt.wantChannel {
					t.Errorf("buildPreviewEmbed() Fields[0].Value = %q, want %q", got.Fields[0].Value, tt.wantChannel)
				}
			}

			// Check original message ID field for edits
			if tt.wantHasMsgID {
				if len(got.Fields) < 2 {
					t.Errorf("buildPreviewEmbed() Fields length = %d, want at least 2 for edit", len(got.Fields))
				} else {
					if got.Fields[1].Name != "Original Message" {
						t.Errorf("buildPreviewEmbed() Fields[1].Name = %q, want Original Message", got.Fields[1].Name)
					}
					if !strings.Contains(got.Fields[1].Value, tt.wantMsgID) {
						t.Errorf("buildPreviewEmbed() Fields[1].Value = %q, should contain %q", got.Fields[1].Value, tt.wantMsgID)
					}
				}
			} else if tt.data.IsEdit && len(got.Fields) > 1 {
				t.Errorf("buildPreviewEmbed() Fields length = %d, should not have Original Message field when ID is empty", len(got.Fields))
			}
		})
	}
}

// ==================== BENCHMARK TESTS ====================

func BenchmarkValidateContent(b *testing.B) {
	var content string = strings.Repeat("a", 1000)
	for i := 0; i < b.N; i++ {
		ValidateContent(content)
	}
}

func BenchmarkValidateEditPermission(b *testing.B) {
	var proxyMsg *storage.ProxyMessage = &storage.ProxyMessage{
		OwnerID: "user123",
	}
	for i := 0; i < b.N; i++ {
		ValidateEditPermission(proxyMsg, "user123")
	}
}

func BenchmarkFormatProxiedContent(b *testing.B) {
	var content string = strings.Repeat("Hello World ", 50)
	for i := 0; i < b.N; i++ {
		FormatProxiedContent(content, "TestUser")
	}
}

func BenchmarkExtractMessageIDFromURL(b *testing.B) {
	var url string = "https://discord.com/channels/123/456/789"
	for i := 0; i < b.N; i++ {
		ExtractMessageIDFromURL(url)
	}
}

func BenchmarkRenderPreviewResponse(b *testing.B) {
	var data PreviewData = PreviewData{
		Content:         "Test content for preview",
		TargetChannel:   "channel123",
		IsEdit:          false,
		ConfirmButtonID: "confirm_test",
		CancelButtonID:  "cancel_test",
	}
	for i := 0; i < b.N; i++ {
		RenderPreviewResponse(data)
	}
}
