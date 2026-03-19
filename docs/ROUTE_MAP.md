# Route Map - Commands and Flows

Where commands are defined, how they flow, and how handlers are wired.

## Command Structure

Commands use subcommands under a top-level group. The main category is `/compose`.

### Implemented Commands

| Route | Purpose |
|-------|---------|
| `/compose create` | Start a new draft with initial content |
| `/compose propose` | Submit a proposed change to an existing proxied message |
| `/compose post` | Confirm and post the current draft |

All MVP compose commands are fully implemented.

## Flow: Compose -> Preview -> Post (Implemented)

```
User: /compose create (content)
  -> handlers: validate content
  -> handlers: render preview
  -> Bot: ephemeral response with preview + [Post] [Cancel] buttons

User: clicks Post (or /compose post)
  -> handlers: post to target channel
  -> storage: save metadata (guild, channel, msg ID, owner, flags)

User: clicks Cancel
  -> Bot: dismiss preview, discard draft
```

**Status**: Fully implemented in `internal/handlers/`.

## Flow: Basic Edit (Implemented)

```
User: /compose propose (target message, new content)
  -> handlers: verify requester is original owner (MVP)
  -> handlers: render edited preview
  -> Bot: ephemeral preview + [Apply] [Cancel] buttons

User: clicks Apply
  -> handlers: edit original proxied message
  -> storage: update metadata (last edited by, last edited at)
```

**Status**: Fully implemented in `internal/handlers/edit.go`.

## Handler Wiring (Implemented)

| Flow Step | Handler | Called From |
|-----------|---------|-------------|
| Validate content | `handlers.ValidateContent` | compose create, compose propose |
| Validate permissions | `handlers.CanUseCompose` | all compose subcommands |
| Render preview | `handlers.RenderPreviewResponse` | compose create, compose propose |
| Post message | `handlers.PostProxiedMessage` | compose post, button handler |
| Edit message | `handlers.EditProxiedMessage` | edit apply button handler |
| Check edit permission | `handlers.ValidateEditPermission` | compose propose |

## Interaction Routing

All interactions flow through `internal/events/interaction_create.go`. The handler receives every interaction type (slash commands, buttons, message context commands, etc.) and routes to the correct definition and execution.

### Type System

All interaction types and their routing:

| Interaction Type | Project Type | Identification | Naming Convention |
|------------------|--------------|----------------|-------------------|
| Slash command | `TSlashCommand` | `data.name` | `context-action` (e.g. `compose-create`) |
| Message context menu | `TMessageCommand` | `data.name` | `context-action` |
| User context menu | `TUserCommand` | `data.name` | `context-action` |
| Button | `TButton` | `data.custom_id` | `button_<context>_<action>` |
| Select menu | `TSelectMenu` | `data.custom_id` | `select_<context>_<action>` |
| Modal submit | `TModalSubmit` | `data.custom_id` | `modal_<context>_<action>` |
| Autocomplete | (none) | command + option | Handled in slash handler or shared handler |

Const lists define all valid values per type. Maps route types to definitions (e.g. `MCommandDefinitions`, `MButtonDefinitions`, `MSelectMenuDefinitions`, `MModalSubmitDefinitions`).

The bot looks up the interaction by its type in the appropriate map and invokes the associated definition/execute logic.

### Flow

1. **Slash commands**: Synced on startup via `registry.SyncCommands(session, guildID)`. Fetches existing, diffs against desired definitions, bulk overwrites only when changed. Scope: `--guild=<id>` (dev) or `--global` (prod). InteractionCreate receives the event, maps `ApplicationCommandData().Name` to `TSlashCommand`, looks up in `CommandDefinitions`, and runs the execute function.
2. **Button clicks** (Post, Cancel, Apply): InteractionCreate receives the event, reads `MessageComponentData().CustomID`, maps to `TButton`, looks up in `MButtonDefinitions`, and runs the execute function.
3. **Select menus**: InteractionCreate reads `MessageComponentData().CustomID`, maps to `TSelectMenu`, looks up in `MSelectMenuDefinitions`, runs execute. Same pattern for string/user/role/channel/mentionable selects.
4. **Modal submits**: InteractionCreate reads `ModalSubmitData().CustomID`, maps to `TModalSubmit`, looks up in `MModalSubmitDefinitions`, runs execute.
5. **Message/User context commands**: InteractionCreate reads `ApplicationCommandData().Name`, maps to `TMessageCommand` or `TUserCommand`, looks up in the appropriate map, runs execute.
6. **Autocomplete**: Handled within slash command handler or shared handler keyed by command+option.

See [ARCHITECTURE.md](./ARCHITECTURE.md#interaction-type-system) for the full type system. See [PROJECT_MAP.md](./PROJECT_MAP.md) for event handler file locations.

## MVP Restrictions

- Only the original requester can edit.
- Posting uses webhooks (custom avatar/username per message); MVP may restrict identity options.
- No voting/approval workflows yet.

## Future Routes (Out of Scope for MVP)

- `/admin` or `/config` - guild configuration (uses `TSlashCommand`; no new types)
- Voting/approval flows - see [Post-MVP Infrastructure](#post-mvp-infrastructure)
- `/message edit` - alternative to `/compose propose` for edits (message context command "Edit this message")

## Post-MVP Infrastructure

Planned extensions (see [PROJECT_STATUS.md](./PROJECT_STATUS.md) for backlog):

- **Voting**: Reaction handlers (`MESSAGE_REACTION_ADD`), approval buttons (`TButton`), state machine (Draft -> Pending -> Approved/Rejected), vote storage
- **Admin/Config**: Uses `TSlashCommand`; no new interaction types
- **Collaborative editing**: Edit button on messages, permission middleware, optional message context command "Edit this message"

## See Also

- [DEPLOYMENT.md](./DEPLOYMENT.md) - Command registration details and bot setup
- [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) - Debugging flows and common issues
