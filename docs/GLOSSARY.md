# Glossary - Terms and Jargon

Definitions for terms used across the project and docs.


| Term                | Definition                                                                                                                                                                                                       |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Proxy message**   | A message posted on behalf of a user. The bot uses channel webhooks to send these, which support custom avatar and username per message. Attribution (e.g. "Requested by @User") may appear in the message body. |
| **Compose**         | The flow of drafting, previewing, and posting a proxy message. Also the `/compose` command group.                                                                                                                |
| **Draft**           | In-progress message content and metadata before posting. Stored in memory or session until Post or Cancel.                                                                                                       |
| **Proxied message** | Same as proxy message. Used in storage and handler naming.                                                                                                                                                       |
| **Webhook**         | Discord channel webhook. Used to post proxy messages with custom avatar and username per message, so the visible author can match the requested identity.                                                        |
| **Ephemeral**       | A response visible only to the user who triggered it. Used for previews and errors.                                                                                                                              |
| **Registry**        | `internal/commands/registry.go` - command definitions and startup sync. Fetches existing commands, diffs against desired, bulk overwrites only when changed.                                                                 |
| **Handlers**        | `internal/handlers/` - reusable logic (preview, post, permissions, validation) used by commands.                                                                                                                 |
| **MVP**             | Minimum Viable Product. First shippable version: compose, preview, post, basic edit.                                                                                                                             |
| **Intent**          | Discord gateway intent. `IntentsGuildMessages` allows receiving messages in guilds.                                                                                                                              |
| **Slash command**   | Discord application command (e.g. `/compose create`). Preferred over message-based commands.                                                                                                                     |
| **Event handler**   | Code in `internal/events/` that receives Discord gateway events (InteractionCreate, GuildCreate, GuildDelete, Error) and routes or processes them.                                                               |
| **TSlashCommand**   | Custom type `string` whose value is the slash command name (e.g. `"compose-create"`). Used to route slash interactions to definitions via maps.                                                                   |
| **TMessageCommand** | Custom type `string` for message context menu (right-click message). Identified by `data.name`.                                                                                                                  |
| **TUserCommand**    | Custom type `string` for user context menu (right-click user). Identified by `data.name`.                                                                                                                         |
| **TButton**         | Custom type `string` whose value is the button `custom_id` (e.g. `"button_compose-create_post"`). Used to route button interactions to definitions via maps.                                                     |
| **TSelectMenu**     | Custom type `string` for string/user/role/channel/mentionable select menus. Identified by `data.custom_id`. Naming: `select_<context>_<action>`.                                                                   |
| **TModalSubmit**    | Custom type `string` for modal form submissions. Identified by `data.custom_id`. Naming: `modal_<context>_<action>`.                                                                                              |
| **MCommandDefinitions** | Map type `map[TSlashCommand]SCommandDef`. Routes slash command types to their definition and execute logic.                                                                                              |
| **SCommandDef**     | Struct holding a slash command definition (ApplicationCommand), Execute function, and optional Autocomplete function. Used as the value type in MCommandDefinitions.                                             |
| **SButtonDef**      | Struct holding a button execute function. Used as the value type in MButtonDefinitions.                                                                                                                          |
| **MButtonDefinitions** | Map type `map[TButton]SButtonDef`. Routes button custom_ids to their execute logic.                                                                                                                         |
| **ID naming convention** | Buttons: `button_<context>_<action>`. Select menus: `select_<context>_<action>`. Modals: `modal_<context>_<action>`. Example: `button_compose-create_post`, `select_vote_approve`, `modal_compose-create_confirm`. |
| **Error handling**  | Errors come from REST API responses (HTTP + JSON) and gateway close codes, not a dedicated Discord "Error" event. Categorization: Transient (429, 502 - retry), Permanent auth (40001 - no retry), Permanent resource (10003, 10008 - clear user message), Validation (50035 - field-specific). See `internal/events/error.go`. |
| **Guild lifecycle** | GuildCreate: store guild metadata (id, name), per-guild config (allowed roles, default channel, logging channel). GuildDelete: remove or soft-delete guild config and proxy metadata. Orphaned messages fail on edit; handlers treat unknown guild/404 appropriately. |


