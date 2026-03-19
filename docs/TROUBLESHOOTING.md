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

- Default storage is SQLite (persistent across restarts)
- Database file: guildmessageproxy.db (or use DATABASE_PATH env var)
- Use --memory flag for in-memory testing mode (resets on restart)
- No database setup required - SQLite works out of the box

### Database Permission Issues

- If SQLite fails to write: Check file permissions on guildmessageproxy.db
- Run with --memory flag to bypass database issues: `make run args="--memory"`


## Permission Error Reference

When using compose commands, you may encounter these permission-related errors:

| Error Message | Cause | Solution |
| ------------- | ------- | ---------- |
| Cannot access this channel. The bot may lack permissions, or the channel no longer exists. Try again or contact a server admin. | Bot cannot access the channel | Verify the bot has View Channel permission, or the channel still exists |
| Cannot verify your permissions in this channel. Try again or use a different channel. | Failed to retrieve user's channel permissions | Try the command again, or use a different channel where you have permissions |
| You need 'Send Messages' permission in this channel to use this command. | User lacks Send Messages permission | Ask a server admin to grant you Send Messages permission in this channel |
| This command requires an allowed role. Ask a server admin which roles can use compose commands. | Guild has role restrictions configured | Contact a server admin to learn which roles are allowed, or request assignment to an allowed role |
| This channel is restricted. Use compose commands in an allowed channel instead. | Channel is in the restricted list | Use a different channel that is not restricted for compose commands |
| This channel is not allowed for compose commands. Use a permitted channel or ask a server admin to add this channel. | Channel whitelist is active and this channel is not included | Use a channel in the allowed list, or ask a server admin to add this channel to the whitelist |


## Discord API Error Reference

| Error Code | Meaning | Solution |
| ---------- | ------- | ---------- |
| 429 | Rate limited | Wait and retry |
| 40001 | Unauthorized | Check bot token |
| 10003 | Unknown channel | Verify channel ID |
| 10008 | Unknown message | Message was deleted or never existed |
| 50035 | Invalid form body | Check message content length (max 2000) |
