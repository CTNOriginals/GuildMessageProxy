package handlers

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// TestHasAnyRole tests the hasAnyRole helper function.
func TestHasAnyRole(t *testing.T) {
	tests := []struct {
		name        string
		memberRoles []string
		allowedRoles []string
		want        bool
	}{
		{
			name:         "user_has_matching_role",
			memberRoles:  []string{"role1", "role2", "role3"},
			allowedRoles: []string{"role2"},
			want:         true,
		},
		{
			name:         "user_has_no_matching_roles",
			memberRoles:  []string{"role1", "role2"},
			allowedRoles: []string{"role3", "role4"},
			want:         false,
		},
		{
			name:         "user_has_multiple_matching_roles",
			memberRoles:  []string{"role1", "role2", "role3"},
			allowedRoles: []string{"role2", "role3"},
			want:         true,
		},
		{
			name:         "empty_member_roles",
			memberRoles:  []string{},
			allowedRoles: []string{"role1"},
			want:         false,
		},
		{
			name:         "empty_allowed_roles",
			memberRoles:  []string{"role1"},
			allowedRoles: []string{},
			want:         false,
		},
		{
			name:         "both_empty",
			memberRoles:  []string{},
			allowedRoles: []string{},
			want:         false,
		},
		{
			name:         "single_matching_role",
			memberRoles:  []string{"role1"},
			allowedRoles: []string{"role1"},
			want:         true,
		},
		{
			name:         "partial_role_id_match",
			memberRoles:  []string{"123456789"},
			allowedRoles: []string{"123456"},
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool = hasAnyRole(tt.memberRoles, tt.allowedRoles)
			if got != tt.want {
				t.Errorf("hasAnyRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsChannelRestricted tests the isChannelRestricted helper function.
func TestIsChannelRestricted(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		config    *storage.GuildConfig
		want      bool
	}{
		{
			name:      "channel_in_restricted_list",
			channelID: "channel123",
			config: &storage.GuildConfig{
				RestrictedChannels: []string{"channel123", "channel456"},
			},
			want: true,
		},
		{
			name:      "channel_not_in_restricted_list",
			channelID: "channel789",
			config: &storage.GuildConfig{
				RestrictedChannels: []string{"channel123", "channel456"},
			},
			want: false,
		},
		{
			name:      "nil_config",
			channelID: "channel123",
			config:    nil,
			want:      false,
		},
		{
			name:      "empty_restricted_list",
			channelID: "channel123",
			config: &storage.GuildConfig{
				RestrictedChannels: []string{},
			},
			want: false,
		},
		{
			name:      "empty_channel_id",
			channelID: "",
			config: &storage.GuildConfig{
				RestrictedChannels: []string{"channel123"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool = isChannelRestricted(tt.channelID, tt.config)
			if got != tt.want {
				t.Errorf("isChannelRestricted() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsChannelAllowed tests the isChannelAllowed helper function.
func TestIsChannelAllowed(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		config    *storage.GuildConfig
		want      bool
	}{
		{
			name:      "channel_in_allowed_list",
			channelID: "channel123",
			config: &storage.GuildConfig{
				AllowedChannels: []string{"channel123", "channel456"},
			},
			want: true,
		},
		{
			name:      "channel_not_in_allowed_list",
			channelID: "channel789",
			config: &storage.GuildConfig{
				AllowedChannels: []string{"channel123", "channel456"},
			},
			want: false,
		},
		{
			name:      "empty_allowed_list_allows_all",
			channelID: "channel123",
			config: &storage.GuildConfig{
				AllowedChannels: []string{},
			},
			want: true,
		},
		{
			name:      "nil_config_allows_all",
			channelID: "channel123",
			config:    nil,
			want:      true,
		},
		{
			name:      "nil_allowed_channels_allows_all",
			channelID: "channel123",
			config: &storage.GuildConfig{
				AllowedChannels: nil,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool = isChannelAllowed(tt.channelID, tt.config)
			if got != tt.want {
				t.Errorf("isChannelAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCanUseCompose_WithRoleChecks tests CanUseCompose with role-based permissions.
func TestCanUseCompose_WithRoleChecks(t *testing.T) {
	tests := []struct {
		name          string
		perms         int64
		memberRoles   []string
		allowedRoles  []string
		restrictedChs []string
		allowedChs    []string
		guildConfig   *storage.GuildConfig
		wantAllowed   bool
		wantError     string
	}{
		{
			name:         "user_has_allowed_role",
			perms:        discordgo.PermissionSendMessages,
			memberRoles:  []string{"role1", "role2"},
			allowedRoles: []string{"role2"},
			wantAllowed:  true,
			wantError:    "",
		},
		{
			name:         "user_lacks_allowed_role",
			perms:        discordgo.PermissionSendMessages,
			memberRoles:  []string{"role1", "role3"},
			allowedRoles: []string{"role2", "role4"},
			wantAllowed:  false,
			wantError:    "You need one of the allowed roles to use this command. Contact server admins.",
		},
		{
			name:         "no_allowed_roles_configured",
			perms:        discordgo.PermissionSendMessages,
			memberRoles:  []string{"role1"},
			allowedRoles: []string{},
			wantAllowed:  true,
			wantError:    "",
		},
		{
			name:          "channel_restricted",
			perms:         discordgo.PermissionSendMessages,
			memberRoles:   []string{},
			allowedRoles:  []string{},
			restrictedChs: []string{"channel123"},
			wantAllowed:   false,
			wantError:     "This channel is restricted from using compose commands.",
		},
		{
			name:         "channel_not_in_whitelist",
			perms:        discordgo.PermissionSendMessages,
			memberRoles:  []string{},
			allowedRoles: []string{},
			allowedChs:   []string{"channel456", "channel789"},
			wantAllowed:  false,
			wantError:    "This channel is not allowed for compose commands.",
		},
		{
			name:         "channel_in_whitelist",
			perms:        discordgo.PermissionSendMessages,
			memberRoles:  []string{},
			allowedRoles: []string{},
			allowedChs:   []string{"channel123"},
			wantAllowed:  true,
			wantError:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockSession *MockDiscordSession = &MockDiscordSession{
				ChannelFunc: func(channelID string) (*discordgo.Channel, error) {
					return &discordgo.Channel{ID: "channel123", Name: "test-channel"}, nil
				},
				UserChannelPermissionsFunc: func(userID, channelID string) (int64, error) {
					return tt.perms, nil
				},
			}

			var mockStore *MockStore = NewMockStore()
			if tt.allowedRoles != nil || tt.restrictedChs != nil || tt.allowedChs != nil {
				var config storage.GuildConfig = storage.GuildConfig{
					GuildID:            "guild123",
					AllowedRoles:       tt.allowedRoles,
					RestrictedChannels: tt.restrictedChs,
					AllowedChannels:    tt.allowedChs,
				}
				mockStore.SaveGuildConfig(config)
			}

			var result PermissionResult = CanUseCompose(mockSession, "guild123", "channel123", "user123", mockStore, tt.memberRoles)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanUseCompose() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if result.Error != tt.wantError {
				t.Errorf("CanUseCompose() Error = %q, want %q", result.Error, tt.wantError)
			}
		})
	}
}

// TestCanUseCompose_ConfigError tests that permission check continues when config fetch fails.
func TestCanUseCompose_ConfigError(t *testing.T) {
	var mockSession *MockDiscordSession = &MockDiscordSession{
		ChannelFunc: func(channelID string) (*discordgo.Channel, error) {
			return &discordgo.Channel{ID: "channel123", Name: "test-channel"}, nil
		},
		UserChannelPermissionsFunc: func(userID, channelID string) (int64, error) {
			return discordgo.PermissionSendMessages, nil
		},
	}

	var mockStore *MockStore = NewMockStore()
	mockStore.getError = errors.New("database error")

	var result PermissionResult = CanUseCompose(mockSession, "guild123", "channel123", "user123", mockStore, []string{})

	if !result.Allowed {
		t.Errorf("CanUseCompose() should allow when config fetch fails, got Allowed = %v", result.Allowed)
	}
}
