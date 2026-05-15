package history

import (
	"testing"
	"time"
)

func seedTagStore(t *testing.T) *Store {
	t.Helper()
	s := tempStore(t)

	records := []Record{
		{JobName: "backup", Status: "success", StartedAt: time.Now(), Meta: map[string]string{"tags": "daily, infra"}},
		{JobName: "backup", Status: "failure", StartedAt: time.Now(), Meta: map[string]string{"tags": "daily, infra"}},
		{JobName: "report", Status: "success", StartedAt: time.Now(), Meta: map[string]string{"tags": "daily"}},
		{JobName: "cleanup", Status: "success", StartedAt: time.Now(), Meta: map[string]string{"tags": "weekly"}},
		{JobName: "deploy", Status: "success", StartedAt: time.Now(), Meta: nil},
	}
	for _, r := range records {
		if err := s.Append(r); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return s
}

func TestTagsForRecord_WithTags(t *testing.T) {
	r := Record{Meta: map[string]string{"tags": "alpha, beta, gamma"}}
	tags := TagsForRecord(r)
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
	if tags[1] != "beta" {
		t.Errorf("expected beta, got %s", tags[1])
	}
}

func TestTagsForRecord_NoMeta(t *testing.T) {
	r := Record{Meta: nil}
	if tags := TagsForRecord(r); tags != nil {
		t.Errorf("expected nil, got %v", tags)
	}
}

func TestTagsForRecord_EmptyTagValue(t *testing.T) {
	r := Record{Meta: map[string]string{"tags": ""}}
	if tags := TagsForRecord(r); tags != nil {
		t.Errorf("expected nil for empty tag value, got %v", tags)
	}
}

func TestBuildTagIndex(t *testing.T) {
	s := seedTagStore(t)
	idx, err := BuildTagIndex(s)
	if err != nil {
		t.Fatalf("BuildTagIndex: %v", err)
	}

	dailyJobs := idx["daily"]
	if len(dailyJobs) != 2 {
		t.Errorf("expected 2 jobs for 'daily', got %d: %v", len(dailyJobs), dailyJobs)
	}

	if len(idx["weekly"]) != 1 || idx["weekly"][0] != "cleanup" {
		t.Errorf("expected [cleanup] for 'weekly', got %v", idx["weekly"])
	}

	if _, ok := idx["infra"]; !ok {
		t.Error("expected 'infra' tag in index")
	}
}

func TestFilterByTag(t *testing.T) {
	s := seedTagStore(t)

	records, err := FilterByTag(s, "weekly")
	if err != nil {
		t.Fatalf("FilterByTag: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].JobName != "cleanup" {
		t.Errorf("expected cleanup, got %s", records[0].JobName)
	}
}

func TestFilterByTag_MultipleMatches(t *testing.T) {
	s := seedTagStore(t)

	records, err := FilterByTag(s, "daily")
	if err != nil {
		t.Fatalf("FilterByTag: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records for 'daily', got %d", len(records))
	}
}

func TestFilterByTag_NoMatch(t *testing.T) {
	s := seedTagStore(t)

	records, err := FilterByTag(s, "nonexistent")
	if err != nil {
		t.Fatalf("FilterByTag: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}
