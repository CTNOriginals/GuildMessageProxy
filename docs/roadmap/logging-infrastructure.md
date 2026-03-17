# Console Logging Infrastructure

Design for terminal-based logging in GuildMessageProxy. Aligns with [infrastructure.md](./infrastructure.md) error handling and `internal/events/` error flow.

---

## 1. Overview

### Purpose and Scope

Console logging provides operational visibility for GuildMessageProxy during development, debugging, and production monitoring. It captures:

- Bot lifecycle events (startup, shutdown, connection state)
- Guild membership changes (join, leave)
- User interactions (commands, button clicks, modal submits)
- Storage operations (reads, writes, errors)
- Error conditions with context for troubleshooting

### Relationship to Error Handling

The logging system extends the error handling in `internal/events/`:

- `LogError()` helper currently writes errors to console via `log.Printf`
- Error categorization (per infrastructure.md) determines log level and handling
- `RespondWithError()` provides user feedback while logging captures technical details

### Philosophy

**Human-readable now, structured-ready for future.**

- MVP uses plain text logs via standard `log` package
- Messages should be clear and scannable for developers
- Design allows migration to `log/slog` for structured JSON logs post-MVP
- Include contextual fields (guild_id, user_id) in consistent positions

---

## 2. Logging Levels

Define severity hierarchy with bot-specific usage patterns.

| Level | When to Use | Bot Example |
|-------|-------------|-------------|
| **Fatal** | Unrecoverable startup failures | Missing `DISCORD_TOKEN`, invalid `.env` file, database connection failure on boot |
| **Error** | Operation failures requiring attention | REST API 500 errors, storage write failures, webhook creation failures |
| **Warn** | Recoverable issues, degraded operation | Rate limit hits, unknown command received, validation failures, message edit conflicts |
| **Info** | Normal operations | Bot started, command executed, guild joined, message posted successfully |
| **Debug** | Diagnostics for troubleshooting | Message routing decisions, storage cache hits/misses, API response details, interaction payload inspection |

### Level Selection Guidelines

**Fatal:** Use sparingly. Only when the bot cannot function and continuing would cause harm or confusion. Always exit immediately after logging.

**Error:** Something failed that affects user experience or data integrity. The bot continues running but an operator should investigate.

**Warn:** Something unexpected happened but the bot recovered. May indicate configuration issues or edge cases.

**Info:** Routine operations that confirm the bot is working correctly. Keep concise to avoid noise.

**Debug:** Detailed information useful only during development or specific investigations. Disabled in production.

---

## 3. Loggable Events

### Startup and Shutdown

| Event | Level | Fields |
|-------|-------|--------|
| Bot starting | Info | version, build_time, go_version |
| Configuration loaded | Info | env_file, storage_type |
| Discord session opened | Info | gateway_url |
| Commands registered | Info | count, scope (guild/global) |
| Ready event received | Info | session_id, connected_guilds |
| Shutdown initiated | Info | signal_received |
| Shutdown complete | Info | cleanup_duration_ms |

### Guild Lifecycle

| Event | Level | Fields |
|-------|-------|--------|
| GuildCreate received | Info | guild_id, guild_name, member_count |
| Guild config loaded | Debug | guild_id, allowed_roles_count |
| GuildDelete received | Info | guild_id, reason (if available) |
| Guild data cleaned up | Debug | guild_id, records_affected |

### Commands and Interactions

| Event | Level | Fields |
|-------|-------|--------|
| Interaction received | Debug | interaction_id, type, guild_id, user_id |
| Command execution started | Info | command_name, user_id, guild_id |
| Command execution completed | Info | command_name, duration_ms |
| Unknown interaction type | Warn | type, interaction_id |
| Button clicked | Info | button_id, user_id, guild_id |
| Modal submitted | Info | modal_id, user_id, guild_id |

### Storage Operations

| Event | Level | Fields |
|-------|-------|--------|
| Storage read | Debug | key, hit/miss |
| Storage write | Debug | key, size_bytes |
| Storage delete | Debug | key |
| Storage error | Error | operation, key, error_category |

### Errors

