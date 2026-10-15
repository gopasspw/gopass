package gogit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// worktree returns the repository worktree.
func (g *GoGit) worktree() (*git.Worktree, error) {
	return g.repo.Worktree()
}

// rel returns the path relative to the store root, or "." for the root itself.
func (g *GoGit) rel(path string) string {
	if path == g.fs.Path() {
		return "."
	}

	path = strings.TrimPrefix(path, g.fs.Path()+string(filepath.Separator))
	path = strings.TrimPrefix(path, g.fs.Path()+"/")
	path = strings.TrimPrefix(path, "./")
	if path == "" || path == "." {
		return "."
	}

	return filepath.ToSlash(path)
}

// Add adds the listed files to the git index.
func (g *GoGit) Add(_ context.Context, files ...string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}

	wt, err := g.worktree()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return wt.AddWithOptions(&git.AddOptions{All: true})
	}

	for _, f := range files {
		rel := g.rel(f)
		if rel == "." {
			if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
				return err
			}

			continue
		}
		// go-git checks AddOptions.All before AddOptions.Path, so combining
		// them would stage the entire worktree. Use Path alone for explicit
		// paths so a single "Save secret" commit cannot sweep in unrelated
		// files.
		if err := wt.AddWithOptions(&git.AddOptions{Path: rel}); err != nil {
			return fmt.Errorf("failed to add %q: %w", rel, err)
		}
	}

	return nil
}

// TryAdd calls Add and returns nil if the git repo was not initialized.
func (g *GoGit) TryAdd(ctx context.Context, files ...string) error {
	err := g.Add(ctx, files...)
	if err == nil || errors.Is(err, store.ErrGitNotInit) {
		return nil
	}

	return err
}

// HasStagedChanges returns true if there are any staged changes which can be committed.
func (g *GoGit) HasStagedChanges(_ context.Context) bool {
	wt, err := g.worktree()
	if err != nil {
		return false
	}
	st, err := wt.StatusWithOptions(git.StatusOptions{Strategy: git.Preload})
	if err != nil {
		// If we cannot determine the status assume there is something to commit
		// so we do not silently drop changes.
		return true
	}

	for _, fs := range st {
		if fs.Staging != git.Unmodified && fs.Staging != git.Untracked {
			return true
		}
	}

	return false
}

// ListUntrackedFiles lists untracked files.
func (g *GoGit) ListUntrackedFiles(_ context.Context) []string {
	wt, err := g.worktree()
	if err != nil {
		return nil
	}
	st, err := wt.StatusWithOptions(git.StatusOptions{Strategy: git.Preload})
	if err != nil {
		return nil
	}

	uf := make([]string, 0, len(st))
	for path, fs := range st {
		if fs.Worktree == git.Untracked {
			uf = append(uf, path)
		}
	}
	sort.Strings(uf)

	return uf
}

// Commit creates a new git commit with the given commit message.
func (g *GoGit) Commit(ctx context.Context, msg string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}
	if !g.HasStagedChanges(ctx) {
		return store.ErrGitNothingToCommit
	}

	// Honour commit.gpgsign by shelling out to the git CLI for this commit
	// only. If that is not possible (no git binary) fall back to an unsigned
	// commit rather than failing the user's write, but only when the user
	// explicitly opted in via gogit.allow-unsigned-commits.
	if g.shouldSign() {
		if err := g.commitSignedCLI(ctx, msg); err == nil {
			return nil
		}
		if !g.allowUnsignedCommits(ctx) {
			return fmt.Errorf("failed to create a signed commit and gogit.allow-unsigned-commits is not set; refusing to create an unsigned commit")
		}
		out.Warningf(ctx, "Failed to create a signed commit, creating an unsigned commit instead.")
	}

	sig := g.commitSignature(ctx)
	wt, err := g.worktree()
	if err != nil {
		return err
	}

	_, err = wt.Commit(msg, &git.CommitOptions{
		Author:            sig,
		Committer:         sig,
		AllowEmptyCommits: false,
	})
	if errors.Is(err, git.ErrEmptyCommit) {
		return store.ErrGitNothingToCommit
	}

	return err
}

