package storage

import (
	"encoding/json"
	"time"
)

// Guild stores minimal metadata about a Discord guild.
type Guild struct {
	ID   string
	Name string
}

// ProxyMessage stores metadata about a proxied message.
type ProxyMessage struct {
	GuildID      string
	ChannelID    string
	MessageID    string
	OwnerID      string
	Content      string     // stored for edit reference
	CreatedAt    time.Time
	LastEditedAt *time.Time // nil if never edited
	LastEditedBy string     // empty if never edited
	WebhookID    string     // for editing via webhook
	WebhookToken string     // for editing via webhook
}

// GuildConfig stores per-guild configuration settings.
type GuildConfig struct {
	GuildID            string
	AllowedRoles       []string // role IDs that can use compose commands
	DefaultChannel     string   // default target channel for compose
	LogChannel         string   // channel for audit logs
	RestrictedChannels []string // channel IDs blacklisted from compose
	AllowedChannels    []string // channel IDs whitelisted for compose (empty = all allowed)
}

// Store defines the interface for persistence operations.
// Design allows easy swapping of implementations (in-memory, database, etc.)
type Store interface {
	// Guild operations
	SaveGuild(guildID, name string) error
	GetGuild(guildID string) (*Guild, error)
	DeleteGuild(guildID string) error

	// Guild config operations
	SaveGuildConfig(config GuildConfig) error
	GetGuildConfig(guildID string) (*GuildConfig, error)

	// Proxy message operations
	SaveProxyMessage(msg ProxyMessage) error
	GetProxyMessage(guildID, messageID string) (*ProxyMessage, error)
	UpdateProxyMessage(msg ProxyMessage) error
	DeleteProxyMessage(guildID, messageID string) error
}

// SerializeStringSlice converts a string slice to a JSON string for storage.
// Returns "[]" for nil or empty slices.
func SerializeStringSlice(slice []string) (string, error) {
	if len(slice) == 0 {
		return "[]", nil
	}
	var bytes, err = json.Marshal(slice)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// DeserializeStringSlice parses a JSON string into a string slice.
// Returns empty slice for empty or invalid input (with error for invalid JSON).
func DeserializeStringSlice(data string) ([]string, error) {
	if data == "" || data == "null" {
		return []string{}, nil
	}
	var result []string
	var err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
