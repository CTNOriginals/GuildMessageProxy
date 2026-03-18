# GuildMessageProxy

![Go](https://img.shields.io/badge/Go-1.25-blue?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

A Discord bot written in Go that allows you to compose messages with custom styling and markdown, then post them via webhooks. After posting, you can propose edits to the message without needing the original author to manually edit it.

## Features

- **Compose messages** with full Discord markdown support
- **Preview before posting** with Post/Cancel buttons
- **Post directly** without preview for quick messages
- **Propose edits** to existing proxied messages (original author only for MVP)
- **Webhook-based posting** for clean message attribution

### Commands

| Command | Description |
|---------|-------------|
| `/compose create <content> [channel]` | Create a draft with preview before posting |
| `/compose post <content> [channel]` | Post a message directly without preview |
| `/compose propose <message> <content>` | Propose an edit to an existing proxied message (owner only) |

## Prerequisites

- [Go](https://golang.org/dl/) 1.25 or higher
- A Discord bot token ([Discord Developer Portal](https://discord.com/developers/applications))

## Installation

1. Clone the repository:
```bash
git clone https://github.com/CTNOriginals/GuildMessageProxy.git
cd GuildMessageProxy
```

2. Copy the environment template and fill in your values:
```bash
cp .env.example .env
```

3. Edit `.env` with your bot credentials:
```
TOKEN=your_bot_token_here
CLIENT_ID=your_application_id_here

# Optional: Development settings
DEV_GUILD_ID=your_test_guild_id
DEV_CHANNEL_LOG_ID=logging_channel_id
DEV_CHANNEL_ERROR_ID=error_channel_id
```

## Quick Start

Already have a Discord bot token? Get started in 3 steps:

1. **Set your token** in `.env`:
   ```
   TOKEN=your_bot_token_here
   ```

2. **Run the bot**:
   ```bash
   make run
   ```

3. **Try a command** in Discord:
   ```
   /compose create Hello, World!
   ```

## Running

```bash
# Run normally
make run

# Run with a custom token (overrides .env)
make run args="-t YOUR_TOKEN"

# Run and watch for file changes (requires wgo: https://github.com/bokwoon95/wgo)
make wrun
```

## Available Make Targets

```bash
make help      # Display all available commands
make run       # Run the bot
make wrun      # Run with file watching (requires wgo)
make test      # Run tests
make build     # Build binary to ./build/
make lint      # Run golangci-lint
make tidy      # Run go mod tidy
make version   # Show current version
```

## Architecture

- **Language:** Go 1.25
- **Discord library:** [discordgo](https://github.com/bwmarrin/discordgo)
- **Storage:** In-memory (with interface for future persistence)
- **Posting:** Discord webhooks for flexible message attribution

## Documentation

For detailed documentation, see the [`docs/`](./docs/) folder:

- [Architecture Overview](./docs/ARCHITECTURE.md) - Package layout and design
- [Project Map](./docs/PROJECT_MAP.md) - Directory structure
- [Command Routes](./docs/ROUTE_MAP.md) - Command flows and handlers
- [Roadmap](./docs/roadmap/) - Future plans and notes

## Project Status

MVP is **complete**. All core features are implemented:
- Full command suite (`/compose create`, `/compose post`, `/compose propose`)
- Preview system with interactive buttons
- Webhook-based message posting
- Edit proposal workflow
- Permission checking
- Structured logging

## Planned Features

The following command groups and features are planned for future releases:

### `/compose` - Enhanced Composition
Enhanced message composition capabilities:

| Planned Command | Description |
|-----------------|-------------|
| `/compose draft` (renamed from `create`) | Create a draft with preview |
| `/compose send` (renamed from `post`) | Send directly without preview |
| `/compose edit` (renamed from `propose`) | Propose edits to existing messages |
| `/compose identity` | Choose posting identity (self, bot, or other with safeguards) |
| `/compose schedule` | Queue message for future delivery |
| `/compose clone <message>` | Use existing message as template |

Features:
- **Persistent drafts** - Save drafts across bot restarts
- **Collaborative drafts** - Multiple users contribute to a single draft
- **Custom webhooks** - Per-message avatars and usernames
- **Embed builder** - Multi-step forms for complex embeds
- **Attachments support** - Include images and files
- **Thread/forum support** - Post to threads and forum posts
- **Bulk posting** - Send to multiple channels at once
- **Cross-guild posting** - Post across multiple guilds (with permissions)

### `/message` - Message Management
Manage existing proxy messages:

| Planned Command | Description |
|-----------------|-------------|
| `/message delete <message>` | Delete a proxied message |
| `/message info <message>` | View message metadata and history |
| `/message history [user]` | List messages by user or in channel |
| `/message restore <message>` | Recover and re-post a deleted message |
| `/message pin <message>` | Pin a proxy message |
| `/message unpin <message>` | Unpin a proxy message |
| `/message react <message> <emoji>` | Auto-add reactions to a message |
| `/message expire <message> <duration>` | Auto-delete after set duration |

### `/template` - Message Templates
Save and reuse message formats:

| Planned Command | Description |
|-----------------|-------------|
| `/template save <name>` | Save current draft or message as template |
| `/template load <name>` | Load a saved template into a draft |
| `/template list` | View all saved templates |
| `/template delete <name>` | Remove a template |
| `/template share <name>` | Share template with other users |

### `/draft` - Draft Management
Manage pending drafts:

| Planned Command | Description |
|-----------------|-------------|
| `/draft list` | View all your pending drafts |
| `/draft resume <draft_id>` | Continue working on a draft |
| `/draft delete <draft_id>` | Discard a draft |
| `/draft share <draft_id> [user]` | Let others contribute to your draft |

### `/vote` - Voting & Governance
Approval workflows for important messages:

| Planned Command | Description |
|-----------------|-------------|
| `/vote start <message>` | Initiate a vote on a draft or proposal |
| `/vote status <message>` | Check current vote status |
| `/vote end <message>` | End voting and apply result |
| `/vote configure` | Set thresholds and rules for votes |

Features:
- **Approval workflows** - Vote to approve/reject before posting
- **Edit approval gates** - Require votes for edits/deletions
- **Configurable thresholds** - Different requirements per action type

### `/config` - Guild Configuration
Server-specific settings:

| Planned Command | Description |
|-----------------|-------------|
| `/config role <role>` | Set which roles can use commands |
| `/config channel <channel>` | Set default target channel |
| `/config restrict <channel>` | Blacklist a channel |
| `/config allow <channel>` | Whitelist a channel |
| `/config defaults` | View current guild settings |

Features:
- **Role-based access** - Restrict commands to specific roles
- **Per-guild settings** - Customize defaults per server
- **Channel restrictions** - Blacklist/whitelist specific channels

### `/audit` - Audit & History
View logs and search messages:

| Planned Command | Description |
|-----------------|-------------|
| `/audit search [user] [date]` | Find messages by criteria |
| `/audit log <message>` | View full history of a message |
| `/audit export` | Export audit logs to file |
| `/audit recent [count]` | View recent bot actions |

Features:
- **Database storage** - Persistent storage replacing in-memory
- **Audit logs** - Full logging of all bot actions
- **Search & filtering** - Find messages by content, author, date
- **Import/export** - Backup and migrate message history

### `/stats` - Analytics
Usage statistics and analytics:

| Planned Command | Description |
|-----------------|-------------|
| `/stats user [user]` | Stats for a specific user |
| `/stats guild` | Guild-wide statistics |
| `/stats popular` | Most active channels and users |
| `/stats trends` | Activity over time |

Features:
- **Usage analytics** - Statistics on activity and content
- **Popular content** - Most used templates and channels

## License

See [LICENSE.md](./LICENSE.md).