// TryCommit calls Commit and returns nil if there was nothing to commit or if
// the git repo was not initialized.
func (g *GoGit) TryCommit(ctx context.Context, msg string) error {
	err := g.Commit(ctx, msg)
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrGitNothingToCommit) || errors.Is(err, store.ErrGitNotInit) {
		debug.Log("Nothing to commit or git not initialized. Ignoring.")

		return nil
	}

	return err
}

// fetch fetches from the named remote. The remote URL is stripped of any
// userinfo before it is handed to go-git: go-git's HTTP transport embeds the
// request URL (userinfo included) in its errors, which would leak a password or
// token into the terminal and the debug log. Credentials are passed via Auth
// only.
func (g *GoGit) fetch(ctx context.Context, remote string) error {
	if err := g.repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: remote,
		RemoteURL:  stripUserInfo(g.remoteURL(remote)),
		Auth:       g.auth(ctx, remote),
		Force:      false,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	return nil
}

// PushPull pushes or pulls the repo from its remote.
func (g *GoGit) PushPull(ctx context.Context, op, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")

		return nil
	}
	if !g.IsInitialized() {
		debug.Log("Git in %s is not initialized. Can not push/pull", g.Path())

		return store.ErrGitNotInit
	}

	remote, branch = g.resolveRemoteBranch(remote, branch)

	if g.remoteURL(remote) == "" {
		debug.Log("No remote %q configured", remote)

		return store.ErrGitNoRemote
	}

	if op == "pull" {
		return g.pullRemote(ctx, remote, branch)
	}

	// Mirror gitfs: pull before pushing so a remote that is ahead does not make
	// every push fail with a non-fast-forward error. A failed pull is only a
	// warning; the push below reports the real error.
	if err := g.pullRemote(ctx, remote, branch); err != nil {
		out.Warningf(ctx, "Failed to pull before git push: %s", err)
	}

	if uf := g.ListUntrackedFiles(ctx); len(uf) > 0 {
		out.Warningf(ctx, "Found untracked files: %+v", uf)
	}

	return g.pushRemote(ctx, remote, branch)
}

// pushRemote pushes the local branch to the remote. The remote URL is stripped
// of userinfo for the same reason as in fetch.
func (g *GoGit) pushRemote(ctx context.Context, remote, branch string) error {
	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
	err := g.repo.PushContext(ctx, &git.PushOptions{
		RemoteName: remote,
		RemoteURL:  stripUserInfo(g.remoteURL(remote)),
		RefSpecs:   []config.RefSpec{refSpec},
		Auth:       g.auth(ctx, remote),
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	return nil
}

// Push pushes to the git remote.
func (g *GoGit) Push(ctx context.Context, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")

		return nil
	}

	return g.PushPull(ctx, "push", remote, branch)
}

// TryPush calls Push and returns nil if the git repo was not initialized.
func (g *GoGit) TryPush(ctx context.Context, remote, branch string) error {
	err := g.Push(ctx, remote, branch)
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrGitNotInit) || errors.Is(err, store.ErrGitNoRemote) {
		debug.Log("Git not initialized or no remote. Ignoring.")

		return nil
	}

	return err
}

// Pull pulls from the git remote.
func (g *GoGit) Pull(ctx context.Context, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")

		return nil
	}

	return g.PushPull(ctx, "pull", remote, branch)
}

// pullRemote fetches and integrates the remote branch. Fast-forwards when
// possible, otherwise performs a file-level merge (see merge.go).
func (g *GoGit) pullRemote(ctx context.Context, remote, branch string) error {
	if err := g.fetch(ctx, remote); err != nil {
		return err
	}

	remoteRef, err := g.repo.Reference(plumbing.NewRemoteReferenceName(remote, branch), true)
	if err != nil {
		// A missing remote branch is only benign when the remote is genuinely
		// empty. Otherwise (e.g. a typo in the branch name) report the error
		// instead of silently doing nothing.
		if g.remoteIsEmpty(ctx, remote) {
			debug.Log("remote %s is empty, nothing to pull", remote)

			return nil
		}

		return fmt.Errorf("remote branch %s/%s not found: %w", remote, branch, err)
	}
	remoteHash := remoteRef.Hash()

	head, err := g.repo.Head()
	if err != nil {
		// unborn HEAD, adopt remote
		return g.resetTo(remoteHash)
	}
	if head.Hash() == remoteHash {
		return nil
	}

	localCommit, err := g.repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}
	remoteCommit, err := g.repo.CommitObject(remoteHash)
	if err != nil {
		return err
	}

	ff, err := localCommit.IsAncestor(remoteCommit)
	if err != nil {
		return err
	}
	if ff {
		return g.resetTo(remoteHash)
	}

	// The remote is an ancestor of local (unpushed local commits): there is
	// nothing to integrate. Without this check the merge below would create a
	// redundant empty two-parent merge commit on every pull.
	if remoteIsAncestor, err := remoteCommit.IsAncestor(localCommit); err != nil {
		return err
	} else if remoteIsAncestor {
		return nil
	}

	return g.mergeRemote(ctx, remoteHash)
}

