package gogit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
)

// changeInfo describes the state of a single path in a tree diff.
type changeInfo struct {
	hash    plumbing.Hash
	mode    filemode.FileMode
	deleted bool
}

// present reports whether the path changed on this side (i.e. exists in the
// diff). A zero changeInfo means "unchanged".
func (c changeInfo) present() bool {
	return c.hash != plumbing.ZeroHash || c.deleted
}

// same reports whether two change infos describe the identical tree entry,
// including its file mode.
func (c changeInfo) same(o changeInfo) bool {
	return c.hash == o.hash && c.mode == o.mode && c.deleted == o.deleted
}

// treeChanges returns the per-path changes between base and tree. A nil base is
// treated as an empty tree, so every file in tree is reported as an insertion.
func treeChanges(base, tree *object.Tree) (map[string]changeInfo, error) {
	changes := make(map[string]changeInfo)
	if tree == nil {
		return changes, nil
	}

	if base == nil {
		iter := tree.Files()
		defer iter.Close()

		if err := iter.ForEach(func(f *object.File) error {
			changes[f.Name] = changeInfo{hash: f.Hash, mode: f.Mode}

			return nil
		}); err != nil {
			return nil, err
		}

		return changes, nil
	}

	cs, err := object.DiffTree(base, tree)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		action, err := c.Action()
		if err != nil {
			return nil, err
		}

		switch action {
		case merkletrie.Insert:
			changes[c.To.Name] = changeInfo{hash: c.To.TreeEntry.Hash, mode: c.To.TreeEntry.Mode}
		case merkletrie.Delete:
			changes[c.From.Name] = changeInfo{deleted: true}
		case merkletrie.Modify:
			changes[c.To.Name] = changeInfo{hash: c.To.TreeEntry.Hash, mode: c.To.TreeEntry.Mode}
		}
	}

	return changes, nil
}

// mergeRemote performs a file-level merge of remoteHash into the current HEAD.
// It is correct for gopass because the files are encrypted blobs: a
// content-level three-way merge is meaningless, so changes are resolved per
// path. A true conflict never discards a secret: the local version is kept and
// the remote version is materialised as <name>.conflict-<shortsha>.
//
// The merge is transactional: every remote path is validated and the whole
// resolution is computed before anything is written. If applying the plan
// fails, the index and worktree are restored to HEAD.
//
// The result is a two-parent merge commit so that a later user of the real git
// CLI sees valid, connected history.
func (g *GoGit) mergeRemote(ctx context.Context, remoteHash plumbing.Hash) error {
	head, err := g.repo.Head()
	if err != nil {
		return err
	}
	localHash := head.Hash()

	// Refuse to merge over uncommitted local changes: the plan writes remote
	// paths with fs.Set/fs.Delete, which would overwrite uncommitted edits and
	// sweep staged-but-uncommitted changes into the merge commit.
	if dirty, err := g.hasUncommittedChanges(); err != nil {
		return err
	} else if dirty {
		return fmt.Errorf("refusing to merge with uncommitted changes; commit or discard them and pull again")
	}

	localCommit, err := g.repo.CommitObject(localHash)
	if err != nil {
		return err
	}
	remoteCommit, err := g.repo.CommitObject(remoteHash)
	if err != nil {
		return err
	}

	// 1. best common ancestor
	bases, err := localCommit.MergeBase(remoteCommit)
	if err != nil {
		return err
	}
	var baseTree *object.Tree
	if len(bases) > 0 {
		if baseTree, err = bases[0].Tree(); err != nil {
			return err
		}
	}

	localTree, err := localCommit.Tree()
	if err != nil {
		return err
	}
	remoteTree, err := remoteCommit.Tree()
	if err != nil {
		return err
	}

	// 2. validate every path in the remote tree before diffing it. A remote
	// commit must never be able to write outside the store root or into .git.
	//
	// This has to happen before the diff: from go-git v5.19.3 on, object.DiffTree
	// walks the tree with path validation disabled and surfaces a rejected name
	// only later, through Tree.Tree, which flattens the error to
	// ErrDirectoryNotFound. Validating here keeps the offending path in the
	// message and does not depend on go-git's internal walker behaviour.
	if err := validateTreePaths(remoteTree); err != nil {
		return err
	}

	// 3. per-side changes vs base
	localChanges, err := treeChanges(baseTree, localTree)
	if err != nil {
		return err
	}
	remoteChanges, err := treeChanges(baseTree, remoteTree)
	if err != nil {
		return err
	}

	paths := unionPaths(localChanges, remoteChanges)

	// 4. validate every changed path again before touching the worktree. The
	// tree walk above already covers the remote tree; this is defense in depth
	// for paths that only appear in the diff.
	for _, p := range paths {
		if err := validateRemotePath(p); err != nil {
			return err
		}
	}

	// 4. resolve every path in memory first. Nothing is written until the whole
	// merge is known to be applicable, so a failure leaves the worktree and
	// index untouched.
	plan, conflicts, err := resolveMergePlan(paths, localChanges, remoteChanges)
	if err != nil {
		return err
	}

	// 5. apply the plan. From here on the worktree is modified; on failure we
	// restore the index and worktree to HEAD.
	if err := g.applyMergePlan(ctx, plan, conflicts, remoteTree, remoteHash); err != nil {
		if rerr := g.restoreToHead(planPaths(plan, conflicts)); rerr != nil {
			out.Warningf(ctx, "Failed to restore worktree after failed merge: %s", rerr)
		}

		return err
	}

	if len(conflicts) > 0 {
		out.Warningf(ctx, "Merge conflict in %d file(s); the remote versions were kept as *.conflict-<sha>:", len(conflicts))
		for _, p := range conflicts {
			out.Warningf(ctx, "  - %s", p)
		}
	}

	// 6. two-parent merge commit so real git users see valid history
	msg := fmt.Sprintf("Merge remote-tracking branch into %s (gopass go-git backend)", head.Name().Short())

	return g.commitMerge(ctx, msg, localHash, remoteHash)
}

