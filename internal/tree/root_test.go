package tree

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestRoot(t *testing.T) {
	color.NoColor = true

	r := New("gopass")
	r.AddTemplate("foo")
	r.AddFile("foo/bar/baz", "")
	r.AddFile("foo/bar/zab", "")
	r.AddMount("mnt/m1", "/tmp/m1")
	r.AddFile("mnt/m1/foo", "")
	r.AddFile("mnt/m1/foo/bar", "")
	t.Logf("%+#v", r)
	assert.Equal(t, `gopass
├── foo (template)
│   └── bar
│       ├── baz
│       └── zab
└── mnt
    └── m1 (/tmp/m1)
        └── foo
            └── bar
`, r.Format(-1))

	assert.Equal(t, []string{
		"foo/bar/baz",
		"foo/bar/zab",
		"mnt/m1/foo/bar",
	}, r.List(-1))
	assert.Equal(t, []string{
		"foo",
		"foo/bar",
		"mnt",
		"mnt/m1",
		"mnt/m1/foo",
	}, r.ListFolders(-1))
	f, err := r.FindFolder("mnt/m1")
	assert.NoError(t, err)
	assert.Equal(t, `gopass
└── foo
    └── bar
`, f.Format(-1))
}
