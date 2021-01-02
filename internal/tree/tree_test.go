package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	t1 := NewTree()
	t2 := NewTree()

	assert.True(t, t1.Equals(t2))

	t1.Insert(&Node{Name: "foo"})
	assert.False(t, t1.Equals(t2))
	t2.Insert(&Node{Name: "foo"})
	assert.True(t, t1.Equals(t2))
}
