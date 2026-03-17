package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// HandleGuildCreate returns a handler for GuildCreate events.
// Stores guild metadata and ensures default config exists when bot joins or reconnects.
func HandleGuildCreate(store storage.Store) func(*discordgo.Session, *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		// Upsert guild metadata
		var err error = store.SaveGuild(g.ID, g.Name)
		if err != nil {
			log.Printf("Failed to save guild %s (%s): %v", g.ID, g.Name, err)
			return
		}

		// Ensure default config exists
		var config *storage.GuildConfig
		config, err = store.GetGuildConfig(g.ID)
		if err != nil {
			log.Printf("Failed to get guild config for %s: %v", g.ID, err)
			return
		}

		if config == nil {
			err = store.SaveGuildConfig(storage.GuildConfig{GuildID: g.ID})
			if err != nil {
				log.Printf("Failed to save default guild config for %s: %v", g.ID, err)
				return
			}
			log.Printf("Created default config for guild %s (%s)", g.ID, g.Name)
		}

		log.Printf("Guild ready: %s (%s)", g.Name, g.ID)
	}
}
