package tree

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

// Folder is intermediate tree node
type Folder struct {
	Name    string // Name is the displayed name of this folder
	Path    string // Path is only used for mounts, it's the on-disk path
	Root    bool   // Root is used for the root node to remove any prefix
	HasTpl  bool
	Entries map[string]Entry // the sub-entries, prevents having files and folder w/ same name
}

// IsFile always return false
func (f *Folder) IsFile() bool { return false }

// IsDir always returns true
func (f *Folder) IsDir() bool { return true }

// IsMount returns true if the path is non-empty
func (f *Folder) IsMount() bool { return f.Path != "" }

// List returns a flattened list of all sub nodes
func (f Folder) List(maxDepth int) []string {
	return f.list("", maxDepth, 0)
}

// Format returns a pretty printed tree
func (f *Folder) Format(maxDepth int) string {
	return f.format("", true, maxDepth, 0)
}

// String implement fmt.Stringer
func (f *Folder) String() string {
	return f.Name
}

// AddFile adds a new file
func (f *Folder) AddFile(name string, contentType string) error {
	return f.addFile(strings.Split(name, string(filepath.Separator)), contentType)
}

// AddMount adds a new mount
func (f *Folder) AddMount(name, path string) error {
	return f.addMount(strings.Split(name, string(filepath.Separator)), path)
}

func (f *Folder) AddTemplate(name string) error {
	return f.addTemplate(strings.Split(name, string(filepath.Separator)))
}

// newFolder creates a new, initialized folder
func newFolder(name string) *Folder {
	return &Folder{
		Name:    name,
		Path:    "",
		Entries: make(map[string]Entry, 10),
	}
}

// newMount creates a new, initialized folder (with a path, i.e. a mount)
func newMount(name, path string) *Folder {
	f := newFolder(name)
	f.Path = path
	return f
}

// list returns a flattened list of all sub entries with their full path
// in the tree, e.g. foo/bar/baz
func (f *Folder) list(prefix string, maxDepth, curDepth int) []string {
	out := make([]string, 0, 10)
	if maxDepth > 0 && curDepth > maxDepth {
		return out
	}

	if !f.Root {
		if prefix != "" {
			prefix += string(filepath.Separator)
		}
		prefix += f.Name
	}
	for _, key := range sortedKeys(f.Entries) {
		out = append(out, f.Entries[key].list(prefix, maxDepth, curDepth+1)...)
	}
	return out
}

// format returns a pretty printed string of all nodes in and below
// this node, e.g. ├── baz
func (f *Folder) format(prefix string, last bool, maxDepth, curDepth int) string {
	if maxDepth > 0 && curDepth > maxDepth {
		return ""
	}

	out := &bytes.Buffer{}
	// only the root node has no prefix
	if !f.Root {
		// all other nodes inherit their ancestors prefix
		out = bytes.NewBufferString(prefix)
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
	}

	// any mount will be colored and include the on-disk path
	if f.IsMount() {
		_, _ = out.WriteString(colMount(f.Name + " (" + f.Path + ")"))
	} else {
		_, _ = out.WriteString(colDir(f.Name))
	}
	// mark templates
	if f.HasTpl {
		_, _ = out.WriteString(" " + colTpl("(template)"))
	}
	// finish this folders output
	_, _ = out.WriteString("\n")
	// let our children format themselfes
	for i, key := range sortedKeys(f.Entries) {
		last := i == len(f.Entries)-1
		_, _ = out.WriteString(f.Entries[key].format(prefix, last, maxDepth, curDepth+1))
	}
	return out.String()
}

// getFolder returns a direct sub-folder within this folder.
// name MUST NOT include filepath separators. If there is no
// such folder a new one is created with that name.
func (f *Folder) getFolder(name string) *Folder {
	if next, found := f.Entries[name]; found && next.IsDir() {
		return next.(*Folder)
	}
	next := newFolder(name)
	f.Entries[name] = next
	return next
}

// FindFolder returns a sub-tree or nil, if the subtree does not exist
func (f *Folder) FindFolder(name string) *Folder {
	return f.findFolder(strings.Split(strings.TrimSuffix(name, "/"), "/"))
}

// findFolder recursively tries to find the named sub-folder
func (f *Folder) findFolder(path []string) *Folder {
	if len(path) < 1 {
		return f
	}
	name := path[0]
	if next, found := f.Entries[name]; found && next.IsDir() {
		if f, ok := next.(*Folder); ok {
			return f.findFolder(path[1:])
		}
	}
	return nil
}

// addFile adds new file
func (f *Folder) addFile(path []string, contentType string) error {
	if len(path) < 1 {
		return fmt.Errorf("Path must not be empty")
	}
	name := path[0]
	if len(path) == 1 {
		if _, found := f.Entries[name]; found {
			return fmt.Errorf("File %s exists", name)
		}
		f.Entries[name] = File{
			Name: name,
			Metadata: map[string]string{
				"Content-Type": contentType,
			},
		}
		return nil
	}
	next := f.getFolder(name)
	return next.addFile(path[1:], contentType)
}

// addMount adds a new mount (folder with non-empty on-disk path)
func (f *Folder) addMount(path []string, dest string) error {
	if len(path) < 1 {
		return fmt.Errorf("Path must not be empty")
	}
	name := path[0]
	if len(path) == 1 {
		if e, found := f.Entries[name]; found {
			if e.IsFile() {
				return fmt.Errorf("File %s exists", name)
			}
		}
		f.Entries[name] = newMount(name, dest)
		return nil
	}
	next := f.getFolder(name)
	return next.addMount(path[1:], dest)
}

func (f *Folder) addTemplate(path []string) error {
	if len(path) < 1 {
		return fmt.Errorf("Path must not be empty")
	}
	name := path[0]
	if len(path) == 1 {
		if e, found := f.Entries[name]; found {
			if sf, ok := e.(*Folder); ok {
				sf.HasTpl = true
			}
		}
		return nil
	}
	return f.getFolder(name).addTemplate(path[1:])
}
