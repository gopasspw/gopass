package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeEquals(t *testing.T) {
	n1 := Node{Name: "node1", Leaf: true}
	n2 := Node{Name: "node1", Leaf: true}
	n3 := Node{Name: "node2", Leaf: false}

	assert.True(t, n1.Equals(n2))
	assert.False(t, n1.Equals(n3))
}

func TestNodeMerge(t *testing.T) {
	n1 := Node{Name: "node1", Leaf: true, Template: true}
	n2 := Node{Name: "node1", Leaf: false, Template: false, Mount: true, Path: "/mnt"}

	merged := n1.Merge(n2)

	assert.Equal(t, "node1", merged.Name)
	assert.False(t, merged.Leaf)
	assert.False(t, merged.Template)
	assert.True(t, merged.Mount)
	assert.Equal(t, "/mnt", merged.Path)
}

func TestNodeFormat(t *testing.T) {
	n := Node{Name: "node1", Leaf: true}
	expected := "└── node1\n"
	result := n.format("", true, INF, 0)

	assert.Equal(t, expected, result)
}

func TestNodeLen(t *testing.T) {
	n := Node{Name: "node1", Leaf: true}
	assert.Equal(t, 1, n.Len())

	subtree := &Tree{Nodes: []*Node{{Name: "child1", Leaf: true}, {Name: "child2", Leaf: true}}}
	n.Subtree = subtree
	assert.Equal(t, 3, n.Len())
}

func TestNodeList(t *testing.T) {
	n := Node{Name: "node1", Leaf: true}
	expected := []string{"node1"}
	result := n.list("", INF, 0, true)

	assert.Equal(t, expected, result)

	subtree := &Tree{Nodes: []*Node{{Name: "child1", Leaf: true}, {Name: "child2", Leaf: true}}}
	n.Subtree = subtree
	expected = []string{"node1", "node1/child1", "node1/child2"}
	result = n.list("", INF, 0, true)

	assert.Equal(t, expected, result)
}
