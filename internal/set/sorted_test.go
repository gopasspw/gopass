package set

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSorted(t *testing.T) {
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	in := append(want, want...)
	rand.Shuffle(len(in), func(i, j int) {
		in[i], in[j] = in[j], in[i]
	})
	assert.Equal(t, want, Sorted(in))
}

func TestSortedFiltered(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	in = append(in, in...)
	rand.Shuffle(len(in), func(i, j int) {
		in[i], in[j] = in[j], in[i]
	})

	want := []int{2, 4, 6, 8, 10}
	assert.Equal(t, want, SortedFiltered(in, func(i int) bool {
		return i%2 == 0
	}))
}
