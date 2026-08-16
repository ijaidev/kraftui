package log

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
