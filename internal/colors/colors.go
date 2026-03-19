package colors

// Discord embed color constants for consistent UI theming.
// All colors use hex format (0xRRGGBB) compatible with discordgo.MessageEmbed.Color.
const (
	// Primary blue for compose operations and general info
	Primary = 0x3498db

	// Orange for edit operations
	Edit = 0xe67e22

	// Yellow/orange for warnings (e.g., draft expiring soon)
	Warning = 0xf39c12

	// Red for error states
	Error = 0xe74c3c

	// Green for success states
	Success = 0x2ecc71
)
