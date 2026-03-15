## Architecture Notes - GuildMessageProxy

This document is a stub for capturing architectural decisions as the project grows.
It focuses on how the Go code is organized and how the bot interacts with Discord.

---

### 1. Entry Point

- Current entry point: `cmd/bot/main.go`.
- Responsibilities of `main` (target state):
  - Load configuration (environment variables, flags).
  - Initialize the Discord session (token, intents).
  - Wire up command and event handlers.
  - Start the bot and manage graceful shutdown.

As features expand, `main` should remain thin and delegate real work to internal packages.

---

### 2. Planned Package Layout (Initial Sketch)

This is an initial idea; names and boundaries can be adjusted as design becomes clearer.

- `internal/commands`
  - Command-first layout where each **top-level category and its subcommands** keep definition and execution together.
  - Suggested shape:
    - `internal/commands/compose.go` contains:
      - Definition of the `compose` command group and its subcommands (for example: `create`, `set`, `propose`, `post`).
      - Handler/execute functions for those subcommands, with access to shared helpers from `internal/handlers`.
      - Any command-specific validation and option parsing.
    - Other command groups (for example, `admin`, `config`) get their own files with a similar pattern.
    - `internal/commands/registry.go` holds command definitions and sync logic. On startup: fetch existing commands, compare with desired definitions, bulk overwrite only when different. No separate registration binary.
  - Keep shared helpers out of here unless they are truly command-specific.
- `internal/handlers`
  - Reusable building blocks used by multiple commands (this is the modular part, without naming it after a single command).
  - Examples of what would live here:
    - Rendering a preview message payload.
    - Posting or updating a proxied message.
    - Permission checks (who can create, set, propose, or post).
    - Input validation shared across subcommands.
  - This package is where the "compose/preview/post" steps become reusable functions, independent of how the commands are structured.
- `internal/storage`
  - Persistence for:
    - Proxy message metadata (guild, channel, message ID, owner, flags).
    - Minimal per-guild configuration needed for MVP.
  - Start with an in-memory implementation; design interfaces so storage can be swapped for a database later.

Discord integration:

- Proxy messages are posted via channel webhooks (custom avatar and username per message). Handlers create/use webhooks per channel as needed.
- For now, avoid a dedicated "Discord API wrapper" package.
  - Most `discordgo` calls will naturally live in the command handlers or `internal/handlers`.
  - If common patterns emerge (permission error translation, message building helpers, consistent response helpers), introduce a small, narrowly named helper package later (for example `internal/discordutil`) instead of a broad `discordapi`.

This layout is meant to keep user-facing entry points (`commands`) close to their execution logic, while keeping reusable logic in `handlers` and persistence concerns in `storage`.

---

### 3. Infrastructure (Pre-MVP)

**This infrastructure must be developed before main features.** It covers event handlers, interaction type systems, and error handling that make the frontend user experience work. All of it should be in place and functional before building compose, post, edit, etc.

- **internal/events package**
  - `interaction_create.go` - Receives all interaction types (slash, buttons, select menus, modals, message/user context commands). Routes to the correct definition and execution based on type and ID.
  - `guild_create.go` - Stores guild metadata and per-guild config when the bot joins a guild.
  - `guild_delete.go` - Removes or soft-deletes guild config and proxy metadata when the bot leaves.
  - `ready.go` - Optional: bot startup confirmation, log ready state.
  - `error.go` - Handles errors from REST API responses and gateway close codes (see Error handling below).

- **Interaction type system**
  - Custom types: `TSlashCommand`, `TButton`, `TSelectMenu`, `TModalSubmit`, `TMessageCommand`, `TUserCommand`.
  - Const lists for each type. ID conventions: `button_<context>_<action>`, `select_<context>_<action>`, `modal_<context>_<action>`.
  - Maps route types to definitions: `MCommandDefinitions`, `MButtonDefinitions`, `MSelectMenuDefinitions`, `MModalSubmitDefinitions`, etc.
  - See [docs/roadmap/infrastructure.md](./infrastructure.md) for the full design.

- **Guild lifecycle and storage**
  - GuildCreate: Store guild metadata (id, name), per-guild config (allowed roles, default channel, logging channel). Use upsert; GuildCreate can fire on re-availability.
  - GuildDelete: Remove or soft-delete guild config and proxy metadata. Document policy: delete, soft-delete, or retention. Orphaned messages fail on edit; handlers treat unknown guild/404 appropriately.

- **Error handling**
  - Discord does NOT send a dedicated gateway "Error" event for REST failures. Errors come from: (a) REST API responses (HTTP + JSON), (b) Gateway close codes, (c) Gateway opcodes.
  - Categorization: Transient (429, 502 - retry), Permanent auth (40001 - no retry), Permanent resource (10003, 10008 - clear user message), Validation (50035 - field-specific).
  - Flow: log to terminal, inform user, optionally send formatted error embed to logging channel.

The infrastructure is designed for extensibility so post-MVP features (additional buttons, context commands, select menus, modals, voting, etc.) are supported without major refactors.

---

### 4. Discord Intents and Events (MVP)

- Intents:
  - Currently: `discordgo.IntentsGuildMessages`.
  - As slash commands and interactions are added, additional intents or application command handling may be needed.
- Event handling:
  - Prioritize slash commands and interaction-based flows for predictability and better UX.
  - Keep message-based commands (if any) minimal and clearly separated.

### 5. Command Registration (Startup Sync)

Commands sync on every bot startup rather than via a separate registration binary:

1. Fetch existing commands (GET) for the target scope (guild or global).
2. Compare desired definitions with fetched (name, description, options). Ignore Discord-only fields (id, version).
3. Only if different: bulk overwrite (PUT) with the full desired list.

Use `--guild=<id>` for development (instant propagation) or `--global` for production. Optional `--no-sync` skips sync when commands are known-good.

As new features are specified in `docs/roadmap`, this file should be updated with concrete interfaces, key structs, and important invariants.

---

### 6. Post-MVP Infrastructure

Planned extensions beyond MVP:

- **Voting**: Reaction handlers (`MESSAGE_REACTION_ADD`), approval buttons (`TButton`), state machine (Draft -> Pending -> Approved/Rejected), vote storage.
- **Admin/Config**: Uses `TSlashCommand`; no new interaction types.
- **Collaborative editing**: Edit button on messages, permission middleware, optional message context command "Edit this message".

See [docs/roadmap/infrastructure.md](./infrastructure.md#post-mvp-infrastructure) for details.

