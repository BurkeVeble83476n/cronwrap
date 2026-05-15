package history

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func seedReplayStore(t *testing.T) *Store {
	t.Helper()
	s := tempStore(t)
	now := time.Now()
	records := []Record{
		{JobName: "backup", Status: "success", ExitCode: 0, StartedAt: now.Add(-3 * time.Hour), Duration: 2 * time.Second},
		{JobName: "backup", Status: "failure", ExitCode: 1, StartedAt: now.Add(-2 * time.Hour), Duration: 1 * time.Second},
		{JobName: "cleanup", Status: "success", ExitCode: 0, StartedAt: now.Add(-1 * time.Hour), Duration: 500 * time.Millisecond},
	}
	for _, r := range records {
		if err := s.Append(r); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return s
}

func TestReplay_AllRecords(t *testing.T) {
	s := seedReplayStore(t)
	var buf bytes.Buffer
	res, err := Replay(s, ReplayOptions{Writer: &buf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Printed != 3 {
		t.Errorf("expected 3 printed, got %d", res.Printed)
	}
	if !strings.Contains(buf.String(), "backup") {
		t.Error("expected 'backup' in output")
	}
}

func TestReplay_FilterByJobName(t *testing.T) {
	s := seedReplayStore(t)
	var buf bytes.Buffer
	res, err := Replay(s, ReplayOptions{JobName: "cleanup", Writer: &buf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Printed != 1 {
		t.Errorf("expected 1 printed, got %d", res.Printed)
	}
	if strings.Contains(buf.String(), "backup") {
		t.Error("did not expect 'backup' in filtered output")
	}
}

func TestReplay_Limit(t *testing.T) {
	s := seedReplayStore(t)
	var buf bytes.Buffer
	res, err := Replay(s, ReplayOptions{Limit: 2, Writer: &buf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Printed != 2 {
		t.Errorf("expected 2 printed, got %d", res.Printed)
	}
}

func TestReplay_SinceFilter(t *testing.T) {
	s := seedReplayStore(t)
	var buf bytes.Buffer
	cutoff := time.Now().Add(-90 * time.Minute)
	res, err := Replay(s, ReplayOptions{Since: cutoff, Writer: &buf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Printed != 1 {
		t.Errorf("expected 1 record after cutoff, got %d", res.Printed)
	}
}

func TestReplay_EmptyStore(t *testing.T) {
	s := tempStore(t)
	var buf bytes.Buffer
	res, err := Replay(s, ReplayOptions{Writer: &buf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Printed != 0 {
		t.Errorf("expected 0 printed, got %d", res.Printed)
	}
}
