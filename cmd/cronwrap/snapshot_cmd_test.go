package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/history"
)

func writeTempSnapshotDB(t *testing.T) (dbPath string) {
	t.Helper()
	dir := t.TempDir()
	dbPath = filepath.Join(dir, "history.db")
	store, err := history.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	for i := 0; i < 2; i++ {
		_ = store.Append(history.Record{
			ID:        fmt.Sprintf("r%d", i),
			JobName:   "nightly",
			StartedAt: time.Now().UTC(),
			Status:    "success",
		})
	}
	return dbPath
}

func TestRunSnapshotTake_CreatesFile(t *testing.T) {
	dbPath := writeTempSnapshotDB(t)
	out := filepath.Join(t.TempDir(), "snap.json")

	err := runSnapshot([]string{"take", "--db", dbPath, "--out", out})
	if err != nil {
		t.Fatalf("runSnapshot take: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var snap history.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if len(snap.Records) != 2 {
		t.Errorf("expected 2 records, got %d", len(snap.Records))
	}
}

func TestRunSnapshotRestore_ReplacesDB(t *testing.T) {
	dbPath := writeTempSnapshotDB(t)
	out := filepath.Join(t.TempDir(), "snap.json")
	_ = runSnapshot([]string{"take", "--db", dbPath, "--out", out})

	newDB := filepath.Join(t.TempDir(), "new.db")
	err := runSnapshot([]string{"restore", "--db", newDB, "--src", out})
	if err != nil {
		t.Fatalf("runSnapshot restore: %v", err)
	}

	store, _ := history.NewStore(newDB)
	records, _ := store.ReadAll()
	if len(records) != 2 {
		t.Errorf("expected 2 restored records, got %d", len(records))
	}
}

func TestRunSnapshot_UnknownSubcommand(t *testing.T) {
	if err := runSnapshot([]string{"bogus"}); err == nil {
		t.Error("expected error for unknown sub-command")
	}
}

func TestRunSnapshot_NoArgs(t *testing.T) {
	if err := runSnapshot([]string{}); err == nil {
		t.Error("expected error with no args")
	}
}
