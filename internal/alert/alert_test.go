package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/alert"
)

func TestShouldAlert_NonZeroExitCode(t *testing.T) {
	if !alert.ShouldAlert(1, time.Second, 0) {
		t.Error("expected ShouldAlert=true for non-zero exit code")
	}
}

func TestShouldAlert_ExceedsThreshold(t *testing.T) {
	if !alert.ShouldAlert(0, 10*time.Second, 5*time.Second) {
		t.Error("expected ShouldAlert=true when duration exceeds threshold")
	}
}

func TestShouldAlert_WithinThreshold(t *testing.T) {
	if alert.ShouldAlert(0, 2*time.Second, 5*time.Second) {
		t.Error("expected ShouldAlert=false when duration is within threshold")
	}
}

func TestShouldAlert_NoThreshold(t *testing.T) {
	if alert.ShouldAlert(0, 10*time.Second, 0) {
		t.Error("expected ShouldAlert=false when threshold is zero and exit code is 0")
	}
}

func TestBuildAlert_FailureLevel(t *testing.T) {
	a := alert.BuildAlert("backup", 2, 3*time.Second, 0)
	if a.Level != alert.LevelError {
		t.Errorf("expected level ERROR, got %s", a.Level)
	}
	if a.JobName != "backup" {
		t.Errorf("unexpected job name: %s", a.JobName)
	}
	if !strings.Contains(a.Message, "exit code 2") {
		t.Errorf("expected message to contain exit code, got: %s", a.Message)
	}
}

func TestBuildAlert_ThresholdExceeded(t *testing.T) {
	a := alert.BuildAlert("report", 0, 20*time.Second, 10*time.Second)
	if a.Level != alert.LevelWarn {
		t.Errorf("expected level WARN, got %s", a.Level)
	}
	if !strings.Contains(a.Message, "threshold") {
		t.Errorf("expected threshold message, got: %s", a.Message)
	}
}

func TestBuildAlert_Success(t *testing.T) {
	a := alert.BuildAlert("sync", 0, 1*time.Second, 5*time.Second)
	if a.Level != alert.LevelInfo {
		t.Errorf("expected level INFO, got %s", a.Level)
	}
}

func TestLogNotifier_Notify(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewLogNotifier(&buf)

	a := alert.BuildAlert("cleanup", 1, 2*time.Second, 0)
	if err := n.Notify(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"ERROR", "cleanup", "exit_code=1"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got: %s", want, out)
		}
	}
}

func TestNewLogNotifier_DefaultsToStderr(t *testing.T) {
	n := alert.NewLogNotifier(nil)
	if n.Out == nil {
		t.Error("expected non-nil writer when nil is passed")
	}
}