// resolveMergePlan computes the per-path resolution of a merge without touching
// the worktree. It returns the changes to apply and the list of true conflicts.
func resolveMergePlan(paths []string, localChanges, remoteChanges map[string]changeInfo) (map[string]changeInfo, []string, error) {
	plan := make(map[string]changeInfo, len(paths))
	conflicts := []string{}
	for _, p := range paths {
		lc, rc := localChanges[p], remoteChanges[p]

		// Recipient changes are security relevant: never resolve a genuine
		// conflict automatically, require manual resolution.
		if isRecipientPath(p) && lc.present() && rc.present() && !lc.same(rc) {
			return nil, nil, fmt.Errorf("conflicting recipient change in %q: resolve .gpg-id manually and pull again", p)
		}

		switch {
		case lc.present() && !rc.present():
			// changed only locally -> already committed in HEAD, nothing to do.
			// In particular a local-only deletion must not be staged again: the
			// path is already absent from HEAD's index, so staging it would fail
			// with "entry not found".
		case rc.present() && !lc.present():
			// changed only remotely -> apply remote content
			plan[p] = rc
		case lc.same(rc):
			// changed identically on both sides -> no-op
		case lc.deleted != rc.deleted:
			// delete vs modify -> keep the modification (never drop a secret)
			if lc.deleted {
				plan[p] = rc
			}
		default:
			// changed on both sides, different content -> conflict
			conflicts = append(conflicts, p)
		}
	}

	return plan, conflicts, nil
}

// applyMergePlan writes the resolved changes and the conflict files into the
// worktree and stages them.
func (g *GoGit) applyMergePlan(ctx context.Context, plan map[string]changeInfo, conflicts []string, remoteTree *object.Tree, remoteHash plumbing.Hash) error {
	paths := make([]string, 0, len(plan))
	for p := range plan {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		if err := g.applyChange(ctx, p, plan[p], remoteTree); err != nil {
			return err
		}
	}

	sort.Strings(conflicts)
	for _, p := range conflicts {
		if err := g.materializeConflict(ctx, p, remoteTree, remoteHash); err != nil {
			return err
		}
	}

	return nil
}

// applyChange materialises a single resolved change and stages it.
func (g *GoGit) applyChange(ctx context.Context, p string, ci changeInfo, remoteTree *object.Tree) error {
	if ci.deleted {
		// The file may already be gone (e.g. the local side deleted it), so a
		// missing file is not an error.
		if err := g.fs.Delete(ctx, p); err != nil && !os.IsNotExist(err) {
			return err
		}

		return g.stageDelete(p)
	}

	if err := g.writeTreeEntry(ctx, p, ci, remoteTree); err != nil {
		return err
	}

	return g.stagePath(p)
}

