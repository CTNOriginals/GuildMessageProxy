package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// ConfigRoleDefinition for setting allowed roles.
var ConfigRoleDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ConfigRole),
	Description: "Set which roles can use compose commands",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionRole,
			Name:        "role",
			Description: "Role to allow for compose commands",
			Required:    true,
		},
	},
}

// ConfigChannelDefinition for setting default channel.
var ConfigChannelDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ConfigChannel),
	Description: "Set default target channel for compose commands",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Default channel for compose messages",
			Required:    true,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
		},
	},
}

// ConfigRestrictDefinition for blacklisting channels.
var ConfigRestrictDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ConfigRestrict),
	Description: "Blacklist a channel from compose commands",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Channel to blacklist",
			Required:    true,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
		},
	},
}

// ConfigAllowDefinition for whitelisting channels.
var ConfigAllowDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ConfigAllow),
	Description: "Whitelist a channel for compose commands",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Channel to whitelist",
			Required:    true,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
		},
	},
}

// ConfigDefaultsDefinition for viewing current settings.
var ConfigDefaultsDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(ConfigDefaults),
	Description: "View current guild configuration settings",
}

// isGuildAdmin checks if the user has ManageGuild permission.
func isGuildAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		return false
	}
	return perms&discordgo.PermissionManageGuild != 0
}

// respondWithEmbed sends an embed response to the user.
func respondWithEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, title string, description string, fields []*discordgo.MessageEmbedField) {
	var embed = &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Fields:      fields,
		Color:       0x3498db, // Blue color
	}

	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		logging.Error("Failed to send embed response",
			logging.Err("error", err),
			logging.String("user_id", i.Member.User.ID),
		)
	}
}

// stringSliceContains checks if a string slice contains a value.
func stringSliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// removeFromStringSlice removes a value from a string slice.
func removeFromStringSlice(slice []string, value string) []string {
	var result []string
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}

// getOrCreateGuildConfig retrieves or creates a guild config.
func getOrCreateGuildConfig(guildID string) (*storage.GuildConfig, error) {
	var config, err = Store.GetGuildConfig(guildID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		config = &storage.GuildConfig{
			GuildID:            guildID,
			AllowedRoles:       []string{},
			DefaultChannel:     "",
			LogChannel:         "",
			RestrictedChannels: []string{},
			AllowedChannels:    []string{},
		}
	}
	return config, nil
}

// ConfigRoleExecute handles the config-role command.
func ConfigRoleExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isGuildAdmin(s, i) {
		respondWithError(s, i, "You need the Manage Guild permission to use this command.", nil)
		return
	}

	var guildID string = i.GuildID
	var data discordgo.ApplicationCommandInteractionData = i.ApplicationCommandData()

	var roleID string = ""
	for _, option := range data.Options {
		if option.Name == "role" {
			if option.RoleValue(s, guildID) != nil {
				roleID = option.RoleValue(s, guildID).ID
			}
		}
	}

	if roleID == "" {
		respondWithError(s, i, "Invalid role provided.", nil)
		return
	}

	var config, err = getOrCreateGuildConfig(guildID)
	if err != nil {
		respondWithError(s, i, "Failed to load guild configuration.", err)
		return
	}

	if stringSliceContains(config.AllowedRoles, roleID) {
		respondWithError(s, i, "This role is already allowed.", nil)
		return
	}

	config.AllowedRoles = append(config.AllowedRoles, roleID)

	err = Store.SaveGuildConfig(*config)
	if err != nil {
		respondWithError(s, i, "Failed to save configuration.", err)
		return
	}

	var roleMention string = fmt.Sprintf("<@&%s>", roleID)
	respondToUser(s, i, fmt.Sprintf("Role %s has been added to the allowed roles list.", roleMention))

	logging.Info("Config role added",
		logging.String("guild_id", guildID),
		logging.String("role_id", roleID),
	)
}

