# Templates - Adding New Features

Patterns and templates for adding commands, handlers, and storage. Use these so new code lands in the right place.

## Adding a New Event Handler

1. Create `internal/events/<event_name>.go` (e.g. `guild_create.go`).
2. Implement a handler function with the correct discordgo event signature.
3. Register the handler in `main.go` via `session.AddHandler`.

### Template: Event Handler Script

```go
package events

import "github.com/bwmarrin/discordgo"

// HandleGuildCreate is called when the bot joins a guild.
func HandleGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	// Update database: register guild, default config, etc.
}
```

In `main.go`:

```go
session.AddHandler(events.HandleGuildCreate)
```

## Adding a New Slash Command Type and Definition

1. Add a new const to the `TSlashCommand` const block (e.g. in `internal/commands/types.go` or alongside command definitions).
2. Add the command definition and execute function in the appropriate command file (e.g. `compose.go`).
3. Add an entry to `CommandDefinitions` (or equivalent map) mapping the new type to its definition and execute function.
4. Ensure the command is included in the registry sync list.

### Template: Slash Command Type and Definition

```go
// In types.go - add const to TSlashCommand block
const (
	ComposeDraft TSlashCommand = "compose-draft"
	ComposeEdit  TSlashCommand = "compose-edit"
	// Add new: MyNewCommand TSlashCommand = "my-new-command"
)

// In command file (e.g., compose.go) - define command definition
var MyNewCommandDefinition *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        string(MyNewCommand),
	Description: "Description of what this command does",
	Options:     []*discordgo.ApplicationCommandOption{
		// ... command options
	},
}

// In command file's init() function - register to CommandDefinitions
CommandDefinitions[MyNewCommand] = SCommandDef{
	Definition:   MyNewCommandDefinition,
	Execute:      MyNewCommandExecute,
	Autocomplete: nil, // or autocomplete function if needed
}
```

## Adding a New Button Type

1. Add a new const to the `TButton` const block with the `custom_id` value.
2. Follow the convention: `button_<context>_<action>` (e.g. `button_compose_preview_post`).
3. Add the button handler/execute logic.
4. Add an entry to `ButtonDefinitions` map in the command file's `init()` function.

### Template: Button Type

```go
const (
	ButtonComposePreviewPost   TButton = "button_compose_preview_post"
	ButtonComposePreviewCancel TButton = "button_compose_preview_cancel"
	// Add new: ButtonContextAction TButton = "button_context_action"
)
```

When creating the button in a response, use the const value as `CustomID`.

## Adding a New Select Menu Type

1. Add a new const to the `TSelectMenu` const block with the `custom_id` value.
2. Follow the convention: `select_<context>_<action>` (e.g. `select_vote_approve`).
3. Add the select menu handler/execute logic.
4. Add an entry to `MSelectMenuDefinitions` (or equivalent map).

### Template: Select Menu Type

```go
const (
    VoteApprove  TSelectMenu = "select_vote_approve"
    VoteReject   TSelectMenu = "select_vote_reject"
    // Add new: MyContextAction TSelectMenu = "select_my-context_action"
)
```

When creating the select menu in a response, use the const value as `CustomID`.

## Adding a New Modal Submit Type

1. Add a new const to the `TModalSubmit` const block with the `custom_id` value.
2. Follow the convention: `modal_<context>_<action>` (e.g. `modal_compose-create_confirm`).
3. Add the modal submit handler/execute logic.
4. Add an entry to `MModalSubmitDefinitions` (or equivalent map).

### Template: Modal Submit Type

```go
const (
    ComposeCreateConfirm TModalSubmit = "modal_compose-create_confirm"
    // Add new: MyContextAction TModalSubmit = "modal_my-context_action"
)
```

When creating the modal in a response, use the const value as `CustomID`.

## Adding a Message or User Context Command

