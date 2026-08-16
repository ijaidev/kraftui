// Package log provides the process-wide structured logger.
package log

import (
	"io"
	"log/slog"
	"os"

	charmlog "github.com/charmbracelet/log"
)

// G is the process-wide logger. It defaults to slog's standard logger until
// Configure is called with application settings.
var G = slog.Default()

// Configure initializes G from the resolved log type, level, and quiet flag.
func Configure(kind LogType, level LogLevel, suppress bool) {
	if suppress {
		G = slog.New(slog.NewTextHandler(io.Discard, nil))
		slog.SetDefault(G)
		return
	}

	G = newLogger(kind, level, os.Stderr)
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
