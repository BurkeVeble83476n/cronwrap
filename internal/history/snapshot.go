package history

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot represents a point-in-time copy of all history records.
type Snapshot struct {
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
	Records   []Record  `json:"records"`
}

// TakeSnapshot reads all records from the store and writes a snapshot
// to the given destination path as a JSON file.
func TakeSnapshot(store *Store, destPath string) (*Snapshot, error) {
	records, err := store.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("snapshot: read records: %w", err)
	}

	snap := &Snapshot{
		CreatedAt: time.Now().UTC(),
		Version:   CurrentVersion,
		Records:   records,
	}

	f, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("snapshot: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return nil, fmt.Errorf("snapshot: encode: %w", err)
	}

	return snap, nil
}

// LoadSnapshot reads a snapshot from the given path.
func LoadSnapshot(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open: %w", err)
	}
	defer f.Close()

	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return nil, fmt.Errorf("snapshot: decode: %w", err)
	}

	return &snap, nil
}

// RestoreSnapshot writes all records from a snapshot into the target store,
// replacing its existing contents.
func RestoreSnapshot(snap *Snapshot, store *Store) error {
	if err := os.Truncate(store.path, 0); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("snapshot: truncate store: %w", err)
	}

	for i := range snap.Records {
		if err := store.Append(snap.Records[i]); err != nil {
			return fmt.Errorf("snapshot: restore record %d: %w", i, err)
		}
	}

	return nil
}
