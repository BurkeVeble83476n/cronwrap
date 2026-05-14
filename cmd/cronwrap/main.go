// Command cronwrap is a drop-in cron job wrapper that adds structured logging,
// alerting, and execution history to any shell command.
//
// Usage:
//
//	cronwrap [flags] -- <command> [args...]
//
// Example:
//
//	cronwrap --job-name backup --config /etc/cronwrap.yaml -- /usr/local/bin/backup.sh
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/yourorg/cronwrap/internal/alert"
	"github.com/yourorg/cronwrap/internal/config"
	"github.com/yourorg/cronwrap/internal/history"
	"github.com/yourorg/cronwrap/internal/runner"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		jobName    = flag.String("job-name", "", "Unique name for this cron job (required)")
		configPath = flag.String("config", "", "Path to cronwrap config file (optional)")
		timeout    = flag.Duration("timeout", 0, "Maximum execution time (e.g. 30s, 5m). 0 means no timeout.")
	)
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "cronwrap: no command specified\n")
		fmt.Fprintf(os.Stderr, "Usage: cronwrap [flags] -- <command> [args...]\n")
		flag.PrintDefaults()
		return 2
	}

	if *jobName == "" {
		// Fall back to the base name of the command being wrapped.
		*jobName = filepath.Base(args[0])
	}

	// Load configuration (falls back to defaults when path is empty).
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cronwrap: failed to load config: %v\n", err)
		return 2
	}

	// Set up structured logger.
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	// Open history store.
	store, err := history.NewStore(cfg.HistoryPath)
	if err != nil {
		logger.Error("failed to open history store", "error", err)
		return 2
	}

	// Respect OS signals for graceful cancellation.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	logger.Info("job started", "job", *jobName, "command", args)
	start := time.Now()

	result := runner.Run(ctx, args[0], args[1:]...)

	duration := time.Since(start)
	logger.Info("job finished",
		"job", *jobName,
		"exit_code", result.ExitCode,
		"duration_ms", duration.Milliseconds(),
	)

	// Persist execution record.
	record := history.Record{
		JobName:   *jobName,
		StartedAt: start,
		Duration:  duration,
		ExitCode:  result.ExitCode,
		Stdout:    result.Stdout,
		Stderr:    result.Stderr,
	}
	if appendErr := store.Append(record); appendErr != nil {
		logger.Warn("failed to write history record", "error", appendErr)
	}

	// Evaluate and dispatch alerts.
	notifier := alert.NewLogNotifier(logger)
	if alert.ShouldAlert(result, cfg.Alert) {
		a := alert.BuildAlert(*jobName, result, duration)
		if notifyErr := notifier.Notify(ctx, a); notifyErr != nil {
			logger.Warn("alert notification failed", "error", notifyErr)
		}
	}

	return result.ExitCode
}
