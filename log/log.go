// Package log provides the process-wide structured logger.
package log

import (
	"io"
	"log/slog"
	"os"

	charmlog "github.com/charmbracelet/log"
	"github.com/ijaidev/kraftui/config"
)

// LogType controls the logger output format.
type LogType = config.LogType

const (
	LogTypeBasic = config.LogTypeBasic
	LogTypeFancy = config.LogTypeFancy
	LogTypeJSON  = config.LogTypeJSON
)

// LogLevel controls the minimum emitted severity.
type LogLevel = config.LogLevel

const (
	LogLevelDebug = config.LogLevelDebug
	LogLevelInfo  = config.LogLevelInfo
	LogLevelWarn  = config.LogLevelWarn
	LogLevelError = config.LogLevelError
)

// G is the process-wide logger. It defaults to slog's standard logger until
// log.Configure() is called with application settings.
var G = slog.Default()

// Configure initializes G after application configuration has been loaded.
func Configure() {
	if config.SuppressLogs() {
		G = slog.New(slog.NewTextHandler(io.Discard, nil))
		slog.SetDefault(G)
		return
	}

	G = newLogger(config.CurrentLogType(), config.CurrentLogLevel(), os.Stderr)
	slog.SetDefault(G)
}

func newLogger(kind LogType, level LogLevel, output io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slogLevel(level)}

	switch kind {
	case LogTypeJSON:
		return slog.New(slog.NewJSONHandler(output, opts))
	case LogTypeFancy:
		return slog.New(charmlog.NewWithOptions(output, charmlog.Options{
			Level:           charmLevel(level),
			ReportTimestamp: true,
		}))
	default:
		return slog.New(slog.NewTextHandler(output, opts))
	}
}

func slogLevel(level LogLevel) slog.Level {
	switch level {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func charmLevel(level LogLevel) charmlog.Level {
	switch level {
	case LogLevelDebug:
		return charmlog.DebugLevel
	case LogLevelWarn:
		return charmlog.WarnLevel
	case LogLevelError:
		return charmlog.ErrorLevel
	default:
		return charmlog.InfoLevel
	}
}
