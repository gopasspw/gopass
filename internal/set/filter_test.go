package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	out := Filter(in, 6, 7, 8, 9)

	assert.Equal(t, []int{1, 2, 3, 4, 5}, out)
}
