//go:build fossil

package fossilfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	dir := t.TempDir()

	marker := filepath.Join(dir, CheckoutMarker)
	_, err := os.Create(marker)
	require.NoError(t, err)

	f, err := New(dir)
	require.NoError(t, err)
	assert.NotNil(t, f)
}

func TestClone(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	repo := "https://example.com/repo.fossil"

	f, err := Clone(ctx, repo, dir)
	require.NoError(t, err)
	assert.NotNil(t, f)
}

func TestInit(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()

	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)
	assert.NotNil(t, f)
}

func TestAdd(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	err = f.Add(ctx, "testfile")
	require.NoError(t, err)
}

func TestCommit(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	err = f.Commit(ctx, "Initial commit")
	require.NoError(t, err)
}

func TestPush(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	err = f.Push(ctx, "origin", "main")
	require.NoError(t, err)
}

func TestPull(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	err = f.Pull(ctx, "origin", "main")
	require.NoError(t, err)
}

func TestAddRemote(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	err = f.AddRemote(ctx, "origin", "https://example.com/repo.fossil")
	require.NoError(t, err)
}

func TestRemoveRemote(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	err = f.RemoveRemote(ctx, "origin")
	require.NoError(t, err)
}

func TestRevisions(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	revs, err := f.Revisions(ctx, "testfile")
	require.NoError(t, err)
	assert.NotNil(t, revs)
}

func TestGetRevision(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	content, err := f.GetRevision(ctx, "testfile", "1")
	require.NoError(t, err)
	assert.NotNil(t, content)
}

func TestStatus(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	status, err := f.Status(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)
}

func TestCompact(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	require.NoError(t, err)

	err = f.Compact(ctx)
	require.NoError(t, err)
}
