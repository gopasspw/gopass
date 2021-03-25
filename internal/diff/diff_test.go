package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
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
		a, r := List(tc.old, tc.new)
		assert.Equal(t, tc.added, a)
		assert.Equal(t, tc.removed, r)
	}
}