| Event | Level | Fields |
|-------|-------|--------|
| Transient error | Warn | category, retry_count, error_code |
| Permanent auth error | Fatal | error_code, context |
| Permanent resource error | Error | error_code, resource_type, resource_id |
| Validation error | Warn | error_code, field, user_id |
| REST API error | Error | endpoint, status_code, discord_error_code |

---

## 4. Contextual Information

### Standard Fields

Include these fields consistently across log entries:

**Temporal**
- `timestamp` - ISO8601 format with timezone (e.g., `2026-01-15T09:23:47Z`)

**Discord Context**
- `guild_id` - Server where interaction occurred
- `channel_id` - Channel where interaction occurred
- `user_id` - Discord user who triggered the interaction
- `interaction_id` - Discord interaction identifier for tracing

**Bot Context**
- `version` - Bot version string
- `build` - Build identifier or commit hash

**Error Context** (when logging errors)
- `category` - Error category per infrastructure.md (transient, permanent_auth, permanent_resource, validation)
- `retry_count` - Number of retry attempts for transient errors
- `discord_error_code` - Discord JSON error code when available

### Field Ordering

Recommended field order in log messages:

```
[LEVEL] [timestamp] message | field1=value1 field2=value2 ...
```

Example:
```
[INFO] [2026-01-15T09:23:47Z] Command executed | command=compose-create user_id=123456 guild_id=789012 duration_ms=45
```

---

## 5. Output Channels

### stdout and stderr

| Destination | Usage |
|-------------|-------|
| **stdout** | Info, Debug logs (normal and diagnostic output) |
| **stderr** | Warn, Error, Fatal logs (problems requiring attention) |

This separation allows operators to:
- Pipe stdout to /dev/null in production (suppress debug)
- Forward stderr to alerting systems
- Filter logs by severity in container environments

### Future Discord Channel Integration

Post-MVP feature: forward select logs to a Discord channel.

| Log Type | Channel | Format |
|----------|---------|--------|
| Errors | `#bot-errors` | Rich embed with error details, stack context |
| Guild joins/leaves | `#bot-activity` | Compact text notification |
| Security events | `#bot-security` | Embeds with user IDs, action taken |

---

## 6. Error Logging by Category

Aligns with [infrastructure.md](./infrastructure.md) section 5 error categorization.

### Transient Errors

**Level:** Warn

**Examples:** 429 (rate limit), 502 (server error), 503 (unavailable)

**Fields:**
- `retry_count` - Current retry attempt
- `max_retries` - Maximum retry limit
- `backoff_ms` - Current backoff duration
- `discord_error_code` - Discord JSON error code

**Example:**
```
[WARN] [2026-01-15T09:23:47Z] Transient error, will retry | category=transient retry_count=2 max_retries=3 backoff_ms=1000 discord_error_code=429 endpoint=/channels/123/messages
```

### Permanent Auth Errors

**Level:** Fatal

**Examples:** 40001 (unauthorized), 40004 (disallowed gateway intent), invalid token

**Fields:**
- `discord_error_code` - Discord error code
- `context` - What operation triggered the error

**Example:**
```
[FATAL] [2026-01-15T09:23:47Z] Authentication failed, shutting down | discord_error_code=40001 context=session.Open
```

**Action:** Log and exit. These indicate configuration issues that require operator intervention.

### Permanent Resource Errors

**Level:** Error

**Examples:** 10003 (unknown channel), 10008 (unknown message), 10013 (unknown user)

**Fields:**
- `discord_error_code` - Discord error code
- `resource_type` - Type of resource (channel, message, user, guild)
- `resource_id` - ID of the missing resource
- `action_taken` - How the bot handled it (cleared_message, ignored, notified_user)

**Example:**
```
[ERROR] [2026-01-15T09:23:47Z] Resource not found | category=permanent_resource discord_error_code=10008 resource_type=message resource_id=456789 action_taken=cleared_user_message
```

### Validation Errors

**Level:** Warn

**Examples:** 50035 (invalid form body), 50016 (too many attachments)

**Fields:**
- `discord_error_code` - Discord error code
- `field` - Field that failed validation (if known)
- `user_id` - User who submitted invalid data
- `validation_message` - Human-readable validation failure

