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
    - `internal/commands/registry.go` (or `register.go`) wires all command group definitions into a single registration function.
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

Discord integration helpers:

- For now, avoid a dedicated "Discord API wrapper" package.
  - Most `discordgo` calls will naturally live in the command handlers or `internal/handlers`.
  - If common patterns emerge (permission error translation, message building helpers, consistent response helpers), introduce a small, narrowly named helper package later (for example `internal/discordutil`) instead of a broad `discordapi`.

This layout is meant to keep user-facing entry points (`commands`) close to their execution logic, while keeping reusable logic in `handlers` and persistence concerns in `storage`.

---

### 3. Discord Intents and Events (MVP)

- Intents:
  - Currently: `discordgo.IntentsGuildMessages`.
  - As slash commands and interactions are added, additional intents or application command handling may be needed.
- Event handling:
  - Prioritize slash commands and interaction-based flows for predictability and better UX.
  - Keep message-based commands (if any) minimal and clearly separated.

As new features are specified in `docs/roadmap`, this file should be updated with concrete interfaces, key structs, and important invariants.

