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
// In types or commands package
const (
	ComposeCreate TSlashCommand = "compose-create"
	ComposeEdit   TSlashCommand = "compose-edit"
	// Add new: MyNewCommand TSlashCommand = "my-new-command"
)

// In CommandDefinitions map
var CommandDefinitions MCommandDefinitions = MCommandDefinitions{
	ComposeCreate: {Definition: ComposeCreateDefinition, Execute: ComposeCreateExecute},
	MyNewCommand:  {Definition: MyNewCommandDefinition, Execute: MyNewCommandExecute},
	// ...
}
```

## Adding a New Button Type

1. Add a new const to the `TButton` const block with the `custom_id` value.
2. Follow the convention: `button_<context>_<action>` (e.g. `button_compose-create_post`).
3. Add the button handler/execute logic.
4. Add an entry to the button definitions map (if one exists) or wire the handler in InteractionCreate.

### Template: Button Type

```go
const (
	ComposeCreatePost   TButton = "button_compose-create_post"
	ComposeCreateCancel TButton = "button_compose-create_cancel"
	// Add new: MyContextAction TButton = "button_my-context_action"
)
```

When creating the button in a response, use the const value as `CustomID`.

## Checklist: New Interaction Type

- [ ] Add const to the appropriate type list (TSlashCommand, TButton, etc.)
- [ ] Add definition and execute function
- [ ] Add entry to the routing map (CommandDefinitions, button map, etc.)
- [ ] Wire in InteractionCreate if routing logic needs updating
- [ ] Update ROUTE_MAP.md with the new route
- [ ] Add tests if applicable

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
3. Accept `*discordgo.Session` and interaction/context as needed.

### Template: Handler Function

```go
package handlers

import "github.com/bwmarrin/discordgo"

// DoSomething performs X. Used by commands A and B.
func DoSomething(s *discordgo.Session, guildID string, input string) (*discordgo.MessageSend, error) {
	// Validate, build, return
	return &discordgo.MessageSend{Content: input}, nil
}
```

## Adding Storage

1. Define interface in `internal/storage/interface.go`.
2. Implement in-memory version in `internal/storage/memory.go`.
3. Use interface in handlers so implementation can be swapped later.

### Template: Storage Interface

```go
package storage

type ProxyMessage struct {
	GuildID   string
	ChannelID string
	MessageID string
	OwnerID   string
	// ...
}

type Store interface {
	SaveProxyMessage(m ProxyMessage) error
	GetProxyMessage(guildID, messageID string) (*ProxyMessage, error)
}
```

## Adding a New Roadmap Doc

1. Create `docs/roadmap/<topic>.md` (e.g. `feature-voting-system.md`).
2. Add a line to `docs/roadmap/overview.md` under "Roadmap Docs Index" if it is a major planning doc.

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
