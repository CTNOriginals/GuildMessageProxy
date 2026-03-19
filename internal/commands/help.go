package commands

import (
	"github.com/bwmarrin/discordgo"
)

// ComposeHelpDefinition is the command definition for compose-help.
var ComposeHelpDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ComposeHelp),
	Description: "Show help for compose commands",
}

// ComposeHelpExecute handles the compose-help command.
// Sends an ephemeral help message explaining all compose commands.
func ComposeHelpExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var fields []*discordgo.MessageEmbedField

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "/compose draft <content> [channel]",
		Value:  "Create a message preview before posting. Shows you exactly how it will look before sending.",
		Inline: false,
	})

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "/compose send <content> [channel]",
		Value:  "Send a message immediately (skips the preview step).",
		Inline: false,
	})

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "/compose edit <message> <content>",
		Value:  "Edit a message you posted via /compose. Use the message link or ID.",
		Inline: false,
	})

	var embed *discordgo.MessageEmbed = &discordgo.MessageEmbed{
		Title:       "Compose Commands Help",
		Description: "Post messages on behalf of your server. Useful for announcements, updates, and collaborative posting.",
		Fields:      fields,
		Color:       0x3498db, // Blue color
	}

	// Add footer with helpful tips
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Tip: Use 'draft' to preview first. Drafts expire after 24 hours. For edits, paste a Discord message link or use the message ID.",
	}

	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		respondWithError(s, i, "Failed to send help message.", err)
	}
}

func init() {
	CommandDefinitions[ComposeHelp] = SCommandDef{
		Definition:   ComposeHelpDefinition,
		Execute:      ComposeHelpExecute,
		Autocomplete: nil,
	}
}
