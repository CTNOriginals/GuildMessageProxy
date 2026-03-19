# Architecture - Package Layout and Conventions

How the Go code is organized and how the bot interacts with Discord.

## Entry Point

- **File**: `cmd/bot/main.go`
- **Current state**:
  - Loads `.env` via godotenv
  - Token via `-t` flag or `TOKEN` env var
  - CLI flags implemented: `--guild=<id>` for dev, `--global` for prod, `--no-sync` to skip sync
  - Creates discordgo session with `IntentsGuildMessages | IntentsGuilds`
  - Event handlers wired: Ready, InteractionCreate, GuildCreate, GuildDelete
  - Command sync on startup with diff detection
  - Graceful shutdown with structured runtime logging via `internal/logging`

`main` should stay thin. Delegate real work to `internal/` packages.

**Implementation cross-references**:
- Command sync logic: `internal/commands/registry.go`
- Event handlers: `internal/events/*.go`
- Storage: `internal/storage/*.go`

## Package Layout

### internal/commands

Command-first layout. Each top-level category and its subcommands live together.

| File | Contents |
|------|----------|
| `compose.go` | [EXISTS] `/compose` group + subcommands (draft, send, edit). Handler/execute functions. Command-specific validation. Aliases create, post, propose maintained for backward compatibility. |
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

#### Permission Checking

The `CanUseCompose` function in `internal/handlers/permissions.go` validates user permissions before allowing compose command execution. Checks are performed in order:

1. **Channel access** - Bot can view the target channel
2. **Permission retrieval** - Bot can fetch user's channel permissions
3. **Send Messages** - User has discordgo.PermissionSendMessages
4. **Allowed roles** - User has at least one configured allowed role (if guild has role restrictions)
5. **Channel restrictions** - Channel is not in the restricted list
6. **Channel whitelist** - Channel is in the allowed list (if guild has whitelist configured)

Failed checks return a `PermissionResult` with a descriptive error message. See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md#permission-error-reference) for error message explanations.

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

- SQLite storage for production use.
- In-memory implementation available for testing (configurable via `STORAGE_TYPE` env var).

## Draft Management

Drafts are temporary message compositions stored in memory before posting.

- **Service**: `DraftService` in `internal/commands/draft_service.go`
- **Thread Safety**: All operations protected by `sync.RWMutex`
- **TTL**: Drafts expire after 24 hours (`DraftTTL`)
- **Cleanup**: Background goroutine runs hourly to clean expired drafts
- **Key Pattern**: `userID:guildID` ensures one draft per user per guild

The service provides:
- `Get(userID, guildID)` - Retrieve existing draft (read lock)
- `Save(draft)` - Store new or updated draft (write lock)
- `Delete(userID, guildID)` - Remove draft (write lock)
- `CleanupExpired() int` - Remove expired drafts, return count

### internal/logging

Structured logging package providing consistent, configurable log output across the application.

| Feature | Description |
|---------|-------------|
| Log levels | Fatal, Error, Warn, Info, Debug (configurable via `LOG_LEVEL` env var) |
| Structured fields | Key-value pairs for consistent context (guild_id, channel_id, user_id, error, etc.) |
| Output channels | stdout for Info/Debug, stderr for Error/Warn/Fatal |
| Format | `[LEVEL] [YY-MM-DD HH:MM:SS] message` with indented fields |
| Context helpers | `WithContext()` for standard Discord context fields |

Used by all event handlers and `main.go` to replace standard `log` package usage. See `internal/logging/logging.go` for implementation and `internal/logging/levels.go` for level definitions.

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

