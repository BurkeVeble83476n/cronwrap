// Package history — Watch subsystem
//
// Watch provides a lightweight polling mechanism that emits WatchEvent values
// whenever new records are appended to a history Store.  It is intended for
// long-running processes (dashboards, alerting sidecars) that want to react to
// job completions in near-real-time without subscribing to filesystem events.
//
// Usage:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	ch, err := history.Watch(ctx, store, history.WatchOptions{
//		JobName:  "nightly-backup",
//		Interval: 10 * time.Second,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	for ev := range ch {
//		fmt.Printf("new record: %+v\n", ev.Record)
//	}
//
// The channel is closed automatically when the context is cancelled.
package history
