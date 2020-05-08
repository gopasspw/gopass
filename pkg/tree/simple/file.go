package simple

// File is a leaf node in the tree
type File struct {
	Name     string
	Metadata map[string]string
}

// String implement fmt.Stringer
func (f File) String() string {
	return f.Name
}

// format will format this leaf node for pretty printing
func (f File) format(prefix string, last bool, maxDepth int, curDepth int) string {
	if maxDepth >= 0 && ((maxDepth == 0 && curDepth > 1) || (curDepth > maxDepth+1)) {
		return ""
	}

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
