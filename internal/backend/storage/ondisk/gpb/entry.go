package gpb

import (
	"sort"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// SortedRevisions returns a list of sorted revisions
func (e *Entry) SortedRevisions() []*Revision {
	sort.Sort(ByRevision(e.Revisions))
	return e.Revisions
}

// Latest returns the latest revision
func (e *Entry) Latest() *Revision {
	sort.Sort(ByRevision(e.Revisions))
	return e.Revisions[len(e.Revisions)-1]
}

// IsDeleted returns true is an entry was marked as deleted
func (e *Entry) IsDeleted() bool {
	return e.Latest().GetTombstone()
}

// Delete marks an entry as deleted
func (e *Entry) Delete(msg string) bool {
	if e.IsDeleted() {
		return false
	}
	e.Revisions = append(e.Revisions, &Revision{
		Created:   timestamppb.Now(),
		Message:   msg,
		Tombstone: true,
	})
	return true
}

// Equals compares two entries
func (e *Entry) Equals(other *Entry) bool {
	if e == nil || other == nil {
		return false
	}
	if e.Name != other.Name {
		return false
	}
	if len(e.Revisions) != len(other.Revisions) {
		return false
	}
	for i, r := range e.Revisions {
		or := other.Revisions[i]
		if !r.Equals(or) {
			return false
		}
	}
	return true
}
