package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	t.Parallel()

	t1 := NewTree()
	t2 := NewTree()

	assert.True(t, t1.Equals(t2))

	_ = t1.Insert(&Node{Name: "foo"})
	assert.False(t, t1.Equals(t2))

	_ = t2.Insert(&Node{Name: "foo"})
	assert.True(t, t1.Equals(t2))
}

func TestTreeInsert(t *testing.T) {
	t.Parallel()

	tree := NewTree()
	node := &Node{Name: "foo"}

	insertedNode := tree.Insert(node)
	assert.Equal(t, node, insertedNode)
	assert.Len(t, tree.Nodes, 1)
	assert.Equal(t, "foo", tree.Nodes[0].Name)
}

func TestTreeString(t *testing.T) {
	t.Parallel()

	tree := NewTree()
	assert.Equal(t, "Tree<0 nodes>", tree.String())

	tree.Insert(&Node{Name: "foo"})
	assert.Equal(t, "Tree<1 nodes>", tree.String())
}

func TestTreeSort(t *testing.T) {
	t.Parallel()

	tree := NewTree()
	tree.Insert(&Node{Name: "foo"})
	tree.Insert(&Node{Name: "bar"})
	tree.Insert(&Node{Name: "baz"})

	tree.Sort()

	assert.Equal(t, "bar", tree.Nodes[0].Name)
	assert.Equal(t, "baz", tree.Nodes[1].Name)
	assert.Equal(t, "foo", tree.Nodes[2].Name)
}

func TestTreeFindPositionFor(t *testing.T) {
	t.Parallel()

	tree := NewTree()
	tree.Insert(&Node{Name: "foo"})
	tree.Insert(&Node{Name: "bar"})

	pos, node := tree.findPositionFor("foo")
	assert.Equal(t, 1, pos)
	assert.NotNil(t, node)
	assert.Equal(t, "foo", node.Name)

	pos, node = tree.findPositionFor("bar")
	assert.Equal(t, 0, pos)
	assert.NotNil(t, node)

	// does not exist
	pos, node = tree.findPositionFor("baz")
	assert.Equal(t, 1, pos)
	assert.Nil(t, node)
}
