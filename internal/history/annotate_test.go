package history

import (
	"testing"
	"time"
)

func seedAnnotateStore(t *testing.T) (string, string) {
	t.Helper()
	path := tempStore(t)
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	r := Record{
		ID:        "rec-001",
		JobName:   "backup",
		StartTime: time.Now(),
		ExitCode:  0,
		Status:    "success",
	}
	if err := store.Append(r); err != nil {
		t.Fatalf("Append: %v", err)
	}
	return path, r.ID
}

func TestAnnotate_AddsAnnotation(t *testing.T) {
	path, id := seedAnnotateStore(t)
	if err := Annotate(path, id, "ticket", "JIRA-42"); err != nil {
		t.Fatalf("Annotate: %v", err)
	}
	anns, err := GetAnnotations(path, id)
	if err != nil {
		t.Fatalf("GetAnnotations: %v", err)
	}
	if len(anns) != 1 || anns[0].Key != "ticket" || anns[0].Value != "JIRA-42" {
		t.Errorf("unexpected annotations: %+v", anns)
	}
}

func TestAnnotate_UpdatesExistingKey(t *testing.T) {
	path, id := seedAnnotateStore(t)
	_ = Annotate(path, id, "note", "first")
	_ = Annotate(path, id, "note", "second")
	anns, err := GetAnnotations(path, id)
	if err != nil {
		t.Fatalf("GetAnnotations: %v", err)
	}
	if len(anns) != 1 || anns[0].Value != "second" {
		t.Errorf("expected single updated annotation, got %+v", anns)
	}
}

func TestAnnotate_EmptyKeyReturnsError(t *testing.T) {
	path, id := seedAnnotateStore(t)
	if err := Annotate(path, id, "", "value"); err == nil {
		t.Error("expected error for empty key, got nil")
	}
}

func TestAnnotate_UnknownIDReturnsError(t *testing.T) {
	path, _ := seedAnnotateStore(t)
	if err := Annotate(path, "no-such-id", "k", "v"); err == nil {
		t.Error("expected error for unknown id, got nil")
	}
}

func TestGetAnnotations_EmptyWhenNoneSet(t *testing.T) {
	path, id := seedAnnotateStore(t)
	anns, err := GetAnnotations(path, id)
	if err != nil {
		t.Fatalf("GetAnnotations: %v", err)
	}
	if len(anns) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(anns))
	}
}