// remoteIsEmpty reports whether the named remote has no branches at all.
func (g *GoGit) remoteIsEmpty(ctx context.Context, remote string) bool {
	iter, err := g.repo.References()
	if err != nil {
		return false
	}
	defer iter.Close()

	prefix := "refs/remotes/" + remote + "/"
	empty := true
	_ = iter.ForEach(func(ref *plumbing.Reference) error {
		if strings.HasPrefix(ref.Name().String(), prefix) {
			empty = false

			return storer.ErrStop
		}

		return nil
	})

	return empty
}

// resetTo fast-forwards the worktree and HEAD to the given commit. It uses
// go-git's MergeReset (the mode PullContext uses) rather than HardReset so that
// untracked files are preserved and uncommitted edits to tracked files abort
// the reset instead of being silently discarded. An unborn HEAD is materialised
// first, which is what `gopass setup --remote` needs: Init stages the store
// without committing, then the initial pull runs against an unborn HEAD.
func (g *GoGit) resetTo(hash plumbing.Hash) error {
	// Validate the remote tree before it is materialised. The fast-forward path
	// resets the worktree straight to the remote commit without a diff, so the
	// check in mergeRemote does not cover it. go-git's reset reports a rejected
	// name only as an opaque ErrDirectoryNotFound, so the check has to happen
	// here to keep the offending path in the message.
	if err := g.validateCommitTree(hash); err != nil {
		return err
	}

	wt, err := g.worktree()
	if err != nil {
		return err
	}

	if _, err := g.repo.Head(); err != nil {
		// Unborn HEAD: MergeReset would drop index entries that are not in the
		// remote tree (e.g. the files Init staged for `gopass setup --remote`).
		// Snapshot the index and re-add those entries after the reset so the
		// staged content survives and can be committed on top.
		staged, err := g.stagedEntries()
		if err != nil {
			return err
		}

		if err := g.setUnbornHead(hash); err != nil {
			return err
		}
		if err := wt.Reset(&git.ResetOptions{Commit: hash, Mode: git.MergeReset}); err != nil {
			return err
		}

		return g.restoreStagedEntries(staged)
	}

	return wt.Reset(&git.ResetOptions{Commit: hash, Mode: git.MergeReset})
}

// stagedEntries returns the current index entries keyed by path.
func (g *GoGit) stagedEntries() (map[string]*index.Entry, error) {
	idx, err := g.repo.Storer.Index()
	if err != nil {
		return nil, err
	}

	entries := make(map[string]*index.Entry, len(idx.Entries))
	for _, e := range idx.Entries {
		entries[e.Name] = e
	}

	return entries, nil
}

// restoreStagedEntries re-adds the given index entries and materialises their
// worktree files, keeping any entry that the reset already produced (the remote
// tree wins for paths present in both). The reset deletes worktree files whose
// index entries it removed, so the content is restored from the object store.
func (g *GoGit) restoreStagedEntries(staged map[string]*index.Entry) error {
	if len(staged) == 0 {
		return nil
	}

	idx, err := g.repo.Storer.Index()
	if err != nil {
		return err
	}

	for name, e := range staged {
		if _, err := idx.Entry(name); err == nil {
			continue
		}

		if err := g.materializeIndexEntry(e); err != nil {
			return err
		}
		idx.Entries = append(idx.Entries, e)
	}

	return g.repo.Storer.SetIndex(idx)
}

