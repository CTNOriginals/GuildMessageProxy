# Deployment - Self-Hosting Guide

Instructions for running GuildMessageProxy on your own infrastructure.

For project structure, see [PROJECT_MAP.md](./PROJECT_MAP.md). For code architecture, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| [Go](https://golang.org/dl/) | 1.25+ | Build and run the bot |
| [Git](https://git-scm.com/) | Any | Clone the repository |
| Discord account | - | Create bot application |

## Installation

### 1. Clone and Build

```bash
# Clone the repository
git clone https://github.com/CTNOriginals/GuildMessageProxy.git
cd GuildMessageProxy

# Copy environment template
cp .env.example .env

# Build the binary
make build
```

### 2. Configure Environment

Edit `.env` with your bot credentials:

```bash
TOKEN=your_bot_token_here
CLIENT_ID=your_application_id_here

# Optional: Development settings
DEV_GUILD_ID=your_test_guild_id
DEV_CHANNEL_LOG_ID=logging_channel_id
DEV_CHANNEL_ERROR_ID=error_channel_id
```

See [Environment Variable Reference](#environment-variable-reference) for details on each variable.

### 3. Run the Bot

```bash
# Run from source
make run

# Run the built binary
./build/bot.exe

# Run with command-line flags (overrides .env)
make run args="-t YOUR_TOKEN"
```

## Environment Variable Reference

| Variable | Required | Description |
|----------|----------|-------------|
| TOKEN | Yes | Discord bot token from the Developer Portal |
| CLIENT_ID | Yes | Discord application client ID |
| DEV_GUILD_ID | No | Development guild ID for command registration |
| DEV_CHANNEL_LOG_ID | No | Channel ID for bot logs |
| DEV_CHANNEL_ERROR_ID | No | Channel ID for error notifications |

**Security note:** Keep your `TOKEN` secret. Do not commit `.env` to version control.

## Discord Bot Setup

### 1. Create a Discord Application

1. Go to [discord.com/developers/applications](https://discord.com/developers/applications)
2. Click "New Application" and give it a name
3. Navigate to the "Bot" section in the left sidebar
4. Click "Add Bot" and confirm

### 2. Get Your Credentials

| Credential | Location | Copy Value |
|------------|----------|------------|
| TOKEN | Bot page > "Token" section | Click "Reset Token" and copy |
| CLIENT_ID | General Information page | Application ID field |

### 3. Configure OAuth2 Scopes

1. Go to "OAuth2" > "URL Generator"
2. Select these scopes:
   - `bot` - Connect as a bot user
   - `applications.commands` - Register slash commands

### 4. Set Bot Permissions

Under "Bot Permissions", select:

| Permission | Why Needed |
|------------|----------|
| Send Messages | Post proxied messages |
| Manage Messages | Clean up draft previews |
| Manage Webhooks | Create channel webhooks for posting |

Copy the generated URL and open it in a browser to invite the bot to your server.

## Command Registration

Commands must be registered with Discord before they appear. The bot syncs commands on startup.

### Registration Modes

| Flag | Use Case | Propagation Time |
|------|----------|------------------|
| `--guild=<id>` | Development | Instant |
| `--global` | Production | Up to 1 hour |
| `--no-sync` | Skip sync for faster restarts | N/A |

### Examples

```bash
# Development - instant updates in specific guild
make run args="--guild=123456789"

# Production - global commands (takes up to 1 hour)
make run args="--global"

# Skip sync (commands already registered)
make run args="--no-sync"
```

### Checking Command Registration

1. Join your Discord server
2. Type `/` in any channel
3. You should see the `/compose` command group
4. If commands do not appear after 1 hour (global), restart the bot with sync enabled

See [ARCHITECTURE.md](./ARCHITECTURE.md#command-registration-startup-sync) for technical details on the sync system.

## Database Setup

### Current State: In-Memory Storage

The MVP uses in-memory storage for proxy message metadata. This means:

- **No persistence**: Data is lost when the bot restarts
- **No setup required**: Works out of the box
- **Suitable for**: Testing, small deployments, development

### Future: SQLite Persistence

Persistent storage is planned for a future release using SQLite. This will:

- Retain proxy message metadata across restarts
- Enable audit logging and message history
- Support the `/audit` and `/stats` command groups

No database configuration is currently required.

See [PROJECT_MAP.md](./PROJECT_MAP.md#directory-tree-current--planned) for storage implementation details.

## Logging

### Default Output

Logs are sent to stdout and stderr by default:

| Level | Output | Use |
|-------|--------|-----|
| DEBUG, INFO | stdout | General operation |
| WARN, ERROR, FATAL | stderr | Issues and failures |

### Log Format

```
[INFO] [26-03-18 14:30:00] bot starting
       version: dev
       go_version: go1.25.0
```

### Discord Channel Logging

Optional: Configure `DEV_CHANNEL_LOG_ID` and `DEV_CHANNEL_ERROR_ID` to send logs to Discord channels.

| Variable | Level | Purpose |
|----------|-------|---------|
| DEV_CHANNEL_LOG_ID | INFO+ | General bot activity |
| DEV_CHANNEL_ERROR_ID | WARN, ERROR, FATAL | Issues requiring attention |

See [ARCHITECTURE.md](./ARCHITECTURE.md#error-handling) for error categorization and handling.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Bot does not start | Missing TOKEN | Check `.env` file or pass `-t` flag |
| Commands not appearing | Not registered | Wait 1 hour (global) or use `--guild` |
| Permission denied | OAuth2 scopes | Re-invite with `bot` and `applications.commands` |
| Cannot post messages | Missing permissions | Grant Send Messages, Manage Webhooks |

## Production Checklist

- [ ] Created Discord application and bot
- [ ] Invited bot with correct permissions
- [ ] Set `TOKEN` and `CLIENT_ID` in `.env`
- [ ] Registered commands with `--global`
- [ ] Tested all `/compose` commands
- [ ] Considered log channel configuration
- [ ] Bot running on stable infrastructure
