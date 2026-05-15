package history

import (
	"context"
	"time"
)

// WatchOptions configures the Watch polling behavior.
type WatchOptions struct {
	// JobName filters events to a specific job; empty means all jobs.
	JobName string
	// Interval is how often the store is polled for new records.
	Interval time.Duration
}

// WatchEvent is emitted when a new record is detected.
type WatchEvent struct {
	Record Record
}

// Watch polls the store at the given interval and emits new records on the
// returned channel. The channel is closed when ctx is cancelled.
func Watch(ctx context.Context, s *Store, opts WatchOptions) (<-chan WatchEvent, error) {
	if opts.Interval <= 0 {
		opts.Interval = 5 * time.Second
	}

	// Snapshot the current high-water mark so we only emit new records.
	initial, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(initial))
	for _, r := range initial {
		seen[r.ID] = struct{}{}
	}

	ch := make(chan WatchEvent, 16)

	go func() {
		defer close(ch)
		ticker := time.NewTicker(opts.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				all, err := s.ReadAll()
				if err != nil {
					continue
				}
				for _, r := range all {
					if _, ok := seen[r.ID]; ok {
						continue
					}
					if opts.JobName != "" && r.JobName != opts.JobName {
						continue
					}
					seen[r.ID] = struct{}{}
					select {
					case ch <- WatchEvent{Record: r}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}
