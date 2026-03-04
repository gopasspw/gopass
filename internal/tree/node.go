package tree

import (
	"bytes"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Node is a tree node.
type Node struct {
	Name     string
	Leaf     bool
	Template bool
	Mount    bool
	Path     string
	Subtree  *Tree
}

const (
	// INF allows to have a full recursion until the leaves of a tree.
	INF = -1
)

// Nodes is a slice of nodes which can be sorted.
type Nodes []*Node

func (n Nodes) Len() int {
	return len(n)
}

func (n Nodes) Less(i, j int) bool {
	return n[i].Name < n[j].Name
}

func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// Equals compares to another node.
func (n Node) Equals(other Node) bool {
	if n.Name != other.Name {
		return false
	}

	if n.Leaf != other.Leaf {
		return false
	}

	if n.Subtree != nil {
		if other.Subtree == nil {
			return false
		}

		if !n.Subtree.Equals(other.Subtree) {
			return false
		}
	} else if other.Subtree != nil {
		return false
	}

	return true
}

// Merge will merge two nodes into a new node. Does not change either of the two
// input nodes. The merged node will be in the returned node. Implements semantics
// specific to gopass' tree model, i.e. mounts shadow (erase) everything below
// a mount point, nodes within a tree can be leafs (i.e. contain a secret as well
// as subdirectories) and any node can also contain a template.
func (n Node) Merge(other Node) *Node {
	r := Node{
		Name:     n.Name,
		Leaf:     n.Leaf,
		Template: n.Template,
		Mount:    n.Mount,
		Path:     n.Path,
		Subtree:  n.Subtree,
	}

	// During a merge we can't change the name.

	// If either of the nodes is a leaf (i.e. contains a secret) the
	// merged node will be a leaf.
	if other.Leaf {
		r.Leaf = true
	}

	// If either node has a template the merged has a template, too.
	if other.Template {
		r.Template = true
	}

	// Handling of mounts is a bit more tricky. See the comment above.
	// If we're adding a mount to the tree this shadows (erases) everything
	// that was on this branch before a replaces it with the mount.
	// Think of Unix mount semantics here.
	if other.Mount {
		r.Mount = true
		// anything at the mount point, including a secret at the root
		// of the mount point will become inaccessible.
		r.Leaf = false
		r.Path = other.Path
		// existing templates will become invisible
		r.Template = false
		// the subtree from the mount overlays (shadows) the original tree
		r.Subtree = other.Subtree
	}
	// Merging can't change the path (except a mount, see above)
	// If the other node has a subtree we use that, otherwise
	// this method shouldn't have been called in the first place.
	if r.Subtree == nil && other.Subtree != nil {
		r.Subtree = other.Subtree
	}

	debug.V(4).Log("merged %+v and %+v into %+v", n, other, r)

	return &r
}

// format returns a pretty printed string of all nodes in and below
// this node, e.g. `├── baz`.
func (n *Node) format(prefix string, last bool, maxDepth, curDepth int) string {
	if maxDepth > INF && (curDepth > maxDepth+1) {
		return ""
	}

	out := bytes.NewBufferString(prefix)
	// adding either an L or a T, depending if this is the last node
	// or not
	if last {
		_, _ = out.WriteString(symLeaf)
	} else {
		_, _ = out.WriteString(symBranch)
	}
	// the next levels prefix needs to be extended depending if
	// this is the last node in a group or not
	if last {
		prefix += symEmpty
	} else {
		prefix += symVert
	}

	// any mount will be colored and include the on-disk path
	switch {
	case n.Mount:
		_, _ = out.WriteString(colMount(n.Name + " (" + n.Path + ")"))
	case n.Subtree != nil:
		_, _ = out.WriteString(colDir(n.Name + sep))
	default:
		_, _ = out.WriteString(n.Name)
	}
	// mark templates
	if n.Template {
		_, _ = out.WriteString(" " + colTpl("(template)"))
	}
	// mark shadowed entries
	if n.Leaf && n.Subtree != nil && !n.Mount {
		_, _ = out.WriteString(" " + colShadow("(shadowed)"))
	}
	// finish this output
	_, _ = out.WriteString("\n")

	if n.Subtree == nil {
		return out.String()
	}

	// let our children format themselves
	for i, node := range n.Subtree.Nodes {
		last := i == len(n.Subtree.Nodes)-1
		_, _ = out.WriteString(node.format(prefix, last, maxDepth, curDepth+1))
	}

	return out.String()
}

// Len returns the length of this subtree.
func (n *Node) Len() int {
	if n.Subtree == nil {
		return 1
	}

	var l int

	// this node might point to a secret itself so we must account for that
	if n.Leaf {
		l++
	}

	// and for any secret it's subtree might contain
	for _, t := range n.Subtree.Nodes {
		l += t.Len()
	}

	return l
}

func (n *Node) list(prefix string, maxDepth, curDepth int, files bool) []string {
	if maxDepth >= 0 && curDepth > maxDepth {
		return nil
	}

	if prefix != "" {
		prefix += sep
	}

	prefix += n.Name

	out := make([]string, 0, n.Len())
	// if it's a file and we are looking for files
	if n.Leaf && !n.Mount && files {
		// we return the file
		out = append(out, prefix)
	} else if curDepth == maxDepth && n.Subtree != nil {
		// otherwise if we are "at the bottom" and it's not a file
		// we return the directory name with a separator at the end
		return []string{prefix + sep}
	}

	// if we don't have subitems, then it's a leaf and we return
	// (notice that this is what ends the recursion when maxDepth is set to -1)
	if n.Subtree == nil {
		return out
	}

	// this is the part that will list the subdirectories on their own line when using the -d option
	if !files {
		out = append(out, prefix+sep)
	}

	// we keep listing the subtree nodes if we haven't exited yet.
	for _, t := range n.Subtree.Nodes {
		out = append(out, t.list(prefix, maxDepth, curDepth+1, files)...)
	}

	return out
}
