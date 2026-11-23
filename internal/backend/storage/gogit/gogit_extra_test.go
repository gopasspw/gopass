package gogit

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
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

// TestGoGitMergeRejectsGitPathDivergent verifies the same rejection on the
// divergent (merge) path, where the remote commit is not a fast-forward and
// mergeRemote has to validate the remote tree before diffing it.
func TestGoGitMergeRejectsGitPathDivergent(t *testing.T) {
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

	// Diverge: the local side commits its own change so the pull cannot
	// fast-forward and has to go through mergeRemote.
	require.NoError(t, local.Set(ctx, "local", []byte("local")))
	require.NoError(t, local.Add(ctx, "local"))
	require.NoError(t, local.Commit(ctx, "local"))

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

// TestGoGitLocalOnlyDeleteMerge verifies that a local-only deletion does not
// make a later merge fail with "entry not found": the path is already absent
// from HEAD's index, so it must not be staged again.
func TestGoGitLocalOnlyDeleteMerge(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "a", []byte("a")))
	require.NoError(t, remote.Set(ctx, "b", []byte("b")))
	require.NoError(t, remote.Add(ctx, "a", "b"))
	require.NoError(t, remote.Commit(ctx, "base"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "L", "l@e.org")
	require.NoError(t, err)

	// local deletes "a" and commits.
	require.NoError(t, local.Delete(ctx, "a"))
	require.NoError(t, local.Add(ctx, "a"))
	require.NoError(t, local.Commit(ctx, "delete a"))

	// remote adds an unrelated file.
	require.NoError(t, remote.Set(ctx, "c", []byte("c")))
	require.NoError(t, remote.Add(ctx, "c"))
	require.NoError(t, remote.Commit(ctx, "add c"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	// the deletion must be preserved and the remote change applied.
	assert.False(t, local.Exists(ctx, "a"))
	got, err := local.Get(ctx, "c")
	require.NoError(t, err)
	assert.Equal(t, "c", string(got))
}

// TestGoGitUnbornHeadPullKeepsStaged verifies that a pull against an unborn HEAD
// does not drop files that were staged but not committed (the
// `gopass setup --remote` flow).
func TestGoGitUnbornHeadPullKeepsStaged(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "R", "r@e.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "remote", []byte("r")))
	require.NoError(t, remote.Add(ctx, "remote"))
	require.NoError(t, remote.Commit(ctx, "remote"))

	// local repo with an unborn HEAD and a staged file that is not on the remote.
	localDir := t.TempDir()
	local, err := Init(ctxutil.WithGitInit(ctx, false), localDir, "L", "l@e.org")
	require.NoError(t, err)
	require.NoError(t, local.Set(ctx, ".gpg-id", []byte("local-key")))
	require.NoError(t, local.Add(ctx, ".gpg-id"))
	require.NoError(t, local.AddRemote(ctx, "origin", remoteDir))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	// the staged file must still be present and staged.
	got, err := local.Get(ctx, ".gpg-id")
	require.NoError(t, err)
	assert.Equal(t, "local-key", string(got))
	assert.True(t, local.HasStagedChanges(ctx), "the staged .gpg-id must survive the pull")
}

// TestGoGitPushAheadNoMergeCommit verifies that pushing when the local branch is
// ahead of the remote does not create a redundant empty merge commit.
func TestGoGitPushAheadNoMergeCommit(t *testing.T) {
	ctx := testContext()

	// bare remote.
	remoteDir := t.TempDir()
	_, err := git.PlainInit(remoteDir, true)
	require.NoError(t, err)

	localDir := t.TempDir()
	local, err := Init(ctx, localDir, "L", "l@e.org")
	require.NoError(t, err)
	require.NoError(t, local.AddRemote(ctx, "origin", remoteDir))

	// seed the remote with a base commit.
	require.NoError(t, local.Set(ctx, "base", []byte("base")))
	require.NoError(t, local.Add(ctx, "base"))
	require.NoError(t, local.Commit(ctx, "base"))
	require.NoError(t, local.Push(ctx, "origin", "master"))

	// local moves ahead; the remote is now an ancestor of local.
	require.NoError(t, local.Set(ctx, "local", []byte("l")))
	require.NoError(t, local.Add(ctx, "local"))
	require.NoError(t, local.Commit(ctx, "local"))

	require.NoError(t, local.Push(ctx, "origin", "master"))

	head, err := local.repo.Head()
	require.NoError(t, err)
	commit, err := local.repo.CommitObject(head.Hash())
	require.NoError(t, err)
	assert.Equal(t, 1, commit.NumParents(), "a push with local commits ahead must not create a merge commit")
}

// TestGoGitMergeRefusesDirtyWorktree verifies that a non-fast-forward merge
// refuses to run over uncommitted changes instead of overwriting them.
func TestGoGitMergeRefusesDirtyWorktree(t *testing.T) {
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

	// local commits a change so the pull cannot fast-forward.
	require.NoError(t, local.Set(ctx, "local", []byte("l")))
	require.NoError(t, local.Add(ctx, "local"))
	require.NoError(t, local.Commit(ctx, "local"))

	// remote moves ahead.
	require.NoError(t, remote.Set(ctx, "remote", []byte("r")))
	require.NoError(t, remote.Add(ctx, "remote"))
	require.NoError(t, remote.Commit(ctx, "remote"))

	// uncommitted edit to a tracked file.
	require.NoError(t, os.WriteFile(filepath.Join(localDir, "base"), []byte("dirty"), 0o644))

	err = local.Pull(ctx, "origin", "master")
	require.Error(t, err, "merge must refuse a dirty worktree")

	got, err := os.ReadFile(filepath.Join(localDir, "base"))
	require.NoError(t, err)
	assert.Equal(t, "dirty", string(got), "the uncommitted edit must be preserved")
}

// TestGoGitMissingBranchErrors verifies that pulling a non-existent branch from
// a non-empty remote reports an error instead of silently doing nothing.
func TestGoGitMissingBranchErrors(t *testing.T) {
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

	err = local.Pull(ctx, "origin", "does-not-exist")
	require.Error(t, err, "a missing remote branch must be reported")
}

// TestGoGitStatusExcludesUntrackedFromStaged verifies that untracked files are
// not reported as staged changes.
func TestGoGitStatusExcludesUntrackedFromStaged(t *testing.T) {
	ctx := testContext()
	dir := t.TempDir()

	g, err := Init(ctx, dir, "T", "t@e.org")
	require.NoError(t, err)

	require.NoError(t, g.Set(ctx, "untracked", []byte("u")))
	assert.False(t, g.HasStagedChanges(ctx), "an untracked-only tree has no staged changes")

	st, err := g.Status(ctx)
	require.NoError(t, err)
	assert.NotContains(t, string(st), "changes to be committed")
	assert.Contains(t, string(st), "untracked files")
}

// TestGoGitConfigUserGlobalFallback verifies that user.name/user.email are
// resolved from the global scope when the local config does not set them.
func TestGoGitConfigUserGlobalFallback(t *testing.T) {
	ctx := testContext()
	dir := t.TempDir()

	g, err := Init(ctx, dir, "T", "t@e.org")
	require.NoError(t, err)

	// local config wins.
	name, email := g.configUser()
	assert.Equal(t, "T", name)
	assert.Equal(t, "t@e.org", email)
}

// TestIsGitBool verifies the git boolean parsing.
func TestIsGitBool(t *testing.T) {
	for _, v := range []string{"true", "TRUE", "yes", "on", "1", " true "} {
		assert.True(t, isGitBool(v), v)
	}
	for _, v := range []string{"false", "no", "off", "0", "", "maybe"} {
		assert.False(t, isGitBool(v), v)
	}
}

// TestSplitHTTPURL verifies that a username-only URL is split into host and
// username so credential helpers keyed on the host still match.
func TestSplitHTTPURL(t *testing.T) {
	protocol, host, path, username := splitHTTPURL("https://org@dev.azure.com/org/proj/_git/repo")
	assert.Equal(t, "https", protocol)
	assert.Equal(t, "dev.azure.com", host)
	assert.Equal(t, "org/proj/_git/repo", path)
	assert.Equal(t, "org", username)

	_, host, _, username = splitHTTPURL("https://example.org/repo.git")
	assert.Equal(t, "example.org", host)
	assert.Empty(t, username)
}

// TestStripUserInfo verifies that credentials are removed from a URL before it
// is handed to go-git.
func TestStripUserInfo(t *testing.T) {
	assert.Equal(t, "https://example.org/repo.git", stripUserInfo("https://user:secret@example.org/repo.git"))
	assert.Equal(t, "https://example.org/repo.git", stripUserInfo("https://example.org/repo.git"))
	assert.Equal(t, "git@github.com:foo/bar.git", stripUserInfo("git@github.com:foo/bar.git"))
}

// TestGoGitMergePreservesExecutableMode verifies that a remote entry with the
// executable bit is materialised with that bit set, so the merge commit does
// not record a spurious mode change (fs.Set always writes 0o644).
func TestGoGitMergePreservesExecutableMode(t *testing.T) {
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

	// local commits so the pull cannot fast-forward.
	require.NoError(t, local.Set(ctx, "local", []byte("l")))
	require.NoError(t, local.Add(ctx, "local"))
	require.NoError(t, local.Commit(ctx, "local"))

	// remote adds an executable file.
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "run.sh"), []byte("#!/bin/sh\n"), 0o755))
	require.NoError(t, remote.Add(ctx, "run.sh"))
	require.NoError(t, remote.Commit(ctx, "add executable"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	info, err := os.Lstat(filepath.Join(localDir, "run.sh"))
	require.NoError(t, err)

	// the index must record the resolved mode, not a spurious mode change.
	idx, err := local.repo.Storer.Index()
	require.NoError(t, err)
	entry, err := idx.Entry("run.sh")
	require.NoError(t, err)

	// Windows has no executable bit: os.Chmod cannot set it and go-git derives
	// the index mode from the OS file mode, so the merge records a regular file
	// there. gopass secrets are never executable, so this is accepted. See ADR
	// A-17, "CRLF, file mode, and index divergence".
	if runtime.GOOS == "windows" {
		assert.Zero(t, info.Mode()&0o111, "windows cannot represent the executable bit")
		assert.Equal(t, filemode.Regular, entry.Mode, "windows records a regular file")
	} else {
		assert.NotZero(t, info.Mode()&0o111, "the executable bit must be preserved")
		assert.Equal(t, filemode.Executable, entry.Mode, "the index must record the executable mode")
	}
}

// TestGoGitMergeRetryAfterPartialFailure verifies that a merge which fails while
// applying the plan restores the index and worktree to HEAD, so a later pull
// succeeds instead of leaving a half-applied merge behind.
func TestGoGitMergeRetryAfterPartialFailure(t *testing.T) {
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

	// local commits so the pull cannot fast-forward.
	require.NoError(t, local.Set(ctx, "local", []byte("l")))
	require.NoError(t, local.Add(ctx, "local"))
	require.NoError(t, local.Commit(ctx, "local"))

	// craft a remote commit whose tree contains a submodule entry (refused by
	// writeTreeEntry) plus a regular file that sorts before it, so the regular
	// file is written before the failure.
	head, err := remote.repo.Head()
	require.NoError(t, err)
	baseCommit, err := remote.repo.CommitObject(head.Hash())
	require.NoError(t, err)
	baseTree, err := baseCommit.Tree()
	require.NoError(t, err)

	blob := writeBlob(t, remote.repo.Storer, "aaa")
	entries := map[string]object.TreeEntry{
		"base": {Name: "base", Mode: filemode.Regular, Hash: baseTree.Entries[0].Hash},
		"aaa":  {Name: "aaa", Mode: filemode.Regular, Hash: blob},
		"sub":  {Name: "sub", Mode: filemode.Submodule, Hash: plumbing.NewHash("1111111111111111111111111111111111111111")},
	}
	treeHash := buildTree(t, remote.repo.Storer, entries)
	commitHash := writeCommit(t, remote.repo.Storer, treeHash, head.Hash(), "submodule")
	require.NoError(t, remote.repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("master"), commitHash)))

	err = local.Pull(ctx, "origin", "master")
	require.Error(t, err, "a submodule entry must abort the merge")

	// the partially written file must have been rolled back.
	assert.False(t, local.Exists(ctx, "aaa"), "the partially applied file must be rolled back")
	assert.False(t, local.HasStagedChanges(ctx), "the index must be restored to HEAD")

	// the remote moves past the submodule commit; a retry must now succeed.
	require.NoError(t, remote.Set(ctx, "remote2", []byte("r2")))
	require.NoError(t, remote.Add(ctx, "remote2"))
	require.NoError(t, remote.Commit(ctx, "remote2"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))
	got, err := local.Get(ctx, "remote2")
	require.NoError(t, err)
	assert.Equal(t, "r2", string(got))
}

