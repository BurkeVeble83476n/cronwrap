package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yourorg/cronwrap/internal/history"
)

// runReplay implements the `cronwrap replay` sub-command.
// It prints a chronological table of past job executions.
func runReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	jobName := fs.String("job", "", "filter by job name")
	limit := fs.Int("limit", 50, "maximum number of records to show (0 = all)")
	sinceDur := fs.Duration("since", 0, "show records from the last duration (e.g. 24h)")
	dbPath := fs.String("db", "", "path to history database (overrides config)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	path := *dbPath
	if path == "" {
		path = defaultHistoryPath()
	}

	store, err := history.NewStore(path)
	if err != nil {
		return fmt.Errorf("replay: open store: %w", err)
	}

	var since time.Time
	if *sinceDur > 0 {
		since = time.Now().Add(-*sinceDur)
	}

	res, err := history.Replay(store, history.ReplayOptions{
		JobName: *jobName,
		Since:   since,
		Limit:   *limit,
		Writer:  os.Stdout,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\n%d record(s) shown (total in store: %d)\n", res.Printed, res.Total)
	return nil
}

// defaultHistoryPath returns the default path for the history database.
func defaultHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cronwrap/history.jsonl"
	}
	return home + "/.cronwrap/history.jsonl"
}
