package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/ijaidev/kraftui/config"
	"github.com/ijaidev/kraftui/log"
)

type cliArgs struct {
	Port         int          `arg:"--port,env:KRAFTUI_PORT" default:"0" help:"HTTP listen port (0 tries 5200-5204)"`
	LogType      log.LogType  `arg:"--log-type,env:KRAFTUI_LOG_TYPE" default:"fancy" help:"Log output type: basic, fancy, or json"`
	LogLevel     log.LogLevel `arg:"--log-level,env:KRAFTUI_LOG_LEVEL" default:"info" help:"Minimum log level: debug, info, warn, or error"`
	SuppressLogs bool         `arg:"--quiet,env:KRAFTUI_SUPPRESS_LOGS" default:"false" help:"Suppress all log output"`
	KraftBinary  string       `arg:"--kraft-binary,env:KRAFTUI_KRAFT_BINARY" default:"kraft" help:"Path to the Kraft CLI binary"`
	KraftTimeout string       `arg:"--kraft-timeout,env:KRAFTUI_KRAFT_TIMEOUT" default:"15s" help:"Timeout for one Kraft CLI command"`
	Version      bool         `arg:"--version" help:"Print version and exit"`
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
		if values.LogType, err = log.ParseLogType(string(values.LogType)); err != nil {
			return cliOptions{}, fmt.Errorf("--log-type: %w", err)
		}
		if values.LogLevel, err = log.ParseLogLevel(string(values.LogLevel)); err != nil {
			return cliOptions{}, fmt.Errorf("--log-level: %w", err)
		}
		timeout, err := time.ParseDuration(values.KraftTimeout)
		if err != nil {
			return cliOptions{}, fmt.Errorf("--kraft-timeout: invalid duration %q: %w", values.KraftTimeout, err)
		}
		if err := config.ValidateKraftConfig(config.KraftConfig{
			Binary:          values.KraftBinary,
			ExpectedVersion: config.SupportedKraftVersion,
			CommandTimeout:  timeout,
		}); err != nil {
			return cliOptions{}, fmt.Errorf("Kraft configuration: %w", err)
		}
		values.KraftTimeout = timeout.String()
		return cliOptions{values: values, parser: parser}, nil
	case arg.ErrHelp:
		return cliOptions{help: true, parser: parser}, nil
	default:
		return cliOptions{}, err
	}
}

func parseWithEmptyEnvironmentUnset(parser *arg.Parser, args []string) error {
	var empty []string
	for _, name := range []string{"KRAFTUI_PORT", "KRAFTUI_LOG_TYPE", "KRAFTUI_LOG_LEVEL", "KRAFTUI_SUPPRESS_LOGS", "KRAFTUI_KRAFT_BINARY", "KRAFTUI_KRAFT_TIMEOUT"} {
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

func (options cliOptions) kraftConfig() config.KraftConfig {
	timeout, _ := time.ParseDuration(options.values.KraftTimeout)
	return config.KraftConfig{
		Binary:          options.values.KraftBinary,
		ExpectedVersion: config.SupportedKraftVersion,
		CommandTimeout:  timeout,
	}
}

func writeHelp(output io.Writer, parser *arg.Parser) {
	fmt.Fprintln(output, "Configuration precedence: command-line flags, environment variables, then defaults.")
	parser.WriteHelp(output)
}
