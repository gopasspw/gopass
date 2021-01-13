package gjs

import (
	"fmt"
	"sort"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Merge merges two stores
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
	debug.Log("merged store %s", s.Name)
	return nil
}

// Merge merges two entries
func (e *Entry) Merge(other *Entry) *Entry {
	if other == nil {
		return e
	}
	if e == nil {
		return other
	}
	if e.Name != other.Name {
		debug.Log("Name mismatch")
		return e
	}
	debug.Log("merging entry %s, (%d <-> %d revisions)", e.Name, len(e.Revisions), len(other.Revisions))
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
		debug.Log("adding non-equal revision from position %d: %s", i, or)
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
		debug.Log("adding non-equal revision from position %d: %s", i, or)
		e.Revisions = append(e.Revisions, or)
	}
	sort.Sort(ByRevision(e.Revisions))
	debug.Log("merged entry %s, %d revisions", e.Name, len(e.Revisions))
	return e
}
