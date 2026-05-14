// Package config handles loading and validating cronwrap configuration
// from YAML files and environment variables.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level cronwrap configuration.
type Config struct {
	// HistoryPath is the file path where execution history is stored.
	HistoryPath string `yaml:"history_path"`

	// MaxHistoryRecords is the maximum number of records to retain.
	MaxHistoryRecords int `yaml:"max_history_records"`

	// DefaultTimeout is the default execution timeout for jobs.
	DefaultTimeout time.Duration `yaml:"default_timeout"`

	// Alert contains alerting configuration.
	Alert AlertConfig `yaml:"alert"`
}

// AlertConfig holds alerting-related settings.
type AlertConfig struct {
	// OnFailure enables alerts when a job exits with a non-zero code.
	OnFailure bool `yaml:"on_failure"`

	// DurationThreshold triggers an alert when a job exceeds this duration.
	DurationThreshold time.Duration `yaml:"duration_threshold"`

	// LogLevel sets the minimum log level for alert output (info, warn, error).
	LogLevel string `yaml:"log_level"`
}

// Defaults returns a Config populated with sensible default values.
func Defaults() Config {
	return Config{
		HistoryPath:       ".cronwrap/history.jsonl",
		MaxHistoryRecords: 1000,
		DefaultTimeout:    0,
		Alert: AlertConfig{
			OnFailure: true,
			LogLevel:  "error",
		},
	}
}

// Load reads a YAML config file from path and merges it over defaults.
// If path is empty, only defaults are returned.
func Load(path string) (Config, error) {
	cfg := Defaults()
	if path == "" {
		return cfg, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return cfg, fmt.Errorf("config: %w", err)
	}

	return cfg, nil
}

// validate checks that the config values are acceptable.
func (c Config) validate() error {
	if c.MaxHistoryRecords < 0 {
		return fmt.Errorf("max_history_records must be >= 0")
	}
	validLevels := map[string]bool{"info": true, "warn": true, "error": true, "": true}
	if !validLevels[c.Alert.LogLevel] {
		return fmt.Errorf("alert.log_level must be one of: info, warn, error")
	}
	return nil
}
