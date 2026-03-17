package storage

import (
	"fmt"

	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
)

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
		logging.Error("storage write failed",
			logging.String("operation", "SaveGuild"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID cannot be empty")),
		)
		return fmt.Errorf("guildID cannot be empty")
	}

	logging.Debug("storage write",
		logging.String("operation", "SaveGuild"),
		logging.String("key", guildID),
	)

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
		logging.Error("storage error",
			logging.String("operation", "GetGuild"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID cannot be empty")),
		)
		return nil, fmt.Errorf("guildID cannot be empty")
	}
	guild, ok := m.guilds[guildID]
	result := "hit"
	if !ok {
		result = "miss"
	}
	logging.Debug("storage read",
		logging.String("operation", "GetGuild"),
		logging.String("key", guildID),
		logging.String("result", result),
	)
	return guild, nil
}

// DeleteGuild removes guild metadata and associated config.
// Policy: Hard delete on leave (as documented in infrastructure.md).
func (m *MemoryStore) DeleteGuild(guildID string) error {
	if guildID == "" {
		logging.Error("storage error",
			logging.String("operation", "DeleteGuild"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID cannot be empty")),
		)
		return fmt.Errorf("guildID cannot be empty")
	}
	logging.Debug("storage delete",
		logging.String("operation", "DeleteGuild"),
		logging.String("key", guildID),
	)
	delete(m.guilds, guildID)
	delete(m.guildConfigs, guildID)
	return nil
}

// SaveGuildConfig stores or updates guild configuration.
// Uses upsert pattern: overwrites existing config.
func (m *MemoryStore) SaveGuildConfig(config GuildConfig) error {
	if config.GuildID == "" {
		logging.Error("storage error",
			logging.String("operation", "SaveGuildConfig"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("GuildID cannot be empty")),
		)
		return fmt.Errorf("GuildID cannot be empty")
	}
	logging.Debug("storage write",
		logging.String("operation", "SaveGuildConfig"),
		logging.String("key", config.GuildID),
	)
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
		logging.Error("storage error",
			logging.String("operation", "GetGuildConfig"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID cannot be empty")),
		)
		return nil, fmt.Errorf("guildID cannot be empty")
	}
	config, ok := m.guildConfigs[guildID]
	result := "hit"
	if !ok {
		result = "miss"
	}
	logging.Debug("storage read",
		logging.String("operation", "GetGuildConfig"),
		logging.String("key", guildID),
		logging.String("result", result),
	)
	return config, nil
}
