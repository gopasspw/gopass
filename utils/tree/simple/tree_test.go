package simple

import (
	"sort"
	"strings"
	"testing"

	"path/filepath"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

const (
	goldenSubFormat = `└── ing
    ├── a
    └── b
`
)

func getGoldenFormat(t *testing.T) string {
	mustAbsoluteFilepath := func(s string) string {
		path, err := filepath.Abs(s)
		if err != nil {
			t.Errorf("Error during filepath.Absolute: %s", err)
			return "ERROR"
		}
		return path
	}

	return `gopass
├── a
│   ├── b
│   │   └── c
│   │       ├── d
│   │       └── e
│   ├── g
│   │   └── h
│   └── f
└── foo (` + mustAbsoluteFilepath("/tmp/foo") + `)
    ├── bar (` + mustAbsoluteFilepath("/tmp/bar") + `)
    │   └── baz
    └── baz
        └── inga`
}

func TestFormat(t *testing.T) {
	color.NoColor = true
	root := New("gopass")
	mounts := map[string]string{
		"foo":     "/tmp/foo",
		"foo/bar": "/tmp/bar",
	}
	keys := make([]string, 0, len(mounts))
	for k := range mounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := mounts[k]
		absV, err := filepath.Abs(v)
		if err != nil {
			t.Errorf("Error during filepath.Abs: %s", err)
		}
		assert.NoError(t, root.AddMount(k, absV))
	}
	for _, f := range []string{
		"foo/baz/inga",
		"foo/bar/baz",
		"a/b/c/d",
		"a/b/c/e",
		"a/f",
		"a/g/h",
	} {
		assert.NoError(t, root.AddFile(f, "text/plain"))
	}
	got := strings.TrimSpace(root.Format(0))
	want := strings.TrimSpace(getGoldenFormat(t))
	if want != got {
		t.Errorf("Format mismatch:\n---\n%s\n---\n%s\n---", want, got)
	}
}

func TestFormatSubtree(t *testing.T) {
	root := New("gopass")
	for _, f := range []string{
		"foo/bar",
		"foo/baz",
		"baz/ing/a",
		"baz/ing/b",
	} {
		assert.NoError(t, root.AddFile(f, "text/plain"))
	}

	sub, err := root.FindFolder("baz/ing")
	assert.NoError(t, err)

	got := strings.TrimSpace(sub.Format(0))
	want := strings.TrimSpace(goldenSubFormat)
	assert.Equal(t, want, got)
}

func TestGetNonExistingSubtree(t *testing.T) {
	root := New("gopass")
	for _, f := range []string{
		"foo/bar",
		"foo/baz",
		"baz/ing/a",
		"baz/ing/b",
	} {
		assert.NoError(t, root.AddFile(f, "text/plain"))
	}

	sub, err := root.FindFolder("bla")
	assert.Error(t, err)

	// if it doesn't panic we're good
	_ = sub
}
