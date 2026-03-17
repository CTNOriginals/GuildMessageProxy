package storage

import "fmt"

// MemoryStore provides an in-memory implementation of the Store interface.
// Data is lost when the process terminates - suitable for development and MVP.
type MemoryStore struct {
	guilds       map[string]*Guild
	guildConfigs map[string]*GuildConfig
}

// NewMemoryStore creates a new in-memory store with initialized maps.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		guilds:       make(map[string]*Guild),
		guildConfigs: make(map[string]*GuildConfig),
	}
}

// SaveGuild stores or updates guild metadata.
// Uses upsert pattern: overwrites existing data if guild already exists.
func (m *MemoryStore) SaveGuild(guildID, name string) error {
	if guildID == "" {
		return fmt.Errorf("guildID cannot be empty")
	}
	m.guilds[guildID] = &Guild{
		ID:   guildID,
		Name: name,
	}
	return nil
}

// GetGuild retrieves guild metadata by ID.
// Returns nil if guild not found.
func (m *MemoryStore) GetGuild(guildID string) (*Guild, error) {
	if guildID == "" {
		return nil, fmt.Errorf("guildID cannot be empty")
	}
	guild, ok := m.guilds[guildID]
	if !ok {
		return nil, nil
	}
	return guild, nil
}

// DeleteGuild removes guild metadata and associated config.
// Policy: Hard delete on leave (as documented in infrastructure.md).
func (m *MemoryStore) DeleteGuild(guildID string) error {
	if guildID == "" {
		return fmt.Errorf("guildID cannot be empty")
	}
	delete(m.guilds, guildID)
	delete(m.guildConfigs, guildID)
	return nil
}

// SaveGuildConfig stores or updates guild configuration.
// Uses upsert pattern: overwrites existing config.
func (m *MemoryStore) SaveGuildConfig(config GuildConfig) error {
	if config.GuildID == "" {
		return fmt.Errorf("GuildID cannot be empty")
	}
	m.guildConfigs[config.GuildID] = &GuildConfig{
		GuildID:        config.GuildID,
		AllowedRoles:   config.AllowedRoles,
		DefaultChannel: config.DefaultChannel,
		LogChannel:     config.LogChannel,
	}
	return nil
}

// GetGuildConfig retrieves guild configuration by ID.
// Returns nil if config not found.
func (m *MemoryStore) GetGuildConfig(guildID string) (*GuildConfig, error) {
	if guildID == "" {
		return nil, fmt.Errorf("guildID cannot be empty")
	}
	config, ok := m.guildConfigs[guildID]
	if !ok {
		return nil, nil
	}
	return config, nil
}
