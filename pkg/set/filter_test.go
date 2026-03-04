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

func TestFilter_EmptyInput(t *testing.T) {
	t.Parallel()

	in := []int{}
	out := Filter(in, 1, 2, 3)

	assert.Equal(t, []int(nil), out)
}

func TestFilter_NoElementsToRemove(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4, 5}
	out := Filter(in)

	assert.Equal(t, []int{1, 2, 3, 4, 5}, out)
}

func TestFilter_RemoveNonExistentElements(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4, 5}
	out := Filter(in, 6, 7, 8)

	assert.Equal(t, []int{1, 2, 3, 4, 5}, out)
}

func TestContains(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4, 5}

	assert.True(t, Contains(in, 3))
	assert.False(t, Contains(in, 6))
}

func TestContains_EmptyInput(t *testing.T) {
	t.Parallel()

	in := []int{}

	assert.False(t, Contains(in, 1))
}
