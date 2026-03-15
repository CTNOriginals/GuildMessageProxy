# GuildMessageProxy

This bot allows you to compose any message, stylize it in any way you want (make an embed, use complex markdown and so forth), and have the bot post it for you as yourself, the bot, or any other user.
After the message is posted, you can then allow other members to edit the message without needing the original user to manually edit it again.

## Features

### MVP (Current / In Progress)

- Post messages with unique styling and markdown.
- Post messages as yourself, the bot, or any other user.
- Allow others to edit the message without needing the original user to manually edit it again.

### Planned (Out of Scope for MVP)

- Setup a voting system to approve or reject the message from being posted, edited or deleted.

## Configuration

Copy `.env.example` to `.env` and fill in the values. See `.env.example` for required variables (`TOKEN`, `CLIENT_ID`, and optional dev guild/channel IDs).

## Running

```bash
make run
```

Or pass the token via flag: `make run args="-t YOUR_TOKEN"`. See `make help` for other targets (build, test, lint).


