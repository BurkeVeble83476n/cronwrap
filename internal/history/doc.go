// Package history manages the execution history of cron jobs wrapped by cronwrap.
//
// Records are stored as newline-delimited JSON (JSON-L) so that the history
// file can be appended to atomically and streamed efficiently without loading
// the entire file into memory.
//
// Typical usage:
//
//	store, err := history.NewStore("/var/lib/cronwrap/history.jsonl")
//	if err != nil { ... }
//
//	store.Append(history.Record{
//		JobName:  "db-backup",
//		Status:   history.StatusSuccess,
//		ExitCode: 0,
//		...
//	})
//
//	records, err := store.ReadAll()
package history
