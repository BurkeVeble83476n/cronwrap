package history

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func seedExportStore(t *testing.T) *Store {
	t.Helper()
	store := tempStore(t)
	now := time.Now().UTC().Truncate(time.Second)
	records := []Record{
		{JobName: "backup", ExitCode: 0, Status: "success", StartedAt: now.Add(-2 * time.Minute), FinishedAt: now.Add(-time.Minute), Stdout: "ok", Stderr: ""},
		{JobName: "backup", ExitCode: 1, Status: "failure", StartedAt: now.Add(-time.Minute), FinishedAt: now, Stdout: "", Stderr: "err"},
		{JobName: "sync", ExitCode: 0, Status: "success", StartedAt: now.Add(-30 * time.Second), FinishedAt: now, Stdout: "done", Stderr: ""},
	}
	for _, r := range records {
		if err := store.Append(r); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return store
}

func TestExportJSON_AllRecords(t *testing.T) {
	store := seedExportStore(t)
	var buf bytes.Buffer
	if err := ExportJSON(store, &buf, QueryOptions{}); err != nil {
		t.Fatalf("ExportJSON: %v", err)
	}
	var records []Record
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}

func TestExportJSON_FilteredByJobName(t *testing.T) {
	store := seedExportStore(t)
	var buf bytes.Buffer
	opts := QueryOptions{JobName: "sync"}
	if err := ExportJSON(store, &buf, opts); err != nil {
		t.Fatalf("ExportJSON: %v", err)
	}
	var records []Record
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
	if records[0].JobName != "sync" {
		t.Errorf("expected job sync, got %s", records[0].JobName)
	}
}

func TestExportCSV_HeaderAndRows(t *testing.T) {
	store := seedExportStore(t)
	var buf bytes.Buffer
	if err := ExportCSV(store, &buf, QueryOptions{}); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}
	r := csv.NewReader(strings.NewReader(buf.String()))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv read: %v", err)
	}
	// 1 header + 3 data rows
	if len(rows) != 4 {
		t.Errorf("expected 4 rows (header+3), got %d", len(rows))
	}
	if rows[0][0] != "job_name" {
		t.Errorf("expected header job_name, got %s", rows[0][0])
	}
}

func TestExportCSV_EmptyStore(t *testing.T) {
	store := tempStore(t)
	var buf bytes.Buffer
	if err := ExportCSV(store, &buf, QueryOptions{}); err != nil {
		t.Fatalf("ExportCSV empty: %v", err)
	}
	r := csv.NewReader(strings.NewReader(buf.String()))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv read: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected only header row, got %d rows", len(rows))
	}
}
