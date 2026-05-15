package history

import (
	"testing"
	"time"
)

func seedPipelineStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(t.TempDir() + "/pipeline.db")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestRecordPipeline_Success(t *testing.T) {
	s := seedPipelineStore(t)

	result := PipelineResult{
		JobName:   "deploy",
		StartedAt: time.Now(),
		Total:     2 * time.Second,
		Success:   true,
		Steps: []PipelineStep{
			{Name: "build", Command: "make build", ExitCode: 0, Duration: 1 * time.Second},
			{Name: "test", Command: "make test", ExitCode: 0, Duration: 1 * time.Second},
		},
	}

	if err := RecordPipeline(s, result); err != nil {
		t.Fatalf("RecordPipeline: %v", err)
	}

	recs, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if recs[0].Status != "success" {
		t.Errorf("expected status success, got %s", recs[0].Status)
	}
	if recs[0].Meta["pipeline_steps"] != "2" {
		t.Errorf("expected pipeline_steps=2, got %s", recs[0].Meta["pipeline_steps"])
	}
	if recs[0].Meta["step_0_name"] != "build" {
		t.Errorf("expected step_0_name=build, got %s", recs[0].Meta["step_0_name"])
	}
}

func TestRecordPipeline_Failure(t *testing.T) {
	s := seedPipelineStore(t)

	result := PipelineResult{
		JobName:   "ci",
		StartedAt: time.Now(),
		Total:     500 * time.Millisecond,
		Success:   false,
		Steps: []PipelineStep{
			{Name: "lint", Command: "golint ./...", ExitCode: 0, Duration: 100 * time.Millisecond},
			{Name: "build", Command: "go build", ExitCode: 2, Duration: 400 * time.Millisecond},
		},
	}

	if err := RecordPipeline(s, result); err != nil {
		t.Fatalf("RecordPipeline: %v", err)
	}

	recs, _ := s.ReadAll()
	if recs[0].Status != "failure" {
		t.Errorf("expected failure, got %s", recs[0].Status)
	}
	if recs[0].ExitCode != 2 {
		t.Errorf("expected exit code 2, got %d", recs[0].ExitCode)
	}
}

func TestRecordPipeline_EmptyJobName(t *testing.T) {
	s := seedPipelineStore(t)
	err := RecordPipeline(s, PipelineResult{})
	if err == nil {
		t.Error("expected error for empty job name")
	}
}

func TestRecordPipeline_StepMetaKeys(t *testing.T) {
	s := seedPipelineStore(t)

	result := PipelineResult{
		JobName:   "etl",
		StartedAt: time.Now(),
		Total:     300 * time.Millisecond,
		Success:   true,
		Steps: []PipelineStep{
			{Name: "extract", ExitCode: 0, Duration: 100 * time.Millisecond},
		},
	}

	_ = RecordPipeline(s, result)
	recs, _ := s.ReadAll()

	if _, ok := recs[0].Meta["step_0_duration_ns"]; !ok {
		t.Error("expected step_0_duration_ns in meta")
	}
}
