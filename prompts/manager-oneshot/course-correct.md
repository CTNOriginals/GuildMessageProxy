Course-correct the entire GuildMessageProxy project.

## Context

A previous manager initiated changes but ran out of tokens before completion. The current codebase has significant QA and UX issues that require correction. Do not add new features - focus exclusively on refining and fixing what already exists.

## Your Task

1. Analyze the current codebase state to identify all QA and UX issues:
   - Code quality problems (bugs, error handling, edge cases, test coverage)
   - UX problems (confusing flows, missing feedback, poor error messages, unclear UI)
   - Architectural inconsistencies or technical debt
   - Missing or incomplete documentation

2. Decompose issues into work categories:
   - **Code Quality**: bugs, error handling, validation, edge cases, tests
   - **UX Polish**: message clarity, button labels, error messages, user feedback
   - **Architecture**: consistency, interfaces, cleanup
   - **Documentation**: code comments, user-facing docs

3. Assign one project-leader per category using `.cursor/agents/project-leader.md`. Instruct each to:
   - Identify specific issues within their category
   - Plan fixes in sequenced order
   - Delegate to appropriate subagents
   - Report progress and escalate conflicts
   - Focus only on existing features: `/compose create`, `/compose post`, `/compose propose`

4. Coordinate the project-leaders:
   - Track deliverables from each category
   - Resolve cross-cutting dependencies
   - Synchronize fixes that touch shared code

5. Review integrated outputs and deliver a summary of all corrections made.

## Constraints

- No new features - only refine existing `/compose` commands
- Fix must be comprehensive across all categories
- Each project-leader stays within their assigned boundary
- Use `.cursor/skills/compose-prompt` for all subagent delegation
