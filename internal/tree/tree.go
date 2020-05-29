// Package tree implements a tree for displaying hierarchical
// password store entries. It is loosely based on
// https://github.com/restic/restic/blob/master/internal/restic/tree.go
package tree

import (
	"fmt"
	"sort"
)

// Tree is a tree
type Tree struct {
	Nodes []*Node
}

// NewTree creates a new tree
func NewTree() *Tree {
	return &Tree{
		Nodes: []*Node{},
	}
}

// String returns the name of this tree
func (t *Tree) String() string {
	return fmt.Sprintf("Tree<%d nodes>", len(t.Nodes))
}

// Equals compares to another tree
func (t *Tree) Equals(other *Tree) bool {
	if len(t.Nodes) != len(other.Nodes) {
		return false
	}

	for i, node := range t.Nodes {
		if !node.Equals(*other.Nodes[i]) {
			return false
		}
	}

	return true
}

// Insert adds a new node at the right position
func (t *Tree) Insert(node *Node) (*Node, error) {
	pos, found := t.find(node.Name)
	if found != nil {
		return t.Nodes[pos], fmt.Errorf("node %q already preset", node.Name)
	}

	// https://code.google.com/p/go-wiki/wiki/SliceTricks
	t.Nodes = append(t.Nodes, &Node{})
	copy(t.Nodes[pos+1:], t.Nodes[pos:])
	t.Nodes[pos] = node

	return node, nil
}

func (t *Tree) find(name string) (int, *Node) {
	pos := sort.Search(len(t.Nodes), func(i int) bool {
		return t.Nodes[i].Name >= name
	})

	if pos < len(t.Nodes) && t.Nodes[pos].Name == name {
		return pos, t.Nodes[pos]
	}

	return pos, nil
}

// Sort ensures this tree is sorted
func (t *Tree) Sort() {
	list := Nodes(t.Nodes)
	sort.Sort(list)
	t.Nodes = list
}
