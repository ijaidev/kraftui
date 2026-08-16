package config

import (
	"testing"
	"time"

	"github.com/ijaidev/kraftui/log"
)

func TestLoadStoresResolvedValues(t *testing.T) {
	kraftConfig := KraftConfig{Binary: "test-kraft", ExpectedVersion: "0.12.14", CommandTimeout: time.Second}
	Load(8080, log.LogTypeJSON, log.LogLevelError, true, kraftConfig)

	if Port() != 8080 {
		t.Fatalf("Port() = %d, want 8080", Port())
	}
	if CurrentLogType() != log.LogTypeJSON {
		t.Fatalf("CurrentLogType() = %q, want json", CurrentLogType())
	}
	if CurrentLogLevel() != log.LogLevelError {
		t.Fatalf("CurrentLogLevel() = %q, want error", CurrentLogLevel())
	}
	if !SuppressLogs() {
		t.Fatal("SuppressLogs() = false, want true")
	}
	if Kraft() != kraftConfig {
		t.Fatalf("Kraft() = %#v, want %#v", Kraft(), kraftConfig)
	}
}

func TestValidateKraftConfig(t *testing.T) {
	if err := ValidateKraftConfig(DefaultKraftConfig()); err != nil {
		t.Fatalf("ValidateKraftConfig(default) error = %v", err)
	}
	if err := ValidateKraftConfig(KraftConfig{}); err == nil {
		t.Fatal("ValidateKraftConfig(empty) error = nil, want error")
	}
}
