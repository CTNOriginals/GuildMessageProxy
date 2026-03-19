# Project Status

**Source of truth for project state, current work, and upcoming backlog.**

Last updated: March 2026

---

## What Exists Now

Working Discord bot with core compose functionality:


| Command                                                                            | Description                                | Status      |
| ---------------------------------------------------------------------------------- | ------------------------------------------ | ----------- |
| `/compose-draft <content> [channel]`                                               | Create draft with preview before posting   | Implemented |
| `/compose-send <content> [channel]`                                                | Post directly without preview              | Implemented |
| `/compose-edit <message> <content>`                                                | Propose edits to existing proxied messages | Implemented |
| `/compose-help`                                                                    | Show help for compose commands             | Implemented |
| `/message-delete`                                                                  | Delete a proxied message                   | Implemented |
| `/config-role, /config-channel, /config-restrict, /config-allow, /config-defaults` | Guild configuration commands               | Implemented |


**Features:**

- Preview with Post/Cancel buttons
- Edit proposal with Apply/Cancel buttons
- Webhook-based message posting with attribution footer
- Basic permission checking (SendMessages)
- SQLite storage with in-memory option for testing
- Structured logging with configurable levels

---

## Currently In Progress


| Feature             | Description                                            | Phase   |
| ------------------- | ------------------------------------------------------ | ------- |
| `/compose identity` | Choose posting identity (self, bot, or custom webhook) | Phase 1 |
| `/compose schedule` | Queue message for future delivery                      | Phase 1 |
| Custom webhooks     | Per-message avatars and usernames                      | Phase 1 |
| Attachments support | Include images and files in messages                   | Phase 1 |
| `/message info`     | View message metadata and history                      | Phase 2 |
| `/message history`  | List messages by user or in channel                    | Phase 2 |
| Database storage    | Replace in-memory with persistent storage              | Phase 3 |
| Persistent drafts   | Save drafts across bot restarts                        | Phase 3 |


---

## Backlog (Priority Order)

### Phase 1: Enhanced Compose (Next)


| Feature             | Description                                            | Complexity |
| ------------------- | ------------------------------------------------------ | ---------- |
| `/compose identity` | Choose posting identity (self, bot, or custom webhook) | Medium     |
| `/compose schedule` | Queue message for future delivery                      | Medium     |
| Custom webhooks     | Per-message avatars and usernames                      | Low        |
| Attachments support | Include images and files in messages                   | Medium     |


### Phase 2: Message Management


| Feature               | Description                         | Complexity | Status        |
| --------------------- | ----------------------------------- | ---------- | ------------- |
| ~~`/message delete`~~ | ~~Delete a proxied message~~        | ~~Low~~    | ~~Completed~~ |
| `/message info`       | View message metadata and history   | Low        | Pending       |
| `/message history`    | List messages by user or in channel | Medium     | Pending       |


### Phase 3: Persistence & Configuration


| Feature                | Description                                | Complexity | Status        |
| ---------------------- | ------------------------------------------ | ---------- | ------------- |
| Database storage       | Replace in-memory with persistent storage  | High       | Pending       |
| Persistent drafts      | Save drafts across bot restarts            | Medium     | Pending       |
| ~~`/config role`~~     | ~~Restrict commands to specific roles~~    | ~~Low~~    | ~~Completed~~ |
| ~~`/config channel`~~  | ~~Set default target channel~~             | ~~Low~~    | ~~Completed~~ |
| ~~`/config restrict`~~ | ~~Restrict commands to specific channels~~ | ~~Low~~    | ~~Completed~~ |
| ~~`/config allow`~~    | ~~Allow commands in specific channels~~    | ~~Low~~    | ~~Completed~~ |
| ~~`/config defaults`~~ | ~~Set default configuration values~~       | ~~Low~~    | ~~Completed~~ |


### Phase 4: Templates & Collaboration


| Feature                           | Description                                 | Complexity |
| --------------------------------- | ------------------------------------------- | ---------- |
| `/template save/load/list/delete` | Save and reuse message formats              | Medium     |
| `/draft list/resume/delete`       | Manage pending drafts                       | Medium     |
| Collaborative drafts              | Multiple users contribute to a single draft | High       |


### Phase 5: Governance & Analytics


| Feature                  | Description                     | Complexity |
| ------------------------ | ------------------------------- | ---------- |
| `/vote start/status/end` | Approval workflows for messages | High       |
| Audit logs               | Full logging of all bot actions | Medium     |
| `/stats user/guild`      | Usage statistics                | Medium     |


---

## Historical Summary

**MVP Completed: March 2026**

Core compose functionality implemented with preview system, webhook posting, and edit proposals. Infrastructure includes event handlers, command sync, interaction routing, and structured logging.

---

## Reference

- Technical details: [ARCHITECTURE.md](./ARCHITECTURE.md)
- Implementation patterns: [TEMPLATES.md](./TEMPLATES.md)

