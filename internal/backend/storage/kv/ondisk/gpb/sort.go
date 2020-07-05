package gpb

import (
	"fmt"
	"sort"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ByRevision sorts to latest revision to the top, i.e. [0]
type ByRevision []*Revision

func (r ByRevision) Len() int      { return len(r) }
func (r ByRevision) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r ByRevision) Less(i, j int) bool {
	return r[i].Created.AsTime().Before(r[j].Created.AsTime())
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

// Time returns the time a revision was created
func (r *Revision) Time() time.Time {
	return time.Unix(r.Created.GetSeconds(), int64(r.Created.GetNanos()))
}

// ID returns the unique ID of this entry
// TODO: It's not really unique. Need to fix that.
func (r *Revision) ID() string {
	return fmt.Sprintf("%d", r.Created.AsTime().UnixNano())
}
