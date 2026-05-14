package history

import (
	"time"
)

// RetentionPolicy defines rules for pruning old execution records.
type RetentionPolicy struct {
	// MaxAge is the maximum age of records to keep. Zero means no age limit.
	MaxAge time.Duration
	// MaxRecords is the maximum number of records to keep per job name.
	// Zero means no limit.
	MaxRecords int
}

// Prune removes records from the store that violate the retention policy.
// Records are pruned by age first, then by count (keeping the most recent).
func Prune(s *Store, policy RetentionPolicy) error {
	records, err := s.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	now := time.Now()
	filtered := records[:0]

	for _, r := range records {
		if policy.MaxAge > 0 && now.Sub(r.StartedAt) > policy.MaxAge {
			continue
		}
		filtered = append(filtered, r)
	}

	if policy.MaxRecords > 0 {
		// Group by job name and enforce per-job record limit.
		byJob := make(map[string][]Record)
		for _, r := range filtered {
			byJob[r.JobName] = append(byJob[r.JobName], r)
		}
		filtered = filtered[:0]
		for _, recs := range byJob {
			if len(recs) > policy.MaxRecords {
				recs = recs[len(recs)-policy.MaxRecords:]
			}
			filtered = append(filtered, recs...)
		}
	}

	return s.rewrite(filtered)
}
