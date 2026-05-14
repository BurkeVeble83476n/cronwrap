package history_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/history"
)

func seedStore(t *testing.T, records []history.Record) *history.Store {
	t.Helper()
	s := tempStore(t)
	for _, r := range records {
		if err := s.Append(r); err != nil {
			t.Fatalf("seed Append: %v", err)
		}
	}
	return s
}

func TestQuery_FilterByJobName(t *testing.T) {
	now := time.Now().UTC()
	s := seedStore(t, []history.Record{
		{JobName: "alpha", Status: history.StatusSuccess, StartedAt: now},
		{JobName: "beta", Status: history.StatusFailure, StartedAt: now},
		{JobName: "alpha", Status: history.StatusFailure, StartedAt: now},
	})

	results, err := s.Query(history.Filter{JobName: "alpha"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2, got %d", len(results))
	}
}

func TestQuery_FilterByStatus(t *testing.T) {
	now := time.Now().UTC()
	s := seedStore(t, []history.Record{
		{JobName: "j", Status: history.StatusSuccess, StartedAt: now},
		{JobName: "j", Status: history.StatusTimeout, StartedAt: now},
		{JobName: "j", Status: history.StatusSuccess, StartedAt: now},
	})

	results, err := s.Query(history.Filter{Status: history.StatusSuccess})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2, got %d", len(results))
	}
}

func TestQuery_Limit(t *testing.T) {
	now := time.Now().UTC()
	var records []history.Record
	for i := 0; i < 10; i++ {
		records = append(records, history.Record{
			JobName: "job", Status: history.StatusSuccess, StartedAt: now,
		})
	}
	s := seedStore(t, records)

	results, err := s.Query(history.Filter{Limit: 3})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3, got %d", len(results))
	}
}

func TestLast_ReturnsNilWhenEmpty(t *testing.T) {
	s := tempStore(t)
	r, err := s.Last("nonexistent")
	if err != nil {
		t.Fatalf("Last: %v", err)
	}
	if r != nil {
		t.Errorf("expected nil, got %+v", r)
	}
}

func TestLast_ReturnsMostRecent(t *testing.T) {
	now := time.Now().UTC()
	s := seedStore(t, []history.Record{
		{JobName: "myjob", Status: history.StatusSuccess, StartedAt: now.Add(-2 * time.Hour)},
		{JobName: "myjob", Status: history.StatusFailure, StartedAt: now},
	})

	r, err := s.Last("myjob")
	if err != nil {
		t.Fatalf("Last: %v", err)
	}
	if r == nil {
		t.Fatal("expected record, got nil")
	}
	if r.Status != history.StatusFailure {
		t.Errorf("expected last record to be failure, got %q", r.Status)
	}
}
