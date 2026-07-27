package kraft

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ijaidev/kraftui/config"
)

var baseArgs = []string{
	"--no-prompt",
	"--no-check-updates",
	"--no-color",
	"--no-emojis",
	"--log-level", "error",
}

// Client invokes one supported Kraft CLI version.
type Client struct {
	binary          string
	expectedVersion string
	timeout         time.Duration
	runner          Runner
}

// New resolves the configured Kraft binary and prepares a client.
func New(cfg config.KraftConfig) (*Client, error) {
	if err := config.ValidateKraftConfig(cfg); err != nil {
		return nil, err
	}
	binary, err := exec.LookPath(cfg.Binary)
	if err != nil {
		return nil, fmt.Errorf("locating Kraft binary %q: %w", cfg.Binary, err)
	}
	return NewWithRunner(cfg, binary, osRunner{})
}

// NewWithRunner constructs a client for tests or custom process runners.
func NewWithRunner(cfg config.KraftConfig, binary string, runner Runner) (*Client, error) {
	if err := config.ValidateKraftConfig(cfg); err != nil {
		return nil, err
	}
	if strings.TrimSpace(binary) == "" {
		return nil, fmt.Errorf("Kraft binary must not be empty")
	}
	if runner == nil {
		return nil, fmt.Errorf("Kraft runner must not be nil")
	}
	return &Client{
		binary:          binary,
		expectedVersion: cfg.ExpectedVersion,
		timeout:         cfg.CommandTimeout,
		runner:          runner,
	}, nil
}

// Verify checks that the configured binary reports the supported version.
func (c *Client) Verify(ctx context.Context) (string, error) {
	result, err := c.run(ctx, "version")
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(result.Stdout))
	if len(fields) < 2 || fields[0] != "kraft" {
		return "", fmt.Errorf("could not parse Kraft version output")
	}
	if fields[1] != c.expectedVersion {
		return "", fmt.Errorf("Kraft version %s is not supported; expected %s", fields[1], c.expectedVersion)
	}
	return fields[1], nil
}

func (c *Client) run(ctx context.Context, args ...string) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	commandArgs := append(append([]string{}, baseArgs...), args...)
	result, err := c.runner.Run(ctx, c.binary, commandArgs)
	if err == nil {
		return result, nil
	}
	if ctx.Err() != nil {
		return result, fmt.Errorf("kraft %s: %w", strings.Join(args, " "), ctx.Err())
	}
	return result, &CommandError{
		Command: strings.Join(args, " "),
		Stderr:  strings.TrimSpace(string(result.Stderr)),
		Err:     err,
	}
}
