package tree

import (
	"fmt"
	"path/filepath"
)

// File is a leaf node in the tree
type File string

// IsFile always returns true
func (f File) IsFile() bool { return true }

// IsDir always returns false
func (f File) IsDir() bool { return false }

// IsMount always returns false
func (f File) IsMount() bool { return false }

// Add always returns an error
func (f File) Add(Entry) error {
	return fmt.Errorf("%s is a file", f)
}

// String implement fmt.Stringer
func (f File) String() string {
	return string(f)
}

// list returns the full path to this leaf node
func (f File) list(prefix string) []string {
	return []string{filepath.Join(prefix, string(f))}
}

// format will format this leaf node for pretty printing
func (f File) format(prefix string, last bool) string {
	sym := symBranch
	if last {
		sym = symLeaf
	}
	return prefix + sym + string(f) + "\n"
}
