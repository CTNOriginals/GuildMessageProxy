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


