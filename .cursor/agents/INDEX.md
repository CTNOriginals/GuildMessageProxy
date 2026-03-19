# Subagent Index

| Agent | Use when |
|-------|----------|
| manager | Overseeing entire projects; coordinates multiple project-leaders for large initiatives |
| project-leader | Complex multi-step tasks requiring planning, delegation, and end-to-end oversight |
| researcher | Investigating how something works, finding where code lives, gathering context |
| documenter | Creating or updating project docs |
| developer | Building features, adding packages, extending the codebase |
| tester | Running tests, analyzing failures, fixing test issues |
| verifier | Validating completed work, confirming tests pass |
| reviewer | Reviewing PRs, changes, or when the user asks for a code review |
| quality-assurance | New features introduced; validating requirement completeness and release readiness |
| user-experience | Assessing feature usability through hands-on simulation and user workflow testing |

## Shared Instructions

All subagents apply these instructions:

- Follow the rules in `.cursor/rules/` where applicable.
- Use `docs/INDEX.md` to navigate and understand project structure.
- **Apply skills from `.cursor/skills/` when working on relevant tasks.** Skills provide direct guidance - no need to invoke a subagent. For example, apply the `compose-prompt` skill when writing prompts.
- **Supply all required context in the task prompt when delegating to subagents.** Include file contents, code snippets, error messages, and background information. Instruct subagents to use this context instead of reading files. This ensures efficient execution without redundant file operations.
- **Delegate documentation updates to the documenter subagent.** When docs need updates, delegate to **documenter**. Do not handle docs yourself unless you are the documenter. This keeps agents focused on their expertise.
