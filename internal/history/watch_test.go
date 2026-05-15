package history

import (
	"context"
	"testing"
	"time"
)

func seedWatchStore(t *testing.T) *Store {
	t.Helper()
	s := tempStore(t)
	if err := s.Append(Record{
		ID:      "w1",
		JobName: "backup",
		Status:  "success",
		Started: time.Now(),
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return s
}

func TestWatch_EmitsNewRecords(t *testing.T) {
	s := seedWatchStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch, err := Watch(ctx, s, WatchOptions{Interval: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	// Append a new record after Watch has started.
	time.Sleep(80 * time.Millisecond)
	_ = s.Append(Record{ID: "w2", JobName: "backup", Status: "success", Started: time.Now()})

	select {
	case ev := <-ch:
		if ev.Record.ID != "w2" {
			t.Errorf("expected w2, got %s", ev.Record.ID)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for watch event")
	}
}

func TestWatch_FiltersJobName(t *testing.T) {
	s := tempStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch, err := Watch(ctx, s, WatchOptions{JobName: "deploy", Interval: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	time.Sleep(80 * time.Millisecond)
	_ = s.Append(Record{ID: "x1", JobName: "backup", Status: "success", Started: time.Now()})
	_ = s.Append(Record{ID: "x2", JobName: "deploy", Status: "success", Started: time.Now()})

	select {
	case ev := <-ch:
		if ev.Record.JobName != "deploy" {
			t.Errorf("expected deploy, got %s", ev.Record.JobName)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for deploy event")
	}
}

func TestWatch_ClosesOnCancel(t *testing.T) {
	s := tempStore(t)
	ctx, cancel := context.WithCancel(context.Background())

	ch, err := Watch(ctx, s, WatchOptions{Interval: 30 * time.Millisecond})
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	cancel()

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return // channel closed as expected
			}
		case <-deadline:
			t.Fatal("channel not closed after cancel")
		}
	}
}

func TestWatch_DefaultInterval(t *testing.T) {
	s := tempStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Zero interval should default to 5s without panicking.
	_, err := Watch(ctx, s, WatchOptions{})
	if err != nil {
		t.Fatalf("Watch with zero interval: %v", err)
	}
}
