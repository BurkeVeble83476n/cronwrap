// Package history provides execution history storage and querying for cronwrap.
package history

import "fmt"

// Annotation holds a key-value note attached to a history record.
type Annotation struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Annotate adds or updates an annotation on the record identified by id
// within the store at path. Returns an error if the record is not found.
func Annotate(path, id, key, value string) error {
	if key == "" {
		return fmt.Errorf("annotation key must not be empty")
	}
	store, err := NewStore(path)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	records, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("read records: %w", err)
	}
	found := false
	for i := range records {
		if records[i].ID != id {
			continue
		}
		found = true
		if records[i].Meta == nil {
			records[i].Meta = make(map[string]string)
		}
		records[i].Meta["annotation:"+key] = value
		break
	}
	if !found {
		return fmt.Errorf("record %q not found", id)
	}
	return store.ReplaceAll(records)
}

// GetAnnotations returns all annotations stored on the record with the given id.
func GetAnnotations(path, id string) ([]Annotation, error) {
	store, err := NewStore(path)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	records, err := store.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read records: %w", err)
	}
	const prefix = "annotation:"
	for _, r := range records {
		if r.ID != id {
			continue
		}
		var out []Annotation
		for k, v := range r.Meta {
			if len(k) > len(prefix) && k[:len(prefix)] == prefix {
				out = append(out, Annotation{Key: k[len(prefix):], Value: v})
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("record %q not found", id)
}
