// Package config provides configuration loading and validation for cronwrap.
//
// Configuration can be supplied via a YAML file. When no file is provided,
// sensible defaults are used so that cronwrap works out of the box with zero
// configuration.
//
// Example YAML:
//
//	history_path: ~/.cronwrap/history.jsonl
//	max_history_records: 1000
//	default_timeout: 5m
//	alert:
//	  on_failure: true
//	  duration_threshold: 10m
//	  log_level: error
//
// Load merges the file over the defaults returned by Defaults, so any
// omitted fields retain their default values.
package config
