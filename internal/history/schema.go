// Package history manages the execution history of cron jobs,
// providing storage, querying, retention, and export capabilities.
package history

import "time"

// Status represents the outcome of a cron job execution.
type Status string

const (
	// StatusSuccess indicates the job exited with code 0.
	StatusSuccess Status = "success"
	// StatusFailure indicates the job exited with a non-zero code.
	StatusFailure Status = "failure"
	// StatusTimeout indicates the job was killed due to a context deadline.
	StatusTimeout Status = "timeout"
)

// Record captures the full execution context of a single cron job run.
// Records are appended to a newline-delimited JSON file managed by Store.
type Record struct {
	// JobName is the human-readable identifier for the cron job,
	// typically derived from the --name flag passed to cronwrap.
	JobName string `json:"job_name"`

	// Command is the shell command that was executed.
	Command string `json:"command"`

	// StartedAt is the wall-clock time at which execution began.
	StartedAt time.Time `json:"started_at"`

	// FinishedAt is the wall-clock time at which execution ended,
	// regardless of outcome.
	FinishedAt time.Time `json:"finished_at"`

	// DurationMs is the elapsed time in milliseconds, provided as a
	// convenience so consumers do not need to diff the timestamps.
	DurationMs int64 `json:"duration_ms"`

	// ExitCode is the numeric exit status returned by the process.
	// A value of -1 indicates the process was signalled or could not
	// be started.
	ExitCode int `json:"exit_code"`

	// Status is the high-level outcome classification.
	Status Status `json:"status"`

	// Stdout contains the captured standard output of the command,
	// truncated to a reasonable size if necessary.
	Stdout string `json:"stdout,omitempty"`

	// Stderr contains the captured standard error of the command,
	// truncated to a reasonable size if necessary.
	Stderr string `json:"stderr,omitempty"`

	// Error holds a human-readable description of any infrastructure-level
	// error (e.g. failed to start the process). It is distinct from
	// non-zero exit codes captured in ExitCode.
	Error string `json:"error,omitempty"`
}

// Duration returns the execution time as a time.Duration, computed from
// the stored timestamps for cases where DurationMs may not be populated.
func (r *Record) Duration() time.Duration {
	if r.DurationMs > 0 {
		return time.Duration(r.DurationMs) * time.Millisecond
	}
	return r.FinishedAt.Sub(r.StartedAt)
}

// IsSuccess reports whether the record represents a successful execution.
func (r *Record) IsSuccess() bool {
	return r.Status == StatusSuccess
}
