// Package history — tag sub-feature
//
// Tags allow operators to group cron jobs by arbitrary labels (e.g.
// "daily", "infra", "critical") without changing job names.  Tags are
// stored as a comma-separated string in the Record.Meta map under the
// key "tags".
//
// Example usage:
//
//	// Write a tagged record
//	err := store.Append(history.Record{
//		JobName: "db-backup",
//		Status:  "success",
//		Meta:    map[string]string{"tags": "daily, infra"},
//	})
//
//	// Retrieve all records carrying a specific tag
//	records, err := history.FilterByTag(store, "infra")
//
//	// Build an index of tag → job names for dashboards / reports
//	idx, err := history.BuildTagIndex(store)
//	for tag, jobs := range idx {
//		fmt.Printf("%s: %v\n", tag, jobs)
//	}
package history
