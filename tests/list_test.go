package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("list")
	require.NoError(t, err)
	assert.Equal(t, "gopass", out)

	out, err = ts.run("ls")
	require.NoError(t, err)
	assert.Equal(t, "gopass", out)

	ts.initSecrets("")

	list := `
gopass
├── baz
├── fixed/
│   ├── secret
│   └── twoliner
└── foo/
    └── bar
`
	out, err = ts.run("list")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)

	list = `
foo/
└── bar
`
	out, err = ts.run("list foo")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)

	list = `fixed/
foo/
`
	out, err = ts.run("list --folders")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)
}

// regression test for #1628.
func TestListRegressions1628(t *testing.T) {
	t.Parallel()

	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("list")
	require.NoError(t, err)
	assert.Equal(t, "gopass", out)

	_, err = ts.run("insert misc/file1")
	require.NoError(t, err)
	_, err = ts.run("insert misc/folder1/folder2/folder3/file2")
	require.NoError(t, err)

	out, err = ts.run("list")
	require.NoError(t, err)

	exp := `gopass
└── misc/
    ├── file1
    └── folder1/
        └── folder2/
            └── folder3/
                └── file2`
	assert.Equal(t, exp, out)
}
