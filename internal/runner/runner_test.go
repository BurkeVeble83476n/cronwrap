package runner_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/runner"
)

func TestRun_Success(t *testing.T) {
	ctx := context.Background()
	result := runner.Run(ctx, "echo", []string{"hello"})

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if !strings.Contains(result.Stdout, "hello") {
		t.Errorf("expected stdout to contain 'hello', got %q", result.Stdout)
	}
	if result.Err != nil {
		t.Errorf("expected no error, got %v", result.Err)
	}
	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestRun_Failure(t *testing.T) {
	ctx := context.Background()
	result := runner.Run(ctx, "false", []string{})

	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
	if result.Err == nil {
		t.Error("expected an error")
	}
}

func TestRun_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result := runner.Run(ctx, "sleep", []string{"5"})

	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code due to timeout")
	}
	if result.Err == nil {
		t.Error("expected context deadline error")
	}
}

func TestRun_StderrCaptured(t *testing.T) {
	ctx := context.Background()
	result := runner.Run(ctx, "sh", []string{"-c", "echo errout >&2"})

	if !strings.Contains(result.Stderr, "errout") {
		t.Errorf("expected stderr to contain 'errout', got %q", result.Stderr)
	}
}
