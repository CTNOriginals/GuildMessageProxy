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

Event handlers that receive Discord gateway events and route them to the correct logic. All event handlers live in this package. Designed for extensibility so post-MVP features (buttons, context commands, etc.) are supported.

| File | Responsibility |
|------|----------------|
| `interaction_create.go` | Receives all interaction types (slash commands, buttons, message context commands). Routes by interaction type and custom_id/command name to the correct definition and execution. |
| `guild_create.go` | Fired when the bot joins a guild. Updates the database (e.g. register guild, default config). |
| `guild_delete.go` | Fired when the bot leaves a guild. Updates the database (e.g. remove guild data). |
| `error.go` | Handles Discord API error events. Logs to terminal, informs the user who triggered it, optionally sends a formatted error embed to a logging channel. |

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

- `type TSlashCommand string` - value is the command name (e.g. `"compose-create"`)
- `type TButton string` - value is the button `custom_id` (e.g. `"button_compose-create_post"`)
- Const lists for each interaction type (slash commands, buttons, etc.)
- Convention for interaction IDs: buttons use `"button_<context>_<action>"` (e.g. `button_compose-create_post`)

Maps route interactions to definitions:

```go
type MCommandDefinitions map[TSlashCommand]SCommandDef
var CommandDefinitions MCommandDefinitions = MCommandDefinitions{...}
```

The bot identifies an interaction by its type and looks up the definition in the appropriate map. See [GLOSSARY.md](./GLOSSARY.md) for term definitions.

## Error Handling (Discord API Error Event)

Discord emits an error event when API errors occur. The `internal/events` package handles it by:

1. **Logging** - Write the error to the terminal
2. **User feedback** - Inform the user who triggered the error that something went wrong
3. **Logging channel** (optional) - Send a formatted error embed to a configured channel (polish feature)

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
