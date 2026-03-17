# Subagent Index

| Agent | Use when |
|-------|----------|
| project-leader | Large, multi-step tasks requiring planning, delegation, and end-to-end oversight. Invoke only from user; subagents must not delegate to this agent. |
| documenter | Creating or updating project docs (README, ARCHITECTURE, ROUTE_MAP, GLOSSARY, TEMPLATES, roadmap) |
| developer | Building features, adding packages, extending the codebase |
| reviewer | Reviewing PRs, changes, or when the user asks for a code review |
| verifier | Validating completed work, confirming tests pass |
| tester | Running tests, analyzing failures, fixing test issues |
| researcher | Investigating how something works, finding where code lives, gathering context |

## Shared Instructions

All project subagents should apply these instructions:

- Follow the rules in `.cursor/rules/` where applicable.
- Use `docs/INDEX.md` to navigate and understand how things are structured.
- **Apply skills from `.cursor/skills/` when working on relevant tasks.** Skills contain guidance you follow directly - no need to invoke a subagent. For example, apply the `prompt-optimizer` skill when writing prompts.
- **Do not invoke the project-leader subagent.** The project-leader is started only by the user. Other subagents may not delegate to it.

## Documentation Delegation

When documentation updates are needed (new structure, new features, glossary changes, route updates, etc.), delegate to the **documenter** subagent. Do not handle documentation yourself unless you are the documenter. This keeps each agent focused on its expertise.