// writeTreeEntry writes the remote version of a path into the worktree,
// recreating symlinks and refusing submodules.
func (g *GoGit) writeTreeEntry(ctx context.Context, p string, ci changeInfo, remoteTree *object.Tree) error {
	switch ci.mode {
	case filemode.Symlink:
		target, err := blobContent(remoteTree, p)
		if err != nil {
			return err
		}

		return g.fs.SetSymlink(ctx, p, string(target))
	case filemode.Submodule:
		return fmt.Errorf("refusing to materialise submodule %q from remote", p)
	case filemode.Empty, filemode.Dir, filemode.Regular, filemode.Deprecated, filemode.Executable:
		content, err := blobContent(remoteTree, p)
		if err != nil {
			return err
		}
		if err := g.fs.Set(ctx, p, content); err != nil && !errors.Is(err, store.ErrMeaninglessWrite) {
			return err
		}

		return g.applyFileMode(p, ci.mode)
	default:
		return fmt.Errorf("refusing to materialise %q with unsupported mode %s from remote", p, ci.mode)
	}
}

// applyFileMode makes the worktree file at p match the git file mode. fs.Set
// always writes 0o644, so an executable entry from the remote would otherwise
// lose its bit and the merge commit would record a mode change. Windows has no
// executable bit, so os.Chmod leaves it regular there and the index (which
// go-git derives from the OS file mode) records a regular file; gopass secrets
// are never executable, so this is accepted.
func (g *GoGit) applyFileMode(p string, mode filemode.FileMode) error {
	perm := os.FileMode(0o644)
	if mode == filemode.Executable {
		perm = 0o755
	}

	return os.Chmod(filepath.Join(g.fs.Path(), filepath.FromSlash(p)), perm)
}

// materializeConflict keeps the local file untouched and writes the remote blob
// to <name>.conflict-<shortsha>, then stages the new file.
func (g *GoGit) materializeConflict(ctx context.Context, p string, remoteTree *object.Tree, remoteHash plumbing.Hash) error {
	content, err := blobContent(remoteTree, p)
	if err != nil {
		return err
	}

	conflictPath, err := g.uniqueConflictName(ctx, p, remoteHash)
	if err != nil {
		return err
	}

	// Preserve the remote entry type: a symlink conflict must stay a symlink.
	if entry, err := remoteTree.FindEntry(p); err == nil && entry.Mode == filemode.Symlink {
		if err := g.fs.SetSymlink(ctx, conflictPath, string(content)); err != nil {
			return err
		}
	} else if err := g.fs.Set(ctx, conflictPath, content); err != nil && !errors.Is(err, store.ErrMeaninglessWrite) {
		return err
	}

	return g.stagePath(conflictPath)
}

// uniqueConflictName returns a conflict-file name that does not collide with an
// existing file, growing the hash prefix until it is unique.
func (g *GoGit) uniqueConflictName(ctx context.Context, p string, remoteHash plumbing.Hash) (string, error) {
	full := remoteHash.String()
	for n := 7; n <= len(full); n++ {
		name := conflictName(p, full[:n])
		if !g.fs.Exists(ctx, name) {
			return name, nil
		}
	}

	return "", fmt.Errorf("could not find a unique conflict name for %q", p)
}

// stagePath adds the given path (or its deletion) to the index.
func (g *GoGit) stagePath(p string) error {
	wt, err := g.worktree()
	if err != nil {
		return err
	}

	return wt.AddWithOptions(&git.AddOptions{Path: p})
}

// stageDelete removes the given path from the index.
func (g *GoGit) stageDelete(p string) error {
	return g.stagePath(p)
}

// restoreToHead discards any partial merge by resetting the index and worktree
// to HEAD. Only the paths touched by the merge are reset (HardReset with Files),
// so uncommitted changes to unrelated files are preserved and the reset does not
// fail on the unstaged changes the merge itself just created.
func (g *GoGit) restoreToHead(paths []string) error {
	head, err := g.repo.Head()
	if err != nil {
		return err
	}
	wt, err := g.worktree()
	if err != nil {
		return err
	}

	return wt.Reset(&git.ResetOptions{Commit: head.Hash(), Mode: git.HardReset, Files: paths})
}

// planPaths returns the sorted union of the plan and conflict paths.
func planPaths(plan map[string]changeInfo, conflicts []string) []string {
	paths := make([]string, 0, len(plan)+len(conflicts))
	for p := range plan {
		paths = append(paths, p)
	}
	paths = append(paths, conflicts...)
	sort.Strings(paths)

	return paths
}

// hasUncommittedChanges reports whether any tracked file has staged or unstaged
// modifications. Untracked files are ignored: they are not part of the merge.
func (g *GoGit) hasUncommittedChanges() (bool, error) {
	wt, err := g.worktree()
	if err != nil {
		return false, err
	}
	st, err := wt.StatusWithOptions(git.StatusOptions{Strategy: git.Preload})
	if err != nil {
		return false, err
	}

	for _, fs := range st {
		if fs.Staging != git.Unmodified && fs.Staging != git.Untracked {
			return true, nil
		}
		if fs.Worktree != git.Unmodified && fs.Worktree != git.Untracked {
			return true, nil
		}
	}

	return false, nil
}

