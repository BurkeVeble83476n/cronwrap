package history

import (
	"fmt"
	"os"
	"time"
)

// GCOptions configures the garbage collection behaviour for the history store.
type GCOptions struct {
	// MaxAge is the maximum age of records to retain. Records older than this
	// duration will be removed. Zero means no age-based eviction.
	MaxAge time.Duration

	// MaxRecords is the maximum total number of records to retain across all
	// jobs. When the store exceeds this limit the oldest records are removed
	// first. Zero means no count-based eviction.
	MaxRecords int

	// BackupBeforeGC controls whether a timestamped backup of the store file
	// is written before any records are deleted. Useful for auditing.
	BackupBeforeGC bool
}

// GCResult summarises the outcome of a garbage-collection run.
type GCResult struct {
	// Removed is the number of records that were deleted.
	Removed int

	// Remaining is the number of records left in the store after GC.
	Remaining int

	// BackupPath is the path of the backup file created before GC, or an
	// empty string when BackupBeforeGC was false or no records were removed.
	BackupPath string

	// Duration is the wall-clock time taken by the GC run.
	Duration time.Duration
}

// String returns a human-readable summary of the GC result.
func (r GCResult) String() string {
	if r.Removed == 0 {
		return fmt.Sprintf("gc: nothing to remove (%d records retained, took %s)",
			r.Remaining, r.Duration.Round(time.Millisecond))
	}
	msg := fmt.Sprintf("gc: removed %d record(s), %d remaining (took %s)",
		r.Removed, r.Remaining, r.Duration.Round(time.Millisecond))
	if r.BackupPath != "" {
		msg += fmt.Sprintf("; backup written to %s", r.BackupPath)
	}
	return msg
}

// GC runs garbage collection on the store identified by path using the
// supplied options. It is a convenience wrapper around Prune that adds
// optional pre-GC backup and timing.
//
// If both MaxAge and MaxRecords are zero GC is a no-op and returns
// immediately with a zero GCResult.
func GC(path string, opts GCOptions) (GCResult, error) {
	if opts.MaxAge == 0 && opts.MaxRecords == 0 {
		return GCResult{}, nil
	}

	start := time.Now()

	// Count records before pruning so we can calculate how many were removed.
	store, err := NewStore(path)
	if err != nil {
		return GCResult{}, fmt.Errorf("gc: open store: %w", err)
	}

	before, err := store.ReadAll()
	if err != nil {
		return GCResult{}, fmt.Errorf("gc: read store: %w", err)
	}

	if len(before) == 0 {
		return GCResult{Duration: time.Since(start)}, nil
	}

	// Optionally back up the file before we mutate it.
	var backupPath string
	if opts.BackupBeforeGC {
		backupPath, err = backupFile(path)
		if err != nil {
			return GCResult{}, fmt.Errorf("gc: backup: %w", err)
		}
	}

	pruneOpts := PruneOptions{
		MaxAge:     opts.MaxAge,
		MaxRecords: opts.MaxRecords,
	}

	if err := Prune(path, pruneOpts); err != nil {
		// If pruning failed and we made a backup, remove it to avoid clutter.
		if backupPath != "" {
			_ = os.Remove(backupPath)
		}
		return GCResult{}, fmt.Errorf("gc: prune: %w", err)
	}

	after, err := store.ReadAll()
	if err != nil {
		return GCResult{}, fmt.Errorf("gc: read store after prune: %w", err)
	}

	return GCResult{
		Removed:    len(before) - len(after),
		Remaining:  len(after),
		BackupPath: backupPath,
		Duration:   time.Since(start),
	}, nil
}
