package fossilfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTestDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "fossilfs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	return dir
}

func TestNew(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	marker := filepath.Join(dir, CheckoutMarker)
	_, err := os.Create(marker)
	assert.NoError(t, err)

	f, err := New(dir)
	assert.NoError(t, err)
	assert.NotNil(t, f)
}

func TestClone(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	repo := "https://example.com/repo.fossil"

	f, err := Clone(ctx, repo, dir)
	assert.NoError(t, err)
	assert.NotNil(t, f)
}

func TestInit(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()

	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, f)
}

func TestAdd(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	err = f.Add(ctx, "testfile")
	assert.NoError(t, err)
}

func TestCommit(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	err = f.Commit(ctx, "Initial commit")
	assert.NoError(t, err)
}

func TestPush(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	err = f.Push(ctx, "origin", "main")
	assert.NoError(t, err)
}

func TestPull(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	err = f.Pull(ctx, "origin", "main")
	assert.NoError(t, err)
}

func TestAddRemote(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	err = f.AddRemote(ctx, "origin", "https://example.com/repo.fossil")
	assert.NoError(t, err)
}

func TestRemoveRemote(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	err = f.RemoveRemote(ctx, "origin")
	assert.NoError(t, err)
}

func TestRevisions(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	revs, err := f.Revisions(ctx, "testfile")
	assert.NoError(t, err)
	assert.NotNil(t, revs)
}

func TestGetRevision(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	content, err := f.GetRevision(ctx, "testfile", "1")
	assert.NoError(t, err)
	assert.NotNil(t, content)
}

func TestStatus(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	status, err := f.Status(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, status)
}

func TestCompact(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	f, err := Init(ctx, dir, "", "")
	assert.NoError(t, err)

	err = f.Compact(ctx)
	assert.NoError(t, err)
}
