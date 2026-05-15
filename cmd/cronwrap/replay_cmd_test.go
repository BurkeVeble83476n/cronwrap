package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/history"
)

func writeTempHistory(t *testing.T, records []history.Record) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	store, err := history.NewStore(path)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	for _, r := range records {
		if err := store.Append(r); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	return path
}

func TestRunReplay_NoArgs(t *testing.T) {
	path := writeTempHistory(t, []history.Record{
		{JobName: "myjob", Status: "success", ExitCode: 0, StartedAt: time.Now(), Duration: time.Second},
	})
	err := runReplay([]string{"-db", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReplay_WithJobFilter(t *testing.T) {
	path := writeTempHistory(t, []history.Record{
		{JobName: "alpha", Status: "success", ExitCode: 0, StartedAt: time.Now(), Duration: time.Second},
		{JobName: "beta", Status: "failure", ExitCode: 1, StartedAt: time.Now(), Duration: time.Second},
	})
	err := runReplay([]string{"-db", path, "-job", "alpha"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReplay_MissingDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "history.jsonl")
	// NewStore creates parent dirs, so this should succeed with empty output.
	err := runReplay([]string{"-db", path})
	if err != nil {
		t.Fatalf("unexpected error for missing db: %v", err)
	}
}

func TestDefaultHistoryPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := defaultHistoryPath()
	if home != "" && got == ".cronwrap/history.jsonl" {
		t.Error("expected home-based path when home dir is available")
	}
	if got == "" {
		t.Error("expected non-empty default history path")
	}
}
