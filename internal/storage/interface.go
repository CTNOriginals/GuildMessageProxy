package storage

// Guild stores minimal metadata about a Discord guild.
type Guild struct {
	ID   string
	Name string
}

// ProxyMessage stores metadata about a proxied message.
// Flags are placeholders for MVP.
type ProxyMessage struct {
	GuildID   string
	ChannelID string
	MessageID string
	OwnerID   string
	// Flags for MVP - placeholder for future features
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
}
