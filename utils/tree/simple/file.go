package simple

// File is a leaf node in the tree
type File struct {
	Name     string
	Metadata map[string]string
}

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

// String implement fmt.Stringer
func (f File) String() string {
	return f.Name
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
