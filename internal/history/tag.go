package history

import (
	"fmt"
	"sort"
	"strings"
)

// TagIndex maps tag names to the set of job names that carry that tag.
type TagIndex map[string][]string

// TagsForRecord returns the tags stored in a Record's metadata field.
// Tags are expected to be stored as a comma-separated value under the
// key "tags" in Record.Meta.
func TagsForRecord(r Record) []string {
	if r.Meta == nil {
		return nil
	}
	raw, ok := r.Meta["tags"]
	if !ok || raw == "" {
		return nil
	}
	tags := strings.Split(raw, ",")
	for i, t := range tags {
		tags[i] = strings.TrimSpace(t)
	}
	return tags
}

// BuildTagIndex scans all records in the store and returns a TagIndex
// mapping each tag to the sorted, deduplicated list of job names that
// have at least one record carrying that tag.
func BuildTagIndex(s *Store) (TagIndex, error) {
	records, err := s.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("tag: read store: %w", err)
	}

	idx := make(TagIndex)
	seen := make(map[string]map[string]struct{})

	for _, r := range records {
		for _, tag := range TagsForRecord(r) {
			if _, ok := seen[tag]; !ok {
				seen[tag] = make(map[string]struct{})
			}
			seen[tag][r.JobName] = struct{}{}
		}
	}

	for tag, jobs := range seen {
		list := make([]string, 0, len(jobs))
		for j := range jobs {
			list = append(list, j)
		}
		sort.Strings(list)
		idx[tag] = list
	}

	return idx, nil
}

// FilterByTag returns all records from the store that carry the given tag.
func FilterByTag(s *Store, tag string) ([]Record, error) {
	records, err := s.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("tag: read store: %w", err)
	}

	var out []Record
	for _, r := range records {
		for _, t := range TagsForRecord(r) {
			if t == tag {
				out = append(out, r)
				break
			}
		}
	}
	return out, nil
}
