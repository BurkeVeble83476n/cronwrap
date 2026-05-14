// Package runner provides the core command execution engine for cronwrap.
//
// It wraps os/exec to run arbitrary shell commands while capturing:
//   - stdout and stderr output
//   - wall-clock start/end times and duration
//   - process exit code
//   - any execution errors (including context cancellation/timeout)
//
// Usage:
//
//	result := runner.Run(ctx, "my-script.sh", []string{"--flag"})
//	if result.ExitCode != 0 {
//	    // handle failure
//	}
//
// The Result struct is designed to be consumed by downstream modules such as
// the structured logger, alerting system, and execution history store.
package runner

import "strings"
