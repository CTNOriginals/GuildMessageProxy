package logging

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Field represents a structured key-value pair for log entries.
type Field struct {
	Key   string
	Value string
}

// String creates a string field.
func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

// Int creates an integer field.
func Int(key string, val int) Field {
	return Field{Key: key, Value: fmt.Sprintf("%d", val)}
}

// Int64 creates an int64 field.
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: fmt.Sprintf("%d", val)}
}

// Duration creates a duration field in milliseconds.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: fmt.Sprintf("%d", val.Milliseconds())}
}

// Err creates an error field. If err is nil, value is "<nil>".
func Err(key string, err error) Field {
	if err == nil {
		return Field{Key: key, Value: "<nil>"}
	}
	return Field{Key: key, Value: err.Error()}
}

// WithContext creates standard Discord context fields.
// Convenience helper for the common guild_id, channel_id, user_id combination.
func WithContext(guildID, channelID, userID string) []Field {
	fields := make([]Field, 0, 3)
	if guildID != "" {
		fields = append(fields, String("guild_id", guildID))
	}
	if channelID != "" {
		fields = append(fields, String("channel_id", channelID))
	}
	if userID != "" {
		fields = append(fields, String("user_id", userID))
	}
	return fields
}

// logger provides thread-safe logging with level filtering.
type logger struct {
	mu        sync.RWMutex
	minLevel  Level
	formatter func(time.Time, Level, string, []Field) string
}

// defaultLogger is the package-level logger instance.
var defaultLogger = &logger{
	minLevel:  InfoLevel,
	formatter: defaultFormatter,
}

// defaultFormatter formats log entries according to the specification:
// [LEVEL] [timestamp] message
//     field1: value1
//     field2: value2
// Timestamp is ISO8601 UTC with millisecond precision.
func defaultFormatter(t time.Time, level Level, msg string, fields []Field) string {
	// Timestamp format: YY-MM-DD HH:MM:SS
	timestamp := t.UTC().Format("06-01-02 15:04:05")

	// Build the base message: [LEVEL] [timestamp] message
	result := fmt.Sprintf("%s [%s] %s", level.String(), timestamp, msg)

	// Append fields if present, each on its own line with 4-space indent
	if len(fields) > 0 {
		for _, f := range fields {
			result += fmt.Sprintf("\n    %s: %s", f.Key, f.Value)
		}
	}

	return result
}

// SetLevel sets the minimum log level. Entries below this level are discarded.
// Default is InfoLevel (Debug is disabled).
func SetLevel(level Level) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.minLevel = level
}

// GetLevel returns the current minimum log level.
func GetLevel() Level {
	defaultLogger.mu.RLock()
	defer defaultLogger.mu.RUnlock()
	return defaultLogger.minLevel
}

// shouldLog returns true if the given level should be logged.
func (l *logger) shouldLog(level Level) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level <= l.minLevel
}

// log writes a log entry at the specified level.
func (l *logger) log(level Level, msg string, fields []Field) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.RLock()
	formatter := l.formatter
	l.mu.RUnlock()

	formatted := formatter(time.Now(), level, msg, fields)

	// Write to appropriate output: stdout for Info/Debug, stderr for others
	if level.isErrorOutput() {
		fmt.Fprintln(os.Stderr, formatted)
	} else {
		fmt.Fprintln(os.Stdout, formatted)
	}
}

// Fatal logs an unrecoverable startup failure and exits the application.
// Must only be used for conditions where continuing would cause harm or confusion.
func Fatal(msg string, fields ...Field) {
	defaultLogger.log(FatalLevel, msg, fields)
	os.Exit(1)
}

// Error logs an operation failure requiring attention.
// The bot continues running but an operator should investigate.
func Error(msg string, fields ...Field) {
	defaultLogger.log(ErrorLevel, msg, fields)
}

// Warn logs a recoverable issue or degraded operation.
// Something unexpected happened but the bot recovered.
func Warn(msg string, fields ...Field) {
	defaultLogger.log(WarnLevel, msg, fields)
}

// Info logs normal operations that confirm the bot is working correctly.
func Info(msg string, fields ...Field) {
	defaultLogger.log(InfoLevel, msg, fields)
}

// Debug logs detailed diagnostics for troubleshooting.
// Disabled by default; enable by setting LOG_LEVEL=debug or calling SetLevel(DebugLevel).
func Debug(msg string, fields ...Field) {
	defaultLogger.log(DebugLevel, msg, fields)
}

// init reads LOG_LEVEL environment variable to configure initial log level.
// Supported values: fatal, error, warn, info, debug (case-insensitive).
// Invalid or unset values default to InfoLevel.
func init() {
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		if level, err := ParseLevel(levelStr); err == nil {
			SetLevel(level)
		}
	}
}
