package gpb

import "sort"

// ListBlobs lists all blobs
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
