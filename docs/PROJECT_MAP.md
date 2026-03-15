# Project Map - Where Things Live

A spatial map of the GuildMessageProxy codebase. Use this to find where files belong and what exists.

## Directory Tree (Current + Planned)

```
GuildMessageProxy/
|-- cmd/
|   |-- bot/
|   |   |-- main.go              [EXISTS] Entry point, config load, session init
|   |
|-- internal/                     [PLANNED - not yet created]
|   |-- events/
|   |   |-- interaction_create.go [PLANNED] Routes all interactions (slash, buttons, select, modal, context) to definitions
|   |   |-- guild_create.go       [PLANNED] Store guild metadata and config when bot joins
|   |   |-- guild_delete.go       [PLANNED] Remove/soft-delete guild data when bot leaves
|   |   |-- ready.go              [PLANNED] Optional: bot startup confirmation, log ready state
|   |   |-- error.go              [PLANNED] Handle REST/gateway errors (log, user feedback, optional embed)
|   |
|   |-- commands/
|   |   |-- types.go              [PLANNED] TSlashCommand, TButton, TSelectMenu, TModalSubmit, TMessageCommand, TUserCommand + const blocks
|   |   |-- compose.go            [PLANNED] /compose group + subcommands
|   |   |-- registry.go           [PLANNED] Command definitions + startup sync (fetch, diff, bulk overwrite)
|   |   |-- (admin.go, config.go) [FUTURE] Other command groups
|   |
|   |-- handlers/
|   |   |-- preview.go           [PLANNED] Render preview message payload
|   |   |-- post.go              [PLANNED] Post/update proxied message
|   |   |-- permissions.go       [PLANNED] Who can create/set/propose/post
|   |   |-- validation.go        [PLANNED] Shared input validation
|   |
|   |-- storage/
|   |   |-- memory.go            [PLANNED] In-memory proxy metadata
|   |   |-- interface.go         [PLANNED] Storage interface for swap later
|
|-- docs/
|   |-- INDEX.md                 [EXISTS] This index
|   |-- PROJECT_MAP.md           [EXISTS] This file
|   |-- ARCHITECTURE.md          [EXISTS] Package layout
|   |-- ROUTE_MAP.md             [EXISTS] Command routes
|   |-- TEMPLATES.md             [EXISTS] File templates
|   |-- GLOSSARY.md              [EXISTS] Terms and jargon
|   |-- roadmap/
|   |   |-- overview.md          [EXISTS] Project purpose
|   |   |-- mvp-feature-plan.md   [EXISTS] MVP flows
|   |   |-- architecture-notes.md [EXISTS] Go layout notes
|   |   |-- infrastructure.md     [EXISTS] Full infrastructure design (types, naming, guild lifecycle, errors)
|   |   |-- permissions-and-safety-notes.md [EXISTS] Safety
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
| Planning and design docs | `docs/roadmap/` |
| Agent navigation docs | `docs/` (this folder) |

## What Exists vs Planned

| Area | Status |
|------|--------|
| `cmd/bot/main.go` | Exists - connects to Discord, no commands yet |
| `internal/` | Not created - all packages planned |
| Slash commands | Not registered |
| Storage | Not implemented |
| Handlers | Not implemented |
| `docs/` | Exists - agent docs and roadmap |

## Key Files to Know

- **main.go** - Thin entry point. Load config, init session, wire handlers, graceful shutdown.
- **registry.go** (planned) - Command definitions and startup sync: fetch existing, diff against desired, bulk overwrite only when changed. Supports `--guild` (dev) or `--global` (prod).
- **compose.go** (planned) - Main user-facing command group for compose/create/set/propose/post.
- **handlers/** - Shared logic used by multiple commands. Keep commands thin.

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
