package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// HandleGuildCreate returns a handler for GuildCreate events.
// Stores guild metadata and ensures default config exists when bot joins or reconnects.
func HandleGuildCreate(store storage.Store) func(*discordgo.Session, *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		defer recoverPanic("HandleGuildCreate")
		// Upsert guild metadata
		var err error = store.SaveGuild(g.ID, g.Name)
		if err != nil {
			logging.Error("failed to save guild",
				logging.String("guild_id", g.ID),
				logging.String("guild_name", g.Name),
				logging.Err("error", err),
			)
			return
		}

		logging.Info("guild create received",
			logging.String("guild_id", g.ID),
			logging.String("guild_name", g.Name),
			logging.Int("member_count", g.MemberCount),
		)

		// Ensure default config exists
		var config *storage.GuildConfig
		config, err = store.GetGuildConfig(g.ID)
		if err != nil {
			logging.Error("failed to get guild config",
				logging.String("guild_id", g.ID),
				logging.Err("error", err),
			)
			return
		}

		if config == nil {
			err = store.SaveGuildConfig(storage.GuildConfig{GuildID: g.ID})
			if err != nil {
				logging.Error("failed to save default guild config",
					logging.String("guild_id", g.ID),
					logging.Err("error", err),
				)
				return
			}
			logging.Debug("created default guild config",
				logging.String("guild_id", g.ID),
			)
		}

		logging.Info("guild ready",
			logging.String("guild_id", g.ID),
			logging.String("guild_name", g.Name),
		)
	}
}
