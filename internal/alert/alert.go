// Package alert provides alerting mechanisms for cronwrap job failures.
// It supports sending notifications when a job exceeds a duration threshold
// or exits with a non-zero status.
package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert holds the details of a triggered alert.
type Alert struct {
	JobName   string
	Level     Level
	Message   string
	ExitCode  int
	Duration  time.Duration
	Timestamp time.Time
}

// Notifier is the interface that wraps the Notify method.
type Notifier interface {
	Notify(a Alert) error
}

// LogNotifier writes alerts as structured text to an io.Writer.
type LogNotifier struct {
	Out io.Writer
}

// NewLogNotifier returns a LogNotifier that writes to the given writer.
// If w is nil, os.Stderr is used.
func NewLogNotifier(w io.Writer) *LogNotifier {
	if w == nil {
		w = os.Stderr
	}
	return &LogNotifier{Out: w}
}

// Notify writes a formatted alert line to the configured writer.
func (n *LogNotifier) Notify(a Alert) error {
	_, err := fmt.Fprintf(
		n.Out,
		"[%s] level=%s job=%q exit_code=%d duration=%s message=%q\n",
		a.Timestamp.UTC().Format(time.RFC3339),
		a.Level,
		a.JobName,
		a.ExitCode,
		a.Duration.Round(time.Millisecond),
		a.Message,
	)
	return err
}

// ShouldAlert returns true when the result warrants an alert, i.e. the exit
// code is non-zero or the duration exceeds the provided threshold (when > 0).
func ShouldAlert(exitCode int, duration, threshold time.Duration) bool {
	if exitCode != 0 {
		return true
	}
	if threshold > 0 && duration > threshold {
		return true
	}
	return false
}

// BuildAlert constructs an Alert from the given parameters.
func BuildAlert(jobName string, exitCode int, duration time.Duration, threshold time.Duration) Alert {
	level := LevelInfo
	msg := "job completed successfully"

	if exitCode != 0 {
		level = LevelError
		msg = fmt.Sprintf("job failed with exit code %d", exitCode)
	} else if threshold > 0 && duration > threshold {
		level = LevelWarn
		msg = fmt.Sprintf("job exceeded duration threshold of %s", threshold)
	}

	return Alert{
		JobName:   jobName,
		Level:     level,
		Message:   msg,
		ExitCode:  exitCode,
		Duration:  duration,
		Timestamp: time.Now().UTC(),
	}
}
