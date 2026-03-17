package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// HandleGuildDelete returns a handler for GuildDelete events.
// Removes guild data when the bot leaves.
// Policy: Hard delete on leave (as documented in infrastructure.md).
func HandleGuildDelete(store storage.Store) func(*discordgo.Session, *discordgo.GuildDelete) {
	return func(s *discordgo.Session, g *discordgo.GuildDelete) {
		var err error = store.DeleteGuild(g.ID)
		if err != nil {
			log.Printf("Failed to delete guild %s: %v", g.ID, err)
			return
		}

		log.Printf("Removed guild data for %s (bot left or was removed)", g.ID)
	}
}
