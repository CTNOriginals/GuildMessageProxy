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

See [docs/PROJECT_STATUS.md](./docs/PROJECT_STATUS.md) for the current backlog and upcoming features in priority order. Planned command groups include:

- `/compose` enhancements (identity selection, scheduling, attachments)
- `/message` management (delete, info, history)
- `/config` guild settings (roles, channels, restrictions)
- `/template` and `/draft` management
- `/vote` governance workflows
- `/audit` and `/stats` analytics

## License

See [LICENSE.md](./LICENSE.md).
