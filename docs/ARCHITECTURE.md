# Architecture - Package Layout and Conventions

How the Go code is organized and how the bot interacts with Discord.

## Entry Point

- **File**: `cmd/bot/main.go`
- **Current state**:
  - Loads `.env` via godotenv
  - Token via `-t` flag or `TOKEN` env var
  - Creates discordgo session with `IntentsGuildMessages`
  - Opens connection, waits for SIGINT/SIGTERM, then closes
  - No command sync, slash handlers, `--guild`/`--global` flags, or graceful shutdown logging yet
- **Target state** (planned):
  - Load configuration (env vars, flags)
  - Initialize Discord session (token, intents)
  - Sync commands on startup (fetch existing, diff, bulk overwrite if changed)
  - Wire up command and event handlers
  - Start the bot and manage graceful shutdown

`main` should stay thin. Delegate real work to `internal/` packages.

## Package Layout

### internal/commands

Command-first layout. Each top-level category and its subcommands live together.

| File | Contents |
|------|----------|
| `compose.go` | `/compose` group + subcommands (create, set, propose, post). Handler/execute functions. Command-specific validation. |
| `registry.go` | Command definitions and sync logic. On startup: fetch existing commands, compare with desired definitions, bulk overwrite only when different. Called from main. |
| `admin.go`, `config.go` | Future command groups. Same pattern as compose. |

- Keep shared helpers out unless truly command-specific.
- Commands call into `internal/handlers` for reusable logic.

### internal/handlers

Reusable building blocks used by multiple commands.

| Responsibility | Example |
|----------------|---------|
| Preview | Render preview message payload |
| Post/Update | Post or update a proxied message |
| Permissions | Who can create, set, propose, post |
| Validation | Input validation shared across subcommands |

This is where the compose/preview/post steps become reusable functions, independent of command structure.

### internal/events

Event handlers that receive Discord gateway events and route them to the correct logic. All event handlers live in this package. Designed for extensibility so post-MVP features (buttons, context commands, select menus, modals, etc.) are supported.