1. Add a new const to the `TMessageCommand` or `TUserCommand` const block.
2. Use naming: `context-action` (e.g. `message-edit`).
3. Add the command definition (ApplicationCommand with type Message/User) and execute function.
4. Add an entry to the appropriate map (e.g. `MMessageCommandDefinitions`, `MUserCommandDefinitions`).
5. Include in registry sync list.

### Template: Message Context Command

```go
const (
    MessageEdit TMessageCommand = "message-edit"
)

// Definition: ApplicationCommand with Type ApplicationCommandMessage
// Execute: HandleMessageEdit(s *discordgo.Session, i *discordgo.InteractionCreate)
```

## Checklist: New Interaction Type

- [ ] Add const to the appropriate type list (TSlashCommand, TButton, TSelectMenu, TModalSubmit, TMessageCommand, TUserCommand)
- [ ] Add definition and execute function
- [ ] Add entry to the routing map (CommandDefinitions, MButtonDefinitions, MSelectMenuDefinitions, MModalSubmitDefinitions, etc.)
- [ ] Wire in InteractionCreate if routing logic needs updating
- [ ] Update ROUTE_MAP.md with the new route
- [ ] Add tests if applicable

## Checklist: New Definition Struct

When adding a new interaction type that needs its own map (e.g. MSelectMenuDefinitions, MModalSubmitDefinitions):

- [ ] Define the type (e.g. `type TSelectMenu string`)
- [ ] Define the struct (e.g. `type SSelectMenuDef struct { Execute func(...) }`)
- [ ] Define the map type (e.g. `type MSelectMenuDefinitions map[TSelectMenu]SSelectMenuDef`)
- [ ] Add routing branch in InteractionCreate
- [ ] Update ARCHITECTURE.md and PROJECT_MAP.md

### Template: Definition Structs

Located in `internal/commands/definitions.go`:

```go
package commands

import "github.com/bwmarrin/discordgo"

// SCommandDef defines a slash, message, or user context command.
type SCommandDef struct {
	Definition   *discordgo.ApplicationCommand
	Execute      func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Autocomplete func(s *discordgo.Session, i *discordgo.InteractionCreate) // optional
}

// MCommandDefinitions maps slash command types to their definitions.
type MCommandDefinitions map[TSlashCommand]SCommandDef

// SButtonDef defines a button handler.
type SButtonDef struct {
	Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// MButtonDefinitions maps button custom IDs to their handlers.
type MButtonDefinitions map[TButton]SButtonDef

// SSelectMenuDef defines a select menu interaction handler.
type SSelectMenuDef struct {
	Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// MSelectMenuDefinitions maps select menu types to their definitions.
type MSelectMenuDefinitions map[TSelectMenu]SSelectMenuDef

// SModalSubmitDef defines a modal submit interaction handler.
type SModalSubmitDef struct {
	Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// MModalSubmitDefinitions maps modal submit types to their definitions.
type MModalSubmitDefinitions map[TModalSubmit]SModalSubmitDef
```

## Adding a New Command Group

1. Create `internal/commands/<group>.go` (e.g. `admin.go`).
2. Define the command group and subcommands.
3. Add the command definition to `internal/commands/registry.go` (in the desired commands list).
4. Wire handlers from `internal/handlers` or add new handlers if needed.

Commands sync on startup: the registry fetches existing, diffs against the desired list, and bulk overwrites only when definitions differ.

### Template: Command Group File

```go
package commands

import (
	"github.com/bwmarrin/discordgo"
)

// DefineCommandGroup returns the ApplicationCommand for this group.
func DefineCommandGroup() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "groupname",
		Description: "Brief description",
		Options:     []*discordgo.ApplicationCommandOption{
			// subcommands here
		},
	}
}

// HandleGroupName routes subcommand execution.
func HandleGroupName(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Route by i.ApplicationCommandData().Options[0].Name
	// Call handlers as needed
}
```

## Adding a New Handler

