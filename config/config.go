package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/ijaidev/kraftui/log"
)

// SupportedKraftVersion is the exact Kraft CLI version supported by KraftUI.
const SupportedKraftVersion = "0.12.14"

type config struct {
	version      string
	port         int
	logType      log.LogType
	logLevel     log.LogLevel
	suppressLogs bool
	kraft        KraftConfig
}

// KraftConfig controls how KraftUI invokes the local Kraft CLI.
type KraftConfig struct {
	Binary          string
	ExpectedVersion string
	CommandTimeout  time.Duration
}

// DefaultKraftConfig returns the supported local CLI configuration.
func DefaultKraftConfig() KraftConfig {
	return KraftConfig{
		Binary:          "kraft",
		ExpectedVersion: SupportedKraftVersion,
		CommandTimeout:  15 * time.Second,
	}
}

// ValidateKraftConfig rejects incomplete or unsafe Kraft CLI settings.
func ValidateKraftConfig(value KraftConfig) error {
	if strings.TrimSpace(value.Binary) == "" {
		return fmt.Errorf("Kraft binary must not be empty")
	}
	if strings.TrimSpace(value.ExpectedVersion) == "" {
		return fmt.Errorf("Kraft version must not be empty")
	}
	if value.CommandTimeout <= 0 {
		return fmt.Errorf("Kraft command timeout must be greater than zero")
	}
	return nil
}

// version can be overridden at build time:
//
//	go build -ldflags "-X github.com/ijaidev/kraftui/config.version=0.1.0" ./cmd/kraftui
var version = "dev"

var g = config{
	version:      version,
	port:         0,
	logType:      log.LogTypeFancy,
	logLevel:     log.LogLevelInfo,
	suppressLogs: false,
	kraft:        DefaultKraftConfig(),
}

// Load stores resolved runtime configuration.
func Load(port int, logType log.LogType, logLevel log.LogLevel, suppressLogs bool, kraft KraftConfig) {
	g = config{
		version:      version,
		port:         port,
		logType:      logType,
		logLevel:     logLevel,
		suppressLogs: suppressLogs,
		kraft:        kraft,
	}
}

// Version returns the KraftUI application version.
func Version() string {
	return g.version
}

// Port returns the configured HTTP listen port. Zero tries 5200-5204.
func Port() int {
	return g.port
}

// CurrentLogType returns the configured logger output type.
func CurrentLogType() log.LogType {
	return g.logType
}

// CurrentLogLevel returns the configured minimum log level.
func CurrentLogLevel() log.LogLevel {
	return g.logLevel
}

// SuppressLogs reports whether log output is disabled.
func SuppressLogs() bool {
	return g.suppressLogs
}

// Kraft returns the configured local Kraft CLI settings.
func Kraft() KraftConfig {
	return g.kraft
}