| File | Responsibility |
|------|----------------|
| `interaction_create.go` | Receives all interaction types (slash, buttons, select menus, modals, message/user context commands). Routes by interaction type and custom_id/command name to the correct definition and execution. |
| `guild_create.go` | Fired when the bot joins a guild. Stores guild metadata and per-guild config (see [Guild Lifecycle](#guild-lifecycle-and-storage)). |
| `guild_delete.go` | Fired when the bot leaves a guild. Removes or soft-deletes guild config and proxy metadata. |
| `ready.go` | Optional: bot startup confirmation, log ready state. |
| `error.go` | Handles errors from REST API responses and gateway close codes (see [Error Handling](#error-handling)). Logs to terminal, informs the user who triggered it, optionally sends a formatted error embed to a logging channel. |

**Handler registration**: `main` wires handlers via `session.AddHandler` for each event (InteractionCreate, GuildCreate, GuildDelete, Ready, etc.).

See [ROUTE_MAP.md](./ROUTE_MAP.md#interaction-routing) for how InteractionCreate routes to command definitions.

### internal/storage

Persistence for:

- Proxy message metadata (guild, channel, message ID, owner, flags)
- Minimal per-guild config for MVP

- Start with in-memory implementation.
- Design interfaces so storage can be swapped for a database later.

## Command Registration (Startup Sync)

Commands are synced on every bot startup:

1. **Fetch** existing commands (GET) for the target scope (guild or global).
2. **Compare** desired definitions with fetched (name, description, options, etc). Ignore Discord-only fields (id, version).
3. **Sync** only if different: bulk overwrite (PUT) with the full desired list.

This keeps commands in sync with code without a separate registration step. Use `--guild=<id>` for dev (instant propagation) or `--global` for prod (up to 1 hour propagation). Optional `--no-sync` skips sync for faster restarts when commands are known-good.

## Interaction Type System

Custom types route interactions to their definitions and execution:

| Project Type | Discord Type | Identification | Naming Convention |
|--------------|--------------|----------------|-------------------|
| `TSlashCommand` | Application Command (slash) | `data.name` | `context-action` (e.g. `compose-create`) |
| `TMessageCommand` | Message context menu | `data.name` | `context-action` |
| `TUserCommand` | User context menu | `data.name` | `context-action` |
| `TButton` | Button | `data.custom_id` | `button_<context>_<action>` |
| `TSelectMenu` | String/user/role/channel/mentionable select | `data.custom_id` | `select_<context>_<action>` |
| `TModalSubmit` | Modal form submission | `data.custom_id` | `modal_<context>_<action>` |
| (none) | Autocomplete | command + option | Handled within slash command or shared handler keyed by command+option |

**Autocomplete**: No dedicated type. Handled inside slash command handlers or a shared handler keyed by command name + option name.

**Definition structs**:

```go
type SCommandDef struct {
    Definition   *discordgo.ApplicationCommand
    Execute      func(s *discordgo.Session, i *discordgo.InteractionCreate)
    Autocomplete func(s *discordgo.Session, i *discordgo.InteractionCreate) // optional
}
type MCommandDefinitions map[TSlashCommand]SCommandDef

type SButtonDef struct {
    Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
type MButtonDefinitions map[TButton]SButtonDef

type SSelectMenuDef struct { Execute func(...) }
type MSelectMenuDefinitions map[TSelectMenu]SSelectMenuDef

type SModalSubmitDef struct { Execute func(...) }
type MModalSubmitDefinitions map[TModalSubmit]SModalSubmitDef
```

The bot identifies an interaction by its type and looks up the definition in the appropriate map. See [docs/roadmap/infrastructure.md](./roadmap/infrastructure.md) for the full design and [GLOSSARY.md](./GLOSSARY.md) for term definitions.

## Error Handling

Discord does **not** send a dedicated gateway "Error" event for REST failures. Errors come from:

- **(a) REST API responses** - HTTP status + JSON body (e.g. 429 rate limit, 50035 validation)
- **(b) Gateway close codes** - Connection-level (e.g. 4001 reconnect, 4004 invalid token)
- **(c) Gateway opcodes** - Event payloads that indicate failure

The `internal/events` package handles errors by:

1. **Logging** - Write the error to the terminal
2. **User feedback** - Inform the user who triggered it that something went wrong
3. **Logging channel** (optional) - Send a formatted error embed to a configured channel (polish feature)

**Error categorization**:

| Category | Example Codes | Handling |
|----------|---------------|----------|
| Transient | 429 (rate limit), 502 (server error) | Retry with backoff |
| Permanent auth | 40001 (unauthorized) | No retry; log and notify |
| Permanent resource | 10003 (unknown channel), 10008 (unknown message) | Clear user message; handlers treat unknown guild/404 appropriately |
| Validation | 50035 (invalid form body) | Field-specific feedback to user |

See [docs/roadmap/infrastructure.md](./roadmap/infrastructure.md#error-handling) for the full handling flow.

## Guild Lifecycle and Storage

- **GuildCreate**: Store guild metadata (id, name), per-guild config (allowed roles, default channel, logging channel). Use upsert; GuildCreate can fire on re-availability (e.g. bot comes back online).
- **GuildDelete**: Remove or soft-delete guild config and proxy metadata. Document policy: either delete, soft-delete, or retention. Orphaned messages fail on edit; handlers treat unknown guild/404 appropriately.
- **Cleanup policy**: Choose one: hard delete on leave, soft-delete with retention window, or retain for audit. Document in `internal/storage` and `docs/roadmap/infrastructure.md`.

## Discord Integration

- Use `discordgo` directly. No dedicated "Discord API wrapper" package.
- Most discordgo calls live in command handlers or `internal/handlers`.
- **Posting**: Proxy messages are sent via channel webhooks, which support custom avatar and username per message. Handlers create/use webhooks per channel as needed.
- If common patterns emerge (permission error translation, message building helpers), introduce `internal/discordutil` later. Avoid broad `discordapi`.

## Intents and Events (MVP)

- **Intents**: `discordgo.IntentsGuildMessages` (current). May need more for slash commands and interactions.
- **Events**: Wired in `internal/events`. Prioritize slash commands and interaction-based flows. Keep message-based commands minimal. See `internal/events` for InteractionCreate, GuildCreate, GuildDelete, and Error handlers.

## Dependencies

- `github.com/bwmarrin/discordgo` - Discord API
- `github.com/joho/godotenv` - Env loading
- See `go.mod` for versions.

## Code Conventions (Project Rules)

- Use `var` over `:=` when defining new variables (except in `for range`).
- No m-dashes or special unicode characters in text.
