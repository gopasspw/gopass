package gogit

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/gopasspw/gopass/internal/out"
)

// changeInfo describes the state of a single path in a tree diff.
type changeInfo struct {
	hash    plumbing.Hash
	mode    filemode.FileMode
	deleted bool
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
// The result is a two-parent merge commit so that a later user of the real git
// CLI sees valid, connected history.
func (g *GoGit) mergeRemote(ctx context.Context, remoteHash plumbing.Hash) error {
	head, err := g.repo.Head()
	if err != nil {
		return err
	}
	localHash := head.Hash()

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

	// 2. per-side changes vs base
	localChanges, err := treeChanges(baseTree, localTree)
	if err != nil {
		return err
	}
	remoteChanges, err := treeChanges(baseTree, remoteTree)
	if err != nil {
		return err
	}

	// 3. per-path resolution
	conflicts := []string{}
	for _, p := range unionPaths(localChanges, remoteChanges) {
		conflict, err := g.resolveMergePath(ctx, p, localChanges[p], remoteChanges[p], remoteTree, remoteHash)
		if err != nil {
			return err
		}
		if conflict {
			conflicts = append(conflicts, p)
		}
	}

	if len(conflicts) > 0 {
		out.Warningf(ctx, "Merge conflict in %d file(s); the remote versions were kept as *.conflict-<sha>:", len(conflicts))
		for _, p := range conflicts {
			out.Warningf(ctx, "  - %s", p)
		}
	}

	// 4. two-parent merge commit so real git users see valid history
	msg := fmt.Sprintf("Merge remote-tracking branch into %s (gopass go-git backend)", head.Name().Short())
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

// resolveMergePath resolves a single path that changed on at least one side. It
// returns true when the path is a true conflict (local kept, remote written to
// a *.conflict-<sha> file).
func (g *GoGit) resolveMergePath(ctx context.Context, p string, lc, rc changeInfo, remoteTree *object.Tree, remoteHash plumbing.Hash) (bool, error) {
	lok := lc.hash != plumbing.ZeroHash || lc.deleted
	rok := rc.hash != plumbing.ZeroHash || rc.deleted

	switch {
	case lok && !rok:
		// changed only locally -> keep local
		if lc.deleted {
			// local deletion: ensure the index reflects it.
			return false, g.stageDelete(p)
		}

		return false, nil
	case rok && !lok:
		// changed only remotely -> apply remote content
		return false, g.applyRemoteChange(ctx, p, rc, remoteTree)
	case lc.hash == rc.hash && lc.deleted == rc.deleted:
		// changed identically on both sides -> no-op
		return false, nil
	case lc.deleted != rc.deleted:
		// delete vs modify -> keep the modification (never drop a secret)
		if lc.deleted {
			// local deleted, remote modified: restore the remote version.
			return false, g.applyRemoteChange(ctx, p, rc, remoteTree)
		}

		// local modified, remote deleted: keep local; ensure it is staged.
		return false, g.stagePath(p)
	default:
		// changed on both sides, different content -> conflict
		return true, g.materializeConflict(ctx, p, remoteTree, remoteHash)
	}
}

// applyRemoteChange writes the remote version of path into the worktree and
// stages it (or stages its deletion).
func (g *GoGit) applyRemoteChange(ctx context.Context, p string, rc changeInfo, remoteTree *object.Tree) error {
	if rc.deleted {
		if err := g.fs.Delete(ctx, p); err != nil {
			return err
		}

		return g.stageDelete(p)
	}

	content, err := blobContent(remoteTree, p)
	if err != nil {
		return err
	}
	if err := g.fs.Set(ctx, p, content); err != nil {
		return err
	}

	return g.stagePath(p)
}

// materializeConflict keeps the local file untouched and writes the remote blob
// to <name>.conflict-<shortsha>, then stages the new file.
func (g *GoGit) materializeConflict(ctx context.Context, p string, remoteTree *object.Tree, remoteHash plumbing.Hash) error {
	content, err := blobContent(remoteTree, p)
	if err != nil {
		return err
	}

	conflictPath := conflictName(p, remoteHash)
	if err := g.fs.Set(ctx, conflictPath, content); err != nil {
		return err
	}

	return g.stagePath(conflictPath)
}

// stagePath adds the given path (or its deletion) to the index.
func (g *GoGit) stagePath(p string) error {
	wt, err := g.worktree()
	if err != nil {
		return err
	}

	return wt.AddWithOptions(&git.AddOptions{All: true, Path: p})
}

// stageDelete removes the given path from the index.
func (g *GoGit) stageDelete(p string) error {
	return g.stagePath(p)
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
func conflictName(p string, remoteHash plumbing.Hash) string {
	short := remoteHash.String()
	if len(short) > 7 {
		short = short[:7]
	}

	dir, file := path.Split(p)
	ext := path.Ext(file)
	base := strings.TrimSuffix(file, ext)
	if ext == "" {
		return dir + base + ".conflict-" + short
	}

	return dir + base + ".conflict-" + short + ext
}

// unionPaths returns the sorted union of the keys of both maps.
func unionPaths(a, b map[string]changeInfo) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for p := range a {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	for p := range b {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	sort.Strings(out)

	return out
}
