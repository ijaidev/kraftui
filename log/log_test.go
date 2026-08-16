package log

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestGlobalLoggerInitialized(t *testing.T) {
	if G() == nil {
		t.Fatal("G is nil")
	}
}

func TestNewLoggerFormats(t *testing.T) {
	tests := []struct {
		name  string
		kind  LogType
		check func(*testing.T, []byte)
	}{
		{
			name: "basic",
			kind: LogTypeBasic,
			check: func(t *testing.T, output []byte) {
				t.Helper()
				if !bytes.Contains(output, []byte("level=INFO")) {
					t.Fatalf("basic output = %q, want text level", output)
				}
			},
		},
		{
			name: "fancy",
			kind: LogTypeFancy,
			check: func(t *testing.T, output []byte) {
				t.Helper()
				if !bytes.Contains(output, []byte("INFO")) || !bytes.Contains(output, []byte("ready")) {
					t.Fatalf("fancy output = %q, want styled text record", output)
				}
			},
		},
		{
			name: "json",
			kind: LogTypeJSON,
			check: func(t *testing.T, output []byte) {
				t.Helper()
				var record map[string]any
				if err := json.Unmarshal(output, &record); err != nil {
					t.Fatalf("json output is invalid: %v", err)
				}
				if record["level"] != "INFO" {
					t.Fatalf("json level = %v, want INFO", record["level"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			newLogger(tt.kind, LogLevelInfo, &output).Info("ready")
			tt.check(t, output.Bytes())
		})
	}
}

func TestNewLoggerFiltersBelowLevel(t *testing.T) {
	var output bytes.Buffer
	logger := newLogger(LogTypeBasic, LogLevelWarn, &output)

	logger.Info("hidden")
	logger.Warn("shown")

	if strings.Contains(output.String(), "hidden") {
		t.Fatalf("output contains filtered log: %q", output.String())
	}
	if !strings.Contains(output.String(), "shown") {
		t.Fatalf("output missing warning: %q", output.String())
	}
}
