# Implement Console Logging Infrastructure

Implement console logging across GuildMessageProxy following the design in `docs/roadmap/logging-infrastructure.md`. Use standard `log` package (not log/slog). Maintain backward compatibility with existing `LogError(err, context)` signature.

## Phase 1: Create `internal/logging/` Package

**Files to create:**
- `internal/logging/level.go` - level type and constants
- `internal/logging/logger.go` - Logger interface and stdlib implementation
- `internal/logging/fields.go` - field helper functions

**Requirements:**
- Define `type Level uint8` with constants: `LevelFatal`, `LevelError`, `LevelWarn`, `LevelInfo`, `LevelDebug`
- Logger interface: `Log(level Level, message string, fields map[string]string)`
- Timestamp format: ISO8601 UTC with millisecond precision (`2006-01-02T15:04:05.000Z`)
- Log format: `[LEVEL] [timestamp] message | field1=value1 field2=value2`
- Output: stdout for Info/Debug, stderr for Warn/Error/Fatal
- Level strings: 5-char uppercase (`[FATAL]`, `[ERROR]`, `[WARN ]`, `[INFO ]`, `[DEBUG]`)

**Success criteria:**
- Package compiles
- `log.New(os.Stdout, "", 0)` and `log.New(os.Stderr, "", 0)` for writers
- `Logger.Info("test", map[string]string{"key": "value"})` outputs: `[INFO ] [2026-01-15T09:23:47.123Z] test | key=value`

## Phase 2: Extend `internal/events/error.go`

**Files to modify:**
- `internal/events/error.go`

**Add:**
```go
// ErrorCategory aligns with error handling design
type ErrorCategory string

const (
    CategoryTransient         ErrorCategory = "transient"
    CategoryPermanentAuth     ErrorCategory = "permanent_auth"
    CategoryPermanentResource ErrorCategory = "permanent_resource"
    CategoryValidation        ErrorCategory = "validation"
)
```

**Requirements:**
- Maintain existing `LogError(err error, context string)` signature - do not break callers
- Add new `LogErrorWithCategory(err error, category ErrorCategory, context string, fields map[string]string)`
- Category determines log level:
  - `permanent_auth` -> Fatal (and exit)
  - `permanent_resource` -> Error
  - `transient` -> Warn
  - `validation` -> Warn
- Map Discord error codes (10003, 10008, 40001, etc.) to categories

**Success criteria:**
- Existing `LogError(err, context)` calls compile unchanged
- New category-aware logging works with appropriate levels
- `RespondWithError()` calls appropriate logging function

## Phase 3: Add Logging to `cmd/bot/main.go`

**Files to modify:**
- `cmd/bot/main.go`

**Add Info-level logs for:**
- Bot starting (fields: version, go_version)
- Configuration loaded (fields: env_file, storage_type)
- Discord session opened (fields: gateway_url)
- Commands registered (fields: count, scope)
- Ready event received (fields: session_id, connected_guilds)
- Shutdown initiated (fields: signal_received)
- Shutdown complete (fields: cleanup_duration_ms)

**Requirements:**
- Create package-level logger: `var logger logging.Logger`
- Initialize logger in `init()` or early in `main()`
- Pass logger through to other packages or use global (for MVP simplicity)

**Success criteria:**
- Startup produces log output with all fields listed
- Shutdown produces log output with all fields listed
- Logs follow format: `[INFO ] [timestamp] Bot starting | version=... go_version=...`

## Phase 4: Add Logging to `internal/events/` Handlers

**Files to modify:**
- `internal/events/guild_create.go`
- `internal/events/guild_delete.go`
- `internal/events/interaction_create.go`
- `internal/events/ready.go`

**GuildCreate handler:**
- Info: Guild joined (guild_id, guild_name, member_count)
- Debug: Guild config loaded (guild_id, allowed_roles_count)

**GuildDelete handler:**
- Info: Guild left (guild_id)
- Debug: Guild data cleaned (guild_id, records_affected)

**InteractionCreate handler:**
- Debug: Interaction received (interaction_id, type, guild_id, user_id)
- Info: Command execution started (command_name, user_id, guild_id)
- Info: Command execution completed (command_name, duration_ms)
- Warn: Unknown interaction type (type, interaction_id)
- Info: Button clicked (button_id, user_id, guild_id)
- Info: Modal submitted (modal_id, user_id, guild_id)

**Ready handler:**
- Info: Bot ready (already logged in main.go; add session_id here too)

**Success criteria:**
- Each handler produces appropriate logs at correct levels
- Field order: command/interaction_id first, then Discord context, then details
- Duration timing for command execution (time.Since or similar)

## Phase 5: Add Logging to `internal/commands/definitions.go`

**Files to modify:**
- `internal/commands/definitions.go` (or where command routing happens)

**Requirements:**
- Log command execution with category-aware error logging
- Log errors during command execution with proper ErrorCategory
- Use fields: command_name, user_id, guild_id, error (if error)

**Success criteria:**
- Successful command: `[INFO ] ... Command completed | command_name=...`
- Failed command: `[WARN ] ... Command failed | category=validation error_code=50035`

## Phase 6: Verification and Testing

**Verification steps:**
1. Run bot and verify startup logs appear correctly
2. Trigger a command and verify execution logs
3. Trigger an error (e.g., invalid input) and verify Warn logs with category
4. Check that Fatal logs exit the bot (simulate missing DISCORD_TOKEN)
5. Verify stdout/stderr separation: `go run ./cmd/bot 2>/dev/null` suppresses errors
6. Check existing tests still pass

**Review against design doc:**
- All log levels implemented (Fatal, Error, Warn, Info, Debug)
- All error categories implemented
- Timestamp format matches ISO8601 with milliseconds
- Field format matches `key=value` pairs
- Output channels correct (stdout/stderr)

**Commit the work:**
- Use conventional commits per `.cursor/rules/git-commit.mdc`
- Example: `feat(logging): implement console logging infrastructure`
- Include scope: `feat(internal/logging):`, `feat(events):`, `feat(cmd/bot):`

## Key Constraints

- Use standard `log` package, not log/slog
- Keep existing `LogError(err, context)` working
- Follow Go conventions in `.cursor/rules/golang-conventions.mdc`
- All changes must compile and existing tests pass
- MVP: global logger instance acceptable (no dependency injection needed)
