package tree

// Tree is tree-like object supporting pretty printing
type Tree interface {
	List(int) []string
	Format(int) string
	String() string
	AddFile(string, string) error
	AddMount(string, string) error
	AddTemplate(string) error
	FindFolder(string) Tree
	SetRoot(bool)
	SetName(string)
}
