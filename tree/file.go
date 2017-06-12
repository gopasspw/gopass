package tree

import (
	"fmt"
	"path/filepath"
)

// File is a leaf node in the tree
type File struct {
	Name     string
	Metadata map[string]string
}

// IsFile always returns true
func (f File) IsFile() bool { return true }

// IsDir always returns false
func (f File) IsDir() bool { return false }

// IsMount always returns false
func (f File) IsMount() bool { return false }

// IsBinary returns true if this is a binary file
func (f File) IsBinary() bool {
	if f.Metadata == nil {
		return false
	}
	if f.Metadata["Content-Type"] == "application/octet-stream" {
		return true
	}
	return false
}

// Add always returns an error
func (f File) Add(Entry) error {
	return fmt.Errorf("%s is a file", f)
}

// String implement fmt.Stringer
func (f File) String() string {
	return f.Name
}

// list returns the full path to this leaf node
func (f File) list(prefix string, _, _ int) []string {
	return []string{filepath.Join(prefix, f.Name)}
}

// format will format this leaf node for pretty printing
func (f File) format(prefix string, last bool, _, _ int) string {
	sym := symBranch
	if last {
		sym = symLeaf
	}
	ft := ""
	if f.Metadata != nil {
		switch f.Metadata["Content-Type"] {
		case "application/octet-stream":
			ft = " " + colBin("(binary)")
		case "text/yaml":
			ft = " " + colYaml("(yaml)")
		}
	}
	return prefix + sym + f.Name + ft + "\n"
}
