package gjs

import (
	"fmt"
	"strconv"
	"time"
)

// Timestamp is time.Time with JSON encoding
type Timestamp time.Time

// New returns a new timestamp
func New(t time.Time) *Timestamp {
	ts := Timestamp(t)
	return &ts
}

// Now returns time.Now() as a Timestamp
func Now() *Timestamp {
	ts := Timestamp(time.Now().UTC())
	return &ts
}

// MarshalJSON implements the json marshaler
func (t *Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprint(time.Time(*t).Unix())), nil
}

// UnmarshalJSON implements the json unmarshaler
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	*t = Timestamp(time.Unix(int64(ts), 0))
	return nil
}

// Time returns the underlying time.Time
func (t *Timestamp) Time() time.Time {
	return time.Time(*t)
}

// Revision is a single revision of a secret
type Revision struct {
	Created   *Timestamp `json:"created"`
	Message   string     `json:"message"`
	Filename  string     `json:"filename"`
	Tombstone bool       `json:"tombstone"`
}

// Entry is a key-value entry with a number of revisions
type Entry struct {
	Name      string      `json:"name"`
	Revisions []*Revision `json:"revisions"`
}

// Store is secrets store
type Store struct {
	Name    string            `json:"name"`
	Entries map[string]*Entry `json:"entries"`
}