// materializeIndexEntry writes the blob referenced by an index entry back into
// the worktree, recreating symlinks.
func (g *GoGit) materializeIndexEntry(e *index.Entry) error {
	blob, err := g.repo.BlobObject(e.Hash)
	if err != nil {
		return err
	}
	r, err := blob.Reader()
	if err != nil {
		return err
	}
	defer func() {
		_ = r.Close()
	}()

	content, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if e.Mode == filemode.Symlink {
		return g.fs.SetSymlink(context.Background(), e.Name, string(content))
	}

	return g.fs.Set(context.Background(), e.Name, content)
}

// setUnbornHead points the branch HEAD refers to at the given commit, turning
// an unborn HEAD into a born one.
func (g *GoGit) setUnbornHead(hash plumbing.Hash) error {
	head, err := g.repo.Reference(plumbing.HEAD, false)
	if err != nil {
		return err
	}

	target := head.Target()
	if target == "" {
		target = plumbing.NewBranchReferenceName(g.defaultBranch())
	}

	return g.repo.Storer.SetReference(plumbing.NewHashReference(target, hash))
}

// AddRemote adds a new remote.
func (g *GoGit) AddRemote(_ context.Context, remote, url string) error {
	_, err := g.repo.CreateRemote(&config.RemoteConfig{Name: remote, URLs: []string{url}})

	return err
}

// RemoveRemote removes a remote.
func (g *GoGit) RemoveRemote(_ context.Context, remote string) error {
	return g.repo.DeleteRemote(remote)
}

// Status returns a deterministic, git-status-like summary.
func (g *GoGit) Status(_ context.Context) ([]byte, error) {
	if !g.IsInitialized() {
		return nil, store.ErrGitNotInit
	}

	wt, err := g.worktree()
	if err != nil {
		return nil, err
	}
	st, err := wt.StatusWithOptions(git.StatusOptions{Strategy: git.Preload})
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "# on branch %s\n", g.defaultBranch())

	staged := []string{}
	unstaged := []string{}
	untracked := []string{}

	for path, fs := range st {
		switch {
		case fs.Staging != git.Unmodified && fs.Staging != git.Untracked:
			staged = append(staged, fmt.Sprintf("#   %s:   %s", statusVerb(fs.Staging), path))
		case fs.Worktree == git.Untracked:
			untracked = append(untracked, fmt.Sprintf("#   %s", path))
		case fs.Worktree != git.Unmodified:
			unstaged = append(unstaged, fmt.Sprintf("#   %s:   %s", statusVerb(fs.Worktree), path))
		}
	}

	if len(staged) > 0 {
		buf.WriteString("# changes to be committed:\n")
		sort.Strings(staged)
		buf.WriteString(strings.Join(staged, "\n") + "\n")
	}
	if len(unstaged) > 0 {
		buf.WriteString("# changes not staged for commit:\n")
		sort.Strings(unstaged)
		buf.WriteString(strings.Join(unstaged, "\n") + "\n")
	}
	if len(untracked) > 0 {
		buf.WriteString("# untracked files:\n")
		sort.Strings(untracked)
		buf.WriteString(strings.Join(untracked, "\n") + "\n")
	}

	return buf.Bytes(), nil
}

func statusVerb(code git.StatusCode) string {
	switch code {
	case git.Added:
		return "new file"
	case git.Deleted:
		return "deleted"
	case git.Renamed:
		return "renamed"
	case git.Copied:
		return "copied"
	case git.Unmodified, git.Untracked, git.Modified, git.UpdatedButUnmerged:
		return "modified"
	default:
		return "modified"
	}
}

// Compact runs a best-effort repack and prune. The prune handler deletes
// unreachable loose objects (a handler returning nil would make the prune a
// no-op).
func (g *GoGit) Compact(ctx context.Context) error {
	if err := g.repo.RepackObjects(&git.RepackConfig{}); err != nil {
		debug.Log("Failed to repack objects: %s", err)

		return nil
	}

	handler := func(plumbing.Hash) error { return nil }
	if d, ok := g.repo.Storer.(looseObjectDeleter); ok {
		handler = d.DeleteLooseObject
	}
	if err := g.repo.Prune(git.PruneOptions{Handler: handler}); err != nil {
		debug.Log("Failed to prune objects: %s", err)
	}

	return nil
}

// looseObjectDeleter is implemented by go-git's filesystem object storage. It
// is used to actually delete unreachable loose objects during a prune.
type looseObjectDeleter interface {
	DeleteLooseObject(plumbing.Hash) error
}