// ConfigChannelExecute handles the config-channel command.
func ConfigChannelExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isGuildAdmin(s, i) {
		respondWithError(s, i, "You need the Manage Guild permission to use this command.", nil)
		return
	}

	var guildID string = i.GuildID
	var data discordgo.ApplicationCommandInteractionData = i.ApplicationCommandData()

	var channelID string = ""
	for _, option := range data.Options {
		if option.Name == "channel" {
			if option.ChannelValue(s) != nil {
				channelID = option.ChannelValue(s).ID
			}
		}
	}

	if channelID == "" {
		respondWithError(s, i, "Invalid channel provided.", nil)
		return
	}

	var config, err = getOrCreateGuildConfig(guildID)
	if err != nil {
		respondWithError(s, i, "Failed to load guild configuration.", err)
		return
	}

	config.DefaultChannel = channelID

	err = Store.SaveGuildConfig(*config)
	if err != nil {
		respondWithError(s, i, "Failed to save configuration.", err)
		return
	}

	var channelMention string = fmt.Sprintf("<#%s>", channelID)
	respondToUser(s, i, fmt.Sprintf("Default channel set to %s.", channelMention))

	logging.Info("Config default channel updated",
		logging.String("guild_id", guildID),
		logging.String("channel_id", channelID),
	)
}

// ConfigRestrictExecute handles the config-restrict command.
func ConfigRestrictExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isGuildAdmin(s, i) {
		respondWithError(s, i, "You need the Manage Guild permission to use this command.", nil)
		return
	}

	var guildID string = i.GuildID
	var data discordgo.ApplicationCommandInteractionData = i.ApplicationCommandData()

	var channelID string = ""
	for _, option := range data.Options {
		if option.Name == "channel" {
			if option.ChannelValue(s) != nil {
				channelID = option.ChannelValue(s).ID
			}
		}
	}

	if channelID == "" {
		respondWithError(s, i, "Invalid channel provided.", nil)
		return
	}

	var config, err = getOrCreateGuildConfig(guildID)
	if err != nil {
		respondWithError(s, i, "Failed to load guild configuration.", err)
		return
	}

	if stringSliceContains(config.RestrictedChannels, channelID) {
		respondWithError(s, i, "This channel is already restricted.", nil)
		return
	}

	config.RestrictedChannels = append(config.RestrictedChannels, channelID)

	config.AllowedChannels = removeFromStringSlice(config.AllowedChannels, channelID)

	err = Store.SaveGuildConfig(*config)
	if err != nil {
		respondWithError(s, i, "Failed to save configuration.", err)
		return
	}

	var channelMention string = fmt.Sprintf("<#%s>", channelID)
	respondToUser(s, i, fmt.Sprintf("Channel %s has been restricted.", channelMention))

	logging.Info("Config channel restricted",
		logging.String("guild_id", guildID),
		logging.String("channel_id", channelID),
	)
}

// ConfigAllowExecute handles the config-allow command.
func ConfigAllowExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isGuildAdmin(s, i) {
		respondWithError(s, i, "You need the Manage Guild permission to use this command.", nil)
		return
	}

	var guildID string = i.GuildID
	var data discordgo.ApplicationCommandInteractionData = i.ApplicationCommandData()

	var channelID string = ""
	for _, option := range data.Options {
		if option.Name == "channel" {
			if option.ChannelValue(s) != nil {
				channelID = option.ChannelValue(s).ID
			}
		}
	}

	if channelID == "" {
		respondWithError(s, i, "Invalid channel provided.", nil)
		return
	}

	var config, err = getOrCreateGuildConfig(guildID)
	if err != nil {
		respondWithError(s, i, "Failed to load guild configuration.", err)
		return
	}

	if stringSliceContains(config.AllowedChannels, channelID) {
		respondWithError(s, i, "This channel is already allowed.", nil)
		return
	}

	config.AllowedChannels = append(config.AllowedChannels, channelID)

	config.RestrictedChannels = removeFromStringSlice(config.RestrictedChannels, channelID)

	err = Store.SaveGuildConfig(*config)
	if err != nil {
		respondWithError(s, i, "Failed to save configuration.", err)
		return
	}

	var channelMention string = fmt.Sprintf("<#%s>", channelID)
	respondToUser(s, i, fmt.Sprintf("Channel %s has been whitelisted.", channelMention))

	logging.Info("Config channel allowed",
		logging.String("guild_id", guildID),
		logging.String("channel_id", channelID),
	)
}

