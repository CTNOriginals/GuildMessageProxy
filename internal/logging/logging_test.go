package logging

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	var f Field
	f = String("key", "value")
	if f.Key != "key" {
		t.Errorf("String() Key = %q, want %q", f.Key, "key")
	}
	if f.Value != "value" {
		t.Errorf("String() Value = %q, want %q", f.Value, "value")
	}
}

func TestInt(t *testing.T) {
	var f Field
	f = Int("count", 42)
	if f.Key != "count" {
		t.Errorf("Int() Key = %q, want %q", f.Key, "count")
	}
	if f.Value != "42" {
		t.Errorf("Int() Value = %q, want %q", f.Value, "42")
	}
}

func TestInt64(t *testing.T) {
	var f Field
	f = Int64("timestamp", 1234567890123)
	if f.Key != "timestamp" {
		t.Errorf("Int64() Key = %q, want %q", f.Key, "timestamp")
	}
	if f.Value != "1234567890123" {
		t.Errorf("Int64() Value = %q, want %q", f.Value, "1234567890123")
	}
}

func TestDuration(t *testing.T) {
	var f Field
	f = Duration("latency", 1500*time.Millisecond)
	if f.Key != "latency" {
		t.Errorf("Duration() Key = %q, want %q", f.Key, "latency")
	}
	if f.Value != "1500" {
		t.Errorf("Duration() Value = %q, want %q", f.Value, "1500")
	}
}

func TestDuration_Zero(t *testing.T) {
	var f Field
	f = Duration("elapsed", 0)
	if f.Value != "0" {
		t.Errorf("Duration(0) Value = %q, want %q", f.Value, "0")
	}
}

func TestDuration_Seconds(t *testing.T) {
	var f Field
	f = Duration("duration", 5*time.Second)
	if f.Value != "5000" {
		t.Errorf("Duration(5s) Value = %q, want %q", f.Value, "5000")
	}
}

func TestErr_WithError(t *testing.T) {
	var testErr error
	testErr = errors.New("something went wrong")
	var f Field
	f = Err("error", testErr)
	if f.Key != "error" {
		t.Errorf("Err() Key = %q, want %q", f.Key, "error")
	}
	if f.Value != "something went wrong" {
		t.Errorf("Err() Value = %q, want %q", f.Value, "something went wrong")
	}
}

func TestErr_WithNil(t *testing.T) {
	var f Field
	f = Err("error", nil)
	if f.Key != "error" {
		t.Errorf("Err(nil) Key = %q, want %q", f.Key, "error")
	}
	if f.Value != "<nil>" {
		t.Errorf("Err(nil) Value = %q, want %q", f.Value, "<nil>")
	}
}

func TestWithContext_AllFields(t *testing.T) {
	var fields []Field
	fields = WithContext("123456789", "987654321", "111222333")
	if len(fields) != 3 {
		t.Errorf("WithContext() returned %d fields, want 3", len(fields))
	}

	var foundGuildID, foundChannelID, foundUserID bool
	for _, f := range fields {
		switch f.Key {
		case "guild_id":
			foundGuildID = true
			if f.Value != "123456789" {
				t.Errorf("guild_id = %q, want %q", f.Value, "123456789")
			}
		case "channel_id":
			foundChannelID = true
			if f.Value != "987654321" {
				t.Errorf("channel_id = %q, want %q", f.Value, "987654321")
			}
		case "user_id":
			foundUserID = true
			if f.Value != "111222333" {
				t.Errorf("user_id = %q, want %q", f.Value, "111222333")
			}
		default:
			t.Errorf("Unexpected field key: %q", f.Key)
		}
	}

	if !foundGuildID {
		t.Error("WithContext() missing guild_id field")
	}
	if !foundChannelID {
		t.Error("WithContext() missing channel_id field")
	}
	if !foundUserID {
		t.Error("WithContext() missing user_id field")
	}
}

func TestWithContext_EmptyGuildID(t *testing.T) {
	var fields []Field
	fields = WithContext("", "987654321", "111222333")
	if len(fields) != 2 {
		t.Errorf("WithContext() returned %d fields, want 2", len(fields))
	}

	for _, f := range fields {
		if f.Key == "guild_id" {
			t.Error("WithContext() should not include guild_id when empty")
		}
	}
}

