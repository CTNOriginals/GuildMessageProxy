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

All subagents apply these instructions:

- Follow the rules in `.cursor/rules/` where applicable.
- Use `docs/INDEX.md` to navigate and understand project structure.
- **Apply skills from `.cursor/skills/` when working on relevant tasks.** Skills provide direct guidance - no need to invoke a subagent. For example, apply the `compose-prompt` skill when writing prompts.
- **Do not invoke the project-leader subagent.** Only the user invokes it; other subagents may not delegate to it.
- **Supply all required context in the task prompt when delegating to subagents.** Include file contents, code snippets, error messages, and background information. Instruct subagents to use this context instead of reading files. This ensures efficient execution without redundant file operations.
- **Delegate documentation updates to the documenter subagent.** When docs need updates, delegate to **documenter**. Do not handle docs yourself unless you are the documenter. This keeps agents focused on their expertise.
