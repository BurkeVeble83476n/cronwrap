package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/cronwrap/internal/history"
)

func writeTempWatchDB(t *testing.T, records []history.Record) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	return path
}

func TestRunWatch_InvalidInterval(t *testing.T) {
	err := runWatch([]string{"--interval", "notaduration"})
	if err == nil {
		t.Fatal("expected error for invalid interval")
	}
}

func TestRunWatch_MissingDBFlag(t *testing.T) {
	err := runWatch([]string{"--db"})
	if err == nil {
		t.Fatal("expected error when --db has no value")
	}
}

func TestRunWatch_MissingJobFlag(t *testing.T) {
	err := runWatch([]string{"--job"})
	if err == nil {
		t.Fatal("expected error when --job has no value")
	}
}

func TestRunWatch_MissingIntervalFlag(t *testing.T) {
	err := runWatch([]string{"--interval"})
	if err == nil {
		t.Fatal("expected error when --interval has no value")
	}
}

func TestRunWatch_NonExistentDB(t *testing.T) {
	// A missing DB file should surface an error from NewStore.
	err := runWatch([]string{"--db", "/nonexistent/path/history.jsonl", "--interval", "10ms"})
	if err == nil {
		t.Fatal("expected error for non-existent DB directory")
	}
}

func TestDefaultHistoryPath_NotEmpty(t *testing.T) {
	p := defaultHistoryPath()
	if p == "" {
		t.Fatal("defaultHistoryPath returned empty string")
	}
}

// Ensure WatchOptions zero interval is handled gracefully (no panic).
func TestWatch_ZeroInterval_NoPanic(t *testing.T) {
	path := writeTempWatchDB(t, nil)
	s, err := history.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx, cancel := t.Context(), func() {}
	_ = cancel
	ctx2, c2 := withTimeout(ctx, 100*time.Millisecond)
	defer c2()
	_, err = history.Watch(ctx2, s, history.WatchOptions{})
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
}

// withTimeout is a local helper to avoid importing context in tests directly.
func withTimeout(parent interface{ Done() <-chan struct{} }, d time.Duration) (interface {
	Deadline() (time.Time, bool)
	Done() <-chan struct{}
	Err() error
	Value(any) any
}, func()) {
	import_ctx := struct{}{}
	_ = import_ctx
	// Use a real context via the standard library.
	import "context"
	return context.WithTimeout(parent.(context.Context), d)
}
