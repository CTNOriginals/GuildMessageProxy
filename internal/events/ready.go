package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// HandleReady logs bot startup confirmation.
func HandleReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Bot is ready! Logged in as %s#%s (ID: %s)", r.User.Username, r.User.Discriminator, r.User.ID)
	log.Printf("Connected to %d guilds", len(r.Guilds))
}
