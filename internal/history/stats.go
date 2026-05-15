package history

import (
	"math"
	"time"
)

// JobStats holds aggregated statistics for a single job.
type JobStats struct {
	JobName      string
	TotalRuns    int
	SuccessCount int
	FailureCount int
	SuccessRate  float64
	AvgDuration  time.Duration
	MinDuration  time.Duration
	MaxDuration  time.Duration
	LastRun      time.Time
}

// Stats computes aggregated statistics for each job found in the store.
// If jobName is non-empty, only records for that job are considered.
func Stats(s *Store, jobName string) ([]JobStats, error) {
	records, err := s.ReadAll()
	if err != nil {
		return nil, err
	}

	type accumulator struct {
		total    int
		success  int
		durations []time.Duration
		lastRun  time.Time
	}

	acc := make(map[string]*accumulator)

	for _, r := range records {
		if jobName != "" && r.JobName != jobName {
			continue
		}
		a, ok := acc[r.JobName]
		if !ok {
			a = &accumulator{}
			acc[r.JobName] = a
		}
		a.total++
		if r.ExitCode == 0 {
			a.success++
		}
		a.durations = append(a.durations, r.Duration)
		if r.StartedAt.After(a.lastRun) {
			a.lastRun = r.StartedAt
		}
	}

	result := make([]JobStats, 0, len(acc))
	for name, a := range acc {
		st := JobStats{
			JobName:      name,
			TotalRuns:    a.total,
			SuccessCount: a.success,
			FailureCount: a.total - a.success,
			LastRun:      a.lastRun,
		}
		if a.total > 0 {
			st.SuccessRate = math.Round(float64(a.success)/float64(a.total)*10000) / 100
		}
		var sum time.Duration
		st.MinDuration = a.durations[0]
		st.MaxDuration = a.durations[0]
		for _, d := range a.durations {
			sum += d
			if d < st.MinDuration {
				st.MinDuration = d
			}
			if d > st.MaxDuration {
				st.MaxDuration = d
			}
		}
		st.AvgDuration = sum / time.Duration(a.total)
		result = append(result, st)
	}
	return result, nil
}
