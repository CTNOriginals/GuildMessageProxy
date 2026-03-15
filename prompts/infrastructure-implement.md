# Infrastructure Implementation

The infrastructure for GuildMessageProxy is fully documented but not yet implemented in code. This task is to build the skeleton: event system, type system, storage, and main wiring. No main features (compose, post, edit) yet - just the infrastructure that must be functional and testable before those features are built.

The infrastructure enables all future bot features: slash commands, buttons, select menus, modals, context commands, guild lifecycle handling, and error feedback. It must be in place and working before any compose/post/edit logic is added.

## Reference Documentation

All design decisions, types, and flows are documented. Use these as the source of truth:

- **docs/roadmap/infrastructure.md** - Full infrastructure design (interaction types, naming conventions, definition structs, guild lifecycle, error handling)
- **docs/ARCHITECTURE.md** - Package layout, conventions, entry points, command sync
- **docs/PROJECT_MAP.md** - Directory structure, where files live, what exists vs planned
- **docs/TEMPLATES.md** - File templates and patterns for adding new features
- **docs/INDEX.md** - Navigation index for all docs

The documentation above is a **strong baseline** to follow, but it is not entirely set in stone. If implementation reveals outcomes the docs did not anticipate, proactively research a more optimal solution and adjust. Keep the design coherent and document any deviations.

## Phase 1: Core Structure

### internal/commands/types.go

Create the interaction type system:

- `TSlashCommand`, `TButton`, `TSelectMenu`, `TModalSubmit`, `TMessageCommand`, `TUserCommand` - each as `type X string`
- Const blocks for each type. Can start with empty or placeholder consts (e.g. one placeholder slash command to verify sync works)
- Naming conventions from docs/roadmap/infrastructure.md:
  - Slash/context: `context-action` (e.g. `compose-create`)
  - Buttons: `button_<context>_<action>`
  - Select menus: `select_<context>_<action>`
  - Modals: `modal_<context>_<action>`

### internal/storage/interface.go

Define the Store interface for:

- Guild metadata (id, name)
- Proxy metadata (guild, channel, message ID, owner, flags) - struct can be minimal for MVP
- Per-guild config (allowed roles, default channel, logging channel - can be placeholder struct)

Required methods (or equivalent):

- `SaveGuild(guildID, name string)` - upsert guild
- `GetGuild(guildID string)` - get guild metadata
- `DeleteGuild(guildID string)` - remove or soft-delete guild data (document the chosen policy)

### internal/storage/memory.go

In-memory implementation of the Store interface. Use maps keyed by guild ID. This is the only storage implementation for MVP; design the interface so it can be swapped for a database later.

### internal/commands/registry.go

Command sync logic:

- Fetch existing commands (GET) for the target scope (guild or global)
- Compare desired definitions with fetched (name, description, options). Ignore Discord-only fields (id, version)
- Bulk overwrite (PUT) only when different
- Can start with empty command list or one placeholder command to verify sync works
- Support `--guild=<id>` (dev, instant propagation) and `--global` (prod, up to 1 hour propagation)
- Optional `--no-sync` flag to skip sync for faster restarts when commands are known-good

## Phase 2: Event Handlers

### internal/events/interaction_create.go

Route by interaction type to the appropriate definition map:

- Slash commands: `ApplicationCommandData().Name` -> `TSlashCommand` -> `MCommandDefinitions`
- Buttons: `MessageComponentData().CustomID` -> `TButton` -> `MButtonDefinitions`
- Select menus: `MessageComponentData().CustomID` -> `TSelectMenu` -> `MSelectMenuDefinitions`
- Modal submits: `ModalSubmitData().CustomID` -> `TModalSubmit` -> `MModalSubmitDefinitions`
- Message context: `ApplicationCommandData().Name` -> `TMessageCommand` -> `MMessageCommandDefinitions`
- User context: `ApplicationCommandData().Name` -> `TUserCommand` -> `MUserCommandDefinitions`

Maps can be empty initially. Routing logic must be in place. If a handler is not found, log and optionally respond ephemerally to the user.

### internal/events/guild_create.go

When the bot joins a guild (or re-connects):

- Store guild metadata (id, name) via storage interface
- Store default per-guild config
- Use upsert: GuildCreate can fire on re-availability, so overwrite rather than fail on duplicate

### internal/events/guild_delete.go

When the bot leaves a guild:

