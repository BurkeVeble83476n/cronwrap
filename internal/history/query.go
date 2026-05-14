package history

import "time"

// Filter holds optional criteria for selecting records.
type Filter struct {
	JobName string
	Status  Status
	Since   time.Time
	Limit   int
}

// Query returns records from the store that match the given filter.
// Results are returned in chronological order (oldest first).
// A zero-value Filter returns all records up to Limit (0 means no limit).
func (s *Store) Query(f Filter) ([]Record, error) {
	all, err := s.ReadAll()
	if err != nil {
		return nil, err
	}

	var result []Record
	for _, r := range all {
		if f.JobName != "" && r.JobName != f.JobName {
			continue
		}
		if f.Status != "" && r.Status != f.Status {
			continue
		}
		if !f.Since.IsZero() && r.StartedAt.Before(f.Since) {
			continue
		}
		result = append(result, r)
	}

	if f.Limit > 0 && len(result) > f.Limit {
		result = result[len(result)-f.Limit:]
	}
	return result, nil
}

// Last returns the most recent record for the given job name, or nil if none.
func (s *Store) Last(jobName string) (*Record, error) {
	records, err := s.Query(Filter{JobName: jobName, Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	r := records[len(records)-1]
	return &r, nil
}
