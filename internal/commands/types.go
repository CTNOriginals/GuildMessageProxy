package commands

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

// Slash command constants follow naming: context-action (e.g., compose-create)
const (
	// Placeholder slash command for verification sync
	ComposeCreate TSlashCommand = "compose-create"
)

// Button constants follow naming: button_<context>_<action>
// Example: button_compose-create_post, button_compose-create_cancel
const (
	// Placeholder button types for future implementation
	ButtonComposeCreatePost   TButton = "button_compose-create_post"
	ButtonComposeCreateCancel TButton = "button_compose-create_cancel"
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
