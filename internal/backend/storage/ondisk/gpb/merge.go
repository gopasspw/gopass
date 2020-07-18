package gpb

import (
	"fmt"
	"sort"

	"github.com/gopasspw/gopass/internal/debug"
)

// Merge merges
func (s *Store) Merge(other *Store) error {
	if s == nil || other == nil {
		return fmt.Errorf("nil store")
	}
	if s.Name != other.Name {
		return fmt.Errorf("name mismatch")
	}
	for k, sv := range s.Entries {
		ov, found := other.Entries[k]
		if !found {
			continue
		}
		s.Entries[k] = sv.Merge(ov)
	}
	for k, ov := range other.Entries {
		sv, found := s.Entries[k]
		if !found {
			s.Entries[k] = ov
			continue
		}
		s.Entries[k] = ov.Merge(sv)
	}
	return nil
}

// Merge merges
func (e *Entry) Merge(other *Entry) *Entry {
	if e.Name != other.Name {
		debug.Log("Name mismatch")
		return e
	}
	sort.Sort(ByRevision(e.Revisions))
	sort.Sort(ByRevision(other.Revisions))
	for i, r := range e.Revisions {
		if i > len(other.Revisions)-1 {
			break
		}
		or := other.Revisions[i]
		if r.Equals(or) {
			continue
		}
		e.Revisions = append(e.Revisions, or)
	}
	sort.Sort(ByRevision(e.Revisions))
	for i, or := range other.Revisions {
		if i > len(e.Revisions)-1 {
			e.Revisions = append(e.Revisions, or)
			continue
		}
		r := e.Revisions[i]
		if r.Equals(or) {
			continue
		}
		e.Revisions = append(e.Revisions, or)
	}
	sort.Sort(ByRevision(e.Revisions))
	return e
}
