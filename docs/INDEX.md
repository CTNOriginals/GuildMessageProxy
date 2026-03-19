# GuildMessageProxy - Agent Documentation Index

This folder contains documentation to help AI agents and developers navigate the project. Start here when onboarding to the codebase.

## Quick Links


| Document                             | Purpose                                                       |
| ------------------------------------ | ------------------------------------------------------------- |
| [PROJECT_STATUS.md](./PROJECT_STATUS.md) | **Current status, what's being worked on, and backlog** |
| [PROJECT_MAP.md](./PROJECT_MAP.md)   | Directory structure, where files live, what exists vs planned |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Package layout, conventions, entry points, command sync, dependencies |
| [ROUTE_MAP.md](./ROUTE_MAP.md)       | Command routes, flows, handler wiring                         |
| [TEMPLATES.md](./TEMPLATES.md)       | File templates, patterns for adding new features              |
| [GLOSSARY.md](./GLOSSARY.md)         | Terms and jargon (proxy message, compose, ephemeral, etc)     |
| [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) | Common issues, error codes, debugging, permissions           |
| [roadmap/](./roadmap/)               | Historical planning docs (MVP design, infrastructure notes) |
| [.cursor/agents/INDEX.md](../.cursor/agents/INDEX.md) | Subagent index - when to use documenter, developer, reviewer, etc |


## When to Use What

- **"What is the current project status?"** -> [PROJECT_STATUS.md](./PROJECT_STATUS.md)
- **"What is being worked on and what's next?"** -> [PROJECT_STATUS.md](./PROJECT_STATUS.md#currently-in-progress)
- **"Where does X go?"** -> [PROJECT_MAP.md](./PROJECT_MAP.md)
- **"Which subagent should handle this?"** -> [.cursor/agents/INDEX.md](../.cursor/agents/INDEX.md)
- **"How is the code organized?"** -> [ARCHITECTURE.md](./ARCHITECTURE.md)
- **"What commands exist and how do they flow?"** -> [ROUTE_MAP.md](./ROUTE_MAP.md)
- **"How do I add a new command/feature?"** -> [TEMPLATES.md](./TEMPLATES.md)
- **"What does X mean?"** -> [GLOSSARY.md](./GLOSSARY.md)
- **"Something is broken, how do I fix it?"** -> [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
- **"What was the original MVP design?"** -> [roadmap/mvp-feature-plan.md](./roadmap/mvp-feature-plan.md) (historical reference)
- **"What is the full infrastructure design?"** -> [roadmap/infrastructure.md](./roadmap/infrastructure.md)
- **"How does event routing and infrastructure work?"** -> [ARCHITECTURE.md](./ARCHITECTURE.md#internal-events), [ROUTE_MAP.md](./ROUTE_MAP.md#interaction-routing), [roadmap/infrastructure.md](./roadmap/infrastructure.md)


