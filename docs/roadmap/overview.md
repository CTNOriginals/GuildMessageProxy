## GuildMessageProxy - Overview

### Purpose

GuildMessageProxy is a Discord bot that lets server members compose rich messages (embeds, complex markdown, etc.) and have the bot post them on their behalf. After posting, the original message can be collaboratively edited without requiring the original author to manually update it.

### Core Capabilities

- **Rich message composition**: Compose messages with unique styling and markdown, including embeds.
- **Flexible authorship**: Post messages as yourself, as the bot, or (with the right safeguards) as another user.
- **Collaborative editing**: Allow others to propose and apply edits to the posted message without needing the original user online.
- **Approval workflows (planned)**: Support voting/approval to gate posting, editing, or deleting important messages.

### Target Audience

- **Server owners and moderators** who need high-quality announcements or rules messages that may be refined over time.
- **Community teams** (e.g. event organizers, staff teams) who collaborate on drafts before publishing.
- **Servers with higher trust requirements**, where impersonation and abuse must be controlled carefully.

### Usage Scenarios

- Drafting and publishing an announcement that multiple moderators help polish before it goes live.
- Posting a rules/FAQ message that different staff members may later update without re-posting from scratch.
- Running a lightweight internal review process (e.g. staff votes) before a sensitive message is posted or changed.

### Roadmap Docs Index

This `docs/roadmap` folder is the home for planning and design docs.

- **MVP flows and behavior**: `mvp-feature-plan.md`
- **Permissions and safety considerations**: `permissions-and-safety-notes.md`
- **High-level architecture and Go layout**: `architecture-notes.md`

As new features are explored, additional focused docs (for example `feature-voting-system.md`) can be added alongside these.

