package gogit

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoGitAddIsolation verifies that adding an explicit path does not stage
// unrelated files (the AddOptions{All:true, Path:...} bug).
func TestGoGitAddIsolation(t *testing.T) {
	ctx := testContext()
	dir := t.TempDir()

	g, err := Init(ctx, dir, "T", "t@e.org")
	require.NoError(t, err)

	require.NoError(t, g.Set(ctx, "a.gpg", []byte("a1")))
	require.NoError(t, g.Add(ctx, "a.gpg"))
	require.NoError(t, g.Commit(ctx, "add a"))

	// modify a and create an unrelated untracked file.
	require.NoError(t, g.Set(ctx, "a.gpg", []byte("a2")))
	require.NoError(t, g.Set(ctx, "unrelated.gpg", []byte("u")))

	require.NoError(t, g.Add(ctx, "a.gpg"))

	// only a.gpg must be staged; unrelated.gpg must remain untracked.
	wt, err := g.worktree()
	require.NoError(t, err)
	st, err := wt.Status()
	require.NoError(t, err)
	assert.Equal(t, git.Modified, st["a.gpg"].Staging)
	assert.Equal(t, git.Untracked, st["unrelated.gpg"].Worktree)
	assert.Contains(t, g.ListUntrackedFiles(ctx), "unrelated.gpg")
}

// TestGoGitPushPullsFirst verifies that a push against a remote that is ahead
// pulls first instead of failing with a non-fast-forward error.
func TestGoGitPushPullsFirst(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "base", []byte("base")))
	require.NoError(t, remote.Add(ctx, "base"))
	require.NoError(t, remote.Commit(ctx, "base"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "L", "l@e.org")
	require.NoError(t, err)

	// remote moves ahead.
	require.NoError(t, remote.Set(ctx, "remote-only", []byte("r")))
	require.NoError(t, remote.Add(ctx, "remote-only"))
	require.NoError(t, remote.Commit(ctx, "remote-only"))

	// local pushes without pulling first; the backend must pull first.
	require.NoError(t, local.Push(ctx, "origin", "master"))

	got, err := local.Get(ctx, "remote-only")
	require.NoError(t, err)
	assert.Equal(t, "r", string(got))
}

