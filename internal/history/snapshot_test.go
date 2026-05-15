package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func seedSnapshotStore(t *testing.T) *Store {
	t.Helper()
	store := tempStore(t)
	for i := 0; i < 3; i++ {
		_ = store.Append(Record{
			ID:        fmt.Sprintf("snap-%d", i),
			JobName:   "backup",
			StartedAt: time.Now().UTC(),
			Status:    "success",
			ExitCode:  0,
		})
	}
	return store
}

func TestTakeSnapshot_CreatesFile(t *testing.T) {
	store := seedSnapshotStore(t)
	dest := filepath.Join(t.TempDir(), "snap.json")

	snap, err := TakeSnapshot(store, dest)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}
	if len(snap.Records) != 3 {
		t.Errorf("expected 3 records, got %d", len(snap.Records))
	}
	if snap.Version != CurrentVersion {
		t.Errorf("expected version %d, got %d", CurrentVersion, snap.Version)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("snapshot file not created: %v", err)
	}
}

func TestLoadSnapshot_RoundTrip(t *testing.T) {
	store := seedSnapshotStore(t)
	dest := filepath.Join(t.TempDir(), "snap.json")

	_, err := TakeSnapshot(store, dest)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}

	loaded, err := LoadSnapshot(dest)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if len(loaded.Records) != 3 {
		t.Errorf("expected 3 records after load, got %d", len(loaded.Records))
	}
}

func TestRestoreSnapshot_ReplacesContents(t *testing.T) {
	origStore := seedSnapshotStore(t)
	dest := filepath.Join(t.TempDir(), "snap.json")
	snap, _ := TakeSnapshot(origStore, dest)

	targetStore := tempStore(t)
	// pre-populate with one unrelated record
	_ = targetStore.Append(Record{ID: "old", JobName: "old-job", Status: "success"})

	if err := RestoreSnapshot(snap, targetStore); err != nil {
		t.Fatalf("RestoreSnapshot: %v", err)
	}

	records, err := targetStore.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll after restore: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 restored records, got %d", len(records))
	}
}

func TestLoadSnapshot_MissingFile(t *testing.T) {
	_, err := LoadSnapshot("/nonexistent/snap.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