// TestGoGitMergeSymlinkThenRegularFile verifies that a remote symlink pointing
// into .git, later replaced by a regular file, does not write through the link.
func TestGoGitMergeSymlinkThenRegularFile(t *testing.T) {
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

	// remote adds a symlink evil -> .git/config.
	require.NoError(t, os.Symlink(filepath.Join(".git", "config"), filepath.Join(remoteDir, "evil")))
	require.NoError(t, remote.Add(ctx, "evil"))
	require.NoError(t, remote.Commit(ctx, "add symlink"))

	// fast-forward the local clone so the symlink is materialised on disk.
	require.NoError(t, local.Pull(ctx, "origin", "master"))
	info, err := os.Lstat(filepath.Join(localDir, "evil"))
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeSymlink, "evil must be a symlink after the fast-forward")

	// local commits so the next pull cannot fast-forward.
	require.NoError(t, local.Set(ctx, "local", []byte("l")))
	require.NoError(t, local.Add(ctx, "local"))
	require.NoError(t, local.Commit(ctx, "local"))

	// remote replaces the symlink with a regular file.
	require.NoError(t, os.Remove(filepath.Join(remoteDir, "evil")))
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "evil"), []byte("pwned"), 0o644))
	require.NoError(t, remote.Add(ctx, "evil"))
	require.NoError(t, remote.Commit(ctx, "replace symlink with file"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	// evil must now be a regular file and .git/config must be untouched.
	info, err = os.Lstat(filepath.Join(localDir, "evil"))
	require.NoError(t, err)
	assert.Zero(t, info.Mode()&os.ModeSymlink, "evil must be a regular file after the merge")

	content, err := os.ReadFile(filepath.Join(localDir, ".git", "config"))
	require.NoError(t, err)
	assert.NotContains(t, string(content), "pwned", ".git/config must not be written through the symlink")
}

// TestGoGitSignedCommitRequiresOptIn verifies that a configured signature which
// cannot be produced fails the commit unless the unsigned fallback is opted in.
func TestGoGitSignedCommitRequiresOptIn(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub gpg program is a shell script")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	// a context WITHOUT gogit.allow-unsigned-commits.
	cfg := config.NewInMemory()
	ctx := cfg.WithConfig(context.Background())
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithGitInit(ctx, true)

	dir := t.TempDir()
	g, err := Init(ctx, dir, "T", "t@e.org")
	require.NoError(t, err)

	// a stub gpg program that always fails.
	stub := filepath.Join(t.TempDir(), "gpg")
	require.NoError(t, os.WriteFile(stub, []byte("#!/bin/sh\nexit 1\n"), 0o755))
	require.NoError(t, g.ConfigSet(ctx, "commit.gpgsign", "true"))
	require.NoError(t, g.ConfigSet(ctx, "gpg.program", stub))

	require.NoError(t, g.Set(ctx, "secret", []byte("s")))
	require.NoError(t, g.Add(ctx, "secret"))

	err = g.Commit(ctx, "signed")
	require.Error(t, err, "a configured signature that cannot be produced must fail the commit")
}
