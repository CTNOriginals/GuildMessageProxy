package storage

import "time"

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
// All fields are placeholders for MVP.
type GuildConfig struct {
	GuildID        string
	AllowedRoles   []string // placeholder for MVP
	DefaultChannel string   // placeholder for MVP
	LogChannel     string   // placeholder for MVP
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
