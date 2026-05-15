package history

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// ReplayOptions controls which records are replayed.
type ReplayOptions struct {
	JobName string
	Since   time.Time
	Limit   int
	Writer  io.Writer
}

// ReplayResult holds the outcome of a replay operation.
type ReplayResult struct {
	Total    int
	Printed  int
	JobNames []string
}

// Replay prints a human-readable chronological replay of job execution
// records from the store, applying optional filters.
func Replay(s *Store, opts ReplayOptions) (*ReplayResult, error) {
	records, err := s.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("replay: read store: %w", err)
	}

	var filtered []Record
	for _, r := range records {
		if opts.JobName != "" && r.JobName != opts.JobName {
			continue
		}
		if !opts.Since.IsZero() && r.StartedAt.Before(opts.Since) {
			continue
		}
		filtered = append(filtered, r)
	}

	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[len(filtered)-opts.Limit:]
	}

	seen := map[string]struct{}{}
	res := &ReplayResult{Total: len(records), Printed: len(filtered)}

	w := tabwriter.NewWriter(opts.Writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIME\tJOB\tSTATUS\tDURATION\tEXIT")
	for _, r := range filtered {
		seen[r.JobName] = struct{}{}
		status := r.Status
		if status == "" {
			status = "unknown"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
			r.StartedAt.Format(time.RFC3339),
			r.JobName,
			status,
			r.Duration.Round(time.Millisecond),
			r.ExitCode,
		)
	}
	_ = w.Flush()

	for name := range seen {
		res.JobNames = append(res.JobNames, name)
	}
	return res, nil
}
