package gjs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMergeEntry(t *testing.T) {
	now := Now()
	a := &Entry{
		Name: "foo/bar",
		Revisions: []*Revision{
			{
				Created:  now,
				Message:  "Initial commit",
				Filename: "a/b/c.age",
			},
		},
	}
	b := &Entry{
		Name: "foo/bar",
		Revisions: []*Revision{
			{
				Created:  now,
				Message:  "Initial commit",
				Filename: "a/b/c.age",
			},
			{
				Created:  New(now.Time().Add(time.Second)),
				Message:  "Second commit",
				Filename: "a/b/d.age",
			},
		},
	}
	assert.False(t, a.Equals(b))
	c := b.Merge(a)
	assert.Equal(t, b, c)
	d := &Entry{
		Name: "foo/bar",
		Revisions: []*Revision{
			{
				Created:  New(now.Time().Add(2 * time.Second)),
				Message:  "Other initial commit",
				Filename: "a/b/e.age",
			},
		},
	}
	e := d.Merge(c)
	assert.False(t, e.Equals(c))
	assert.Equal(t, 3, len(e.Revisions))
}

func TestMergeStore(t *testing.T) {}
