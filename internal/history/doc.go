// Package history provides persistent storage, querying, retention management,
// schema migration, export, and statistical aggregation for cronwrap job
// execution records.
//
// # Storage
//
// Records are stored as newline-delimited JSON in a single flat file (one JSON
// object per line). The file path is configurable; parent directories are
// created automatically by NewStore.
//
// # Schema versioning
//
// The first line of the file is a special version-stamp record:
//
//	{"_schema_version":2}
//
// MigrateStore detects the current version and applies any necessary
// migrations before the store is used. A timestamped backup of the original
// file is created before every migration.
//
// # Querying
//
// Query and Last provide filtered, ordered access to stored records without
// loading the entire file into memory beyond the working set.
//
// # Retention
//
// Prune enforces configurable retention policies (maximum age and/or maximum
// record count) on a per-job-name basis.
//
// # Export
//
// ExportJSON and ExportCSV serialise filtered record sets to the supplied
// io.Writer for use in dashboards, auditing, or downstream processing.
//
// # Statistics
//
// Stats computes aggregate metrics (total runs, success/failure counts, mean
// duration, last-run timestamp) across all jobs or for a single named job.
package history