// TestGoGitPullPreservesUntracked verifies that a fast-forward pull does not
// delete untracked files (the HardReset bug).
func TestGoGitPullPreservesUntracked(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "base", []byte("base")))
	require.NoError(t, remote.Add(ctx, "base"))
	require.NoError(t, remote.Commit(ctx, "base"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "L", "l@e.org")
	require.NoError(t, err)

	// untracked file in the local worktree.
	require.NoError(t, os.WriteFile(filepath.Join(localDir, "untracked"), []byte("u"), 0o644))

	// remote moves ahead.
	require.NoError(t, remote.Set(ctx, "remote-only", []byte("r")))
	require.NoError(t, remote.Add(ctx, "remote-only"))
	require.NoError(t, remote.Commit(ctx, "remote-only"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	_, err = os.Stat(filepath.Join(localDir, "untracked"))
	require.NoError(t, err, "untracked file must survive a pull")
	got, err := local.Get(ctx, "remote-only")
	require.NoError(t, err)
	assert.Equal(t, "r", string(got))
}

// TestGoGitMergeSymlink verifies that a symlink added on the remote is
// materialised as a symlink, not as a regular file containing the target.
func TestGoGitMergeSymlink(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "target.gpg", []byte("secret")))
	require.NoError(t, remote.Add(ctx, "target.gpg"))
	require.NoError(t, remote.Commit(ctx, "target"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "L", "l@e.org")
	require.NoError(t, err)

	// remote adds a symlink.
	require.NoError(t, os.Symlink("target.gpg", filepath.Join(remoteDir, "alias.gpg")))
	require.NoError(t, remote.Add(ctx, "alias.gpg"))
	require.NoError(t, remote.Commit(ctx, "add symlink"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	info, err := os.Lstat(filepath.Join(localDir, "alias.gpg"))
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeSymlink, "alias.gpg must be a symlink")

	target, err := os.Readlink(filepath.Join(localDir, "alias.gpg"))
	require.NoError(t, err)
	assert.Equal(t, "target.gpg", target)
}

// TestGoGitMergeRejectsGitPath verifies that a remote commit carrying a path
// inside .git is rejected and the worktree is left untouched.
func TestGoGitMergeRejectsGitPath(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "base", []byte("base")))
	require.NoError(t, remote.Add(ctx, "base"))
	require.NoError(t, remote.Commit(ctx, "base"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "L", "l@e.org")
	require.NoError(t, err)

	// craft a remote commit whose tree contains .git/hooks/pre-commit.
	head, err := remote.repo.Head()
	require.NoError(t, err)
	baseCommit, err := remote.repo.CommitObject(head.Hash())
	require.NoError(t, err)
	baseTree, err := baseCommit.Tree()
	require.NoError(t, err)

	blob := writeBlob(t, remote.repo.Storer, "#!/bin/sh\necho pwned\n")
	entries := map[string]object.TreeEntry{
		"base":                  {Name: "base", Mode: filemode.Regular, Hash: baseTree.Entries[0].Hash},
		".git/hooks/pre-commit": {Name: "pre-commit", Mode: filemode.Executable, Hash: blob},
	}
	treeHash := buildTree(t, remote.repo.Storer, entries)
	commitHash := writeCommit(t, remote.repo.Storer, treeHash, head.Hash(), "malicious")
	require.NoError(t, remote.repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("master"), commitHash)))

	err = local.Pull(ctx, "origin", "master")
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".git")

	// the hook must not have been written.
	_, err = os.Stat(filepath.Join(localDir, ".git", "hooks", "pre-commit"))
	assert.True(t, os.IsNotExist(err), "hook must not be materialised")
}

// TestGoGitMergeGpgIDConflict verifies that a concurrent recipient change is
// not resolved automatically.
func TestGoGitMergeGpgIDConflict(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, ".gpg-id", []byte("remote-key")))
	require.NoError(t, remote.Add(ctx, ".gpg-id"))
	require.NoError(t, remote.Commit(ctx, "gpg-id"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "L", "l@e.org")
	require.NoError(t, err)

	// both sides change .gpg-id differently.
	require.NoError(t, local.Set(ctx, ".gpg-id", []byte("local-key")))
	require.NoError(t, local.Add(ctx, ".gpg-id"))
	require.NoError(t, local.Commit(ctx, "local gpg-id"))

	require.NoError(t, remote.Set(ctx, ".gpg-id", []byte("remote-key-2")))
	require.NoError(t, remote.Add(ctx, ".gpg-id"))
	require.NoError(t, remote.Commit(ctx, "remote gpg-id"))

	err = local.Pull(ctx, "origin", "master")
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".gpg-id")

	// local value must be untouched.
	got, err := local.Get(ctx, ".gpg-id")
	require.NoError(t, err)
	assert.Equal(t, "local-key", string(got))
}

// TestGoGitConflictNameCollision verifies that an existing conflict file is not
// clobbered: the hash prefix grows until the name is unique.
func TestGoGitConflictNameCollision(t *testing.T) {
	ctx := testContext()
	dir := t.TempDir()

	g, err := Init(ctx, dir, "T", "t@e.org")
	require.NoError(t, err)

	hash := plumbing.NewHash("1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b")
	first := conflictName("base", hash.String()[:7])
	require.NoError(t, g.Set(ctx, first, []byte("existing")))

	name, err := g.uniqueConflictName(ctx, "base", hash)
	require.NoError(t, err)
	assert.NotEqual(t, first, name)
	assert.True(t, strings.HasPrefix(name, "base.conflict-"))
	assert.Greater(t, len(name), len(first))
}

