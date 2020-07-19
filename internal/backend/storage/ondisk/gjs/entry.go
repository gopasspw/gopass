package gjs

import (
	"sort"
	"time"
)

// GetName returns the name
func (e *Entry) GetName() string {
	return e.Name
}

// GetRevisions returns the slice of revisions
func (e *Entry) GetRevisions() []*Revision {
	if e.Revisions == nil {
		e.Revisions = []*Revision{}
	}
	return e.Revisions
}

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
	return e.Latest().Tombstone
}

// Delete marks an entry as deleted
func (e *Entry) Delete(msg string) bool {
	if e.IsDeleted() {
		return false
	}
	ts := Timestamp(time.Now())
	e.Revisions = append(e.Revisions, &Revision{
		Created:   &ts,
		Message:   msg,
		Tombstone: true,
	})
	return true
}

// Equals returns true if other is identical
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
