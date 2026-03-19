# Project Map - Where Things Live

A spatial map of the GuildMessageProxy codebase. Use this to find where files belong and what exists.

## Directory Tree (Current + Planned)

```
GuildMessageProxy/
|-- cmd/
|   |-- bot/
|   |   |-- main.go              [EXISTS] Entry point, config load, session init
|   |
|-- internal/                     [EXISTS]
|   |-- events/
|   |   |-- interaction_create.go [EXISTS] Routes all interactions (slash, buttons, select, modal, context) to definitions
|   |   |-- guild_create.go       [EXISTS] Store guild metadata and config when bot joins
|   |   |-- guild_delete.go       [EXISTS] Remove/soft-delete guild data when bot leaves
|   |   |-- ready.go              [EXISTS] Optional: bot startup confirmation, log ready state
|   |   |-- error.go              [EXISTS] Handle REST/gateway errors (log, user feedback, optional embed)
|   |
|   |-- commands/
|   |   |-- types.go              [EXISTS] TSlashCommand, TButton, TSelectMenu, TModalSubmit, TMessageCommand, TUserCommand + const blocks
|   |   |-- compose.go            [EXISTS] /compose group + subcommands
|   |   |-- definitions.go        [EXISTS] Command and button definitions
|   |   |-- registry.go           [EXISTS] Command definitions + startup sync (fetch, diff, bulk overwrite)
|   |   |-- config.go             [EXISTS] Guild configuration commands
|   |   |-- message.go            [EXISTS] Message deletion command
|   |   |-- help.go               [EXISTS] Compose help command
|   |
|   |-- handlers/
|   |   |-- preview.go           [EXISTS] Render preview message payload
|   |   |-- post.go              [EXISTS] Post/update proxied message
|   |   |-- edit.go              [EXISTS] Edit proxied message handler
|   |   |-- permissions.go       [EXISTS] Who can create/set/propose/post
|   |   |-- validation.go        [EXISTS] Shared input validation
|   |
|   |-- storage/
|   |   |-- memory.go            [EXISTS] In-memory proxy metadata
|   |   |-- sqlite.go            [EXISTS] SQLite persistence implementation
|   |   |-- interface.go         [EXISTS] Storage interface for swap later
|
|   |-- health/
|   |   |-- health.go            [EXISTS] Health check server for monitoring
|
|   |-- logging/
|   |   |-- logging.go           [EXISTS] Structured logging implementation
|   |   |-- levels.go            [EXISTS] Log level definitions
|
|-- docs/
|   |-- INDEX.md                 [EXISTS] This index
|   |-- PROJECT_MAP.md           [EXISTS] This file
|   |-- ARCHITECTURE.md          [EXISTS] Package layout
|   |-- ROUTE_MAP.md             [EXISTS] Command routes
|   |-- TEMPLATES.md             [EXISTS] File templates
|   |-- GLOSSARY.md              [EXISTS] Terms and jargon
|   |-- DEPLOYMENT.md            [EXISTS] Deployment guide
|   |-- TROUBLESHOOTING.md      [EXISTS] Troubleshooting guide
|
|-- .env.example                 [EXISTS] Env var template
|-- .gitignore                   [EXISTS]
|-- .cursorignore                [EXISTS]
|-- .cursor/
|   |-- agents/                  [EXISTS] Subagent definitions (documenter, developer, etc)
|   |-- rules/                   [EXISTS] Cursor rules
|-- go.mod                       [EXISTS] Module deps
|-- go.sum                       [EXISTS]
|-- Makefile                     [EXISTS] run, build, test, lint, tidy
|-- README.md                    [EXISTS] User-facing project description
```

## File Location Rules

| What | Where |
|------|-------|
| Entry point, config load, session init | `cmd/bot/main.go` |
| Event handlers (InteractionCreate, GuildCreate, GuildDelete, Ready, Error) | `internal/events/` |
| Interaction types (TSlashCommand, TButton, etc.) | `internal/commands/types.go` or `internal/interactions/types.go` |
| Slash command definitions + handlers | `internal/commands/<group>.go` |
| Command definitions + startup sync | `internal/commands/registry.go` |
| Reusable logic (preview, post, permissions) | `internal/handlers/` |
| Persistence (proxy metadata, guild config) | `internal/storage/` |
| Structured logging implementation | `internal/logging/*.go` |
| Health check server | `internal/health/health.go` |
| Agent navigation docs | `docs/` (this folder) |

## What Exists vs Planned

| Area | Status |
|------|--------|
| `cmd/bot/main.go` | Exists - full features |
| `internal/commands/compose.go` | Exists - full implementation |
| `internal/handlers/` | Exists - preview, post, edit, validation, permissions |
| `internal/storage/` | Exists - in-memory with proxy message support |
| `internal/events/` | Exists - interaction routing with button support |
| `internal/logging/` | Exists - structured logging with levels |
| `internal/health/` | Exists - health check server |
| `internal/commands/config.go` | Exists - guild configuration |
| `internal/commands/message.go` | Exists - message deletion |
| `internal/commands/help.go` | Exists - compose help |
| `internal/storage/sqlite.go` | Exists - SQLite persistence |
| `docs/` | Exists - agent navigation docs |

## Key Files to Know

- **main.go** - Thin entry point. Load config, init session, wire handlers, graceful shutdown.
- **registry.go** - Command definitions and startup sync: fetch existing, diff against desired, bulk overwrite only when changed. Supports `--guild` (dev) or `--global` (prod).
- **compose.go** - Main user-facing command group for compose/create/propose/post.
- **config.go** - Guild configuration commands.
- **message.go** - Message management commands.
- **handlers/** - Shared logic used by multiple commands. Keep commands thin.
- **logging/** - Structured logging with configurable levels and output formatting.

## Environment Variables (.env.example)

| Variable | Purpose |
|----------|---------|
| `TOKEN` | Discord bot token |
| `CLIENT_ID` | Discord application client ID |
| `DEV_GUILD_ID` | Development guild for command registration |
| `DEV_CHANNEL_LOG_ID` | Development channel for logs |
| `DEV_CHANNEL_ERROR_ID` | Development channel for errors |

## Out of Scope (Do Not Add Here)

- Discord API wrapper package - use discordgo directly; add `internal/discordutil` only if common patterns emerge.
- Broad `discordapi` package - avoid.
- Message-based commands - prioritize slash commands and interactions.
