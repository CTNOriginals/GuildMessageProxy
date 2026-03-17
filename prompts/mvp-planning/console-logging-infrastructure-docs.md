Create console logging infrastructure documentation (no code implementation). Plan and execute in phases:

**Phase 1: Research**
Delegate to a researcher to:
- Read Go source files and identify current logging usage: where logs are written, what information is logged
- Research best practices: Go console logging standards, Discord bot logging patterns, common third-party libraries
- Document current state: existing logging patterns and what is logged (or not logged)

**Phase 2: Analysis**
Review research findings and determine:
- Logging levels (debug, info, warn, error, fatal) appropriate for each bot operation type
- Loggable events: startup/shutdown, commands, interactions, errors, guild lifecycle, API calls
- Output channels: stdout, stderr, optional Discord logging channel (when each applies)
- Contextual fields: timestamps, interaction IDs, guild/user/channel identifiers, operation context
- Integration: how logging fits into existing error handling in `internal/events/error.go` and error categorization in `docs/roadmap/infrastructure.md`

**Phase 3: Documentation Design**
Delegate to a documenter to create `docs/roadmap/logging-infrastructure.md`:
- **Overview**: Purpose and scope of console logging in GuildMessageProxy
- **Logging Levels**: When to use each level (debug/info/warn/error/fatal) with bot-specific examples
- **Loggable Events**: Categorized list of events to log (startup, commands, interactions, errors, guild lifecycle, API calls)
- **Contextual Information**: Standard fields to include in logs
- **Output Channels**: Where logs go (stdout, stderr, Discord channel) and when each is appropriate
- **Error Logging**: How errors are logged by category (transient vs permanent, matching `docs/roadmap/infrastructure.md` categorization)
- **Integration with Error Handling**: Logging's role in `internal/events` error flow
- **Log Format**: Recommended structure and readability guidelines
- **Post-MVP Extensions**: Planned enhancements (structured logging libraries, log aggregation, audit logs)

**Phase 4: Verification**
Delegate to a reviewer to verify:
- Documentation aligns with Go console logging best practices
- Logging strategy integrates with existing error handling in `internal/events`
- All major bot operations are covered
- Documentation guides future implementation without prescribing specific tools
- Cross-references and structure align with other roadmap documents

Oversee all phases; resolve blockers and ensure final documentation is thorough and actionable.