// ConfigDefaultsExecute handles the config-defaults command.
func ConfigDefaultsExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isGuildAdmin(s, i) {
		respondWithError(s, i, "You need the Manage Guild permission to use this command.", nil)
		return
	}

	var guildID string = i.GuildID

	var config, err = getOrCreateGuildConfig(guildID)
	if err != nil {
		respondWithError(s, i, "Failed to load guild configuration.", err)
		return
	}

	var fields []*discordgo.MessageEmbedField

	var allowedRolesValue string = "*None set - all roles can use compose*"
	if len(config.AllowedRoles) > 0 {
		var roleMentions []string
		for _, roleID := range config.AllowedRoles {
			roleMentions = append(roleMentions, fmt.Sprintf("<@&%s>", roleID))
		}
		allowedRolesValue = strings.Join(roleMentions, ", ")
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Allowed Roles",
		Value:  allowedRolesValue,
		Inline: false,
	})

	var defaultChannelValue string = "*Not set - uses current channel*"
	if config.DefaultChannel != "" {
		defaultChannelValue = fmt.Sprintf("<#%s>", config.DefaultChannel)
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Default Channel",
		Value:  defaultChannelValue,
		Inline: false,
	})

	var restrictedChannelsValue string = "*None restricted*"
	if len(config.RestrictedChannels) > 0 {
		var channelMentions []string
		for _, channelID := range config.RestrictedChannels {
			channelMentions = append(channelMentions, fmt.Sprintf("<#%s>", channelID))
		}
		restrictedChannelsValue = strings.Join(channelMentions, ", ")
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Restricted Channels",
		Value:  restrictedChannelsValue,
		Inline: false,
	})

	var allowedChannelsValue string = "*All channels allowed (whitelist not used)*"
	if len(config.AllowedChannels) > 0 {
		var channelMentions []string
		for _, channelID := range config.AllowedChannels {
			channelMentions = append(channelMentions, fmt.Sprintf("<#%s>", channelID))
		}
		allowedChannelsValue = strings.Join(channelMentions, ", ")
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Whitelisted Channels",
		Value:  allowedChannelsValue,
		Inline: false,
	})

	respondWithEmbed(s, i, "Guild Configuration", "Current settings for this server:", fields)

	logging.Info("Config defaults viewed",
		logging.String("guild_id", guildID),
		logging.String("user_id", i.Member.User.ID),
	)
}

func init() {
	CommandDefinitions[ConfigRole] = SCommandDef{
		Definition:   ConfigRoleDefinition,
		Execute:      ConfigRoleExecute,
		Autocomplete: nil,
	}
	CommandDefinitions[ConfigChannel] = SCommandDef{
		Definition:   ConfigChannelDefinition,
		Execute:      ConfigChannelExecute,
		Autocomplete: nil,
	}
	CommandDefinitions[ConfigRestrict] = SCommandDef{
		Definition:   ConfigRestrictDefinition,
		Execute:      ConfigRestrictExecute,
		Autocomplete: nil,
	}
	CommandDefinitions[ConfigAllow] = SCommandDef{
		Definition:   ConfigAllowDefinition,
		Execute:      ConfigAllowExecute,
		Autocomplete: nil,
	}
	CommandDefinitions[ConfigDefaults] = SCommandDef{
		Definition:   ConfigDefaultsDefinition,
		Execute:      ConfigDefaultsExecute,
		Autocomplete: nil,
	}
}
