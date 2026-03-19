package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// HandleGuildDelete returns a handler for GuildDelete events.
// Removes guild data when the bot leaves.
// Policy: Hard delete on leave (as documented in infrastructure.md).
func HandleGuildDelete(store storage.Store) func(*discordgo.Session, *discordgo.GuildDelete) {
	return func(s *discordgo.Session, g *discordgo.GuildDelete) {
		defer recoverPanic("HandleGuildDelete")
		logging.Info("guild delete received",
			logging.String("guild_id", g.ID),
		)

		var err error = store.DeleteGuild(g.ID)
		if err != nil {
			logging.Error("failed to delete guild",
				logging.String("guild_id", g.ID),
				logging.Err("error", err),
			)
			return
		}

		logging.Debug("guild data cleaned up",
			logging.String("guild_id", g.ID),
		)

		logging.Info("guild removed",
			logging.String("guild_id", g.ID),
			logging.String("reason", "bot left or was removed"),
		)
	}
}
