package gjs

import "sort"

// ListBlobs returns a slice of all contained blogs. For remote sync.
func (s *Store) ListBlobs() []string {
	out := make([]string, 0, len(s.Entries)*10)
	for _, e := range s.Entries {
		for _, r := range e.Revisions {
			out = append(out, r.Filename)
		}
	}
	sort.Strings(out)
	return out
}

// GetName returns the name
func (s *Store) GetName() string {
	return s.Name
}

// GetEntries returns the map of entries
func (s *Store) GetEntries() map[string]*Entry {
	if s.Entries == nil {
		s.Entries = map[string]*Entry{}
	}
	return s.Entries
}
