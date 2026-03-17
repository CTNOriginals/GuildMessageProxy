package events

import (
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/bwmarrin/discordgo"
)

// HandleReady logs bot startup confirmation.
func HandleReady(s *discordgo.Session, r *discordgo.Ready) {
	logging.Info("bot ready",
		logging.String("username", r.User.Username),
		logging.String("session_id", r.SessionID),
		logging.Int("connected_guilds", len(r.Guilds)),
	)
}
