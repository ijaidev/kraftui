package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/ijaidev/kraftui/config"
)

func TestLoadCLIUsesDefaults(t *testing.T) {
	clearEnvironment(t)
	got := loadCLIForTest(t)

	if got.values.Port != 0 || got.values.LogType != config.LogTypeFancy || got.values.LogLevel != config.LogLevelInfo || got.values.SuppressLogs {
		t.Fatalf("options = %#v, want defaults", got.values)
	}
}

func TestLoadCLIUsesEnvironment(t *testing.T) {
	t.Setenv("KRAFTUI_PORT", "8080")
	t.Setenv("KRAFTUI_LOG_TYPE", "FANCY")
	t.Setenv("KRAFTUI_LOG_LEVEL", "WARN")
	t.Setenv("KRAFTUI_SUPPRESS_LOGS", "true")
	got := loadCLIForTest(t)

	if got.values.Port != 8080 || got.values.LogType != config.LogTypeFancy || got.values.LogLevel != config.LogLevelWarn || !got.values.SuppressLogs {
		t.Fatalf("options = %#v, want environment values", got.values)
	}
}

func TestLoadCLIFlagsOverrideEnvironment(t *testing.T) {
	t.Setenv("KRAFTUI_PORT", "8080")
	t.Setenv("KRAFTUI_LOG_TYPE", "basic")
	t.Setenv("KRAFTUI_LOG_LEVEL", "warn")
	t.Setenv("KRAFTUI_SUPPRESS_LOGS", "true")
	got := loadCLIForTest(t, "--port=9090", "--log-type=json", "--log-level=debug", "--quiet=false")

	if got.values.Port != 9090 || got.values.LogType != config.LogTypeJSON || got.values.LogLevel != config.LogLevelDebug || got.values.SuppressLogs {
		t.Fatalf("options = %#v, want flag values", got.values)
	}
}

func TestLoadCLIRejectsInvalidValues(t *testing.T) {
	_, err := loadCLI([]string{"--log-level=trace"})
	if err == nil || !strings.Contains(err.Error(), "invalid log level") {
		t.Fatalf("loadCLI() error = %v, want invalid log level", err)
	}

	t.Setenv("KRAFTUI_PORT", "70000")
	_, err = loadCLI(nil)
	if err == nil || !strings.Contains(err.Error(), "invalid port") {
		t.Fatalf("loadCLI() error = %v, want invalid port", err)
	}
}

func TestLoadCLIHelpSkipsEnvironmentResolution(t *testing.T) {
	t.Setenv("KRAFTUI_LOG_TYPE", "invalid")
	got := loadCLIForTest(t, "--help")
	if !got.help {
		t.Fatal("help = false, want true")
	}
}

func TestWriteHelpIncludesSettings(t *testing.T) {
	got := loadCLIForTest(t)
	var output bytes.Buffer
	writeHelp(&output, got.parser)

	for _, value := range []string{"--port", "KRAFTUI_PORT", "--log-type", "KRAFTUI_LOG_TYPE", "--log-level", "KRAFTUI_LOG_LEVEL", "--quiet", "KRAFTUI_SUPPRESS_LOGS"} {
		if !strings.Contains(output.String(), value) {
			t.Fatalf("help output missing %q:\n%s", value, output.String())
		}
	}
}

func clearEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{"KRAFTUI_PORT", "KRAFTUI_LOG_TYPE", "KRAFTUI_LOG_LEVEL", "KRAFTUI_SUPPRESS_LOGS"} {
		value, present := os.LookupEnv(name)
		_ = os.Unsetenv(name)
		t.Cleanup(func() {
			if present {
				_ = os.Setenv(name, value)
			} else {
				_ = os.Unsetenv(name)
			}
		})
	}
}

func loadCLIForTest(t *testing.T, args ...string) cliOptions {
	t.Helper()
	got, err := loadCLI(args)
	if err != nil {
		t.Fatalf("loadCLI(%q) error = %v", args, err)
	}
	return got
}
