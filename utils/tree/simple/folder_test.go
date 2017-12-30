package simple

import (
	"sort"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
)

func TestFolder(t *testing.T) {
	root := New("gopass")
	if err := root.AddFile("foo/bar", "text/plain"); err != nil {
		t.Fatalf("Failed to add file: %s", err)
	}
	if err := root.AddTemplate("foo"); err != nil {
		t.Fatalf("Failed to add template: %s", err)
	}
	if err := root.AddFile("foo/baz.b64", "application/octet-stream"); err != nil {
		t.Fatalf("Failed to add file: %s", err)
	}
	if err := root.AddFile("foo/zab.yml", "text/yaml"); err != nil {
		t.Fatalf("Failed to add file: %s", err)
	}
	if root.Len() != 3 {
		t.Errorf("Should have 3 entries not %d", root.Len())
	}
	// test list
	lst := root.List(0)
	sort.Strings(lst)
	wants := []string{
		"foo/bar",
		"foo/baz.b64",
		"foo/zab.yml",
	}
	if !cmp.Equal(lst, wants) {
		t.Errorf("'%v' != '%v'", lst, wants)
	}
	// test name
	if root.String() != "gopass" {
		t.Errorf("Wrong name: %s", root.String())
	}
	// test format
	color.NoColor = true
	out := root.Format(2)
	want := `gopass
└── foo (template)
    ├── bar
    ├── baz.b64 (binary)
    └── zab.yml (yaml)
`
	if out != want {
		t.Errorf("\n%s\n != \n%s\n", out, want)
	}
}
