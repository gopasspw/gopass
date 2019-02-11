package fs

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveEmptyParentDirectories(t *testing.T) {
	var tests = []struct {
		name          string
		storeRoot     []string
		subdirs       []string
		expectPresent []string
		expectGone    []string
		prepare       func(string)
	}{
		{
			name:          "only empty subdirs",
			storeRoot:     []string{"store-root"},
			subdirs:       []string{"a", "b", "c"},
			expectPresent: []string{"store-root"},
			expectGone:    []string{"a", "b", "c"},
		},
		{
			name:          "empty subdirs, nested root",
			storeRoot:     []string{"root-parent", "store-root"},
			subdirs:       []string{"a", "b"},
			expectPresent: []string{"root-parent", "store-root"},
			expectGone:    []string{"a", "b"},
		},
		{
			name:          "file in outer dir",
			storeRoot:     []string{"root-parent", "store-root"},
			subdirs:       []string{"a", "b"},
			expectPresent: []string{"root-parent", "store-root", "a", "b"},
			prepare: func(path string) {
				f, _ := os.Create(filepath.Join(path, "some-file"))
				_ = f.Close()
			},
		},
		{
			name:          "file in parent dir",
			storeRoot:     []string{"store-root"},
			subdirs:       []string{"a", "b"},
			expectPresent: []string{"store-root", "a"},
			expectGone:    []string{"b"},
			prepare: func(path string) {
				f, _ := os.Create(filepath.Join(path, "..", "file-in-parent"))
				_ = f.Close()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			td, err := ioutil.TempDir("", "gopass-")
			if err != nil {
				t.Error(err)
			}
			defer func() {
				_ = os.RemoveAll(td)
			}()

			path := filepath.Join(append([]string{td}, test.storeRoot...)...)
			subdir := filepath.Join(append([]string{path}, test.subdirs...)...)

			if err := os.MkdirAll(subdir, 0777); err != nil {
				t.Error(err)
			}

			if test.prepare != nil {
				test.prepare(subdir)
			}

			s := &Store{
				path,
			}
			if err = s.removeEmptyParentDirectories(filepath.Join(subdir, "deletedFile")); err != nil {
				t.Error(err)
			}

			dir := td
			for _, d := range test.expectPresent {
				dir = filepath.Join(dir, d)
				assert.DirExists(t, dir)
			}
			for _, d := range test.expectGone {
				dir = filepath.Join(dir, d)
				if _, err := os.Stat(dir); err == nil || !os.IsNotExist(err) {
					t.Errorf("Directory %s should not exist", dir)
				}
			}
		})
	}
}
