package config

import (
	"fmt"
	"strings"
)

// LogType controls the logger output format.
type LogType string

const (
	LogTypeBasic LogType = "basic"
	LogTypeFancy LogType = "fancy"
	LogTypeJSON  LogType = "json"
)

// ParseLogType converts a string into a validated LogType.
func ParseLogType(value string) (LogType, error) {
	switch parsed := LogType(strings.ToLower(strings.TrimSpace(value))); parsed {
	case LogTypeBasic, LogTypeFancy, LogTypeJSON:
		return parsed, nil
	default:
		return "", fmt.Errorf("invalid log type %q; supported values are basic, fancy, and json", value)
	}
}

// LogLevel controls the minimum emitted severity.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// ParseLogLevel converts a string into a validated LogLevel.
func ParseLogLevel(value string) (LogLevel, error) {
	switch parsed := LogLevel(strings.ToLower(strings.TrimSpace(value))); parsed {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return parsed, nil
	default:
		return "", fmt.Errorf("invalid log level %q; supported values are debug, info, warn, and error", value)
	}
}

type config struct {
	version      string
	port         int
	logType      LogType
	logLevel     LogLevel
	suppressLogs bool
}

// version can be overridden at build time:
//
//	go build -ldflags "-X github.com/ijaidev/kraftui/config.version=0.1.0" ./cmd/kraftui
var version = "dev"

var g = config{
	version:      version,
	port:         0,
	logType:      LogTypeFancy,
	logLevel:     LogLevelInfo,
	suppressLogs: false,
}

// Load stores resolved runtime configuration.
func Load(port int, logType LogType, logLevel LogLevel, suppressLogs bool) {
	g = config{
		version:      version,
		port:         port,
		logType:      logType,
		logLevel:     logLevel,
		suppressLogs: suppressLogs,
	}
}

// KraftVersion returns the configured Kraft version.
func KraftVersion() string {
	return g.version
}

// Port returns the configured HTTP listen port. Zero selects an available port.
func Port() int {
	return g.port
}

// CurrentLogType returns the configured logger output type.
func CurrentLogType() LogType {
	return g.logType
}

// CurrentLogLevel returns the configured minimum log level.
func CurrentLogLevel() LogLevel {
	return g.logLevel
}

// SuppressLogs reports whether log output is disabled.
func SuppressLogs() bool {
	return g.suppressLogs
}