// commitMerge creates the two-parent merge commit, honouring commit.gpgsign the
// same way Commit does.
func (g *GoGit) commitMerge(ctx context.Context, msg string, localHash, remoteHash plumbing.Hash) error {
	if g.shouldSign() {
		if err := g.commitMergeSignedCLI(ctx, msg, remoteHash); err == nil {
			return nil
		}
		if !g.allowUnsignedCommits(ctx) {
			return fmt.Errorf("failed to create a signed merge commit and gogit.allow-unsigned-commits is not set; refusing to create an unsigned commit")
		}
		out.Warningf(ctx, "Failed to create a signed merge commit, creating an unsigned commit instead.")
	}

	wt, err := g.worktree()
	if err != nil {
		return err
	}
	sig := g.commitSignature(ctx)
	_, err = wt.Commit(msg, &git.CommitOptions{
		Author:            sig,
		Committer:         sig,
		Parents:           []plumbing.Hash{localHash, remoteHash},
		AllowEmptyCommits: true,
	})

	return err
}

// blobContent returns the content of the file at path in the given tree.
func blobContent(tree *object.Tree, p string) ([]byte, error) {
	f, err := tree.File(p)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q from remote tree: %w", p, err)
	}
	content, err := f.Contents()
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

// conflictName returns the conflict-file name for a path, inserting the suffix
// before the extension: secrets/api.gpg -> secrets/api.conflict-1a2b3c4.gpg.
func conflictName(p, short string) string {
	dir, file := path.Split(p)
	ext := path.Ext(file)
	base := strings.TrimSuffix(file, ext)
	if ext == "" {
		return dir + base + ".conflict-" + short
	}

	return dir + base + ".conflict-" + short + ext
}

// isRecipientPath reports whether p is a recipient or identity file whose
// concurrent change must never be resolved automatically.
func isRecipientPath(p string) bool {
	switch path.Base(p) {
	case ".gpg-id", ".age-recipients", ".age-ids", ".gopass-mapping":
		return true
	default:
		return false
	}
}

// validateCommitTree validates the tree of the given commit before its content
// is materialised into the worktree. It is used by the fast-forward path, which
// resets the worktree straight to the remote commit without a diff.
func (g *GoGit) validateCommitTree(hash plumbing.Hash) error {
	commit, err := g.repo.CommitObject(hash)
	if err != nil {
		return err
	}
	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	return validateTreePaths(tree)
}

// validateTreePaths rejects a tree that contains any path which must never be
// materialised from a remote: absolute paths, paths escaping the store root,
// and anything inside a VCS metadata directory.
//
// The walk uses go-git's default (validating) tree walker, which rejects
// `.git`-shaped names itself. That is deliberate: it keeps this check working
// across go-git versions. From v5.19.3 on, object.DiffTree walks with path
// validation disabled and surfaces a rejected name only later, through
// Tree.Tree, which flattens the error to ErrDirectoryNotFound. Validating the
// tree up front keeps the offending path in the message and does not depend on
// go-git's internal walker behaviour.
func validateTreePaths(tree *object.Tree) error {
	if tree == nil {
		return nil
	}

	walker := object.NewTreeWalker(tree, true, nil)
	defer walker.Close()

	for {
		name, _, err := walker.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := validateRemotePath(name); err != nil {
			return err
		}
	}
}

// validateRemotePath rejects paths that must never be materialised from a
// remote tree: absolute paths, paths escaping the store root, and anything
// inside the git metadata directory.
func validateRemotePath(p string) error {
	if p == "" {
		return fmt.Errorf("invalid empty path in remote tree")
	}
	if path.IsAbs(p) || filepath.IsAbs(p) {
		return fmt.Errorf("refusing absolute path %q from remote tree", p)
	}
	if vol := filepath.VolumeName(filepath.FromSlash(p)); vol != "" {
		return fmt.Errorf("refusing path %q with a volume name from remote tree", p)
	}

	clean := path.Clean(p)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("refusing path %q escaping the store root", p)
	}

	for _, part := range strings.Split(clean, "/") {
		if fs.IsVCSDirName(part) {
			return fmt.Errorf("refusing path %q inside a VCS metadata directory", p)
		}
	}

	return nil
}

// unionPaths returns the sorted union of the keys of both maps.
func unionPaths(a, b map[string]changeInfo) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	res := make([]string, 0, len(a)+len(b))
	for p := range a {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		res = append(res, p)
	}
	for p := range b {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		res = append(res, p)
	}
	sort.Strings(res)

	return res
}
