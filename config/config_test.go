package config

import (
	"strings"
	"testing"
)

func TestParseLogType(t *testing.T) {
	got, err := ParseLogType(" FaNcY ")
	if err != nil || got != LogTypeFancy {
		t.Fatalf("ParseLogType() = %q, %v; want fancy", got, err)
	}
	if _, err := ParseLogType("plain"); err == nil || !strings.Contains(err.Error(), "invalid log type") {
		t.Fatalf("ParseLogType(invalid) error = %v", err)
	}
}

func TestParseLogLevel(t *testing.T) {
	got, err := ParseLogLevel(" WARN ")
	if err != nil || got != LogLevelWarn {
		t.Fatalf("ParseLogLevel() = %q, %v; want warn", got, err)
	}
	if _, err := ParseLogLevel("trace"); err == nil || !strings.Contains(err.Error(), "invalid log level") {
		t.Fatalf("ParseLogLevel(invalid) error = %v", err)
	}
}

func TestLoadStoresResolvedValues(t *testing.T) {
	Load(8080, LogTypeJSON, LogLevelError, true)

	if Port() != 8080 {
		t.Fatalf("Port() = %d, want 8080", Port())
	}
	if CurrentLogType() != LogTypeJSON {
		t.Fatalf("CurrentLogType() = %q, want json", CurrentLogType())
	}
	if CurrentLogLevel() != LogLevelError {
		t.Fatalf("CurrentLogLevel() = %q, want error", CurrentLogLevel())
	}
	if !SuppressLogs() {
		t.Fatal("SuppressLogs() = false, want true")
	}
}
