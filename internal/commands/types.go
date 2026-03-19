package commands

import "github.com/bwmarrin/discordgo"

// TSlashCommand identifies slash commands by their name.
type TSlashCommand string

// TButton identifies button interactions by their CustomID.
type TButton string

// TSelectMenu identifies select menu interactions by their CustomID.
type TSelectMenu string

// TModalSubmit identifies modal submit interactions by their CustomID.
type TModalSubmit string

// TMessageCommand identifies message context menu commands by their name.
type TMessageCommand string

// TUserCommand identifies user context menu commands by their name.
type TUserCommand string

// Slash command constants follow naming: context-action (e.g., compose-draft)
const (
	// Compose commands (new naming)
	ComposeDraft TSlashCommand = "compose-draft"
	ComposeSend  TSlashCommand = "compose-send"
	ComposeEdit  TSlashCommand = "compose-edit"
	ComposeHelp  TSlashCommand = "compose-help"

	// Backward compatibility aliases (old names)
	ComposeCreate  TSlashCommand = ComposeDraft
	ComposePost    TSlashCommand = ComposeSend
	ComposePropose TSlashCommand = ComposeEdit

	// Config commands (admin-only guild configuration)
	ConfigRole     TSlashCommand = "config-role"
	ConfigChannel  TSlashCommand = "config-channel"
	ConfigRestrict TSlashCommand = "config-restrict"
	ConfigAllow    TSlashCommand = "config-allow"
	ConfigDefaults TSlashCommand = "config-defaults"

	// Message commands (message management)
	MessageDelete TSlashCommand = "message-delete"
)

// Button constants follow naming: button_<context>_<action>
const (
	// Compose preview buttons
	ButtonComposePreviewPost   TButton = "button_compose_preview_post"
	ButtonComposePreviewCancel TButton = "button_compose_preview_cancel"

	// Edit proposal buttons
	ButtonEditPreviewApply  TButton = "button_edit_preview_apply"
	ButtonEditPreviewCancel TButton = "button_edit_preview_cancel"

	// Discard confirmation buttons
	ButtonComposeConfirmDiscard TButton = "button_compose_confirm_discard"
	ButtonComposeKeepDraft      TButton = "button_compose_keep_draft"

	// Retry/cancel buttons after post failure
	ButtonComposeRetryPost        TButton = "button_compose_retry_post"
	ButtonComposeCancelAfterError TButton = "button_compose_cancel_after_error"

	// Retry/cancel buttons after edit failure
	ButtonEditRetryApply       TButton = "button_edit_retry_apply"
	ButtonEditCancelAfterError TButton = "button_edit_cancel_after_error"
)

// Select menu constants follow naming: select_<context>_<action>
// Example: select_vote_approve, select_vote_reject
const (
	// Placeholder select menu types for future implementation
	SelectVoteApprove TSelectMenu = "select_vote_approve"
	SelectVoteReject  TSelectMenu = "select_vote_reject"
)

// Modal submit constants follow naming: modal_<context>_<action>
// Example: modal_compose-create_confirm
const (
	// Placeholder modal types for future implementation
	ModalComposeCreateConfirm TModalSubmit = "modal_compose-create_confirm"
)

// Message context command constants follow naming: context-action
const (
	// Placeholder message command types for future implementation
	MessageEdit TMessageCommand = "message-edit"
)

// User context command constants follow naming: context-action
const (
	// Placeholder user command types for future implementation
	UserInfo TUserCommand = "user-info"
)

// SButtonDef defines a button handler
type SButtonDef struct {
	Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// MButtonDefinitions maps button custom IDs to their handlers
type MButtonDefinitions map[TButton]SButtonDef

// ButtonDefinitions is the source of truth for button handlers
var ButtonDefinitions MButtonDefinitions = make(MButtonDefinitions)
