package kraft

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Result is the captured result of one Kraft process invocation.
type Result struct {
	Stdout []byte
	Stderr []byte
}

// Runner runs a process without involving a shell.
type Runner interface {
	Run(context.Context, string, []string) (Result, error)
}

type osRunner struct{}

func (osRunner) Run(ctx context.Context, binary string, args []string) (Result, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return Result{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
}

// CommandError includes safe diagnostic output from a failed Kraft invocation.
type CommandError struct {
	Command string
	Stderr  string
	Err     error
}

func (e *CommandError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("kraft %s: %s", e.Command, e.Stderr)
	}
	return fmt.Sprintf("kraft %s: %v", e.Command, e.Err)
}

func (e *CommandError) Unwrap() error { return e.Err }
