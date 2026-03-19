Course-correct the GuildMessageProxy project with mandatory QA and UX validation.

## Context

The project has existing functionality with known quality and usability gaps. Project-leaders tend to delegate only to developers while neglecting quality assurance and user experience validation.

## Phase 1: Discovery

Assign project-leaders to analyze the current codebase. Each project-leader must:

1. Identify issues within their category:
   - **Code Quality**: bugs, error handling, validation, test coverage
   - **UX/Product**: confusing flows, unclear messages, missing feedback, usability friction
   - **Architecture**: consistency, interfaces, technical debt
   - **Documentation**: accuracy, completeness, drift from implementation

2. Delegate discovery to applicable readonly subagents:
   - Use `.cursor/agents/quality-assurance.md` for feature completeness validation
   - Use `.cursor/agents/user-experience.md` for product usability evaluation
   - See `.cursor/agents/INDEX.md` for available readonly subagents

3. Require QA and UX subagents to evaluate all existing functionality comprehensively
4. Evaluate all documentation for accuracy against the actual implementation

## Phase 2: Action

When discovery reports return, direct project-leaders to:

1. Prioritize issues by impact (UX friction > code cleanup)
2. Plan fixes in sequenced order
3. Delegate implementation to `.cursor/agents/developer.md`
4. Update documentation progressively as changes are made
5. **Mandatory**: Every implementation task must be followed by QA and UX re-validation

## Phase 3: Continuous Validation

Enforce this workflow for every change:

```
Developer completes work → Update docs → QA validates completeness → UX validates usability
```

Project-leaders must either:
- Start QA and UX subagents themselves after developer work, OR
- Instruct developers to start QA/UX subagents before marking work complete

Documentation updates must accompany every change, not happen as a separate batch at the end.

## Manager Enforcement

Evaluate project-leaders continuously. Correct them immediately when they:

- Delegate only to developers without QA/UX coverage
- Accept developer work without validation reports
- Skip re-validation after fixes
- Fail to act on QA/UX report findings
- Neglect documentation updates or batch them at the end

Correction format: reference the specific instruction violated, state the issue, demand the fix.

## Deliverable

Report: discovered issues, corrections made, validation results from QA and UX subagents, documentation sync status.
