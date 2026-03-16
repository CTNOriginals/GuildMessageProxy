# Infrastructure Implementation

Build the skeleton infrastructure for GuildMessageProxy: event system, type system, storage, and main wiring. Main features (compose, post, edit) are out of scope - implement only the infrastructure that enables them.

This infrastructure supports all future bot features: slash commands, buttons, select menus, modals, context commands, guild lifecycle handling, and error feedback.

## Reference Documentation

Use these as the source of truth:

- **docs/roadmap/infrastructure.md** - Interaction types, naming conventions, definition structs, guild lifecycle, error handling
- **docs/ARCHITECTURE.md** - Package layout, conventions, entry points, command sync
- **docs/PROJECT_MAP.md** - Directory structure, existing vs planned files
- **docs/TEMPLATES.md** - File templates and patterns
- **docs/INDEX.md** - Documentation navigation

The documentation is a baseline, not a rigid specification. If implementation reveals better patterns, adjust proactively and document deviations.

## Phase 1: Core Structure

### internal/commands/types.go

Create the interaction type system:

- Types: `TSlashCommand`, `TButton`, `TSelectMenu`, `TModalSubmit`, `TMessageCommand`, `TUserCommand` (each as `type X string`)
- Const blocks for each type. Start with empty or placeholder consts (one placeholder slash command to verify sync)
- Naming conventions from docs/roadmap/infrastructure.md:
  - Slash/context: `context-action` (e.g. `compose-create`)
  - Buttons: `button_<context>_<action>`
  - Select menus: `select_<context>_<action>`
  - Modals: `modal_<context>_<action>`

### internal/storage/interface.go

Define the Store interface:

- Guild metadata (id, name)
- Proxy metadata (guild, channel, message ID, owner, flags) - use minimal struct for MVP
- Per-guild config (allowed roles, default channel, logging channel) - use placeholder struct

Methods:
- `SaveGuild(guildID, name string)` - upsert guild
- `GetGuild(guildID string)` - get guild metadata
- `DeleteGuild(guildID string)` - remove or soft-delete guild data (document the policy)

### internal/storage/memory.go

In-memory Store implementation using maps keyed by guild ID. Design the interface for easy database swapping later.

### internal/commands/registry.go

Command sync logic:

- Fetch existing commands (GET) for target scope (guild or global)
- Compare desired vs fetched (name, description, options). Ignore Discord-only fields (id, version)
- Bulk overwrite (PUT) only when different
- Start with empty command list or one placeholder to verify sync
- Support `--guild=<id>` (dev, instant propagation) and `--global` (prod, up to 1 hour propagation)
- Optional `--no-sync` flag to skip sync for faster restarts

## Phase 2: Event Handlers

### internal/events/interaction_create.go

Route by interaction type:

- Slash commands: `ApplicationCommandData().Name` -> `TSlashCommand` -> `MCommandDefinitions`
- Buttons: `MessageComponentData().CustomID` -> `TButton` -> `MButtonDefinitions`
- Select menus: `MessageComponentData().CustomID` -> `TSelectMenu` -> `MSelectMenuDefinitions`
- Modal submits: `ModalSubmitData().CustomID` -> `TModalSubmit` -> `MModalSubmitDefinitions`
- Message context: `ApplicationCommandData().Name` -> `TMessageCommand` -> `MMessageCommandDefinitions`
- User context: `ApplicationCommandData().Name` -> `TUserCommand` -> `MUserCommandDefinitions`

Maps can be empty initially. If handler not found, log and optionally respond ephemerally.

### internal/events/guild_create.go

On bot join or re-connect:
- Store guild metadata via storage interface
- Store default per-guild config
- Use upsert (GuildCreate fires on re-availability)

### internal/events/guild_delete.go

On bot leaving a guild:
- Remove or soft-delete guild data via storage interface
- Document the policy in the storage package

### internal/events/ready.go

Log ready state on startup.

### internal/events/error.go

Error handling helper:

- `LogError(err error, context string)` - log to terminal
- `RespondToUser(s *discordgo.Session, i *discordgo.InteractionCreate, msg string)` - ephemeral error message
- Optional: categorize by Discord error code for retry vs no-retry
- No logging channel required for MVP; structure for later addition

## Phase 3: Main Wiring

### cmd/bot/main.go

Entry point:

- Load config: env vars, flags `-t` (token), `--guild=<id>`, `--global`, `--no-sync`
- Initialize storage (memory implementation)
- Create session with intents: `IntentsGuildMessages`, `IntentsGuilds`, and any others needed for interactions
- Wire event handlers via `session.AddHandler`:
  - `events.HandleInteractionCreate`
  - `events.HandleGuildCreate`
  - `events.HandleGuildDelete`
  - `events.HandleReady`
- Sync commands on startup: `registry.SyncCommands(session, guildID)`
- Graceful shutdown on SIGINT/SIGTERM

## Definition Structs and Maps

In the commands package:

- `SCommandDef` - `Definition *discordgo.ApplicationCommand`, `Execute func(...)`, `Autocomplete func(...)` (optional)
- `MCommandDefinitions map[TSlashCommand]SCommandDef`
- `SButtonDef` - `Execute func(...)`
- `MButtonDefinitions map[TButton]SButtonDef`
- `SSelectMenuDef`, `MSelectMenuDefinitions` (empty for MVP)
- `SModalSubmitDef`, `MModalSubmitDefinitions` (empty for MVP)
- `MMessageCommandDefinitions`, `MUserCommandDefinitions` (empty for MVP)

InteractionCreate looks up in the appropriate map and calls `Execute`.

## Guild Lifecycle

- **GuildCreate**: Upsert guild (id, name) and default config
- **GuildDelete**: Delete or soft-delete guild data (document policy)
- Storage interface: SaveGuild, GetGuild, DeleteGuild

## Error Handling

- Helper in events/error.go: LogError, RespondToUser (ephemeral)
- Optional: categorize by Discord error code
- No logging channel for MVP; structure for later addition

## Verification

After implementation, verify:

1. Bot starts, connects, logs Ready
2. Placeholder command syncs (use `--guild` for dev)
3. GuildCreate/GuildDelete updates storage
4. InteractionCreate receives events; logs and optionally responds when handler not found

## Out of Scope

- Compose commands, handlers, post/preview logic
- Database (in-memory storage only)
- Logging channel for errors
- Post-MVP features (voting, reactions)

---

# Instruction

**Task**: Implement the infrastructure above. It must be functional and testable before main features are built.

**Planning and adaptation**:
- Continuously verify implementation aligns with the design
- If documentation gaps or better patterns emerge, research and adjust proactively
- Prefer coherence and maintainability over strict adherence

**Delegation**:
- Delegate to **developer** subagent for implementation
- Delegate to **tester** subagent for tests
- Delegate to **verifier** subagent for validation
- Act as **project-leader** - plan, delegate, oversee, verify, commit

**Process**:
- Read `.cursor/agents/INDEX.md` for subagent usage
- Apply all applicable `.cursor/rules/`
- Use `docs/INDEX.md` to navigate documentation
- Follow `docs/roadmap/infrastructure.md`, `docs/ARCHITECTURE.md`, `docs/PROJECT_MAP.md`, `docs/TEMPLATES.md` as source of truth
