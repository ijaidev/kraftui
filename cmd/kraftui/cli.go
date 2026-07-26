package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	arg "github.com/alexflint/go-arg"
	"github.com/ijaidev/kraftui/config"
)

type cliArgs struct {
	Port         int             `arg:"--port,env:KRAFTUI_PORT" default:"0" help:"HTTP listen port"`
	LogType      config.LogType  `arg:"--log-type,env:KRAFTUI_LOG_TYPE" default:"fancy" help:"Log output type: basic, fancy, or json"`
	LogLevel     config.LogLevel `arg:"--log-level,env:KRAFTUI_LOG_LEVEL" default:"info" help:"Minimum log level: debug, info, warn, or error"`
	SuppressLogs bool            `arg:"--quiet,env:KRAFTUI_SUPPRESS_LOGS" default:"false" help:"Suppress all log output"`
	Version      bool            `arg:"--version" help:"Print version and exit"`
}

type cliOptions struct {
	values cliArgs
	help   bool
	parser *arg.Parser
}

func loadCLI(args []string) (cliOptions, error) {
	var values cliArgs
	parser, err := arg.NewParser(arg.Config{Program: "kraftui", Out: io.Discard}, &values)
	if err != nil {
		return cliOptions{}, err
	}

	err = parseWithEmptyEnvironmentUnset(parser, args)
	switch err {
	case nil:
		if values.Port < 0 || values.Port > 65535 {
			return cliOptions{}, fmt.Errorf("--port: invalid port %q; supported range is 0-65535", strconv.Itoa(values.Port))
		}
		if values.LogType, err = config.ParseLogType(string(values.LogType)); err != nil {
			return cliOptions{}, fmt.Errorf("--log-type: %w", err)
		}
		if values.LogLevel, err = config.ParseLogLevel(string(values.LogLevel)); err != nil {
			return cliOptions{}, fmt.Errorf("--log-level: %w", err)
		}
		return cliOptions{values: values, parser: parser}, nil
	case arg.ErrHelp:
		return cliOptions{help: true, parser: parser}, nil
	default:
		return cliOptions{}, err
	}
}

func parseWithEmptyEnvironmentUnset(parser *arg.Parser, args []string) error {
	var empty []string
	for _, name := range []string{"KRAFTUI_PORT", "KRAFTUI_LOG_TYPE", "KRAFTUI_LOG_LEVEL", "KRAFTUI_SUPPRESS_LOGS"} {
		if value, ok := os.LookupEnv(name); ok && value == "" {
			empty = append(empty, name)
			if err := os.Unsetenv(name); err != nil {
				return err
			}
		}
	}
	defer func() {
		for _, name := range empty {
			_ = os.Setenv(name, "")
		}
	}()
	return parser.Parse(args)
}

func writeHelp(output io.Writer, parser *arg.Parser) {
	fmt.Fprintln(output, "Configuration precedence: command-line flags, environment variables, then defaults.")
	parser.WriteHelp(output)
}
