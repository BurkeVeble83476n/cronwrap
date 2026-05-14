// Package history provides storage and retrieval of cron job execution records.
package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Status represents the outcome of a job execution.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusTimeout Status = "timeout"
)

// Record holds metadata about a single job execution.
type Record struct {
	JobName   string        `json:"job_name"`
	Command   string        `json:"command"`
	Status    Status        `json:"status"`
	ExitCode  int           `json:"exit_code"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration_ns"`
	Stdout    string        `json:"stdout,omitempty"`
	Stderr    string        `json:"stderr,omitempty"`
}

// Store persists execution records to a JSON-lines file.
type Store struct {
	path string
}

// NewStore creates a Store that writes records to the given file path.
// Parent directories are created if they do not exist.
func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}

// Append writes a single Record as a JSON line to the store file.
func (s *Store) Append(r Record) error {
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(r)
}

// ReadAll returns all records stored in the file.
func (s *Store) ReadAll() ([]Record, error) {
	f, err := os.Open(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var records []Record
	dec := json.NewDecoder(f)
	for dec.More() {
		var r Record
		if err := dec.Decode(&r); err != nil {
			return records, err
		}
		records = append(records, r)
	}
	return records, nil
}
