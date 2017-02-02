package tree

import (
	"sort"
	"strings"
	"testing"

	"github.com/fatih/color"
)

const (
	goldenFormat = `gopass
├── a
│   ├── b
│   │   └── c
│   │       ├── d
│   │       └── e
│   ├── f
│   └── g
│       └── h
└── foo (/tmp/foo)
    ├── bar (/tmp/bar)
    │   └── baz
    └── baz
        └── inga`
	goldenSubFormat = `└── ing
    ├── a
    └── b
`
)

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
		if err := root.AddMount(k, v); err != nil {
			t.Fatalf("failed to add mount: %s", err)
		}
	}
	for _, f := range []string{
		"foo/baz/inga",
		"foo/bar/baz",
		"a/b/c/d",
		"a/b/c/e",
		"a/f",
		"a/g/h",
	} {
		if err := root.AddFile(f); err != nil {
			t.Fatalf("failed to add file: %s", err)
		}
	}
	got := strings.TrimSpace(root.Format())
	want := strings.TrimSpace(goldenFormat)
	if want != got {
		t.Errorf("Format mismatch: %s vs %s", want, got)
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
		if err := root.AddFile(f); err != nil {
			t.Fatalf("failed to add file: %s", err)
		}
	}
	sub := root.FindFolder("baz/ing")
	if sub == nil {
		t.Fatalf("failed to find subtree")
	}
	got := strings.TrimSpace(sub.Format())
	want := strings.TrimSpace(goldenSubFormat)
	if want != got {
		t.Errorf("Format mismatch: %s vs %s", want, got)
	}
}
