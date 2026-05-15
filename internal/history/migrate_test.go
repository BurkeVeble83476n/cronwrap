package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeLegacyV0Records(t *testing.T, path string, records []map[string]any) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("writeLegacyV0Records: create: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			t.Fatalf("writeLegacyV0Records: encode: %v", err)
		}
	}
}

func TestMigrateStore_AlreadyCurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// Append a record so the file is non-empty and stamped.
	rec := Record{
		JobName:   "noop",
		StartedAt: time.Now().UTC(),
		ExitCode:  0,
	}
	if err := store.Append(rec); err != nil {
		t.Fatalf("Append: %v", err)
	}

	if err := MigrateStore(path); err != nil {
		t.Fatalf("MigrateStore on current-version store: %v", err)
	}

	records, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll after migration: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record after no-op migration, got %d", len(records))
	}
}

func TestMigrateStore_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")

	// Create an empty file.
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := MigrateStore(path); err != nil {
		t.Fatalf("MigrateStore on empty file: %v", err)
	}
}

func TestMigrateStore_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")

	// Should succeed gracefully when the file does not yet exist.
	if err := MigrateStore(path); err != nil {
		t.Fatalf("MigrateStore on missing file: %v", err)
	}
}

func TestMigrateStore_V0ToCurrentPreservesRecords(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")

	now := time.Now().UTC().Truncate(time.Second)

	// Write raw v0-style records (no schema_version line).
	legacy := []map[string]any{
		{
			"job_name":   "backup",
			"started_at": now.Format(time.RFC3339),
			"exit_code":  0,
			"duration_ms": 120,
		},
		{
			"job_name":   "backup",
			"started_at": now.Add(-time.Hour).Format(time.RFC3339),
			"exit_code":  1,
			"duration_ms": 45,
		},
	}
	writeLegacyV0Records(t, path, legacy)

	if err := MigrateStore(path); err != nil {
		t.Fatalf("MigrateStore: %v", err)
	}

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore after migration: %v", err)
	}
	records, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll after migration: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records after migration, got %d", len(records))
	}
	if records[0].JobName != "backup" {
		t.Errorf("record[0].JobName = %q, want %q", records[0].JobName, "backup")
	}
	if records[1].ExitCode != 1 {
		t.Errorf("record[1].ExitCode = %d, want 1", records[1].ExitCode)
	}
}

func TestDetectVersion_NoVersionLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")

	legacy := []map[string]any{
		{"job_name": "ping", "exit_code": 0},
	}
	writeLegacyV0Records(t, path, legacy)

	v, err := detectVersion(path)
	if err != nil {
		t.Fatalf("detectVersion: %v", err)
	}
	if v != 0 {
		t.Errorf("detectVersion = %d, want 0 for legacy file", v)
	}
}