- Remove or soft-delete guild data via storage interface
- Document the chosen policy (delete, soft-delete, or retention) in the storage package

### internal/events/ready.go

Log ready state on startup. Optional but useful for verification.

### internal/events/error.go

Error handling helper:

- `LogError(err error, context string)` - log to terminal
- `RespondToUser(s *discordgo.Session, i *discordgo.InteractionCreate, msg string)` - ephemeral error message to user
- Optional: categorize by Discord error code for retry vs no-retry
- No logging channel required for MVP; structure so it can be added later
- Wrap REST calls or provide a helper; handle gateway close in main if needed

## Phase 3: Main Wiring

### cmd/bot/main.go

Update the entry point:

- Load config: env vars, flags `-t` (token), `--guild=<id>`, `--global`, `--no-sync`
- Initialize storage (memory implementation)
- Create session with required intents: `IntentsGuildMessages`, `IntentsGuilds` (for guild events), and any others needed for slash commands and interactions
- Wire event handlers via `session.AddHandler`:
  - `events.HandleInteractionCreate`
  - `events.HandleGuildCreate`
  - `events.HandleGuildDelete`
  - `events.HandleReady`
- Sync commands on startup: `registry.SyncCommands(session, guildID)` (or equivalent)
- Graceful shutdown on SIGINT/SIGTERM: log shutdown, close session cleanly

## Definition Structs and Maps

Define in the commands package (or split across types.go and registry.go as appropriate):

- `SCommandDef` - `Definition *discordgo.ApplicationCommand`, `Execute func(...)`, `Autocomplete func(...)` (optional)
- `MCommandDefinitions map[TSlashCommand]SCommandDef`
- `SButtonDef` - `Execute func(...)`
- `MButtonDefinitions map[TButton]SButtonDef`
- `SSelectMenuDef`, `MSelectMenuDefinitions` (can be empty for MVP)
- `SModalSubmitDef`, `MModalSubmitDefinitions` (can be empty for MVP)
- `MMessageCommandDefinitions`, `MUserCommandDefinitions` (can be empty for MVP)

InteractionCreate looks up in the appropriate map and calls `Execute`. If not found, log and optionally respond ephemerally.

## Guild Lifecycle

- **GuildCreate**: Upsert guild (id, name) and default config
- **GuildDelete**: Delete or soft-delete guild data. Document the chosen policy in storage (e.g. in interface.go or memory.go doc comments)
- Storage interface must support: SaveGuild, GetGuild, DeleteGuild (or equivalent names)

## Error Handling

- Helper in events/error.go: LogError, RespondToUser (ephemeral error message)
- Optional: categorize by Discord error code for retry vs no-retry
- No logging channel required for MVP; structure so it can be added later

## Verification

After implementation, the following should work:

1. Bot starts, connects, logs Ready
2. If a placeholder command is added, it syncs (use `--guild` for dev)
3. GuildCreate/GuildDelete update storage (verify with logs or a simple test)
4. InteractionCreate receives events; if no handler found, log and optionally respond

## Out of Scope for This Task

- Compose commands, handlers, post/preview logic
- Database (use in-memory storage only)
- Logging channel for errors
- Post-MVP features (voting, reactions)

---

# Instruction

**Task**: Implement the infrastructure as documented above. The infrastructure must be functional and testable before main features are built.

**Planning and adaptation**:
- Continuously plan and verify that implementation aligns with the intended design. Ensure each piece fits into place as planned.
- If you notice outcomes the initial documentation did not expect (e.g. discordgo quirks, structural friction, clearer patterns), proactively research a more optimal solution before committing to the original plan.
- The docs are a strong baseline, not rigid rules. Prefer coherence and maintainability over strict adherence when the two conflict.

**Delegation**:

- Delegate to the **developer** subagent for implementation
- Delegate to the **tester** subagent for tests (if applicable)
- Delegate to the **verifier** subagent for validation
- Act as **project-leader**: plan, delegate, oversee, verify, commit

**Process**:

- Read `.cursor/agents/INDEX.md` for when to use which subagent
- Apply all applicable `.cursor/rules/`
- Use `docs/INDEX.md` to navigate documentation
- Follow `docs/roadmap/infrastructure.md`, `docs/ARCHITECTURE.md`, `docs/PROJECT_MAP.md`, `docs/TEMPLATES.md` as the source of truth
