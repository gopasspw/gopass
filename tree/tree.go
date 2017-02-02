package tree

import (
	"sort"

	"github.com/fatih/color"
)

const (
	symEmpty  = "    "
	symBranch = "├── "
	symLeaf   = "└── "
	symVert   = "│   "
)

var (
	colMount = color.New(color.FgRed, color.Bold).SprintfFunc()
	colDir   = color.New(color.FgBlue, color.Bold).SprintfFunc()
)

// Entry is any kind of tree node
type Entry interface {
	format(string, bool) string
	list(string) []string
	IsFile() bool
	IsDir() bool
	IsMount() bool
}

// New create a new root folder
func New(name string) *Folder {
	f := newFolder(name)
	f.Root = true
	return f
}

func sortedKeys(m map[string]Entry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
