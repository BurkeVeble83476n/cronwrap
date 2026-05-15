// Package history provides execution history storage and analysis for cronwrap.
package history

import (
	"fmt"
	"time"
)

// Diff represents the change in execution metrics between two records.
type Diff struct {
	JobName        string
	OlderRunAt     time.Time
	NewerRunAt     time.Time
	DurationDelta  time.Duration // positive means newer was slower
	ExitCodeChange bool
	OlderExitCode  int
	NewerExitCode  int
	StatusChange   bool
	OlderStatus    string
	NewerStatus    string
}

// String returns a human-readable summary of the diff.
func (d Diff) String() string {
	if !d.ExitCodeChange && !d.StatusChange && d.DurationDelta == 0 {
		return fmt.Sprintf("[%s] no change between runs at %s and %s",
			d.JobName,
			d.OlderRunAt.Format(time.RFC3339),
			d.NewerRunAt.Format(time.RFC3339),
		)
	}

	msg := fmt.Sprintf("[%s] diff between %s → %s:\n",
		d.JobName,
		d.OlderRunAt.Format(time.RFC3339),
		d.NewerRunAt.Format(time.RFC3339),
	)
	if d.StatusChange {
		msg += fmt.Sprintf("  status:    %s → %s\n", d.OlderStatus, d.NewerStatus)
	}
	if d.ExitCodeChange {
		msg += fmt.Sprintf("  exit_code: %d → %d\n", d.OlderExitCode, d.NewerExitCode)
	}
	if d.DurationDelta != 0 {
		sign := "+"
		if d.DurationDelta < 0 {
			sign = ""
		}
		msg += fmt.Sprintf("  duration:  %s%s\n", sign, d.DurationDelta.Round(time.Millisecond))
	}
	return msg
}

// DiffRecords computes the diff between two consecutive execution records.
// older should be the earlier record, newer the more recent one.
func DiffRecords(older, newer Record) Diff {
	d := Diff{
		JobName:    older.JobName,
		OlderRunAt: older.RunAt,
		NewerRunAt: newer.RunAt,
	}

	if older.ExitCode != newer.ExitCode {
		d.ExitCodeChange = true
		d.OlderExitCode = older.ExitCode
		d.NewerExitCode = newer.ExitCode
	}

	if older.Status != newer.Status {
		d.StatusChange = true
		d.OlderStatus = older.Status
		d.NewerStatus = newer.Status
	}

	d.DurationDelta = newer.Duration - older.Duration
	return d
}
