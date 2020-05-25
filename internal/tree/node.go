package tree

import "bytes"

// Node is a tree node
type Node struct {
	Name     string
	Type     string
	Template bool
	Mount    bool
	Path     string
	Subtree  *Tree
}

// Nodes is a slice of nodes which can be sorted
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

// Equals compares to another node
func (n Node) Equals(other Node) bool {
	if n.Name != other.Name {
		return false
	}
	if n.Type != other.Type {
		return false
	}
	if n.Subtree != nil {
		if other.Subtree == nil {
			return false
		}
		if !n.Subtree.Equals(other.Subtree) {
			return false
		}
	} else {
		if other.Subtree != nil {
			return false
		}
	}
	return true
}

// format returns a pretty printed string of all nodes in and below
// this node, e.g. ├── baz
func (n *Node) format(prefix string, last bool, maxDepth, curDepth int) string {
	if maxDepth >= 0 && ((maxDepth == 0 && curDepth > 1) || (curDepth > maxDepth+1)) {
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
	case n.Type == "dir":
		_, _ = out.WriteString(colDir(n.Name))
	default:
		_, _ = out.WriteString(n.Name)
	}
	// mark templates
	if n.Template {
		_, _ = out.WriteString(" " + colTpl("(template)"))
	}
	// finish this folders output
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

// Len returns the length of this subtree
func (n *Node) Len() int {
	if n.Type == "file" {
		return 1
	}
	var l int
	for _, t := range n.Subtree.Nodes {
		l += t.Len()
	}
	return l
}

func (n *Node) list(prefix string, maxDepth, curDepth int, files bool) []string {
	if maxDepth > 0 && curDepth > maxDepth {
		return nil
	}

	if prefix != "" {
		prefix += sep
	}
	prefix += n.Name

	if n.Type == "file" && files {
		return []string{prefix}
	}

	out := make([]string, 0, n.Len())
	if n.Subtree == nil {
		return out
	}
	if !files {
		out = append(out, prefix)
	}
	for _, t := range n.Subtree.Nodes {
		out = append(out, t.list(prefix, maxDepth, curDepth+1, files)...)
	}
	return out
}
