package handlers

import "github.com/bwmarrin/discordgo"

// DiscordSession defines the interface for Discord session operations used by handlers.
// This interface allows for easy mocking in tests while accepting *discordgo.Session in production.
type DiscordSession interface {
	// Channel returns a channel by ID
	Channel(channelID string) (*discordgo.Channel, error)

	// UserChannelPermissions returns the permissions for a user in a channel
	UserChannelPermissions(userID, channelID string) (int64, error)

	// ChannelWebhooks returns all webhooks for a channel
	ChannelWebhooks(channelID string) ([]*discordgo.Webhook, error)

	// WebhookCreate creates a new webhook in a channel
	WebhookCreate(channelID, name, avatar string) (*discordgo.Webhook, error)

	// WebhookExecute executes a webhook to send a message
	WebhookExecute(webhookID, token string, wait bool, data *discordgo.WebhookParams) (*discordgo.Message, error)

	// WebhookMessageEdit edits a message sent via webhook
	WebhookMessageEdit(webhookID, token, messageID string, data *discordgo.WebhookEdit) (*discordgo.Message, error)

	// BotUser returns the bot's user information
	BotUser() *discordgo.User

	// ChannelTyping triggers the typing indicator in a channel
	ChannelTyping(channelID string) error
}

// sessionWrapper wraps *discordgo.Session to implement DiscordSession interface
type sessionWrapper struct {
	session *discordgo.Session
}

// NewDiscordSession wraps a *discordgo.Session to implement DiscordSession
func NewDiscordSession(s *discordgo.Session) DiscordSession {
	return &sessionWrapper{session: s}
}

func (w *sessionWrapper) Channel(channelID string) (*discordgo.Channel, error) {
	return w.session.Channel(channelID)
}

func (w *sessionWrapper) UserChannelPermissions(userID, channelID string) (int64, error) {
	return w.session.UserChannelPermissions(userID, channelID)
}

func (w *sessionWrapper) ChannelWebhooks(channelID string) ([]*discordgo.Webhook, error) {
	return w.session.ChannelWebhooks(channelID)
}

func (w *sessionWrapper) WebhookCreate(channelID, name, avatar string) (*discordgo.Webhook, error) {
	return w.session.WebhookCreate(channelID, name, avatar)
}

func (w *sessionWrapper) WebhookExecute(webhookID, token string, wait bool, data *discordgo.WebhookParams) (*discordgo.Message, error) {
	return w.session.WebhookExecute(webhookID, token, wait, data)
}

func (w *sessionWrapper) WebhookMessageEdit(webhookID, token, messageID string, data *discordgo.WebhookEdit) (*discordgo.Message, error) {
	return w.session.WebhookMessageEdit(webhookID, token, messageID, data)
}

func (w *sessionWrapper) BotUser() *discordgo.User {
	if w.session.State != nil {
		return w.session.State.User
	}
	return nil
}

func (w *sessionWrapper) ChannelTyping(channelID string) error {
	return w.session.ChannelTyping(channelID)
}
