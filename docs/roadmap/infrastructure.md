# Infrastructure Design - GuildMessageProxy

Single source of truth for infrastructure design: interaction types, naming conventions, definition structs, guild lifecycle, error handling, and post-MVP extensions.

See [ARCHITECTURE.md](../ARCHITECTURE.md) for package layout and [ROUTE_MAP.md](../ROUTE_MAP.md) for routing flows.

---

## 1. Interaction Type System

### Full Type Table

| Discord Type | Project Type | Identification | Naming Convention |
|--------------|--------------|----------------|-------------------|
| Application Command (slash) | `TSlashCommand` | `data.name` | `context-action` (e.g. `compose-create`) |
| Message context menu | `TMessageCommand` | `data.name` | `context-action` |
| User context menu | `TUserCommand` | `data.name` | `context-action` |
| Button | `TButton` | `data.custom_id` | `button_<context>_<action>` |
| Select menu (string/user/role/channel/mentionable) | `TSelectMenu` | `data.custom_id` | `select_<context>_<action>` |
| Modal submit | `TModalSubmit` | `data.custom_id` | `modal_<context>_<action>` |
| Autocomplete | (none) | command + option | Handled in slash handler or shared handler keyed by command+option |

**Autocomplete**: No dedicated type. Handled inside slash command handlers or a shared handler keyed by command name + option name.

---

## 2. ID Naming Conventions

| Component | Pattern | Example |
|-----------|---------|---------|
| Buttons | `button_<context>_<action>` | `button_compose-create_post`, `button_compose-create_cancel` |
| Select menus | `select_<context>_<action>` | `select_vote_approve`, `select_vote_reject` |
| Modals | `modal_<context>_<action>` | `modal_compose-create_confirm` |

Context and action use hyphens for multi-word parts (e.g. `compose-create`, not `compose_create`).

---

## 3. Definition Structs

### SCommandDef (Slash, Message, User Commands)

```go
type SCommandDef struct {
    Definition   *discordgo.ApplicationCommand  // Discord command definition for sync
    Execute      func(s *discordgo.Session, i *discordgo.InteractionCreate)
    Autocomplete func(s *discordgo.Session, i *discordgo.InteractionCreate) // optional
}
```

- **Definition**: Used by registry for sync (fetch, diff, bulk overwrite).
- **Execute**: Called when the command is invoked.
- **Autocomplete**: Optional. Called when user types in an option that supports autocomplete. If nil, no autocomplete for this command.

### SButtonDef

```go
type SButtonDef struct {
    Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
type MButtonDefinitions map[TButton]SButtonDef
```

### SSelectMenuDef

```go
type SSelectMenuDef struct {
    Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
type MSelectMenuDefinitions map[TSelectMenu]SSelectMenuDef
```

### SModalSubmitDef

```go
type SModalSubmitDef struct {
    Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
type MModalSubmitDefinitions map[TModalSubmit]SModalSubmitDef
```

Message and user context commands use the same `SCommandDef` pattern with separate maps (`MMessageCommandDefinitions`, `MUserCommandDefinitions`) if needed, or can be folded into a unified map keyed by type + name.

---

## 4. Guild Lifecycle and Storage

### GuildCreate

When the bot joins a guild (or re-connects and receives GuildCreate):

- Store guild metadata: `id`, `name`.
- Store per-guild config: allowed roles, default channel, logging channel.
- Use **upsert**: GuildCreate can fire on re-availability (e.g. bot comes back online), so overwrite rather than fail on duplicate.

### GuildDelete

When the bot leaves a guild:

- Remove or soft-delete guild config and proxy metadata.
- **Policy**: Choose one and document in `internal/storage`:
  - **Delete**: Hard remove all data.
  - **Soft-delete**: Mark as deleted, retain for audit window.
  - **Retention**: Keep for audit; handlers treat as inactive.

### Orphaned Messages

- Messages in a guild the bot has left will fail on edit (e.g. 10008 unknown message).
- Handlers treat unknown guild/404 appropriately: clear user message, log, do not retry.

---

## 5. Error Handling

Discord does **not** send a dedicated gateway "Error" event for REST failures. Errors come from:

- **(a) REST API responses** - HTTP status + JSON body (e.g. 429 rate limit, 50035 validation).
- **(b) Gateway close codes** - Connection-level (e.g. 4001 reconnect, 4004 invalid token).
- **(c) Gateway opcodes** - Event payloads that indicate failure.

### Error Categorization

| Category | Example Codes | Handling |
|----------|---------------|----------|
| Transient | 429 (rate limit), 502 (server error) | Retry with backoff |
| Permanent auth | 40001 (unauthorized) | No retry; log and notify |
| Permanent resource | 10003 (unknown channel), 10008 (unknown message) | Clear user message; handlers treat unknown guild/404 appropriately |
| Validation | 50035 (invalid form body) | Field-specific feedback to user |

### Handling Flow

1. **Log** - Write the error to the terminal.
2. **User feedback** - Inform the user who triggered it that something went wrong.
3. **Logging channel** (optional) - Send a formatted error embed to a configured channel (polish feature).

---

## 6. Event Handler Registration

`main` wires handlers via `session.AddHandler` for each event:

```go
session.AddHandler(events.HandleInteractionCreate)
session.AddHandler(events.HandleGuildCreate)
session.AddHandler(events.HandleGuildDelete)
session.AddHandler(events.HandleReady)      // optional
// Error handling: wrap REST calls, handle gateway close in Open()
```

There is no single "Error" event handler; errors are surfaced from REST call returns and gateway connection state.

---

## 7. Post-MVP Infrastructure

Planned extensions beyond MVP:

### Voting

- **Reaction handlers**: `MESSAGE_REACTION_ADD` for vote-by-reaction.
- **Approval buttons**: `TButton` for Approve/Reject actions.
- **State machine**: Draft -> Pending -> Approved/Rejected.
- **Vote storage**: Persist votes per message/decision.

### Admin/Config

- Uses `TSlashCommand`; no new interaction types.
- Guild configuration (allowed roles, default channel, logging channel) via slash commands.

### Collaborative Editing

- **Edit button** on messages: `TButton` (e.g. `button_proxy_edit`).
- **Permission middleware**: Check who can edit (owner, allowed roles).
- **Message context command**: Optional "Edit this message" (`TMessageCommand`) for right-click edit.
