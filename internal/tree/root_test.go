package tree

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {
	t.Parallel()

	color.NoColor = true

	r := New("gopass")
	require.NoError(t, r.AddTemplate("foo"))
	require.NoError(t, r.AddFile("foo/bar/baz", ""))
	require.NoError(t, r.AddFile("foo/bar/zab", ""))
	require.NoError(t, r.AddMount("mnt/m1", "/tmp/m1"))
	require.NoError(t, r.AddFile("mnt/m1/foo", ""))
	require.NoError(t, r.AddFile("mnt/m1/foo/bar", ""))
	t.Logf("%+#v", r)
	assert.Equal(t, `gopass
├── foo/ (template) (shadowed)
│   └── bar/
│       ├── baz
│       └── zab
└── mnt/
    └── m1 (/tmp/m1)
        └── foo/ (shadowed)
            └── bar
`, r.Format(INF))

	assert.Equal(t, []string{
		"foo",
		"foo/bar/baz",
		"foo/bar/zab",
		"mnt/m1/foo",
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
	require.NoError(t, err)
	assert.Equal(t, `gopass
└── foo/ (shadowed)
    └── bar
`, f.Format(INF))
}

func TestMountShadow(t *testing.T) {
	t.Parallel()

	color.NoColor = true

	r := New("gopass")
	require.NoError(t, r.AddTemplate("foo"))
	require.NoError(t, r.AddFile("foo/bar/baz", ""))
	require.NoError(t, r.AddFile("foo/bar/zab", ""))
	require.NoError(t, r.AddMount("foo", "/tmp/m1"))
	require.NoError(t, r.AddFile("foo/zab", ""))
	require.NoError(t, r.AddFile("foo/baz", ""))
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
	require.Error(t, err)
}