1. Create or edit a file in `internal/handlers/`.
2. Keep handlers pure where possible (inputs in, outputs out).
3. Accept `DiscordSession` interface for testability, or `*discordgo.Session` for simple cases.
4. Command handlers use signature `func(*discordgo.Session, *discordgo.InteractionCreate)`.

### Template: Handler Function

```go
package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// DoSomething performs X. Used by commands A and B.
// Uses DiscordSession interface for testability.
func DoSomething(s DiscordSession, guildID string, input string, store storage.Store) (*discordgo.Message, error) {
	// Validate, build, return
	return &discordgo.Message{Content: input}, nil
}
```

### Template: Command/Button Handler

```go
package commands

import "github.com/bwmarrin/discordgo"

// MyCommandExecute handles the slash command or button interaction.
// This signature matches what SCommandDef and SButtonDef expect.
func MyCommandExecute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Extract data from interaction
	var data discordgo.ApplicationCommandInteractionData = i.ApplicationCommandData()
	// ... handle command
}
```

### Template: Using DiscordSession Interface

```go
package handlers

import "github.com/bwmarrin/discordgo"

// DiscordSession interface allows mocking in tests.
// Use handlers.NewDiscordSession(s) to wrap *discordgo.Session.
var session DiscordSession = NewDiscordSession(s)

// Now pass 'session' to handler functions for testability.
result := SomeHandler(session, guildID, content)
```

## Adding Storage

1. Define interface in `internal/storage/interface.go`.
2. Implement in-memory version in `internal/storage/memory.go`.
3. Use interface in handlers so implementation can be swapped later.

### Template: Storage Interface

```go
package storage

import "time"

// ProxyMessage stores metadata about a proxied message for edit tracking.
type ProxyMessage struct {
	GuildID      string     // Discord guild ID where message was posted
	ChannelID    string     // Discord channel ID where message was posted
	MessageID    string     // Discord message ID of the proxied message
	OwnerID      string     // Discord user ID of the message author
	Content      string     // Message text for edit reference and history
	CreatedAt    time.Time  // Timestamp when message was first created
	LastEditedAt *time.Time // Timestamp of last edit, nil if never edited
	LastEditedBy string     // Discord user ID of last editor, empty if never edited
	WebhookID    string     // Webhook ID used for editing the proxied message
	WebhookToken string     // Webhook token used for editing the proxied message
}

// Store defines the interface for persistence operations.
// Allows swapping implementations (in-memory for testing, SQLite for production).
type Store interface {
	// Guild operations
	SaveGuild(guildID, name string) error
	GetGuild(guildID string) (*Guild, error)
	DeleteGuild(guildID string) error

	// Guild config operations
	SaveGuildConfig(config GuildConfig) error
	GetGuildConfig(guildID string) (*GuildConfig, error)

	// Proxy message operations
	SaveProxyMessage(msg ProxyMessage) error
	GetProxyMessage(guildID, messageID string) (*ProxyMessage, error)
	UpdateProxyMessage(msg ProxyMessage) error
	DeleteProxyMessage(guildID, messageID string) error
}
```

## Adding New Planning Documentation

For design documents and planning notes, add them to `docs/` directly or update existing docs:

- For current project status and backlog: Update `docs/PROJECT_STATUS.md`
- For architecture changes: Update `docs/ARCHITECTURE.md`
- For route changes: Update `docs/ROUTE_MAP.md`
- For implementation patterns: Update `docs/TEMPLATES.md`

## Checklist: New Compose Subcommand

- [ ] Add option to `compose` command in `internal/commands/compose.go`
- [ ] Add handler branch in compose handler
- [ ] Reuse or add handler in `internal/handlers/`
- [ ] Update `ROUTE_MAP.md` with new route
- [ ] Add tests if applicable

## Checklist: New Top-Level Command Group

- [ ] Create `internal/commands/<group>.go`
- [ ] Add command definition to `internal/commands/registry.go` (desired commands list)
- [ ] Add handler in `main.go` for the new group
- [ ] Update `PROJECT_MAP.md` and `ROUTE_MAP.md`
