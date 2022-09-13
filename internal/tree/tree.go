// Package tree implements a tree for displaying hierarchical
// password store entries. It is loosely based on
// https://github.com/restic/restic/blob/master/internal/restic/tree.go
package tree

import (
	"fmt"
	"sort"

	"github.com/gopasspw/gopass/pkg/debug"
)

// ErrNodePresent is returned when a node with the same name is already present.
var ErrNodePresent = fmt.Errorf("node already present")

// Tree is a tree.
type Tree struct {
	Nodes []*Node
}

// NewTree creates a new tree.
func NewTree() *Tree {
	return &Tree{
		Nodes: []*Node{},
	}
}

// String returns the name of this tree.
func (t *Tree) String() string {
	return fmt.Sprintf("Tree<%d nodes>", len(t.Nodes))
}

// Equals compares to another tree.
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

// Insert adds a new node at the right position.
func (t *Tree) Insert(node *Node) *Node {
	pos, found := t.findPositionFor(node.Name)
	if found != nil {
		debug.Log("merging node (%+v) with existing one (%+v)", found, node)
		m := found.Merge(*node)
		t.Nodes[pos] = m

		return m
	}

	debug.Log("extending subtree for %s", node.Name)
	// insert at the right position, see
	// https://code.google.com/p/go-wiki/wiki/SliceTricks
	t.Nodes = append(t.Nodes, &Node{})
	copy(t.Nodes[pos+1:], t.Nodes[pos:])
	t.Nodes[pos] = node

	return node
}

func (t *Tree) findPositionFor(name string) (int, *Node) {
	pos := sort.Search(len(t.Nodes), func(i int) bool {
		return t.Nodes[i].Name >= name
	})

	if pos < len(t.Nodes) && t.Nodes[pos].Name == name {
		return pos, t.Nodes[pos]
	}

	return pos, nil
}

// Sort ensures this tree is sorted.
func (t *Tree) Sort() {
	list := Nodes(t.Nodes)
	sort.Sort(list)
	t.Nodes = list
}
