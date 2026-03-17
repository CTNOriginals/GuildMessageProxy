# GuildMessageProxy - Agent Documentation Index

This folder contains documentation to help AI agents and developers navigate the project. Start here when onboarding to the codebase.

## Quick Links


| Document                             | Purpose                                                       |
| ------------------------------------ | ------------------------------------------------------------- |
| [PROJECT_MAP.md](./PROJECT_MAP.md)   | Directory structure, where files live, what exists vs planned |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Package layout, conventions, entry points, command sync, dependencies |
| [ROUTE_MAP.md](./ROUTE_MAP.md)       | Command routes, flows, handler wiring                         |
| [TEMPLATES.md](./TEMPLATES.md)       | File templates, patterns for adding new features              |
| [GLOSSARY.md](./GLOSSARY.md)         | Terms and jargon (proxy message, compose, ephemeral, etc)     |
| [roadmap/](./roadmap/)               | Planning docs - feature plans, architecture notes, infrastructure, safety |
| [.cursor/agents/INDEX.md](../.cursor/agents/INDEX.md) | Subagent index - when to use documenter, developer, reviewer, etc |


## When to Use What

- **"Where does X go?"** -> [PROJECT_MAP.md](./PROJECT_MAP.md)
- **"Which subagent should handle this?"** -> [.cursor/agents/INDEX.md](../.cursor/agents/INDEX.md)
- **"How is the code organized?"** -> [ARCHITECTURE.md](./ARCHITECTURE.md)
- **"What commands exist and how do they flow?"** -> [ROUTE_MAP.md](./ROUTE_MAP.md)
- **"How do I add a new command/feature?"** -> [TEMPLATES.md](./TEMPLATES.md)
- **"What does X mean?"** -> [GLOSSARY.md](./GLOSSARY.md)
- **"What is planned for this project?"** -> [roadmap/](./roadmap/)
- **"What is the full infrastructure design?"** -> [roadmap/infrastructure.md](./roadmap/infrastructure.md)
- **"How does event routing and infrastructure work?"** -> [ARCHITECTURE.md](./ARCHITECTURE.md#internal-events), [ROUTE_MAP.md](./ROUTE_MAP.md#interaction-routing), [roadmap/infrastructure.md](./roadmap/infrastructure.md)

## Project State

As of the last doc update, the **infrastructure is COMPLETE**. The bot has:

- Full event handlers (Ready, GuildCreate, GuildDelete, InteractionCreate, Error)
- Command sync system with diff detection (`--guild`, `--global`, `--no-sync` flags)
- Interaction type system for command/data routing
- Storage interface with in-memory implementation
- Placeholder `/compose` command registered
- Graceful shutdown with runtime logging

The **only remaining work for MVP** is the `internal/handlers/` package containing the compose, preview, and post logic. See [roadmap/mvp-feature-plan.md](./roadmap/mvp-feature-plan.md) for handler specifications.

## Roadmap Subfolder

The `roadmap/` folder holds planning documents:

- `overview.md` - Purpose, capabilities, target audience
- `mvp-feature-plan.md` - MVP flows, compose/post/edit behavior
- `architecture-notes.md` - Go package layout, Discord integration
- `permissions-and-safety-notes.md` - Risks, guardrails, governance

