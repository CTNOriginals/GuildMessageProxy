package storage

import (
	"database/sql"
	"fmt"
)

// Migration represents a single database schema migration.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations is the ordered list of all migrations.
// Add new migrations at the end to maintain version order.
var migrations = []Migration{
	{
		Version: 1,
		Name:    "initial_schema",
		SQL: `
			-- Guilds table: minimal metadata about Discord guilds
			CREATE TABLE IF NOT EXISTS guilds (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);

			-- Guild configs table: per-guild configuration settings
			CREATE TABLE IF NOT EXISTS guild_configs (
				guild_id TEXT PRIMARY KEY,
				allowed_roles TEXT DEFAULT '[]', -- JSON array of role IDs
				default_channel TEXT,
				log_channel TEXT,
				restricted_channels TEXT DEFAULT '[]', -- JSON array of channel IDs
				allowed_channels TEXT DEFAULT '[]', -- JSON array of channel IDs
				FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE
			);

			-- Proxy messages table: metadata about proxied messages
			CREATE TABLE IF NOT EXISTS proxy_messages (
				guild_id TEXT NOT NULL,
				channel_id TEXT NOT NULL,
				message_id TEXT NOT NULL,
				owner_id TEXT NOT NULL,
				content TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				last_edited_at TIMESTAMP,
				last_edited_by TEXT,
				webhook_id TEXT,
				webhook_token TEXT,
				PRIMARY KEY (guild_id, message_id),
				FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE
			);

			-- Schema migrations tracking table
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`,
	},
}

// getCurrentVersion returns the highest applied migration version.
// Returns 0 if no migrations have been applied yet.
func getCurrentVersion(db *sql.DB) (int, error) {
	// Check if schema_migrations table exists
	var tableExists int
	err := db.QueryRow(`
		SELECT 1 FROM sqlite_master WHERE type='table' AND name='schema_migrations'
	`).Scan(&tableExists)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to check schema_migrations table: %w", err)
	}

	// Get the current version
	var version int
	err = db.QueryRow(`
		SELECT COALESCE(MAX(version), 0) FROM schema_migrations
	`).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}
	return version, nil
}

// runMigrations executes all pending migrations in order.
// Uses a transaction to ensure atomicity.
func (s *SQLiteStore) runMigrations() error {
	var currentVersion, err = getCurrentVersion(s.db)
	if err != nil {
		return err
	}

	// If no migrations table exists yet, we need to run all migrations
	if currentVersion == 0 {
		// Check if any tables exist (from a pre-migration database)
		var tablesExist int
		err = s.db.QueryRow(`
			SELECT 1 FROM sqlite_master WHERE type='table' AND name='guilds'
		`).Scan(&tablesExist)
		if err != sql.ErrNoRows && err != nil {
			return fmt.Errorf("failed to check existing tables: %w", err)
		}
		// If tables already exist but no version record, treat as migration v1 applied
		if err != sql.ErrNoRows {
			// Tables exist but no migration record - insert initial version
			_, err = s.db.Exec(`
				INSERT INTO schema_migrations (version, applied_at) VALUES (1, CURRENT_TIMESTAMP)
			`)
			if err != nil {
				return fmt.Errorf("failed to record initial migration version: %w", err)
			}
			currentVersion = 1
		}
	}

	// Find pending migrations
	var pendingMigrations []Migration
	for _, m := range migrations {
		if m.Version > currentVersion {
			pendingMigrations = append(pendingMigrations, m)
		}
	}

	if len(pendingMigrations) == 0 {
		return nil
	}

	// Execute pending migrations in a transaction
	var tx *sql.Tx
	tx, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, migration := range pendingMigrations {
		_, err = tx.Exec(migration.SQL)
		if err != nil {
			return fmt.Errorf("migration v%d (%s) failed: %w", migration.Version, migration.Name, err)
		}

		_, err = tx.Exec(`
			INSERT INTO schema_migrations (version, applied_at) VALUES (?, CURRENT_TIMESTAMP)
		`, migration.Version)
		if err != nil {
			return fmt.Errorf("failed to record migration v%d: %w", migration.Version, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit migrations: %w", err)
	}

	return nil
}
