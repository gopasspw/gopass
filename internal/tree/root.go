package tree

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/debug"
)

const (
	symEmpty  = "    "
	symBranch = "├── "
	symLeaf   = "└── "
	symVert   = "│   "
)

var (
	colMount = color.New(color.FgCyan, color.Bold).SprintfFunc()
	colDir   = color.New(color.FgBlue, color.Bold).SprintfFunc()
	colTpl   = color.New(color.FgGreen, color.Bold).SprintfFunc()
	//colBin   = color.New(color.FgYellow, color.Bold).SprintfFunc()
	//colYaml  = color.New(color.FgCyan, color.Bold).SprintfFunc()
	sep = "/" // this should not be platform-agnostic. this is for the CLI interface
)

// Root is the root of a tree
type Root struct {
	Name    string
	Subtree *Tree
}

// New creates a new tree
func New(name string) *Root {
	return &Root{
		Name:    name,
		Subtree: NewTree(),
	}
}

// AddFile adds a new file to the tree
func (r *Root) AddFile(path string, _ string) error {
	return r.insert(path, false, "")
}

// AddMount adds a new mount point to the tree
func (r *Root) AddMount(path, dest string) error {
	return r.insert(path, false, dest)
}

// AddTemplate adds a template to the tree
func (r *Root) AddTemplate(path string) error {
	return r.insert(path, true, "")
}

func (r *Root) insert(path string, template bool, mountPath string) error {
	t := r.Subtree
	p := strings.Split(path, "/")
	for i, e := range p {
		n := &Node{
			Name:    e,
			Type:    "dir",
			Subtree: NewTree(),
		}
		if i == len(p)-1 {
			n.Type = "file"
			n.Subtree = nil
			n.Template = template
			if mountPath != "" {
				n.Mount = true
				n.Path = mountPath
			}
		}
		node, err := t.Insert(n)
		if err != nil {
			debug.Log("failed to insert: %s", err)
		}
		// do we need to extend an existing subtree?
		if i < len(p)-1 && node.Subtree == nil {
			node.Subtree = NewTree()
			node.Type = "dir"
		}
		t = node.Subtree
	}
	return nil
}

// Format returns a pretty printed string of all nodes in and below
// this node, e.g. ├── baz
func (r *Root) Format(maxDepth int) string {
	out := &bytes.Buffer{}

	// any mount will be colored and include the on-disk path
	_, _ = out.WriteString(colDir(r.Name))

	// finish this folders output
	_, _ = out.WriteString("\n")

	// let our children format themselves
	for i, node := range r.Subtree.Nodes {
		last := i == len(r.Subtree.Nodes)-1
		_, _ = out.WriteString(node.format("", last, maxDepth, 1))
	}
	return out.String()
}

// List returns a flat list of all files in this tree
func (r *Root) List(maxDepth int) []string {
	out := make([]string, 0, r.Len())
	for _, t := range r.Subtree.Nodes {
		out = append(out, t.list("", maxDepth, 0, true)...)
	}
	return out
}

// ListFolders returns a flat list of all folders in this tree
func (r *Root) ListFolders(maxDepth int) []string {
	out := make([]string, 0, r.Len())
	for _, t := range r.Subtree.Nodes {
		out = append(out, t.list("", maxDepth, 0, false)...)
	}
	return out
}

// String returns the name of this tree
func (r *Root) String() string {
	return r.Name
}

// FindFolder returns the subtree rooted at path
func (r *Root) FindFolder(path string) (*Root, error) {
	t := r.Subtree
	p := strings.Split(path, "/")
	for _, e := range p {
		_, node := t.find(e)
		if node == nil || node.Type == "file" || node.Subtree == nil {
			return nil, fmt.Errorf("not found")
		}
		t = node.Subtree
	}
	return &Root{Name: r.Name, Subtree: t}, nil
}

// SetName changes the name of this tree
func (r *Root) SetName(n string) {
	r.Name = n
}

// Len returns the number of entries in this folder and all subfolder including
// this folder itself
func (r *Root) Len() int {
	var l int
	for _, t := range r.Subtree.Nodes {
		l += t.Len()
	}
	return l
}
