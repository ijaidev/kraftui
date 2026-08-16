// Package log provides the process-wide structured logger.
package log

import (
	"io"
	"log/slog"
	"os"

	charmlog "github.com/charmbracelet/log"
)

// g is the process-wide logger. It defaults to slog's standard logger until
// Configure is called with application settings.
var g = slog.Default()

// G returns the process-wide logger.
func G() *slog.Logger {
	return g
}

// Configure initializes the process-wide logger from the resolved log type,
// level, and quiet flag.
func Configure(kind LogType, level LogLevel, suppress bool) {
	if suppress {
		g = slog.New(slog.NewTextHandler(io.Discard, nil))
		slog.SetDefault(g)
		return
	}

	g = newLogger(kind, level, os.Stderr)
	slog.SetDefault(g)
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