func TestWithContext_EmptyChannelID(t *testing.T) {
	var fields []Field
	fields = WithContext("123456789", "", "111222333")
	if len(fields) != 2 {
		t.Errorf("WithContext() returned %d fields, want 2", len(fields))
	}

	for _, f := range fields {
		if f.Key == "channel_id" {
			t.Error("WithContext() should not include channel_id when empty")
		}
	}
}

func TestWithContext_EmptyUserID(t *testing.T) {
	var fields []Field
	fields = WithContext("123456789", "987654321", "")
	if len(fields) != 2 {
		t.Errorf("WithContext() returned %d fields, want 2", len(fields))
	}

	for _, f := range fields {
		if f.Key == "user_id" {
			t.Error("WithContext() should not include user_id when empty")
		}
	}
}

func TestWithContext_AllEmpty(t *testing.T) {
	var fields []Field
	fields = WithContext("", "", "")
	if len(fields) != 0 {
		t.Errorf("WithContext() returned %d fields, want 0", len(fields))
	}
}

func TestDefaultFormatter_NoFields(t *testing.T) {
	var ts time.Time
	ts = time.Date(2024, 3, 18, 14, 30, 45, 0, time.UTC)
	var result string
	result = defaultFormatter(ts, InfoLevel, "test message", nil)

	var expectedPrefix string
	expectedPrefix = "[INFO ] [24-03-18 14:30:45] test message"
	if result != expectedPrefix {
		t.Errorf("defaultFormatter() = %q, want %q", result, expectedPrefix)
	}
}

func TestDefaultFormatter_WithFields(t *testing.T) {
	var ts time.Time
	ts = time.Date(2024, 3, 18, 14, 30, 45, 0, time.UTC)
	var fields []Field
	fields = []Field{
		String("key1", "value1"),
		Int("count", 42),
	}
	var result string
	result = defaultFormatter(ts, ErrorLevel, "error occurred", fields)

	var lines []string
	lines = strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Errorf("defaultFormatter() returned %d lines, want 3", len(lines))
	}

	var expectedFirstLine string
	expectedFirstLine = "[ERROR] [24-03-18 14:30:45] error occurred"
	if lines[0] != expectedFirstLine {
		t.Errorf("defaultFormatter() first line = %q, want %q", lines[0], expectedFirstLine)
	}

	var expectedSecondLine string
	expectedSecondLine = "    key1: value1"
	if lines[1] != expectedSecondLine {
		t.Errorf("defaultFormatter() second line = %q, want %q", lines[1], expectedSecondLine)
	}

	var expectedThirdLine string
	expectedThirdLine = "    count: 42"
	if lines[2] != expectedThirdLine {
		t.Errorf("defaultFormatter() third line = %q, want %q", lines[2], expectedThirdLine)
	}
}

func TestDefaultFormatter_TimestampFormat(t *testing.T) {
	var ts time.Time
	ts = time.Date(2024, 12, 1, 0, 5, 9, 0, time.UTC)
	var result string
	result = defaultFormatter(ts, DebugLevel, "msg", nil)

	var expectedPrefix string
	expectedPrefix = "[DEBUG] [24-12-01 00:05:09] msg"
	if result != expectedPrefix {
		t.Errorf("defaultFormatter() = %q, want %q", result, expectedPrefix)
	}
}

func TestDefaultFormatter_UTCConversion(t *testing.T) {
	var ts time.Time
	var loc *time.Location
	loc, _ = time.LoadLocation("America/New_York")
	ts = time.Date(2024, 3, 18, 10, 30, 0, 0, loc) // 10:30 AM ET = 14:30 UTC
	var result string
	result = defaultFormatter(ts, InfoLevel, "msg", nil)

	var expectedPrefix string
	expectedPrefix = "[INFO ] [24-03-18 14:30:00] msg"
	if result != expectedPrefix {
		t.Errorf("defaultFormatter() = %q, want %q", result, expectedPrefix)
	}
}

