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
	colTpl   = color.New(color.FgGreen, color.Bold).SprintfFunc()
	colBin   = color.New(color.FgYellow, color.Bold).SprintfFunc()
	colYaml  = color.New(color.FgCyan, color.Bold).SprintfFunc()
)

// New create a new root folder
func New(name string) *Folder {
	f := newFolder(name)
	f.Root = true
	return f
}

func sortedFolders(m map[string]*Folder) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedFiles(m map[string]*File) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
