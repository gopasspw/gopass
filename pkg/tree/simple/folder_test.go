package simple

import (
	"github.com/fatih/color"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFolder(t *testing.T) {
	root := New("gopass")
	assert.NoError(t, root.AddFile("foo/bar", "text/plain"))
	assert.NoError(t, root.AddTemplate("foo"))
	assert.NoError(t, root.AddFile("foo/baz.b64", "application/octet-stream"))
	assert.NoError(t, root.AddFile("foo/zab.yml", "text/yaml"))
	assert.Equal(t, 3, root.Len())

	// test list
	lst := root.List(0)
	sort.Strings(lst)
	wants := []string{
		"foo/bar",
		"foo/baz.b64",
		"foo/zab.yml",
	}
	assert.Equal(t, wants, lst)

	// test name
	assert.Equal(t, "gopass", root.String())

	// test format
	color.NoColor = true
	out := root.Format(1)
	want := `gopass
└── foo (template)
    ├── bar
    ├── baz.b64 (binary)
    └── zab.yml (yaml)
`
	assert.Equal(t, want, out)

	// test list 1
	root = New("gopass")
	assert.NoError(t, root.AddFile("zab/foozen", "text/plain"))
	assert.NoError(t, root.AddFile("zab/foo/bar", "text/plain"))
	assert.NoError(t, root.AddFile("zab2/foo/baz", "text/plain"))
	assert.NoError(t, root.AddFile("zab2/foo/zen/baz", "text/plain"))

	lst = root.List(1)
	sort.Strings(lst)
	wants = []string{
		"zab/foozen",
	}
	assert.Equal(t, wants, lst)

	// test folders
	lst = root.ListFolders(0)
	wants = []string{
		"zab",
		"zab/foo",
		"zab2",
		"zab2/foo",
		"zab2/foo/zen",
	}
	assert.Equal(t, wants, lst)

	// test folders maxDepth = 1
	lst = root.ListFolders(1)
	wants = []string{
		"zab",
		"zab/foo",
		"zab2",
		"zab2/foo",
	}
	assert.Equal(t, wants, lst)

	out = root.Format(0)
	want = `gopass
├── zab
└── zab2
`
	assert.Equal(t, want, out)

	out = root.Format(1)
	want = `gopass
├── zab
│   ├── foo
│   └── foozen
└── zab2
    └── foo
`
	assert.Equal(t, want, out)

	out = root.Format(2)
	want = `gopass
├── zab
│   ├── foo
│   │   └── bar
│   └── foozen
└── zab2
    └── foo
        ├── zen
        └── baz
`
	assert.Equal(t, want, out)
}
