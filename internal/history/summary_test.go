package history

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func seedSummaryStore(t *testing.T) *Store {
	t.Helper()
	s := tempStore(t)
	now := time.Now()
	records := []Record{
		{JobName: "backup", Status: "success", ExitCode: 0, Duration: 2 * time.Second, StartedAt: now.Add(-4 * time.Hour)},
		{JobName: "backup", Status: "failure", ExitCode: 1, Duration: 3 * time.Second, StartedAt: now.Add(-2 * time.Hour)},
		{JobName: "backup", Status: "success", ExitCode: 0, Duration: 1 * time.Second, StartedAt: now.Add(-1 * time.Hour)},
		{JobName: "cleanup", Status: "success", ExitCode: 0, Duration: 500 * time.Millisecond, StartedAt: now.Add(-30 * time.Minute)},
	}
	for _, r := range records {
		if err := s.Append(r); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return s
}

func TestSummarize_AllJobs(t *testing.T) {
	s := seedSummaryStore(t)
	sums, err := Summarize(s, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sums) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(sums))
	}

	backup := sums[0]
	if backup.JobName != "backup" {
		t.Errorf("expected backup, got %s", backup.JobName)
	}
	if backup.TotalRuns != 3 {
		t.Errorf("expected 3 runs, got %d", backup.TotalRuns)
	}
	if backup.SuccessCount != 2 || backup.FailureCount != 1 {
		t.Errorf("unexpected success/failure counts: %d/%d", backup.SuccessCount, backup.FailureCount)
	}
	if backup.LastStatus != "success" {
		t.Errorf("expected last status success, got %s", backup.LastStatus)
	}
}

func TestSummarize_FilterByJobName(t *testing.T) {
	s := seedSummaryStore(t)
	sums, err := Summarize(s, "cleanup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sums) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(sums))
	}
	if sums[0].TotalRuns != 1 {
		t.Errorf("expected 1 run, got %d", sums[0].TotalRuns)
	}
}

func TestSummarize_EmptyStore(t *testing.T) {
	s := tempStore(t)
	sums, err := Summarize(s, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sums) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(sums))
	}
}

func TestSummarize_SuccessRate(t *testing.T) {
	s := seedSummaryStore(t)
	sums, err := Summarize(s, "backup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sums) == 0 {
		t.Fatal("expected at least one summary")
	}
	got := sums[0].SuccessRate
	want := 66.6
	if got < 66.0 || got > 67.0 {
		t.Errorf("expected success rate ~%.1f%%, got %.2f%%", want, got)
	}
}

func TestPrintSummary_ContainsHeaders(t *testing.T) {
	s := seedSummaryStore(t)
	sums, _ := Summarize(s, "")
	var buf bytes.Buffer
	PrintSummary(&buf, sums)
	out := buf.String()
	for _, hdr := range []string{"JOB", "RUNS", "SUCCESS%", "AVG DURATION", "LAST STATUS"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("output missing header %q", hdr)
		}
	}
	if !strings.Contains(out, "backup") {
		t.Error("output missing job name 'backup'")
	}
}
