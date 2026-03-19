// Package storage defines the persistence layer interface for GuildMessageProxy.
//
// This package provides swappable storage implementations allowing the bot to use
// different backends (in-memory for testing, SQLite for production, etc.)
// without changing business logic. All storage operations are abstracted
// behind the Store interface.
package storage

import (
	"encoding/json"
	"time"
)

// Guild stores minimal metadata about a Discord guild.
type Guild struct {
	// ID is the Discord guild (server) ID.
	ID string
	// Name is the Discord guild name at the time of registration.
	Name string
}

// ProxyMessage stores metadata about a proxied message for edit tracking.
type ProxyMessage struct {
	// GuildID is the Discord guild ID where the message was posted.
	GuildID string
	// ChannelID is the Discord channel ID where the message was posted.
	ChannelID string
	// MessageID is the Discord message ID of the proxied message.
	MessageID string
	// OwnerID is the Discord user ID of the message author.
	OwnerID string
	// Content stores the message text for edit reference and history.
	Content string
	// CreatedAt is the timestamp when the message was first created.
	CreatedAt time.Time
	// LastEditedAt is the timestamp of the last edit, nil if never edited.
	LastEditedAt *time.Time
	// LastEditedBy is the Discord user ID of the last editor, empty if never edited.
	LastEditedBy string
	// WebhookID is the webhook ID used for editing the proxied message.
	WebhookID string
	// WebhookToken is the webhook token used for editing the proxied message.
	WebhookToken string
}

// GuildConfig stores per-guild configuration settings for the compose feature.
type GuildConfig struct {
	// GuildID is the Discord guild ID this configuration applies to.
	GuildID string
	// AllowedRoles contains role IDs that are permitted to use compose commands.
	AllowedRoles []string
	// DefaultChannel is the default target channel for compose messages.
	DefaultChannel string
	// LogChannel is the channel ID where audit logs are sent.
	LogChannel string
	// RestrictedChannels contains channel IDs blacklisted from compose.
	RestrictedChannels []string
	// AllowedChannels contains channel IDs whitelisted for compose. Empty means all channels are allowed.
	AllowedChannels []string
}

// Store defines the interface for persistence operations.
//
// The interface design allows easy swapping of implementations (in-memory for testing,
// SQLite for production) without modifying application code.
//
// Example usage:
//
//	// Testing: use in-memory implementation
//	var testStore storage.Store = storage.NewMemoryStore()
//
//	// Production: use database implementation
//	var dbStore storage.Store = storage.NewSQLiteStore("guildmessageproxy.db")
//
//	// Both implement the same interface, so handlers work identically
//	msg, err := store.GetProxyMessage(guildID, messageID)
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

// SerializeStringSlice converts a string slice to a JSON string for database storage.
//
// Returns "[]" for nil or empty slices to ensure consistent database representation.
// Returns an error if JSON marshaling fails (unlikely for string slices).
//
// Example:
//
//	roles := []string{"123456789", "987654321"}
//	jsonStr, err := storage.SerializeStringSlice(roles)
//	// jsonStr == `["123456789","987654321"]`
//
//	empty, _ := storage.SerializeStringSlice(nil)
//	// empty == "[]"
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
//
// Returns an empty slice for empty input or JSON null values.
// Returns an error if the input contains invalid JSON that cannot be parsed.
//
// Example:
//
//	data := `["role1","role2"]`
//	roles, err := storage.DeserializeStringSlice(data)
//	// roles == []string{"role1", "role2"}
//
//	empty, _ := storage.DeserializeStringSlice("")
//	// empty == []string{}
//
//	invalid, err := storage.DeserializeStringSlice("invalid")
//	// err != nil (JSON parse error)
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
