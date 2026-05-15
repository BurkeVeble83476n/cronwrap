package history

import (
	"strings"
	"testing"
	"time"
)

func makeRecord(job, status string, exitCode int, runAt time.Time, dur time.Duration) Record {
	return Record{
		JobName:  job,
		Status:   status,
		ExitCode: exitCode,
		RunAt:    runAt,
		Duration: dur,
	}
}

func TestDiffRecords_NoChange(t *testing.T) {
	now := time.Now()
	older := makeRecord("backup", "success", 0, now.Add(-time.Hour), 2*time.Second)
	newer := makeRecord("backup", "success", 0, now, 2*time.Second)

	d := DiffRecords(older, newer)

	if d.ExitCodeChange {
		t.Error("expected no exit code change")
	}
	if d.StatusChange {
		t.Error("expected no status change")
	}
	if d.DurationDelta != 0 {
		t.Errorf("expected zero duration delta, got %v", d.DurationDelta)
	}
	if !strings.Contains(d.String(), "no change") {
		t.Errorf("expected 'no change' in output, got: %s", d.String())
	}
}

func TestDiffRecords_StatusChange(t *testing.T) {
	now := time.Now()
	older := makeRecord("sync", "success", 0, now.Add(-time.Hour), time.Second)
	newer := makeRecord("sync", "failure", 1, now, time.Second)

	d := DiffRecords(older, newer)

	if !d.StatusChange {
		t.Error("expected status change")
	}
	if d.OlderStatus != "success" || d.NewerStatus != "failure" {
		t.Errorf("unexpected status values: %s → %s", d.OlderStatus, d.NewerStatus)
	}
	if !d.ExitCodeChange {
		t.Error("expected exit code change")
	}
}

func TestDiffRecords_DurationDelta(t *testing.T) {
	now := time.Now()
	older := makeRecord("report", "success", 0, now.Add(-time.Hour), 1*time.Second)
	newer := makeRecord("report", "success", 0, now, 3*time.Second)

	d := DiffRecords(older, newer)

	if d.DurationDelta != 2*time.Second {
		t.Errorf("expected +2s delta, got %v", d.DurationDelta)
	}
	if !strings.Contains(d.String(), "+2s") {
		t.Errorf("expected '+2s' in output, got: %s", d.String())
	}
}

func TestDiffRecords_JobName(t *testing.T) {
	now := time.Now()
	older := makeRecord("cleanup", "success", 0, now.Add(-time.Minute), time.Second)
	newer := makeRecord("cleanup", "success", 0, now, time.Second)

	d := DiffRecords(older, newer)

	if d.JobName != "cleanup" {
		t.Errorf("expected job name 'cleanup', got %q", d.JobName)
	}
}
