package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLongestCommonPrefix(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Src    string
		Dst    string
		Prefix string
	}{
		{
			Src:    "foo/bar/baz/zab.txt",
			Dst:    "foo/baz/foo.txt",
			Prefix: "foo",
		},
	} {
		prefix := longestCommonPrefix(tc.Src, tc.Dst)
		assert.Equal(t, tc.Prefix, prefix)
	}
}

func TestAddRel(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Src string
		Dst string
		Out string
	}{
		{
			Src: "bar/baz.txt",
			Dst: "baz/foo.txt",
			Out: "../bar/baz.txt",
		},
	} {
		assert.Equal(t, tc.Out, addRel(tc.Src, tc.Dst))
	}
}

func TestLinkRejectsPathTraversal(t *testing.T) {
	base := t.TempDir()
	storeRoot := filepath.Join(base, "store")
	outside := filepath.Join(base, "outside")
	require.NoError(t, os.MkdirAll(storeRoot, 0o700))
	require.NoError(t, os.MkdirAll(outside, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(storeRoot, "secret.gpg"), []byte("encrypted"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.gpg"), []byte("outside"), 0o600))

	s := New(storeRoot)
	ctx := context.Background()

	t.Run("destination escapes store", func(t *testing.T) {
		err := s.Link(ctx, "secret.gpg", "../outside/linked.gpg")
		require.Error(t, err, "Link must reject a destination outside the store root")
		_, statErr := os.Lstat(filepath.Join(outside, "linked.gpg"))
		assert.True(t, os.IsNotExist(statErr), "Link must not create anything outside the store root")
	})

	t.Run("source escapes store", func(t *testing.T) {
		err := s.Link(ctx, "../outside/secret.gpg", "linked.gpg")
		require.Error(t, err, "Link must reject a source outside the store root")
		_, statErr := os.Lstat(filepath.Join(storeRoot, "linked.gpg"))
		assert.True(t, os.IsNotExist(statErr), "Link must not create a link to a source outside the store root")
	})
}
