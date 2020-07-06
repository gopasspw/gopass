package gpb

import (
	"fmt"
	"time"
)

// Time returns the time a revision was created
func (r *Revision) Time() time.Time {
	return time.Unix(r.Created.GetSeconds(), int64(r.Created.GetNanos()))
}

// ID returns the unique ID of this entry
// TODO: It's not really unique. Need to fix that.
func (r *Revision) ID() string {
	return fmt.Sprintf("%d", r.Created.AsTime().UnixNano())
}

// Equals compares
func (r *Revision) Equals(other *Revision) bool {
	if r == nil || other == nil {
		return false
	}
	if r.Time() != other.Time() {
		return false
	}
	if r.GetMessage() != other.GetMessage() {
		return false
	}
	if r.GetFilename() != other.GetFilename() {
		return false
	}
	if r.GetTombstone() != other.GetTombstone() {
		return false
	}

	return true
}
