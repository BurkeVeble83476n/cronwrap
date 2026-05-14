package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/config"
)

func writeYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "cronwrap.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeYAML: %v", err)
	}
	return p
}

func TestDefaults(t *testing.T) {
	cfg := config.Defaults()
	if cfg.HistoryPath == "" {
		t.Error("expected non-empty default HistoryPath")
	}
	if cfg.MaxHistoryRecords <= 0 {
		t.Errorf("expected positive MaxHistoryRecords, got %d", cfg.MaxHistoryRecords)
	}
	if !cfg.Alert.OnFailure {
		t.Error("expected Alert.OnFailure to be true by default")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HistoryPath == "" {
		t.Error("expected default HistoryPath when no file provided")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	path := writeYAML(t, `
history_path: /tmp/test-history.jsonl
max_history_records: 500
default_timeout: 30s
alert:
  on_failure: false
  duration_threshold: 10s
  log_level: warn
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HistoryPath != "/tmp/test-history.jsonl" {
		t.Errorf("HistoryPath = %q", cfg.HistoryPath)
	}
	if cfg.MaxHistoryRecords != 500 {
		t.Errorf("MaxHistoryRecords = %d", cfg.MaxHistoryRecords)
	}
	if cfg.DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultTimeout = %v", cfg.DefaultTimeout)
	}
	if cfg.Alert.OnFailure {
		t.Error("expected Alert.OnFailure false")
	}
	if cfg.Alert.DurationThreshold != 10*time.Second {
		t.Errorf("DurationThreshold = %v", cfg.Alert.DurationThreshold)
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	path := writeYAML(t, `alert:\n  log_level: critical\n`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for invalid log_level")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/cronwrap.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_NegativeMaxHistory(t *testing.T) {
	path := writeYAML(t, "max_history_records: -1\n")
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected validation error for negative max_history_records")
	}
}
