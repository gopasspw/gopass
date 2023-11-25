package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAndGet(t *testing.T) {
	t.Parallel()

	initialContent := []byte(`initial file content`)
	otherContent := []byte(`other file content`)
	ctx := config.NewNoWrites().WithConfig(context.Background())

	path := t.TempDir()

	s := &Store{path}

	fileHasContent := func(filename string, content []byte) {
		written, _ := s.Get(ctx, filename)
		assert.Equalf(t, content, written, "content of file")
	}

	filename := filepath.Join("a", "b", "file")

	// when file does not exist
	fileHasContent(filename, nil)

	// when the folder does not exist
	_ = s.Set(ctx, filename, initialContent)
	fileHasContent(filename, initialContent)

	// overwrite file
	_ = s.Set(ctx, filename, otherContent)
	fileHasContent(filename, otherContent)

	// when folder already exists, with unclean path
	_ = s.Set(ctx, filepath.Join("a", ".", "b", "..", "other"), initialContent)
	fileHasContent(filepath.Join("a", "other"), initialContent)
}

func TestMove(t *testing.T) {
	t.Parallel()

	initialContent := []byte(`initial file content`)
	otherContent := []byte(`other file content`)
	ctx := config.NewNoWrites().WithConfig(context.Background())

	path := t.TempDir()

	s := &Store{path}

	fileHasContent := func(filename string, content []byte) {
		written, _ := s.Get(ctx, filename)
		assert.Equalf(t, content, written, "content of file")
	}

	filename := "src"

	// when file does not exist
	fileHasContent(filename, nil)

	// when the folder does not exist
	_ = s.Set(ctx, filename, initialContent)
	fileHasContent(filename, initialContent)

	// move file
	require.NoError(t, s.Move(ctx, filename, "dst1", true))
	fileHasContent("dst1", initialContent)

	// overwrite file
	_ = s.Set(ctx, "dst2", otherContent)
	fileHasContent("dst2", otherContent)

	// move file
	require.NoError(t, s.Move(ctx, "dst1", "dst2", true))
	fileHasContent("dst1", nil)
	fileHasContent("dst2", initialContent)

	// copy file
	require.NoError(t, s.Move(ctx, "dst2", "dst3", false))
	fileHasContent("dst2", initialContent)
	fileHasContent("dst3", initialContent)
}

func TestRemoveEmptyParentDirectories(t *testing.T) {
	t.Parallel()

	tests := []struct {
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
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			td := t.TempDir()

			path := filepath.Join(append([]string{td}, test.storeRoot...)...)
			subdir := filepath.Join(append([]string{path}, test.subdirs...)...)

			if err := os.MkdirAll(subdir, 0o777); err != nil {
				t.Error(err)
			}

			if test.prepare != nil {
				test.prepare(subdir)
			}

			s := &Store{
				path,
			}
			if err := s.removeEmptyParentDirectories(filepath.Join(subdir, "deletedFile")); err != nil {
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

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		location  []string
		toDelete  []string
		shouldErr bool
	}{
		{
			name:     "simple paths",
			location: []string{"a", "b"},
			toDelete: []string{"a", "b", "secret"},
		},
		{
			name:      "non-existent file",
			toDelete:  []string{"a", "b", "other"},
			location:  []string{"a", "b"},
			shouldErr: true,
		},
		{
			name:      "deletion of non-empty dir not allowed",
			toDelete:  []string{"a"},
			location:  []string{"a"},
			shouldErr: true,
		},
		{
			name:     "unclean path, with parent",
			location: []string{"a"},
			toDelete: []string{"a", "..", "a", "secret"},
		},
		{
			name:     "unclean path, with dots",
			location: []string{"a"},
			toDelete: []string{".", "a", ".", ".", "secret"},
		},
		{
			name:     "unclean path, with dots and parent",
			location: []string{"a", "b"},
			toDelete: []string{".", "a", ".", "b", "..", ".", "b", "secret"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path := t.TempDir()

			subdir := filepath.Join(append([]string{path}, test.location...)...)
			if err := os.MkdirAll(subdir, 0o777); err != nil {
				t.Error(err)
			}

			file := filepath.Join(subdir, "secret")
			if f, err := os.Create(file); err != nil {
				t.Error(err)
			} else {
				_ = f.Close()
			}

			store := &Store{
				path,
			}
			err := store.Delete(config.NewNoWrites().WithConfig(context.Background()), filepath.Join(test.toDelete...))

			if test.shouldErr {
				if err == nil {
					t.Error("Deletion should fail")
				}
			} else {
				if err != nil {
					t.Error("Deletion should not fail")
				}
				if _, err = os.Stat(file); !os.IsNotExist(err) {
					t.Error(err)
				}
			}
		})
	}
}
