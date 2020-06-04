package tree

import (
	"testing"

	"gotest.tools/assert"
)

func TestTree(t *testing.T) {
	t1 := NewTree()
	t2 := NewTree()

	assert.Equal(t, true, t1.Equals(t2))

	t1.Insert(&Node{Name: "foo"})
	assert.Equal(t, false, t1.Equals(t2))
	t2.Insert(&Node{Name: "foo"})
	assert.Equal(t, true, t1.Equals(t2))
}
