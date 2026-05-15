// Package history provides execution history storage, querying, and replay
// for cronwrap job records.
//
// # Replay
//
// The Replay function renders a human-readable, chronological table of past
// job executions directly to any io.Writer. It is designed to be used by CLI
// sub-commands that need to surface historical context at a glance.
//
// Usage:
//
//	res, err := history.Replay(store, history.ReplayOptions{
//		JobName: "backup",
//		Since:   time.Now().Add(-24 * time.Hour),
//		Limit:   20,
//		Writer:  os.Stdout,
//	})
//
// ReplayOptions fields:
//   - JobName: when non-empty, only records for that job are shown.
//   - Since:   when non-zero, records older than this time are excluded.
//   - Limit:   when > 0, only the N most recent matching records are shown.
//   - Writer:  destination for the rendered table (required).
package history
