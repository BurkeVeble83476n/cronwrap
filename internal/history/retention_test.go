package history

import (
	"testing"
	"time"
)

func seedRetentionStore(t *testing.T, records []Record) *Store {
	t.Helper()
	s := tempStore(t)
	for _, r := range records {
		if err := s.Append(r); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return s
}

func TestPrune_ByMaxAge(t *testing.T) {
	now := time.Now()
	records := []Record{
		{JobName: "job", StartedAt: now.Add(-48 * time.Hour), ExitCode: 0},
		{JobName: "job", StartedAt: now.Add(-1 * time.Hour), ExitCode: 0},
	}
	s := seedRetentionStore(t, records)

	err := Prune(s, RetentionPolicy{MaxAge: 24 * time.Hour})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	all, _ := s.ReadAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 record after prune, got %d", len(all))
	}
	if all[0].StartedAt.Unix() != records[1].StartedAt.Unix() {
		t.Error("expected the recent record to be retained")
	}
}

func TestPrune_ByMaxRecords(t *testing.T) {
	now := time.Now()
	records := []Record{
		{JobName: "job", StartedAt: now.Add(-3 * time.Hour), ExitCode: 0},
		{JobName: "job", StartedAt: now.Add(-2 * time.Hour), ExitCode: 0},
		{JobName: "job", StartedAt: now.Add(-1 * time.Hour), ExitCode: 0},
	}
	s := seedRetentionStore(t, records)

	err := Prune(s, RetentionPolicy{MaxRecords: 2})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	all, _ := s.ReadAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 records after prune, got %d", len(all))
	}
}

func TestPrune_EmptyStore(t *testing.T) {
	s := tempStore(t)
	if err := Prune(s, RetentionPolicy{MaxAge: time.Hour, MaxRecords: 10}); err != nil {
		t.Fatalf("Prune on empty store: %v", err)
	}
}

func TestPrune_MultipleJobs(t *testing.T) {
	now := time.Now()
	records := []Record{
		{JobName: "a", StartedAt: now.Add(-3 * time.Hour)},
		{JobName: "a", StartedAt: now.Add(-2 * time.Hour)},
		{JobName: "b", StartedAt: now.Add(-3 * time.Hour)},
		{JobName: "b", StartedAt: now.Add(-2 * time.Hour)},
		{JobName: "b", StartedAt: now.Add(-1 * time.Hour)},
	}
	s := seedRetentionStore(t, records)

	err := Prune(s, RetentionPolicy{MaxRecords: 1})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	all, _ := s.ReadAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 records (1 per job), got %d", len(all))
	}
}
