package runner

import (
	"context"
	"os/exec"
	"time"
)

// Result holds the outcome of a command execution.
type Result struct {
	Command   string
	Args      []string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	ExitCode  int
	Stdout    string
	Stderr    string
	Err       error
}

// Run executes the given command with args, respecting the provided context.
// It captures stdout, stderr, timing, and exit code.
func Run(ctx context.Context, command string, args []string) Result {
	result := Result{
		Command:   command,
		Args:      args,
		StartTime: time.Now(),
	}

	cmd := exec.CommandContext(ctx, command, args...)

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()
	result.Err = runErr

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
