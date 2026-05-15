package history

import (
	"testing"
	"time"
)

func seedStatsStore(t *testing.T) *Store {
	t.Helper()
	s := tempStore(t)
	now := time.Now().UTC()
	records := []Record{
		{JobName: "backup", ExitCode: 0, Duration: 2 * time.Second, StartedAt: now.Add(-4 * time.Hour)},
		{JobName: "backup", ExitCode: 0, Duration: 3 * time.Second, StartedAt: now.Add(-3 * time.Hour)},
		{JobName: "backup", ExitCode: 1, Duration: 1 * time.Second, StartedAt: now.Add(-2 * time.Hour)},
		{JobName: "cleanup", ExitCode: 0, Duration: 500 * time.Millisecond, StartedAt: now.Add(-1 * time.Hour)},
		{JobName: "cleanup", ExitCode: 1, Duration: 800 * time.Millisecond, StartedAt: now},
	}
	for _, r := range records {
		if err := s.Append(r); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return s
}

func TestStats_AllJobs(t *testing.T) {
	s := seedStatsStore(t)
	stats, err := Stats(s, "")
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 job entries, got %d", len(stats))
	}
}

func TestStats_FilterByJobName(t *testing.T) {
	s := seedStatsStore(t)
	stats, err := Stats(s, "backup")
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(stats))
	}
	st := stats[0]
	if st.TotalRuns != 3 {
		t.Errorf("TotalRuns: want 3, got %d", st.TotalRuns)
	}
	if st.SuccessCount != 2 {
		t.Errorf("SuccessCount: want 2, got %d", st.SuccessCount)
	}
	if st.FailureCount != 1 {
		t.Errorf("FailureCount: want 1, got %d", st.FailureCount)
	}
	wantRate := 66.67
	if st.SuccessRate != wantRate {
		t.Errorf("SuccessRate: want %.2f, got %.2f", wantRate, st.SuccessRate)
	}
	if st.MinDuration != 1*time.Second {
		t.Errorf("MinDuration: want 1s, got %v", st.MinDuration)
	}
	if st.MaxDuration != 3*time.Second {
		t.Errorf("MaxDuration: want 3s, got %v", st.MaxDuration)
	}
	if st.AvgDuration != 2*time.Second {
		t.Errorf("AvgDuration: want 2s, got %v", st.AvgDuration)
	}
}

func TestStats_EmptyStore(t *testing.T) {
	s := tempStore(t)
	stats, err := Stats(s, "")
	if err != nil {
		t.Fatalf("Stats on empty store: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats, got %d", len(stats))
	}
}

func TestStats_LastRun(t *testing.T) {
	s := seedStatsStore(t)
	stats, err := Stats(s, "cleanup")
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if stats[0].LastRun.IsZero() {
		t.Error("LastRun should not be zero")
	}
}
