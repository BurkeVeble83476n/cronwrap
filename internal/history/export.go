package history

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Format represents the output format for exported records.
type Format string

const (
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
)

// ExportJSON writes all records from the store as a JSON array to w.
func ExportJSON(store *Store, w io.Writer, opts QueryOptions) error {
	records, err := Query(store, opts)
	if err != nil {
		return fmt.Errorf("export json: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(records); err != nil {
		return fmt.Errorf("export json encode: %w", err)
	}
	return nil
}

// ExportCSV writes all records from the store as CSV to w.
// Columns: job_name, exit_code, status, started_at, finished_at, duration_ms, stdout, stderr
func ExportCSV(store *Store, w io.Writer, opts QueryOptions) error {
	records, err := Query(store, opts)
	if err != nil {
		return fmt.Errorf("export csv: %w", err)
	}

	cw := csv.NewWriter(w)
	header := []string{"job_name", "exit_code", "status", "started_at", "finished_at", "duration_ms", "stdout", "stderr"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("export csv header: %w", err)
	}

	for _, r := range records {
		row := []string{
			r.JobName,
			fmt.Sprintf("%d", r.ExitCode),
			r.Status,
			r.StartedAt.Format(time.RFC3339),
			r.FinishedAt.Format(time.RFC3339),
			fmt.Sprintf("%d", r.FinishedAt.Sub(r.StartedAt).Milliseconds()),
			r.Stdout,
			r.Stderr,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("export csv row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}
