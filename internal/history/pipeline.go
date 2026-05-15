// Package history provides execution history storage and querying for cronwrap.
package history

import (
	"fmt"
	"time"
)

// PipelineStep represents a single step in a multi-stage pipeline run.
type PipelineStep struct {
	Name     string        `json:"name"`
	Command  string        `json:"command"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration_ns"`
	Stdout   string        `json:"stdout,omitempty"`
	Stderr   string        `json:"stderr,omitempty"`
}

// PipelineResult holds the aggregated result of a multi-step pipeline.
type PipelineResult struct {
	JobName   string         `json:"job_name"`
	Steps     []PipelineStep `json:"steps"`
	StartedAt time.Time      `json:"started_at"`
	Total     time.Duration  `json:"total_ns"`
	Success   bool           `json:"success"`
}

// RecordPipeline appends a PipelineResult to the store as a single Record,
// encoding step details into the record's Meta field.
func RecordPipeline(s *Store, result PipelineResult) error {
	if result.JobName == "" {
		return fmt.Errorf("pipeline: job name must not be empty")
	}

	status := "success"
	exitCode := 0
	if !result.Success {
		status = "failure"
		// Use the exit code of the first failed step.
		for _, step := range result.Steps {
			if step.ExitCode != 0 {
				exitCode = step.ExitCode
				break
			}
		}
	}

	meta := map[string]string{
		"pipeline_steps": fmt.Sprintf("%d", len(result.Steps)),
	}
	for i, step := range result.Steps {
		meta[fmt.Sprintf("step_%d_name", i)] = step.Name
		meta[fmt.Sprintf("step_%d_exit_code", i)] = fmt.Sprintf("%d", step.ExitCode)
		meta[fmt.Sprintf("step_%d_duration_ns", i)] = fmt.Sprintf("%d", step.Duration.Nanoseconds())
	}

	rec := Record{
		JobName:   result.JobName,
		StartedAt: result.StartedAt,
		Duration:  result.Total,
		ExitCode:  exitCode,
		Status:    status,
		Meta:      meta,
	}

	return s.Append(rec)
}
