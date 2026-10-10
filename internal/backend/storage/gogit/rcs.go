package gogit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
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
		if err := wt.AddWithOptions(&git.AddOptions{All: true, Path: rel}); err != nil {
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
		if fs.Staging != git.Unmodified {
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
	// commit rather than failing the user's write.
	if g.shouldSign() {
		if err := g.commitSignedCLI(ctx, msg); err == nil {
			return nil
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

// fetch fetches from the named remote.
func (g *GoGit) fetch(ctx context.Context, remote string) error {
	if err := g.repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: remote,
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

	if err := g.fetch(ctx, remote); err != nil {
		out.Warningf(ctx, "Failed to pull before git push: %s", err)
	}

	if uf := g.ListUntrackedFiles(ctx); len(uf) > 0 {
		out.Warningf(ctx, "Found untracked files: %+v", uf)
	}

	return g.pushRemote(ctx, remote, branch)
}

// pushRemote pushes the local branch to the remote.
func (g *GoGit) pushRemote(ctx context.Context, remote, branch string) error {
	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
	err := g.repo.PushContext(ctx, &git.PushOptions{
		RemoteName: remote,
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
		// Nothing fetched (e.g. empty remote). Nothing to do.
		debug.Log("remote ref %s/%s not found: %s", remote, branch, err)

		return nil
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

	if ff, err := localCommit.IsAncestor(remoteCommit); err == nil && ff {
		return g.resetTo(remoteHash)
	}

	return g.mergeRemote(ctx, remoteHash)
}

// resetTo hard-resets the working tree and HEAD to the given commit.
func (g *GoGit) resetTo(hash plumbing.Hash) error {
	wt, err := g.worktree()
	if err != nil {
		return err
	}

	return wt.Reset(&git.ResetOptions{Commit: hash, Mode: git.HardReset})
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
		case fs.Staging != git.Unmodified:
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

// Compact runs a best-effort repack and prune.
func (g *GoGit) Compact(ctx context.Context) error {
	if err := g.repo.RepackObjects(&git.RepackConfig{}); err != nil {
		debug.Log("Failed to repack objects: %s", err)

		return nil
	}
	if err := g.repo.Prune(git.PruneOptions{Handler: func(plumbing.Hash) error { return nil }}); err != nil {
		debug.Log("Failed to prune objects: %s", err)
	}

	return nil
}
