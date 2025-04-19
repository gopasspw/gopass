package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapFunc(t *testing.T) {
	t.Parallel()

	assert.Equal(t, map[int]bool{1: true, 2: true, 3: true}, Map([]int{1, 2, 3}))
}

func TestApplyFunc(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []int{2, 3, 4}, Apply([]int{1, 2, 3}, func(i int) int { return i + 1 }))
}
