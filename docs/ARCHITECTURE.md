# Architecture - Package Layout and Conventions

How the Go code is organized and how the bot interacts with Discord.

## Entry Point

- **File**: `cmd/bot/main.go`
- **Responsibilities** (target state):
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

## Discord Integration

- Use `discordgo` directly. No dedicated "Discord API wrapper" package.
- Most discordgo calls live in command handlers or `internal/handlers`.
- **Posting**: Proxy messages are sent via channel webhooks, which support custom avatar and username per message. Handlers create/use webhooks per channel as needed.
- If common patterns emerge (permission error translation, message building helpers), introduce `internal/discordutil` later. Avoid broad `discordapi`.

## Intents and Events (MVP)

- **Intents**: `discordgo.IntentsGuildMessages` (current). May need more for slash commands and interactions.
- **Events**: Prioritize slash commands and interaction-based flows. Keep message-based commands minimal.

## Dependencies

- `github.com/bwmarrin/discordgo` - Discord API
- `github.com/joho/godotenv` - Env loading
- See `go.mod` for versions.

## Code Conventions (Project Rules)

- Use `var` over `:=` when defining new variables (except in `for range`).
- No m-dashes or special unicode characters in text.