The bot identifies an interaction by its type and looks up the definition in the appropriate map. See [ROUTE_MAP.md](./ROUTE_MAP.md#interaction-routing) for routing details and [GLOSSARY.md](./GLOSSARY.md) for term definitions.

## Error Handling

Discord does **not** send a dedicated gateway "Error" event for REST failures. Errors come from:

- **(a) REST API responses** - HTTP status + JSON body (e.g. 429 rate limit, 50035 validation)
- **(b) Gateway close codes** - Connection-level (e.g. 4001 reconnect, 4004 invalid token)
- **(c) Gateway opcodes** - Event payloads that indicate failure

The `internal/events` package handles errors by:

1. **Logging** - Write errors via `internal/logging` package with structured context (guild_id, channel_id, user_id, error details)
2. **User feedback** - Inform the user who triggered it that something went wrong
3. **Logging channel** (optional) - Send a formatted error embed to a configured channel (polish feature)

**Structured logging approach**:
- Errors are logged with appropriate severity level (Error for failures, Warn for recoverable issues)
- Context fields attach Discord IDs and error details for traceability
- See `internal/events/error.go` for error categorization and logging integration

**Defensive coding in error handlers**:
Error handling code may be called when interaction data is incomplete. Use nil checks before accessing nested fields:

```go
var userID string
if i.Member != nil && i.Member.User != nil {
    userID = i.Member.User.ID
}
```

This pattern is used in `respondWithError()` and other error logging contexts where the interaction structure may not be fully populated.

**Error categorization**:

| Category | Example Codes | Handling |
|----------|---------------|----------|
| Transient | 429 (rate limit), 502 (server error), 503 (service unavailable) | Retry with backoff |
| Permanent auth | 40001 (unauthorized), 40004 (disallowed intent) | No retry; log and notify |
| Permanent resource | 10003 (unknown channel), 10008 (unknown message), 10013 (unknown user) | Clear user message; handlers treat unknown guild/404 appropriately |
| Validation | 50035 (invalid form body), 50016 (too many attachments) | Field-specific feedback to user |

See [internal/events/error.go](../internal/events/error.go) for implementation details and [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for common errors.

## Guild Lifecycle and Storage

- **GuildCreate**: Store guild metadata (id, name), per-guild config (allowed roles, default channel, logging channel). Use upsert; GuildCreate can fire on re-availability (e.g. bot comes back online).
- **GuildDelete**: Remove or soft-delete guild config and proxy metadata. Document policy: either delete, soft-delete, or retention. Orphaned messages fail on edit; handlers treat unknown guild/404 appropriately.
- **Cleanup policy**: Choose one: hard delete on leave, soft-delete with retention window, or retain for audit. Document in `internal/storage`.

## Discord Integration

- Use `discordgo` directly. No dedicated "Discord API wrapper" package.
- Most discordgo calls live in command handlers or `internal/handlers`.
- **Posting**: Proxy messages are sent via channel webhooks, which support custom avatar and username per message. Handlers create/use webhooks per channel as needed.
- If common patterns emerge (permission error translation, message building helpers), introduce `internal/discordutil` later. Avoid broad `discordapi`.

## Intents and Events (MVP)

- **Intents**: `discordgo.IntentsGuildMessages | discordgo.IntentsGuilds`
- **Events**: Wired in `internal/events`. Prioritize slash commands and interaction-based flows. Keep message-based commands minimal. See `internal/events` for InteractionCreate, GuildCreate, GuildDelete, and Ready handlers.

## Dependencies

- `github.com/bwmarrin/discordgo` - Discord API
- `github.com/joho/godotenv` - Env loading
- See `go.mod` for versions.

## Concurrency and Thread Safety

The bot handles concurrent Discord interactions. Shared state must be protected.

### In-Memory Storage with Mutex Protection

Maps used by multiple handlers require explicit synchronization.

**Pattern used in `internal/commands/compose.go`:**

```go
var (
    draftStore   = make(map[string]*Draft)
    draftStoreMu sync.RWMutex
)
```

**Access patterns:**

| Operation | Mutex Method |
|-----------|--------------|
| Read-only | `RLock()` / `RUnlock()` |
| Write (add/update/delete) | `Lock()` / `Unlock()` |

**Example:**

```go
// Write - exclusive lock
draftStoreMu.Lock()
draftStore[key] = &draft
draftStoreMu.Unlock()

// Read - shared lock
draftStoreMu.RLock()
draft, exists := draftStore[key]
draftStoreMu.RUnlock()
```

**Scope notes:**

- Keep critical sections small - release lock before calling Discord API
- Use deferred unlocks only when function scope matches critical section
- For multi-step operations (read-then-write), hold Lock throughout to prevent race conditions

## Code Conventions (Project Rules)

- Use `var` over `:=` when defining new variables (except in `for range`).
- No m-dashes or special unicode characters in text.
