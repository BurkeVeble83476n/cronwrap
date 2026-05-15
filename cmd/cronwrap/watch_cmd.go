package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/cronwrap/internal/history"
)

// runWatch implements the `cronwrap watch` sub-command.
// It polls the history store and prints new records as they arrive.
func runWatch(args []string) error {
	dbPath := defaultHistoryPath()
	jobFilter := ""
	interval := 5 * time.Second

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--db":
			if i+1 >= len(args) {
				return fmt.Errorf("--db requires a value")
			}
			i++
			dbPath = args[i]
		case "--job":
			if i+1 >= len(args) {
				return fmt.Errorf("--job requires a value")
			}
			i++
			jobFilter = args[i]
		case "--interval":
			if i+1 >= len(args) {
				return fmt.Errorf("--interval requires a value")
			}
			i++
			var err error
			interval, err = time.ParseDuration(args[i])
			if err != nil {
				return fmt.Errorf("invalid interval %q: %w", args[i], err)
			}
		}
	}

	store, err := history.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ch, err := history.Watch(ctx, store, history.WatchOptions{
		JobName:  jobFilter,
		Interval: interval,
	})
	if err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	fmt.Fprintf(os.Stderr, "watching %s (interval=%s)…\n", dbPath, interval)
	for ev := range ch {
		r := ev.Record
		fmt.Printf("%s\t%s\t%s\t%s\n",
			r.Started.Format(time.RFC3339),
			r.JobName,
			r.Status,
			r.Duration,
		)
	}
	return nil
}
