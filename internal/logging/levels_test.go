package logging

import (
	"testing"
)

func TestParseLevel_ValidLevels(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"fatal", FatalLevel},
		{"error", ErrorLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"info", InfoLevel},
		{"debug", DebugLevel},
		{"FATAL", FatalLevel},
		{"ERROR", ErrorLevel},
		{"WARN", WarnLevel},
		{"WARNING", WarnLevel},
		{"INFO", InfoLevel},
		{"DEBUG", DebugLevel},
		{"Fatal", FatalLevel},
		{"Error", ErrorLevel},
		{"Warn", WarnLevel},
		{"Warning", WarnLevel},
		{"Info", InfoLevel},
		{"Debug", DebugLevel},
		{"FaTaL", FatalLevel},
		{"ErRoR", ErrorLevel},
		{"WaRn", WarnLevel},
		{"WaRnInG", WarnLevel},
		{"InFo", InfoLevel},
		{"DeBuG", DebugLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var level Level
			var err error
			level, err = ParseLevel(tt.input)
			if err != nil {
				t.Errorf("ParseLevel(%q) returned unexpected error: %v", tt.input, err)
			}
			if level != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, level, tt.expected)
			}
		})
	}
}

func TestParseLevel_InvalidLevels(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"unknown"},
		{"invalid"},
		{"trace"},
		{"verbose"},
		{"critical"},
		{"panic"},
		{"info "},
		{" info"},
		{"123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var level Level
			var err error
			level, err = ParseLevel(tt.input)
			if err == nil {
				t.Errorf("ParseLevel(%q) expected error, got nil", tt.input)
			}
			if level != InfoLevel {
				t.Errorf("ParseLevel(%q) = %v for invalid level, want %v", tt.input, level, InfoLevel)
			}
		})
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{FatalLevel, "[FATAL]"},
		{ErrorLevel, "[ERROR]"},
		{WarnLevel, "[WARN ]"},
		{InfoLevel, "[INFO ]"},
		{DebugLevel, "[DEBUG]"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			var result string
			result = tt.level.String()
			if result != tt.expected {
				t.Errorf("Level(%d).String() = %q, want %q", tt.level, result, tt.expected)
			}
		})
	}
}

func TestLevel_String_Length(t *testing.T) {
	// All level strings should be 7 characters: [ + 5 chars + ]
	var levels []Level
	levels = []Level{FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel}

	for _, level := range levels {
		var s string
		s = level.String()
		if len(s) != 7 {
			t.Errorf("Level.String() = %q, length = %d, want 7", s, len(s))
		}
	}
}

func TestLevel_isErrorOutput(t *testing.T) {
	tests := []struct {
		level    Level
		expected bool
	}{
		{FatalLevel, true},
		{ErrorLevel, true},
		{WarnLevel, true},
		{InfoLevel, false},
		{DebugLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			var result bool
			result = tt.level.isErrorOutput()
			if result != tt.expected {
				t.Errorf("Level(%d).isErrorOutput() = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}
