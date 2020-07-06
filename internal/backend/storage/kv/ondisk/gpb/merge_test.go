package gpb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMergeEntry(t *testing.T) {
	now := timestamppb.Now()
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
				Created:  timestamppb.New(now.AsTime().Add(time.Second)),
				Message:  "Second commit",
				Filename: "a/b/d.age",
			},
		},
	}
	assert.Equal(t, false, a.Equals(b))
	c := b.Merge(a)
	assert.Equal(t, b, c)
	d := &Entry{
		Name: "foo/bar",
		Revisions: []*Revision{
			{
				Created:  timestamppb.New(now.AsTime().Add(2 * time.Second)),
				Message:  "Other initial commit",
				Filename: "a/b/e.age",
			},
		},
	}
	e := d.Merge(c)
	assert.Equal(t, false, e.Equals(c))
	assert.Equal(t, 3, len(e.Revisions))
}

func TestMergeStore(t *testing.T) {}
