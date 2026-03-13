# Templates - Adding New Features

Patterns and templates for adding commands, handlers, and storage. Use these so new code lands in the right place.

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
