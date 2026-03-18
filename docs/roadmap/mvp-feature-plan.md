## MVP Feature Plan - GuildMessageProxy

## Status: COMPLETE

All MVP features have been implemented and are in production.

---

This document describes the first version of GuildMessageProxy focused on composing, previewing, and posting proxy messages, plus a minimal editing flow.

The goal is to ship a usable core with clear behavior before adding voting, advanced governance, or complex persistence.

---

### 1. User-Facing MVP Features

Commands are grouped into intuitive categories using subcommands. For composing and managing a proxied message, the main category is `/compose`.

- **Compose and edit proxied messages**
  - Command group: `/compose ...`
  - Initial subcommands (names are examples and can be refined later):
    - `/compose create` - start a new draft with initial content.
    - `/compose set` - adjust properties of the draft (channel, allow_edits flag, etc).
    - `/compose propose` - submit a proposed change to an existing proxied message.
    - `/compose post` - confirm and post the current draft.
  - For the absolute MVP, this can be collapsed into a smaller set (for example, `/compose create` + `/compose post`) as long as the structure leaves room to grow.
  - Editing can initially be exposed either as:
    - A separate subcommand like `/compose propose` targeting an existing message, or
    - A dedicated group such as `/message edit`, depending on what feels clearer once implemented.

---

### 2. Core Flows

#### 2.1 Compose → Preview → Post

1. **Initiate**
   - User runs a compose command in a guild text channel, for example:
     - `/compose create` with minimal required input: target content (string).
2. **Compose**
   - For MVP, treat content as a single text field that may contain markdown.
   - Embeds and advanced layouts can be added later as separate options or additional subcommands.
3. **Preview**
   - Bot responds ephemerally with:
     - A rendered preview of the message (as close as possible to the final output).
     - Metadata summary: target channel, posting identity. Posting uses channel webhooks (custom avatar/username per message); for MVP, identity options may be restricted.
     - Buttons: `Post` and `Cancel`.
4. **Post**
   - On `Post` (or via a `/compose post` subcommand, depending on the final UX):
     - Bot sends the final message to the target channel.
     - Bot stores metadata needed for edits (guild ID, channel ID, message ID, requesting user ID, timestamp, flags).
   - On `Cancel`:
     - Bot dismisses the preview and discards the draft.

Assumption: Posting will use channel webhooks, which support custom avatar and username per message. For MVP, the visible author identity may be restricted (e.g. bot-only or text attribution like "Requested by @User") until governance is in place.

#### 2.2 Basic Edit Flow

1. **Initiate edit**
   - User (initially: the original requester) triggers an edit via:
     - A subcommand such as `/compose propose` pointing at an existing message, or
     - An interaction button under the original proxied message (preferred UX, but may be added after the first command-based version).
2. **Propose new content**
   - User provides new message content as a text field.
3. **Preview edited message**
   - Bot displays an ephemeral preview of how the edited message will look.
   - Buttons: `Apply` and `Cancel`.
4. **Apply edit**
   - On `Apply`, bot edits the original proxied message.
   - Update any stored metadata if necessary (for example, last edited by, last edited at).

MVP restriction: Only the original requester can edit. Broader editing permissions and voting will be handled in later iterations.

---

### 3. Minimal Configuration for MVP

- **Per-bot configuration**
  - Bot token and basic environment configuration via `.env` (already used in `cmd/bot/main.go`).
  - Command sync scope: `--guild=<id>` for dev (instant) or `--global` for prod. Commands sync on startup (fetch, diff, bulk overwrite when changed).
- **Per-guild configuration (MVP)**
  - Keep configuration extremely simple:
    - Optional list of roles allowed to use the `compose` command group at all.
    - Default target channel behavior (post to the same channel where the command was invoked).

Persistence of configuration beyond memory (e.g. database, files) can be postponed if necessary, but the flows should be designed so that adding storage later is straightforward.

---

### 4. Common Roadblocks to Consider (MVP Scope)

For each of these, MVP should at least define a basic behavior, even if it is not fully configurable yet:

- **Discord rate limits**
  - Avoid repeated rapid edits/posting; surface friendly error messages if Discord returns rate-limit errors.
- **Permission errors**
  - If the bot lacks `SendMessages` or `ManageMessages` in the target channel, respond ephemerally with a clear explanation.
- **Message length and embed limits**
  - Validate content length before sending; warn the user if they approach Discord limits.
- **Editing after long delays**
  - If the message can no longer be edited (e.g. due to Discord constraints or lost metadata), provide a clear error and suggest re-posting.

These should be captured as behavior notes, even if the implementation is initially minimal.

---

### 5. Out of Scope for MVP

The following are explicitly **not** in the first iteration, but should be mentioned here so they are not forgotten:

- Voting / approval workflows for posting, editing, or deleting.
- Posting **as** another user in a way that looks identical to them without clear attribution.
- Complex multi-step forms for building rich embeds.
- Persistent audit logs and dashboards.

Each of these will get its own feature document in `docs/roadmap` once the core MVP flows are solid.

