package history

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// JobSummary holds a human-readable summary for a single job.
type JobSummary struct {
	JobName      string
	TotalRuns    int
	SuccessCount int
	FailureCount int
	SuccessRate  float64
	AvgDuration  time.Duration
	LastStatus   string
	LastRun      time.Time
}

// Summarize computes a JobSummary for each distinct job name found in the
// store. Records are filtered by jobName when non-empty.
func Summarize(s *Store, jobName string) ([]JobSummary, error) {
	records, err := s.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("summary: read all: %w", err)
	}

	type agg struct {
		total    int
		success  int
		failure  int
		durSum   time.Duration
		lastStat string
		lastRun  time.Time
	}

	buckets := make(map[string]*agg)
	order := []string{}

	for _, r := range records {
		if jobName != "" && r.JobName != jobName {
			continue
		}
		a, exists := buckets[r.JobName]
		if !exists {
			a = &agg{}
			buckets[r.JobName] = a
			order = append(order, r.JobName)
		}
		a.total++
		if r.ExitCode == 0 {
			a.success++
		} else {
			a.failure++
		}
		a.durSum += r.Duration
		if r.StartedAt.After(a.lastRun) {
			a.lastRun = r.StartedAt
			a.lastStat = r.Status
		}
	}

	out := make([]JobSummary, 0, len(order))
	for _, name := range order {
		a := buckets[name]
		var rate float64
		if a.total > 0 {
			rate = float64(a.success) / float64(a.total) * 100
		}
		out = append(out, JobSummary{
			JobName:      name,
			TotalRuns:    a.total,
			SuccessCount: a.success,
			FailureCount: a.failure,
			SuccessRate:  rate,
			AvgDuration:  a.durSum / time.Duration(a.total),
			LastStatus:   a.lastStat,
			LastRun:      a.lastRun,
		})
	}
	return out, nil
}

// PrintSummary writes a formatted table of JobSummary entries to w.
func PrintSummary(w io.Writer, summaries []JobSummary) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "JOB\tRUNS\tSUCCESS\tFAILURE\tSUCCESS%\tAVG DURATION\tLAST STATUS\tLAST RUN")
	for _, s := range summaries {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%.1f%%\t%s\t%s\t%s\n",
			s.JobName,
			s.TotalRuns,
			s.SuccessCount,
			s.FailureCount,
			s.SuccessRate,
			s.AvgDuration.Round(time.Millisecond),
			s.LastStatus,
			s.LastRun.Format(time.RFC3339),
		)
	}
	tw.Flush()
}
