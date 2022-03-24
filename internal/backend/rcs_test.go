package backend

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	repo := filepath.Join(td, "repo")
	require.NoError(t, os.MkdirAll(repo, 0o700))

	store := filepath.Join(td, "store")
	require.NoError(t, os.MkdirAll(store, 0o700))

	cmd := exec.Command("git", "init", repo)
	assert.NoError(t, cmd.Run())

	r, err := Clone(ctx, GitFS, repo, store)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestInitRCS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	gitDir := filepath.Join(td, "git")
	assert.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o700))

	r, err := InitStorage(ctx, GitFS, gitDir)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
