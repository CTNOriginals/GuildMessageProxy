# Route Map - Commands and Flows

Where commands are defined, how they flow, and how handlers are wired.

## Command Structure

Commands use subcommands under a top-level group. The main category is `/compose`.

### Planned Commands

| Route | Purpose |
|-------|---------|
| `/compose create` | Start a new draft with initial content |
| `/compose set` | Adjust draft properties (channel, allow_edits, etc) |
| `/compose propose` | Submit a proposed change to an existing proxied message |
| `/compose post` | Confirm and post the current draft |

MVP may collapse to `/compose create` + `/compose post` initially. Structure should allow growth.

## Flow: Compose -> Preview -> Post

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

## Flow: Basic Edit

```
User: /compose propose (target message, new content)
  -> handlers: verify requester is original owner (MVP)
  -> handlers: render edited preview
  -> Bot: ephemeral preview + [Apply] [Cancel] buttons

User: clicks Apply
  -> handlers: edit original proxied message
  -> storage: update metadata (last edited by, last edited at)
```

## Handler Wiring (Planned)

| Flow Step | Handler | Called From |
|-----------|---------|-------------|
| Validate content | `handlers.ValidateContent` | compose create, compose propose |
| Render preview | `handlers.RenderPreview` | compose create, compose propose |
| Post message | `handlers.PostProxiedMessage` | compose post |
| Edit message | `handlers.EditProxiedMessage` | compose propose (Apply) |
| Permission check | `handlers.CanUseCompose` | all compose subcommands |

## Interaction Routing

- Slash commands: synced on startup via `registry.SyncCommands(session, guildID)`. Fetches existing, diffs against desired definitions, bulk overwrites only when changed. Scope: `--guild=<id>` (dev) or `--global` (prod).
- Button clicks (Post, Cancel, Apply): handled via `session.AddHandler` for `InteractionCreate`. Route by `interaction.Type` and `interaction.MessageComponentData().CustomID`.

## MVP Restrictions

- Only the original requester can edit.
- Posting uses webhooks (custom avatar/username per message); MVP may restrict identity options.
- No voting/approval workflows yet.

## Future Routes (Out of Scope for MVP)

- `/admin` or `/config` - guild configuration
- Voting/approval flows
- `/message edit` - alternative to `/compose propose` for edits
