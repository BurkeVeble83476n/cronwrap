package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yourorg/cronwrap/internal/history"
)

// runSnapshot handles the "snapshot" sub-command.
//
//	cronwrap snapshot --db <path> --out <dest>
//	cronwrap snapshot restore --db <path> --src <file>
func runSnapshot(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("snapshot: expected sub-command: take|restore")
	}

	switch args[0] {
	case "take":
		return runSnapshotTake(args[1:])
	case "restore":
		return runSnapshotRestore(args[1:])
	default:
		return fmt.Errorf("snapshot: unknown sub-command %q", args[0])
	}
}

func runSnapshotTake(args []string) error {
	fs := flag.NewFlagSet("snapshot take", flag.ContinueOnError)
	dbPath := fs.String("db", defaultHistoryPath(), "history DB path")
	outPath := fs.String("out", "", "destination snapshot file (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *outPath == "" {
		*outPath = fmt.Sprintf("cronwrap-snapshot-%s.json", time.Now().Format("20060102-150405"))
	}

	store, err := history.NewStore(*dbPath)
	if err != nil {
		return fmt.Errorf("snapshot take: open store: %w", err)
	}

	snap, err := history.TakeSnapshot(store, *outPath)
	if err != nil {
		return fmt.Errorf("snapshot take: %w", err)
	}

	fmt.Fprintf(os.Stdout, "snapshot: wrote %d records to %s\n", len(snap.Records), *outPath)
	return nil
}

func runSnapshotRestore(args []string) error {
	fs := flag.NewFlagSet("snapshot restore", flag.ContinueOnError)
	dbPath := fs.String("db", defaultHistoryPath(), "history DB path")
	srcPath := fs.String("src", "", "snapshot file to restore (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *srcPath == "" {
		return fmt.Errorf("snapshot restore: --src is required")
	}

	snap, err := history.LoadSnapshot(*srcPath)
	if err != nil {
		return fmt.Errorf("snapshot restore: load: %w", err)
	}

	store, err := history.NewStore(*dbPath)
	if err != nil {
		return fmt.Errorf("snapshot restore: open store: %w", err)
	}

	if err := history.RestoreSnapshot(snap, store); err != nil {
		return fmt.Errorf("snapshot restore: %w", err)
	}

	fmt.Fprintf(os.Stdout, "snapshot: restored %d records from %s\n", len(snap.Records), *srcPath)
	return nil
}
