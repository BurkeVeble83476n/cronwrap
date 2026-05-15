// Package history provides execution history storage and querying for cronwrap.
//
// # Pipeline Support
//
// The pipeline sub-feature allows multi-step job runs to be recorded as a
// single history entry, preserving per-step metadata for later inspection.
//
// Usage:
//
//	result := history.PipelineResult{
//		JobName:   "deploy",
//		StartedAt: time.Now(),
//		Total:     duration,
//		Success:   true,
//		Steps: []history.PipelineStep{
//			{Name: "build", Command: "make build", ExitCode: 0, Duration: buildDur},
//			{Name: "push",  Command: "make push",  ExitCode: 0, Duration: pushDur},
//		},
//	}
//
//	if err := history.RecordPipeline(store, result); err != nil {
//		log.Fatal(err)
//	}
//
// Each step's name, exit code, and duration are stored as Meta keys on the
// resulting Record, making them accessible via tag and query filters.
package history
