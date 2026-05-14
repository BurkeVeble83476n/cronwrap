package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/history"
)

func tempStore(t *testing.T) *history.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := history.NewStore(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestAppendAndReadAll(t *testing.T) {
	s := tempStore(t)

	rec := history.Record{
		JobName:   "backup",
		Command:   "tar -czf /tmp/backup.tar.gz /data",
		Status:    history.StatusSuccess,
		ExitCode:  0,
		StartedAt: time.Now().UTC(),
		Duration:  2 * time.Second,
	}

	if err := s.Append(rec); err != nil {
		t.Fatalf("Append: %v", err)
	}

	records, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].JobName != rec.JobName {
		t.Errorf("JobName mismatch: got %q, want %q", records[0].JobName, rec.JobName)
	}
	if records[0].Status != history.StatusSuccess {
		t.Errorf("Status mismatch: got %q", records[0].Status)
	}
}

func TestReadAll_EmptyFile(t *testing.T) {
	s := tempStore(t)
	records, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll on missing file: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestAppend_MultipleRecords(t *testing.T) {
	s := tempStore(t)
	for i := 0; i < 5; i++ {
		err := s.Append(history.Record{
			JobName:  "job",
			Status:   history.StatusFailure,
			ExitCode: 1,
		})
		if err != nil {
			t.Fatalf("Append[%d]: %v", i, err)
		}
	}
	records, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 5 {
		t.Errorf("expected 5 records, got %d", len(records))
	}
}

func TestNewStore_CreatesParentDirs(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "a", "b", "c", "history.jsonl")
	_, err := history.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore with nested dirs: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Errorf("parent dir not created: %v", err)
	}
}