func TestDefaultFormatter_AllFieldTypes(t *testing.T) {
	var ts time.Time
	ts = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	var fields []Field
	fields = []Field{
		String("str", "text"),
		Int("int", 42),
		Int64("int64", 1234567890123),
		Duration("dur", 100*time.Millisecond),
		Err("err", nil),
		Err("err2", errors.New("test error")),
	}
	var result string
	result = defaultFormatter(ts, InfoLevel, "mixed fields", fields)

	var expectedSubstrings []string
	expectedSubstrings = []string{
		"str: text",
		"int: 42",
		"int64: 1234567890123",
		"dur: 100",
		"err: <nil>",
		"err2: test error",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(result, substr) {
			t.Errorf("defaultFormatter() missing expected substring: %q in %q", substr, result)
		}
	}
}

func TestSetLevel_GetLevel(t *testing.T) {
	// Save original level
	var originalLevel Level
	originalLevel = GetLevel()
	defer SetLevel(originalLevel) // Restore after test

	var testLevels []Level
	testLevels = []Level{FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel}
	for _, level := range testLevels {
		SetLevel(level)
		var got Level
		got = GetLevel()
		if got != level {
			t.Errorf("SetLevel(%v); GetLevel() = %v, want %v", level, got, level)
		}
	}
}

func TestSetLevel_GetLevel_Concurrent(t *testing.T) {
	// Save original level
	var originalLevel Level
	originalLevel = GetLevel()
	defer SetLevel(originalLevel)

	var wg sync.WaitGroup
	var numGoroutines int
	numGoroutines = 100
	var iterations int
	iterations = 50

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			var levels []Level
			levels = []Level{FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel}
			for j := 0; j < iterations; j++ {
				var level Level
				level = levels[(id+j)%len(levels)]
				SetLevel(level)
				var _ Level
				_ = GetLevel()
			}
		}(i)
	}

	wg.Wait()
	// Test passed if no panic or data race occurs
}

func TestShouldLog(t *testing.T) {
	// Save original level
	var originalLevel Level
	originalLevel = GetLevel()
	defer SetLevel(originalLevel)

	var tests []struct {
		setLevel     Level
		logLevel     Level
		shouldOutput bool
	}
	tests = []struct {
		setLevel     Level
		logLevel     Level
		shouldOutput bool
	}{
		{DebugLevel, DebugLevel, true},
		{DebugLevel, InfoLevel, true},
		{DebugLevel, WarnLevel, true},
		{DebugLevel, ErrorLevel, true},
		{DebugLevel, FatalLevel, true},
		{InfoLevel, DebugLevel, false},
		{InfoLevel, InfoLevel, true},
		{InfoLevel, WarnLevel, true},
		{InfoLevel, ErrorLevel, true},
		{InfoLevel, FatalLevel, true},
		{ErrorLevel, DebugLevel, false},
		{ErrorLevel, InfoLevel, false},
		{ErrorLevel, WarnLevel, false},
		{ErrorLevel, ErrorLevel, true},
		{ErrorLevel, FatalLevel, true},
		{FatalLevel, DebugLevel, false},
		{FatalLevel, InfoLevel, false},
		{FatalLevel, WarnLevel, false},
		{FatalLevel, ErrorLevel, false},
		{FatalLevel, FatalLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.setLevel.String()+"_"+tt.logLevel.String(), func(t *testing.T) {
			SetLevel(tt.setLevel)
			var got bool
			got = defaultLogger.shouldLog(tt.logLevel)
			if got != tt.shouldOutput {
				t.Errorf("shouldLog(%v) with minLevel=%v = %v, want %v",
					tt.logLevel, tt.setLevel, got, tt.shouldOutput)
			}
		})
	}
}

func TestLogger_shouldLog_RaceFree(t *testing.T) {
	// Save original level
	var originalLevel Level
	originalLevel = GetLevel()
	defer SetLevel(originalLevel)

	var wg sync.WaitGroup
	var numGoroutines int
	numGoroutines = 100

	wg.Add(numGoroutines * 2)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			var _ bool
			_ = defaultLogger.shouldLog(InfoLevel)
		}()
		go func() {
			defer wg.Done()
			SetLevel(DebugLevel)
		}()
	}

	wg.Wait()
	// Test passed if no data race occurs
}
