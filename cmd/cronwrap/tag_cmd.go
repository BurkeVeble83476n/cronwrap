package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/user/cronwrap/internal/history"
)

// runTagIndex prints a tag → job-names index to stdout.
// Called when the user runs: cronwrap tags index
func runTagIndex(storePath string) error {
	s, err := history.NewStore(storePath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}

	idx, err := history.BuildTagIndex(s)
	if err != nil {
		return err
	}

	if len(idx) == 0 {
		fmt.Fprintln(os.Stdout, "no tags found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TAG\tJOBS")
	for tag, jobs := range idx {
		fmt.Fprintf(w, "%s\t%v\n", tag, jobs)
	}
	return w.Flush()
}

// runTagFilter prints records matching a tag as JSON.
// Called when the user runs: cronwrap tags filter <tag>
func runTagFilter(storePath, tag string) error {
	if tag == "" {
		return fmt.Errorf("tag name must not be empty")
	}

	s, err := history.NewStore(storePath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}

	records, err := history.FilterByTag(s, tag)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}
