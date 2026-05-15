package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// schemaVersion is the current version of the history file format.
// Increment this when making breaking changes to the record structure.
const schemaVersion = 1

// fileHeader is written as the first line of a history file to identify
// the schema version. This allows future migrations to detect and upgrade
// older files automatically.
type fileHeader struct {
	SchemaVersion int `json:"schema_version"`
}

// MigrateStore inspects the history file at the given path and applies any
// necessary schema migrations to bring it up to the current version.
// If the file does not exist or is empty, no migration is performed.
// A backup of the original file is created before any migration is applied.
func MigrateStore(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || (err == nil && info.Size() == 0) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("migrate: stat %s: %w", path, err)
	}

	records, err := readAllRaw(path)
	if err != nil {
		return fmt.Errorf("migrate: read records: %w", err)
	}
	if len(records) == 0 {
		return nil
	}

	// Detect version from the first record. Pre-versioned files lack the
	// schema_version field and default to version 0.
	detected := detectVersion(records[0])
	if detected >= schemaVersion {
		return nil // already current
	}

	// Back up the original file before mutating it.
	if err := backupFile(path); err != nil {
		return fmt.Errorf("migrate: backup: %w", err)
	}

	// Apply migrations sequentially from detected version to current.
	for v := detected; v < schemaVersion; v++ {
		records, err = applyMigration(v, records)
		if err != nil {
			return fmt.Errorf("migrate: apply v%d->v%d: %w", v, v+1, err)
		}
	}

	// Rewrite the file with migrated records.
	if err := rewriteFile(path, records); err != nil {
		return fmt.Errorf("migrate: rewrite: %w", err)
	}
	return nil
}

// detectVersion returns the schema version encoded in a raw JSON record.
// Returns 0 for records that pre-date versioning.
func detectVersion(raw json.RawMessage) int {
	var h fileHeader
	if err := json.Unmarshal(raw, &h); err != nil {
		return 0
	}
	return h.SchemaVersion
}

// applyMigration transforms records from schema version `from` to `from+1`.
// Add a new case here whenever schemaVersion is incremented.
func applyMigration(from int, records []json.RawMessage) ([]json.RawMessage, error) {
	switch from {
	case 0:
		// v0 -> v1: stamp each record with schema_version=1.
		return stampVersion(records, 1)
	default:
		return nil, fmt.Errorf("no migration defined for version %d", from)
	}
}

// stampVersion injects a schema_version field into each raw JSON object.
func stampVersion(records []json.RawMessage, version int) ([]json.RawMessage, error) {
	out := make([]json.RawMessage, 0, len(records))
	for _, raw := range records {
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		m["schema_version"] = version
		stamped, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		out = append(out, json.RawMessage(stamped))
	}
	return out, nil
}

// backupFile copies path to path.bak, overwriting any existing backup.
func backupFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path+".bak", data, 0o644)
}

// rewriteFile writes migrated records back to path as newline-delimited JSON.
func rewriteFile(path string, records []json.RawMessage) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}
