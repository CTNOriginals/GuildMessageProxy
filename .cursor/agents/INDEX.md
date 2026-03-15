# Subagent Index

| Agent | Use when |
|-------|----------|
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

## Documentation Delegation

When documentation updates are needed (new structure, new features, glossary changes, route updates, etc.), delegate to the **documenter** subagent. Do not handle documentation yourself unless you are the documenter. This keeps each agent focused on its expertise.
