# Project Status

**Source of truth for project state, current work, and upcoming backlog.**

Last updated: March 2026

---

## What Exists Now

Working Discord bot with core compose functionality:

| Command | Description |
|---------|-------------|
| `/compose create <content> [channel]` | Create draft with preview before posting |
| `/compose post <content> [channel]` | Post directly without preview |
| `/compose propose <message> <content>` | Propose edits to existing proxied messages |

**Features:**
- Preview with Post/Cancel buttons
- Edit proposal with Apply/Cancel buttons
- Webhook-based message posting with attribution footer
- Basic permission checking (SendMessages)
- In-memory storage with interface for future persistence
- Structured logging with configurable levels

---

## Currently In Progress

*Nothing actively being worked on.*

---

## Backlog (Priority Order)

### Phase 1: Enhanced Compose (Next)

| Feature | Description | Complexity |
|---------|-------------|------------|
| `/compose identity` | Choose posting identity (self, bot, or custom webhook) | Medium |
| `/compose schedule` | Queue message for future delivery | Medium |
| Custom webhooks | Per-message avatars and usernames | Low |
| Attachments support | Include images and files in messages | Medium |

### Phase 2: Message Management

| Feature | Description | Complexity |
|---------|-------------|------------|
| `/message delete` | Delete a proxied message | Low |
| `/message info` | View message metadata and history | Low |
| `/message history` | List messages by user or in channel | Medium |

### Phase 3: Persistence & Configuration

| Feature | Description | Complexity |
|---------|-------------|------------|
| Database storage | Replace in-memory with persistent storage | High |
| Persistent drafts | Save drafts across bot restarts | Medium |
| `/config role` | Restrict commands to specific roles | Low |
| `/config channel` | Set default target channel | Low |

### Phase 4: Templates & Collaboration

| Feature | Description | Complexity |
|---------|-------------|------------|
| `/template save/load/list/delete` | Save and reuse message formats | Medium |
| `/draft list/resume/delete` | Manage pending drafts | Medium |
| Collaborative drafts | Multiple users contribute to a single draft | High |

### Phase 5: Governance & Analytics

| Feature | Description | Complexity |
|---------|-------------|------------|
| `/vote start/status/end` | Approval workflows for messages | High |
| Audit logs | Full logging of all bot actions | Medium |
| `/stats user/guild` | Usage statistics | Medium |

---

## Historical Summary

**MVP Completed: March 2026**

Core compose functionality implemented with preview system, webhook posting, and edit proposals. Infrastructure includes event handlers, command sync, interaction routing, and structured logging.

---

## Reference

- Technical details: [ARCHITECTURE.md](./ARCHITECTURE.md)
- Implementation patterns: [TEMPLATES.md](./TEMPLATES.md)
