# Troubleshooting Guide

Common issues and solutions for GuildMessageProxy bot operation.

See also: [DEPLOYMENT.md](./DEPLOYMENT.md) for setup and deployment issues.


## Common Issues and Solutions

### Bot won't start

- Check TOKEN is set in .env or via -t flag
- Verify .env file is in project root
- Check for syntax errors in environment file

### Commands not appearing

- Use `--guild=<id>` for instant registration in dev
- Wait up to 1 hour for global commands to propagate
- Check bot has `applications.commands` scope
- Verify command sync succeeded in logs

### Can't post messages

- Bot needs Send Messages permission in target channel
- Bot needs Manage Webhooks permission
- Check channel-specific permissions override guild permissions

### Edit proposals failing

- Only the original poster can edit (MVP restriction)
- Message must be a proxied message (posted via bot)
- Check message ID format or use full message URL


## Debug Logging Configuration

- Set LOG_LEVEL environment variable: DEBUG, INFO, WARN, ERROR, FATAL
- Default is INFO
- Logs include structured fields: guild_id, channel_id, user_id, error


## How to Check if Commands are Registered

- Look for "commands registered" in startup logs
- Use `--no-sync` flag to skip sync and verify existing commands work
- In Discord, type `/` and look for "compose" commands


## How to Verify Bot Permissions

Check bot has these permissions in guild:

- Send Messages
- Manage Messages
- Manage Webhooks
- Use Slash Commands (implicit)

Use Discord's bot permission calculator for invite URL.


## Database Troubleshooting

- Current storage is in-memory (resets on restart)
- No database setup required currently
- Lost drafts on restart is expected behavior


## Discord API Error Reference

| Error Code | Meaning | Solution |
| ---------- | ------- | ---------- |
| 429 | Rate limited | Wait and retry |
| 40001 | Unauthorized | Check bot token |
| 10003 | Unknown channel | Verify channel ID |
| 10008 | Unknown message | Message was deleted or never existed |
| 50035 | Invalid form body | Check message content length (max 2000) |
