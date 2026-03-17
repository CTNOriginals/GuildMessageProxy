// Package logging provides structured console logging for GuildMessageProxy.
// It supports five log levels (Fatal, Error, Warn, Info, Debug) with configurable
// minimum level and environment-based configuration.
package logging

import (
	"fmt"
	"strings"
)

// Level represents the severity of a log entry.
type Level int

const (
	// FatalLevel is for unrecoverable startup failures. The application must exit after logging.
	FatalLevel Level = iota
	// ErrorLevel is for operation failures requiring attention.
	ErrorLevel
	// WarnLevel is for recoverable issues and degraded operation.
	WarnLevel
	// InfoLevel is for normal operations.
	InfoLevel
	// DebugLevel is for diagnostics and troubleshooting. Disabled in production.
	DebugLevel
)

// levelNames maps levels to their uppercase display names.
// Names are 4 characters for alignment: FATAL, ERROR, WARN, INFO, DEBUG
var levelNames = map[Level]string{
	FatalLevel: "FATAL",
	ErrorLevel: "ERROR",
	WarnLevel:  "WARN",
	InfoLevel:  "INFO",
	DebugLevel: "DEBUG",
}

// levelOutputs maps levels to their output destination.
// stdout receives Info and Debug; stderr receives Warn, Error, and Fatal.
var levelOutputs = map[Level]outputDestination{
	FatalLevel: stderr,
	ErrorLevel: stderr,
	WarnLevel:  stderr,
	InfoLevel:  stdout,
	DebugLevel: stdout,
}

// outputDestination represents where log entries are written.
type outputDestination int

const (
	stdout outputDestination = iota
	stderr
)

// String returns the bracketed, 5-character-wide level name for display.
// Example: [FATAL], [ERROR], [WARN ], [INFO ], [DEBUG]
func (l Level) String() string {
	name := levelNames[l]
	// Pad to 5 characters for alignment
	return fmt.Sprintf("[%s]", name+strings.Repeat(" ", 5-len(name)))
}

// isErrorOutput returns true if this level writes to stderr.
func (l Level) isErrorOutput() bool {
	return levelOutputs[l] == stderr
}

// ParseLevel converts a string level name to a Level.
// Case-insensitive. Returns InfoLevel and an error for unknown levels.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	default:
		return InfoLevel, fmt.Errorf("unknown log level: %s", s)
	}
}
