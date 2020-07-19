package gjs

import (
	"fmt"
	"time"
)

// Time returns the time a revision was created
func (r *Revision) Time() time.Time {
	return r.GetCreated().Time()
}

// ID returns the unique ID of this entry
// TODO: It's not really unique. Need to fix that.
func (r *Revision) ID() string {
	return fmt.Sprintf("%d", r.GetCreated().Time().UnixNano())
}

// Equals returns true if other is the same
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

// GetCreated returns the creation timestampo
func (r *Revision) GetCreated() *Timestamp {
	if r == nil {
		return nil
	}
	return r.Created
}

// GetMessage returns the commit message
func (r *Revision) GetMessage() string {
	return r.Message
}

// GetFilename returns the blob filename
func (r *Revision) GetFilename() string {
	return r.Filename
}

// GetTombstone returns true if this entry was deleted.
// Important for merging.
func (r *Revision) GetTombstone() bool {
	return r.Tombstone
}