// TestGoGitRevisionsWithMerge verifies that Revisions does not list commits
// that never touched the path, even across merge commits.
func TestGoGitRevisionsWithMerge(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "base", []byte("base")))
	require.NoError(t, remote.Add(ctx, "base"))
	require.NoError(t, remote.Commit(ctx, "base"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "L", "l@e.org")
	require.NoError(t, err)

	// local touches "local-only", remote touches "remote-only".
	require.NoError(t, local.Set(ctx, "local-only", []byte("l")))
	require.NoError(t, local.Add(ctx, "local-only"))
	require.NoError(t, local.Commit(ctx, "local-only"))

	require.NoError(t, remote.Set(ctx, "remote-only", []byte("r")))
	require.NoError(t, remote.Add(ctx, "remote-only"))
	require.NoError(t, remote.Commit(ctx, "remote-only"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	// "local-only" was only ever touched by the local commit; the merge commit
	// must not appear in its history.
	revs, err := local.Revisions(ctx, "local-only")
	require.NoError(t, err)
	require.Len(t, revs, 1)
	assert.Equal(t, "local-only", revs[0].Subject)
}

// TestValidateRemotePath exercises the remote path validation rules.
func TestValidateRemotePath(t *testing.T) {
	for _, tc := range []struct {
		path string
		ok   bool
	}{
		{"secrets/api.gpg", true},
		{"a/b/c.gpg", true},
		{".gpg-id", true},
		{"", false},
		{"/etc/passwd", false},
		{"../escape", false},
		{"a/../../escape", false},
		{".git/config", false},
		{".git/hooks/pre-commit", false},
		{".GIT/config", false},
	} {
		err := validateRemotePath(tc.path)
		if tc.ok {
			assert.NoError(t, err, tc.path)
		} else {
			assert.Error(t, err, tc.path)
		}
	}
}

// TestURLUserInfo verifies that a username-only URL does not short-circuit the
// credential helper and that userinfo is percent-decoded.
func TestURLUserInfo(t *testing.T) {
	u, p, ok := urlUserInfo("https://user:pass@example.org/repo.git")
	assert.True(t, ok)
	assert.Equal(t, "user", u)
	assert.Equal(t, "pass", p)

	// username only -> not ok, so the credential helper is consulted.
	_, _, ok = urlUserInfo("https://org@dev.azure.com/org/proj/_git/repo")
	assert.False(t, ok)

	// percent-encoded password.
	u, p, ok = urlUserInfo("https://user:p%40ss@example.org/repo.git")
	assert.True(t, ok)
	assert.Equal(t, "user", u)
	assert.Equal(t, "p@ss", p)

	_, _, ok = urlUserInfo("https://example.org/repo.git")
	assert.False(t, ok)
}

// TestRedactURL verifies that passwords are removed from URLs.
func TestRedactURL(t *testing.T) {
	assert.Equal(t, "https://user:REDACTED@example.org/repo.git", redactURL("https://user:secret@example.org/repo.git"))
	assert.Equal(t, "https://example.org/repo.git", redactURL("https://example.org/repo.git"))
	assert.Equal(t, "git@github.com:foo/bar.git", redactURL("git@github.com:foo/bar.git"))
}

// TestGoGitWithoutGitBinary verifies that the core backend operations work with
// no git binary on PATH. (Cloning a local path is excluded: go-git's file
// transport shells out to git-upload-pack.)
func TestGoGitWithoutGitBinary(t *testing.T) {
	t.Setenv("PATH", "")

	ctx := testContext()
	dir := t.TempDir()

	g, err := Init(ctx, dir, "T", "t@e.org")
	require.NoError(t, err)

	require.NoError(t, g.Set(ctx, "secret", []byte("s3cr3t")))
	require.NoError(t, g.Add(ctx, "secret"))
	require.NoError(t, g.Commit(ctx, "add secret"))

	got, err := g.Get(ctx, "secret")
	require.NoError(t, err)
	assert.Equal(t, "s3cr3t", string(got))

	revs, err := g.Revisions(ctx, "secret")
	require.NoError(t, err)
	require.Len(t, revs, 1)
	assert.Equal(t, "add secret", revs[0].Subject)

	content, err := g.GetRevision(ctx, "secret", "latest")
	require.NoError(t, err)
	assert.Equal(t, "s3cr3t", string(content))

	// a signed commit must fall back to an unsigned one without git.
	require.NoError(t, g.ConfigSet(ctx, "commit.gpgsign", "true"))
	require.NoError(t, g.Set(ctx, "secret", []byte("s3cr3t2")))
	require.NoError(t, g.Add(ctx, "secret"))
	require.NoError(t, g.Commit(ctx, "update secret"))
}

// TestGoGitPullDirtyWorktree verifies that a pull with uncommitted changes to a
// tracked file aborts instead of silently discarding the edit.
func TestGoGitPullDirtyWorktree(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "base", []byte("base")))
	require.NoError(t, remote.Add(ctx, "base"))
	require.NoError(t, remote.Commit(ctx, "base"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "L", "l@e.org")
	require.NoError(t, err)

	// uncommitted edit to a tracked file.
	require.NoError(t, os.WriteFile(filepath.Join(localDir, "base"), []byte("dirty"), 0o644))

	// remote moves ahead.
	require.NoError(t, remote.Set(ctx, "remote", []byte("r")))
	require.NoError(t, remote.Add(ctx, "remote"))
	require.NoError(t, remote.Commit(ctx, "remote"))

	err = local.Pull(ctx, "origin", "master")
	require.Error(t, err, "pull must abort on a dirty worktree")

	// the local edit must be preserved.
	got, err := os.ReadFile(filepath.Join(localDir, "base"))
	require.NoError(t, err)
	assert.Equal(t, "dirty", string(got))
}

// writeBlob stores a blob object and returns its hash.
func writeBlob(t *testing.T, s storer.EncodedObjectStorer, content string) plumbing.Hash {
	t.Helper()

	obj := s.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	w, err := obj.Writer()
	require.NoError(t, err)
	_, err = w.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	h, err := s.SetEncodedObject(obj)
	require.NoError(t, err)

	return h
}

// buildTree builds a (possibly nested) tree from a flat map of path -> entry.
func buildTree(t *testing.T, s storer.EncodedObjectStorer, entries map[string]object.TreeEntry) plumbing.Hash {
	t.Helper()

	dirs := map[string]map[string]object.TreeEntry{}
	tree := object.Tree{}
	for name, e := range entries {
		if i := strings.Index(name, "/"); i >= 0 {
			dir, rest := name[:i], name[i+1:]
			if dirs[dir] == nil {
				dirs[dir] = map[string]object.TreeEntry{}
			}
			dirs[dir][rest] = e

			continue
		}
		tree.Entries = append(tree.Entries, object.TreeEntry{Name: name, Mode: e.Mode, Hash: e.Hash})
	}
	for dir, sub := range dirs {
		tree.Entries = append(tree.Entries, object.TreeEntry{Name: dir, Mode: filemode.Dir, Hash: buildTree(t, s, sub)})
	}
	sort.Sort(object.TreeEntrySorter(tree.Entries))

	obj := s.NewEncodedObject()
	require.NoError(t, tree.Encode(obj))
	h, err := s.SetEncodedObject(obj)
	require.NoError(t, err)

	return h
}

// writeCommit stores a commit object with the given tree and parent.
func writeCommit(t *testing.T, s storer.EncodedObjectStorer, tree, parent plumbing.Hash, msg string) plumbing.Hash {
	t.Helper()

	sig := object.Signature{Name: "Test", Email: "test@example.org"}
	commit := &object.Commit{
		Author:    sig,
		Committer: sig,
		Message:   msg,
		TreeHash:  tree,
	}
	if parent != plumbing.ZeroHash {
		commit.ParentHashes = []plumbing.Hash{parent}
	}

	obj := s.NewEncodedObject()
	require.NoError(t, commit.Encode(obj))
	h, err := s.SetEncodedObject(obj)
	require.NoError(t, err)

	return h
}