**Example:**
```
[WARN] [2026-01-15T09:23:47Z] Validation failed | category=validation discord_error_code=50035 field=content user_id=123456 validation_message=content exceeds 2000 characters
```

---

## 7. Integration with Error Handling

### Extending LogError()

The current `LogError()` in `internal/events/` accepts `(err error, context string)`. Extend to support error categorization:

```go
// ErrorCategory aligns with infrastructure.md categorization
type ErrorCategory string

const (
    CategoryTransient       ErrorCategory = "transient"
    CategoryPermanentAuth   ErrorCategory = "permanent_auth"
    CategoryPermanentResource ErrorCategory = "permanent_resource"
    CategoryValidation      ErrorCategory = "validation"
)

// LogError logs an error with category and retry context
// Note: This is a proposed future extension; current implementation uses LogError(err, context)
func LogError(err error, category ErrorCategory, context string, retryCount int)
```

### Integration with RespondWithError()

Current flow:
1. Error occurs
2. `LogError()` writes to console
3. `RespondToUser()` sends ephemeral message
4. User is informed

Enhanced flow with categorization:
1. Error occurs
2. Categorize error (transient/permanent/validation)
3. `LogError()` writes with appropriate level and fields
4. `RespondWithError()` sends user-friendly message
5. If transient, retry logic executes

### REST Call Wrappers

Add logging to REST API wrappers in `internal/events/`:

```go
// Example wrapper pattern
func LoggedRESTCall(operation string, call func() error) error {
    err := call()
    if err != nil {
        category := CategorizeDiscordError(err)
        LogError(err, category, operation, 0)
    }
    return err
}
```

---

## 8. Log Format

### Timestamp Format

Use ISO8601 with UTC timezone and millisecond precision:

```
2006-01-02T15:04:05.000Z
```

Example: `2026-01-15T09:23:47.123Z`

### Level Display

Bracketed uppercase, 5 characters wide (left-aligned):

```
[FATAL]
[ERROR]
[WARN ]
[INFO ]
[DEBUG]
```

### Message Format

**Standard entry:**
```
[LEVEL] [timestamp] message | field1=value1 field2=value2
```

**Multi-line messages:**
For errors with stack traces or large payloads, use newline continuation with indent:

```
[ERROR] [2026-01-15T09:23:47Z] REST API failure |
    endpoint=/channels/123/messages
    status=500
    response={"message": "Internal Server Error"}
```

### Field Ordering Convention

1. Core identifiers (command, interaction_id)
2. Discord context (guild_id, channel_id, user_id)
3. Operation details (duration_ms, retry_count)
4. Error details (category, discord_error_code)

---

## 9. Post-MVP Extensions

### Structured Logging with log/slog

Migrate to `log/slog` (Go 1.21+ standard library) for JSON output:

```go
// Example structured log
slog.Info("command executed",
    slog.String("command", "compose-create"),
    slog.String("user_id", userID),
    slog.Duration("duration", duration),
)
```

**Benefits:**
- Machine-parseable output
- Consistent field typing
- Built-in level filtering
- Handler ecosystem (file, syslog, external services)

### Log Aggregation

Integration with log aggregation services:

- File-based logging with rotation
- Structured JSON output for log shippers
- Correlation IDs for distributed tracing
- Log level configuration via environment variable

### Discord Logging Channel

Forward select log levels to Discord channels:

| Feature | Implementation |
|---------|----------------|
| Error embeds | Rich embeds with stack traces, user context |
| Guild activity | Compact notifications for joins/leaves |
| Security alerts | Immediate notification for auth failures, suspicious patterns |
| Configurable levels | Per-channel log level filtering |

### Audit Logs

Persistent audit trail for compliance:

- Who posted what message and when
- Edit history with before/after snapshots
- Deletion records with actor and reason
- Retention policy alignment with guild requirements

---

## See Also

- [infrastructure.md](./infrastructure.md) - Error categorization and handling flow
- `internal/events/` - Error handling implementation
- [ARCHITECTURE.md](../ARCHITECTURE.md) - Package layout and conventions
