package tree

// Tree is tree-like object supporting pretty printing
type Tree interface {
	List(int) []string
	ListFolders(int) []string
	Format(int) string
	String() string
	AddFile(string, string) error
	AddMount(string, string) error
	AddTemplate(string) error
	FindFolder(string) (Tree, error)
	SetRoot(bool)
	SetName(string)
	Len() int
}
