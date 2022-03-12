package tree

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestRoot(t *testing.T) {
	t.Parallel()

	color.NoColor = true

	r := New("gopass")
	assert.NoError(t, r.AddTemplate("foo"))
	assert.NoError(t, r.AddFile("foo/bar/baz", ""))
	assert.NoError(t, r.AddFile("foo/bar/zab", ""))
	assert.NoError(t, r.AddMount("mnt/m1", "/tmp/m1"))
	assert.NoError(t, r.AddFile("mnt/m1/foo", ""))
	assert.NoError(t, r.AddFile("mnt/m1/foo/bar", ""))
	t.Logf("%+#v", r)
	assert.Equal(t, `gopass
├── foo/ (template)
│   └── bar/
│       ├── baz
│       └── zab
└── mnt/
    └── m1 (/tmp/m1)
        └── foo/
            └── bar
`, r.Format(INF))

	assert.Equal(t, []string{
		"foo/bar/baz",
		"foo/bar/zab",
		"mnt/m1/foo/bar",
	}, r.List(INF))
	assert.Equal(t, []string{
		"foo/",
		"foo/bar/",
		"mnt/",
		"mnt/m1/",
		"mnt/m1/foo/",
	}, r.ListFolders(INF))

	f, err := r.FindFolder("mnt/m1")
	assert.NoError(t, err)
	assert.Equal(t, `gopass
└── foo/
    └── bar
`, f.Format(INF))
}

func TestMountShadow(t *testing.T) {
	t.Parallel()

	color.NoColor = true

	r := New("gopass")
	assert.NoError(t, r.AddTemplate("foo"))
	assert.NoError(t, r.AddFile("foo/bar/baz", ""))
	assert.NoError(t, r.AddFile("foo/bar/zab", ""))
	assert.NoError(t, r.AddMount("foo", "/tmp/m1"))
	assert.NoError(t, r.AddFile("foo/zab", ""))
	assert.NoError(t, r.AddFile("foo/baz", ""))
	t.Logf("%+#v", r)
	assert.Equal(t, `gopass
└── foo (/tmp/m1)
    ├── baz
    └── zab
`, r.Format(INF))

	assert.Equal(t, []string{
		"foo/baz",
		"foo/zab",
	}, r.List(INF))
	assert.Equal(t, []string{
		"foo/",
	}, r.ListFolders(INF))

	_, err := r.FindFolder("mnt/m1")
	assert.Error(t, err)
}
