package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStat(t *testing.T) {
	for _, tc := range []struct {
		old     []string
		new     []string
		added   int
		removed int
	}{
		{
			old:   []string{"foo", "bar"},
			new:   []string{"foo", "bar", "baz"},
			added: 1,
		},
		{
			old:     []string{"foo", "bar", "baz"},
			new:     []string{"foo", "bar"},
			removed: 1,
		},
		{
			old:     []string{"foo", "baz"},
			new:     []string{"foo", "bar"},
			added:   1,
			removed: 1,
		},
	} {
		a, r := Stat(tc.old, tc.new)
		assert.Equal(t, tc.added, a)
		assert.Equal(t, tc.removed, r)
	}
}

func TestList(t *testing.T) {
	for _, tc := range []struct {
		old     []string
		new     []string
		added   []string
		removed []string
	}{
		{
			old:     []string{"foo", "bar"},
			new:     []string{"foo", "bar", "baz"},
			added:   []string{"baz"},
			removed: nil,
		},
		{
			old:     []string{"foo", "bar", "baz"},
			new:     []string{"foo", "bar"},
			added:   nil,
			removed: []string{"baz"},
		},
		{
			old:     []string{"foo", "baz"},
			new:     []string{"foo", "bar"},
			added:   []string{"bar"},
			removed: []string{"baz"},
		},
	} {
		a, r := List(tc.old, tc.new)
		assert.Equal(t, tc.added, a)
		assert.Equal(t, tc.removed, r)
	}
}

func TestListToMap(t *testing.T) {
	for _, tc := range []struct {
		l []string
		m map[string]struct{}
	}{
		{
			l: []string{"foo", "bar"},
			m: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
		},
		{
			l: []string{"foo", "bar", "baz", "baz"},
			m: map[string]struct{}{
				"foo": {},
				"bar": {},
				"baz": {},
			},
		},
	} {
		m := listToMap(tc.l)
		assert.Equal(t, tc.m, m)
	}
}
