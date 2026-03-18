package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore provides a SQLite-based implementation of the Store interface.
// Data persists across process restarts. Uses prepared statements for efficiency.
type SQLiteStore struct {
	db *sql.DB

	// Prepared statements for common operations
	stmtSaveGuild        *sql.Stmt
	stmtGetGuild         *sql.Stmt
	stmtDeleteGuild      *sql.Stmt
	stmtSaveGuildConfig  *sql.Stmt
	stmtGetGuildConfig   *sql.Stmt
	stmtSaveProxyMessage *sql.Stmt
	stmtGetProxyMessage  *sql.Stmt
	stmtUpdateProxyMessage *sql.Stmt
	stmtDeleteProxyMessage *sql.Stmt
}

// NewSQLiteStore creates a new SQLite store with the given database path.
// Initializes the database connection and runs migrations.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	var db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		logging.Error("failed to open sqlite database",
			logging.String("path", dbPath),
			logging.Err("error", err),
		)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign key support
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		logging.Error("failed to enable foreign keys",
			logging.Err("error", err),
		)
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	var store = &SQLiteStore{db: db}

	// Run migrations
	err = store.runMigrations()
	if err != nil {
		logging.Error("failed to run database migrations",
			logging.Err("error", err),
		)
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Prepare statements
	err = store.prepareStatements()
	if err != nil {
		logging.Error("failed to prepare statements",
			logging.Err("error", err),
		)
		db.Close()
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	logging.Info("sqlite store initialized",
		logging.String("path", dbPath),
	)
	return store, nil
}

// prepareStatements initializes all prepared statements for common operations.
func (s *SQLiteStore) prepareStatements() error {
	var err error

	s.stmtSaveGuild, err = s.db.Prepare(`
		INSERT INTO guilds (id, name, joined_at) VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name
	`)
	if err != nil {
		return fmt.Errorf("prepare SaveGuild: %w", err)
	}

	s.stmtGetGuild, err = s.db.Prepare(`
		SELECT id, name FROM guilds WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare GetGuild: %w", err)
	}

	s.stmtDeleteGuild, err = s.db.Prepare(`
		DELETE FROM guilds WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare DeleteGuild: %w", err)
	}

	s.stmtSaveGuildConfig, err = s.db.Prepare(`
		INSERT INTO guild_configs (
			guild_id, allowed_roles, default_channel, log_channel,
			restricted_channels, allowed_channels
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET
			allowed_roles = excluded.allowed_roles,
			default_channel = excluded.default_channel,
			log_channel = excluded.log_channel,
			restricted_channels = excluded.restricted_channels,
			allowed_channels = excluded.allowed_channels
	`)
	if err != nil {
		return fmt.Errorf("prepare SaveGuildConfig: %w", err)
	}

	s.stmtGetGuildConfig, err = s.db.Prepare(`
		SELECT guild_id, allowed_roles, default_channel, log_channel,
			restricted_channels, allowed_channels
		FROM guild_configs WHERE guild_id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare GetGuildConfig: %w", err)
	}

	s.stmtSaveProxyMessage, err = s.db.Prepare(`
		INSERT INTO proxy_messages (
			guild_id, channel_id, message_id, owner_id, content,
			created_at, last_edited_at, last_edited_by, webhook_id, webhook_token
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, message_id) DO UPDATE SET
			channel_id = excluded.channel_id,
			owner_id = excluded.owner_id,
			content = excluded.content,
			created_at = excluded.created_at,
			last_edited_at = excluded.last_edited_at,
			last_edited_by = excluded.last_edited_by,
			webhook_id = excluded.webhook_id,
			webhook_token = excluded.webhook_token
	`)
	if err != nil {
		return fmt.Errorf("prepare SaveProxyMessage: %w", err)
	}

	s.stmtGetProxyMessage, err = s.db.Prepare(`
		SELECT guild_id, channel_id, message_id, owner_id, content,
			created_at, last_edited_at, last_edited_by, webhook_id, webhook_token
		FROM proxy_messages WHERE guild_id = ? AND message_id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare GetProxyMessage: %w", err)
	}

	s.stmtUpdateProxyMessage, err = s.db.Prepare(`
		UPDATE proxy_messages SET
			channel_id = ?,
			owner_id = ?,
			content = ?,
			created_at = ?,
			last_edited_at = ?,
			last_edited_by = ?,
			webhook_id = ?,
			webhook_token = ?
		WHERE guild_id = ? AND message_id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare UpdateProxyMessage: %w", err)
	}

	s.stmtDeleteProxyMessage, err = s.db.Prepare(`
		DELETE FROM proxy_messages WHERE guild_id = ? AND message_id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare DeleteProxyMessage: %w", err)
	}

	return nil
}

// Close closes the database connection and all prepared statements.
func (s *SQLiteStore) Close() error {
	// Close all prepared statements
	var stmts = []*sql.Stmt{
		s.stmtSaveGuild,
		s.stmtGetGuild,
		s.stmtDeleteGuild,
		s.stmtSaveGuildConfig,
		s.stmtGetGuildConfig,
		s.stmtSaveProxyMessage,
		s.stmtGetProxyMessage,
		s.stmtUpdateProxyMessage,
		s.stmtDeleteProxyMessage,
	}

	for _, stmt := range stmts {
		if stmt != nil {
			stmt.Close()
		}
	}

	return s.db.Close()
}

// SaveGuild stores or updates guild metadata.
// Uses upsert pattern: overwrites existing data if guild already exists.
func (s *SQLiteStore) SaveGuild(guildID, name string) error {
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

	var joinedAt = time.Now()
	_, err := s.stmtSaveGuild.Exec(guildID, name, joinedAt)
	if err != nil {
		logging.Error("storage write failed",
			logging.String("operation", "SaveGuild"),
			logging.String("key", guildID),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to save guild: %w", err)
	}
	return nil
}

// GetGuild retrieves guild metadata by ID.
// Returns nil if guild not found.
func (s *SQLiteStore) GetGuild(guildID string) (*Guild, error) {
	if guildID == "" {
		logging.Error("storage error",
			logging.String("operation", "GetGuild"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID cannot be empty")),
		)
		return nil, fmt.Errorf("guildID cannot be empty")
	}

	var guild = &Guild{}
	err := s.stmtGetGuild.QueryRow(guildID).Scan(&guild.ID, &guild.Name)
	if err == sql.ErrNoRows {
		logging.Debug("storage read",
			logging.String("operation", "GetGuild"),
			logging.String("key", guildID),
			logging.String("result", "miss"),
		)
		return nil, nil
	}
	if err != nil {
		logging.Error("storage read failed",
			logging.String("operation", "GetGuild"),
			logging.String("key", guildID),
			logging.Err("error", err),
		)
		return nil, fmt.Errorf("failed to get guild: %w", err)
	}

	logging.Debug("storage read",
		logging.String("operation", "GetGuild"),
		logging.String("key", guildID),
		logging.String("result", "hit"),
	)
	return guild, nil
}

// DeleteGuild removes guild metadata and associated config (cascade).
// Policy: Hard delete on leave.
func (s *SQLiteStore) DeleteGuild(guildID string) error {
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

	_, err := s.stmtDeleteGuild.Exec(guildID)
	if err != nil {
		logging.Error("storage delete failed",
			logging.String("operation", "DeleteGuild"),
			logging.String("key", guildID),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to delete guild: %w", err)
	}
	return nil
}

// SaveGuildConfig stores or updates guild configuration.
// Uses upsert pattern: overwrites existing config.
func (s *SQLiteStore) SaveGuildConfig(config GuildConfig) error {
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

	var allowedRolesJSON, err = json.Marshal(config.AllowedRoles)
	if err != nil {
		logging.Error("storage write failed",
			logging.String("operation", "SaveGuildConfig"),
			logging.String("key", config.GuildID),
			logging.String("error_category", "serialization"),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to marshal allowed_roles: %w", err)
	}

	var restrictedChannelsJSON []byte
	restrictedChannelsJSON, err = json.Marshal(config.RestrictedChannels)
	if err != nil {
		logging.Error("storage write failed",
			logging.String("operation", "SaveGuildConfig"),
			logging.String("key", config.GuildID),
			logging.String("error_category", "serialization"),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to marshal restricted_channels: %w", err)
	}

	var allowedChannelsJSON []byte
	allowedChannelsJSON, err = json.Marshal(config.AllowedChannels)
	if err != nil {
		logging.Error("storage write failed",
			logging.String("operation", "SaveGuildConfig"),
			logging.String("key", config.GuildID),
			logging.String("error_category", "serialization"),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to marshal allowed_channels: %w", err)
	}

	_, err = s.stmtSaveGuildConfig.Exec(
		config.GuildID,
		string(allowedRolesJSON),
		config.DefaultChannel,
		config.LogChannel,
		string(restrictedChannelsJSON),
		string(allowedChannelsJSON),
	)
	if err != nil {
		logging.Error("storage write failed",
			logging.String("operation", "SaveGuildConfig"),
			logging.String("key", config.GuildID),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to save guild config: %w", err)
	}
	return nil
}

// GetGuildConfig retrieves guild configuration by ID.
// Returns nil if config not found.
func (s *SQLiteStore) GetGuildConfig(guildID string) (*GuildConfig, error) {
	if guildID == "" {
		logging.Error("storage error",
			logging.String("operation", "GetGuildConfig"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID cannot be empty")),
		)
		return nil, fmt.Errorf("guildID cannot be empty")
	}

	var config = &GuildConfig{}
	var allowedRolesJSON string
	var restrictedChannelsJSON string
	var allowedChannelsJSON string

	err := s.stmtGetGuildConfig.QueryRow(guildID).Scan(
		&config.GuildID,
		&allowedRolesJSON,
		&config.DefaultChannel,
		&config.LogChannel,
		&restrictedChannelsJSON,
		&allowedChannelsJSON,
	)
	if err == sql.ErrNoRows {
		logging.Debug("storage read",
			logging.String("operation", "GetGuildConfig"),
			logging.String("key", guildID),
			logging.String("result", "miss"),
		)
		return nil, nil
	}
	if err != nil {
		logging.Error("storage read failed",
			logging.String("operation", "GetGuildConfig"),
			logging.String("key", guildID),
			logging.Err("error", err),
		)
		return nil, fmt.Errorf("failed to get guild config: %w", err)
	}

	// Unmarshal JSON arrays
	if err = json.Unmarshal([]byte(allowedRolesJSON), &config.AllowedRoles); err != nil {
		logging.Error("storage read failed",
			logging.String("operation", "GetGuildConfig"),
			logging.String("key", guildID),
			logging.String("error_category", "deserialization"),
			logging.Err("error", err),
		)
		return nil, fmt.Errorf("failed to unmarshal allowed_roles: %w", err)
	}

	if err = json.Unmarshal([]byte(restrictedChannelsJSON), &config.RestrictedChannels); err != nil {
		logging.Error("storage read failed",
			logging.String("operation", "GetGuildConfig"),
			logging.String("key", guildID),
			logging.String("error_category", "deserialization"),
			logging.Err("error", err),
		)
		return nil, fmt.Errorf("failed to unmarshal restricted_channels: %w", err)
	}

	if err = json.Unmarshal([]byte(allowedChannelsJSON), &config.AllowedChannels); err != nil {
		logging.Error("storage read failed",
			logging.String("operation", "GetGuildConfig"),
			logging.String("key", guildID),
			logging.String("error_category", "deserialization"),
			logging.Err("error", err),
		)
		return nil, fmt.Errorf("failed to unmarshal allowed_channels: %w", err)
	}

	logging.Debug("storage read",
		logging.String("operation", "GetGuildConfig"),
		logging.String("key", guildID),
		logging.String("result", "hit"),
	)
	return config, nil
}

// SaveProxyMessage stores or updates proxy message metadata.
// Uses upsert pattern: overwrites existing data if message already exists.
func (s *SQLiteStore) SaveProxyMessage(msg ProxyMessage) error {
	if msg.GuildID == "" || msg.MessageID == "" {
		logging.Error("storage error",
			logging.String("operation", "SaveProxyMessage"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID and messageID cannot be empty")),
		)
		return fmt.Errorf("guildID and messageID cannot be empty")
	}

	var key = msg.GuildID + ":" + msg.MessageID
	logging.Debug("storage write",
		logging.String("operation", "SaveProxyMessage"),
		logging.String("key", key),
	)

	var lastEditedAt interface{}
	if msg.LastEditedAt != nil {
		lastEditedAt = *msg.LastEditedAt
	} else {
		lastEditedAt = nil
	}

	_, err := s.stmtSaveProxyMessage.Exec(
		msg.GuildID,
		msg.ChannelID,
		msg.MessageID,
		msg.OwnerID,
		msg.Content,
		msg.CreatedAt,
		lastEditedAt,
		msg.LastEditedBy,
		msg.WebhookID,
		msg.WebhookToken,
	)
	if err != nil {
		logging.Error("storage write failed",
			logging.String("operation", "SaveProxyMessage"),
			logging.String("key", key),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to save proxy message: %w", err)
	}
	return nil
}

// GetProxyMessage retrieves proxy message metadata by guild ID and message ID.
// Returns nil if message not found.
func (s *SQLiteStore) GetProxyMessage(guildID, messageID string) (*ProxyMessage, error) {
	if guildID == "" || messageID == "" {
		logging.Error("storage error",
			logging.String("operation", "GetProxyMessage"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID and messageID cannot be empty")),
		)
		return nil, fmt.Errorf("guildID and messageID cannot be empty")
	}

	var key = guildID + ":" + messageID
	var msg = &ProxyMessage{}
	var lastEditedAt sql.NullTime

	err := s.stmtGetProxyMessage.QueryRow(guildID, messageID).Scan(
		&msg.GuildID,
		&msg.ChannelID,
		&msg.MessageID,
		&msg.OwnerID,
		&msg.Content,
		&msg.CreatedAt,
		&lastEditedAt,
		&msg.LastEditedBy,
		&msg.WebhookID,
		&msg.WebhookToken,
	)
	if err == sql.ErrNoRows {
		logging.Debug("storage read",
			logging.String("operation", "GetProxyMessage"),
			logging.String("key", key),
			logging.String("result", "miss"),
		)
		return nil, nil
	}
	if err != nil {
		logging.Error("storage read failed",
			logging.String("operation", "GetProxyMessage"),
			logging.String("key", key),
			logging.Err("error", err),
		)
		return nil, fmt.Errorf("failed to get proxy message: %w", err)
	}

	if lastEditedAt.Valid {
		msg.LastEditedAt = &lastEditedAt.Time
	}

	logging.Debug("storage read",
		logging.String("operation", "GetProxyMessage"),
		logging.String("key", key),
		logging.String("result", "hit"),
	)
	return msg, nil
}

// UpdateProxyMessage updates existing proxy message metadata.
// Returns error if message does not exist.
func (s *SQLiteStore) UpdateProxyMessage(msg ProxyMessage) error {
	if msg.GuildID == "" || msg.MessageID == "" {
		logging.Error("storage error",
			logging.String("operation", "UpdateProxyMessage"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID and messageID cannot be empty")),
		)
		return fmt.Errorf("guildID and messageID cannot be empty")
	}

	var key = msg.GuildID + ":" + msg.MessageID

	// Check if message exists
	var exists int
	err := s.db.QueryRow(
		"SELECT 1 FROM proxy_messages WHERE guild_id = ? AND message_id = ?",
		msg.GuildID, msg.MessageID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		logging.Error("storage error",
			logging.String("operation", "UpdateProxyMessage"),
			logging.String("error_category", "not_found"),
			logging.String("key", key),
			logging.Err("error", fmt.Errorf("proxy message not found")),
		)
		return fmt.Errorf("proxy message not found: %s", key)
	}
	if err != nil {
		logging.Error("storage error",
			logging.String("operation", "UpdateProxyMessage"),
			logging.String("key", key),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to check message existence: %w", err)
	}

	logging.Debug("storage write",
		logging.String("operation", "UpdateProxyMessage"),
		logging.String("key", key),
	)

	var lastEditedAt interface{}
	if msg.LastEditedAt != nil {
		lastEditedAt = *msg.LastEditedAt
	} else {
		lastEditedAt = nil
	}

	_, err = s.stmtUpdateProxyMessage.Exec(
		msg.ChannelID,
		msg.OwnerID,
		msg.Content,
		msg.CreatedAt,
		lastEditedAt,
		msg.LastEditedBy,
		msg.WebhookID,
		msg.WebhookToken,
		msg.GuildID,
		msg.MessageID,
	)
	if err != nil {
		logging.Error("storage write failed",
			logging.String("operation", "UpdateProxyMessage"),
			logging.String("key", key),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to update proxy message: %w", err)
	}
	return nil
}

// DeleteProxyMessage removes proxy message metadata.
// No error if message does not exist.
func (s *SQLiteStore) DeleteProxyMessage(guildID, messageID string) error {
	if guildID == "" || messageID == "" {
		logging.Error("storage error",
			logging.String("operation", "DeleteProxyMessage"),
			logging.String("error_category", "validation"),
			logging.Err("error", fmt.Errorf("guildID and messageID cannot be empty")),
		)
		return fmt.Errorf("guildID and messageID cannot be empty")
	}

	var key = guildID + ":" + messageID
	logging.Debug("storage delete",
		logging.String("operation", "DeleteProxyMessage"),
		logging.String("key", key),
	)

	_, err := s.stmtDeleteProxyMessage.Exec(guildID, messageID)
	if err != nil {
		logging.Error("storage delete failed",
			logging.String("operation", "DeleteProxyMessage"),
			logging.String("key", key),
			logging.Err("error", err),
		)
		return fmt.Errorf("failed to delete proxy message: %w", err)
	}
	return nil
}
