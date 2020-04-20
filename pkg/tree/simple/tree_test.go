package simple

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	goldenSubFormat = `└── ing
    ├── a
    └── b
`
)

func getGoldenFormat(t *testing.T) string {
	mustAbsoluteFilepath := func(s string) string {
		t.Helper()
		path, err := filepath.Abs(s)
		if err != nil {
			assert.NoError(t, err)
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
└── foo (` + mustAbsoluteFilepath(filepath.Join(os.TempDir(), "foo")) + `)
    ├── bar (` + mustAbsoluteFilepath(filepath.Join(os.TempDir(), "bar")) + `)
    │   └── baz
    └── baz
        └── inga`
}

func TestFormat(t *testing.T) {
	color.NoColor = true
	root := New("gopass")
	tmpDir := os.TempDir()
	mounts := map[string]string{
		"foo":                    filepath.Join(tmpDir, "foo"),
		filepath.Join("foo/bar"): filepath.Join(tmpDir, "bar"),
	}
	keys := make([]string, 0, len(mounts))
	for k := range mounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := mounts[k]
		absV, err := filepath.Abs(v)
		assert.NoError(t, err)
		assert.NoError(t, root.AddMount(k, absV))
	}
	for _, f := range []string{
		filepath.Join("foo", "baz", "inga"),
		filepath.Join("foo", "bar", "baz"),
		filepath.Join("a", "b", "c", "d"),
		filepath.Join("a", "b", "c", "e"),
		filepath.Join("a", "f"),
		filepath.Join("a", "g", "h"),
	} {
		assert.NoError(t, root.AddFile(f, "text/plain"))
	}
	got := strings.TrimSpace(root.Format(root.Len()))
	want := strings.TrimSpace(getGoldenFormat(t))
	if want != got {
		t.Errorf("Format mismatch:\n---\n%s\n---\n%s\n---", want, got)
	}
}

func TestFormatSubtree(t *testing.T) {
	root := New("gopass")
	for _, f := range []string{
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("baz", "ing", "a"),
		filepath.Join("baz", "ing", "b"),
	} {
		assert.NoError(t, root.AddFile(f, "text/plain"))
	}

	sub, err := root.FindFolder(filepath.Join("baz", "ing"))
	require.NoError(t, err)

	got := strings.TrimSpace(sub.Format(0))
	want := strings.TrimSpace(goldenSubFormat)
	assert.Equal(t, want, got)
}

func TestGetNonExistingSubtree(t *testing.T) {
	root := New("gopass")
	for _, f := range []string{
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("baz", "ing", "a"),
		filepath.Join("baz", "ing", "b"),
	} {
		assert.NoError(t, root.AddFile(f, "text/plain"))
	}

	sub, err := root.FindFolder("bla")
	assert.Error(t, err)

	// if it doesn't panic we're good
	_ = sub
}
