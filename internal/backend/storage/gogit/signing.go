package gogit

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	gopassconfig "github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// This file isolates commit-signing policy. go-git can sign with an in-process
// OpenPGP key, but it does not use gpg-agent, pinentry or smartcards, and it
// cannot use a passphrase-protected agent key. gopass therefore shells out to
// the git CLI for signed commits (hybrid mode) when the config requests
// signing and a git binary is available, and otherwise creates an unsigned
// commit rather than failing the user's write.

// signingOption returns the effective value of a git config option, consulting
// the local scope first and falling back to the global scope. This honours a
// user-wide commit.gpgsign=true without mutating the repository's stored config
// (ConfigScoped merges the global scope into the local config in place).
func (g *GoGit) signingOption(section, option string) string {
	if g.repo == nil {
		return ""
	}

	if local, err := g.repo.Config(); err == nil && local.Raw != nil {
		if v := local.Raw.Section(section).Option(option); v != "" {
			return v
		}
	}

	if global, err := config.LoadConfig(config.GlobalScope); err == nil && global.Raw != nil {
		return global.Raw.Section(section).Option(option)
	}

	return ""
}

// shouldSign reports whether the config asks for signed commits. It consults
// the global scope as well, so a user-wide commit.gpgsign=true is honoured.
// git accepts yes/on/1 (case-insensitively) in addition to true.
func (g *GoGit) shouldSign() bool {
	return isGitBool(g.signingOption("commit", "gpgsign"))
}

// allowUnsignedCommits reports whether the user opted into silently creating an
// unsigned commit when a configured signature cannot be produced. Without it a
// signing failure is a hard error, so a team relying on signed history never
// loses it silently.
func (g *GoGit) allowUnsignedCommits(ctx context.Context) bool {
	return gopassconfig.Bool(ctx, "gogit.allow-unsigned-commits")
}

// isGitBool reports whether a git config value is truthy. git accepts true,
// yes, on and 1, case-insensitively.
func isGitBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "yes", "on", "1":
		return true
	default:
		return false
	}
}

// gpgFormat returns the configured gpg.format ("openpgp" or "ssh").
func (g *GoGit) gpgFormat() string {
	return g.signingOption("gpg", "format")
}

// commitSignedCLI stages the already-added changes and creates a signed commit
// by shelling out to the git CLI. It returns an error when git is unavailable or
// the commit fails, in which case the caller falls back to an unsigned commit.
func (g *GoGit) commitSignedCLI(ctx context.Context, msg string) error {
	return g.gitCommitCLI(ctx, msg)
}

// commitMergeSignedCLI creates a signed two-parent merge commit by shelling out
// to the git CLI. The remote parent is written to .git/MERGE_HEAD so the CLI
// records the same merge shape as the pure-Go path. (Passing `-p <hash>` to
// `git commit` would be parsed as `--patch`, not as a parent.)
func (g *GoGit) commitMergeSignedCLI(ctx context.Context, msg string, remoteHash plumbing.Hash) error {
	mergeHead := g.fs.Path() + "/.git/MERGE_HEAD"
	if err := os.WriteFile(mergeHead, []byte(remoteHash.String()+"\n"), 0o600); err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(mergeHead)
	}()

	return g.gitCommitCLI(ctx, msg)
}

// gitCommitCLI runs `git commit -S` in the store directory. Hooks are disabled
// with --no-verify: gopass documents that git hooks never run, and a hook
// rejection must not be mistaken for a signing failure (which would silently
// fall back to an unsigned commit and bypass the hook).
func (g *GoGit) gitCommitCLI(ctx context.Context, msg string) error {
	if _, err := exec.LookPath("git"); err != nil {
		debug.Log("git binary not available for signed commit")

		return err
	}

	if g.gpgFormat() == "ssh" {
		out.Warningf(ctx, "SSH commit signing (gpg.format=ssh) is not supported by the gogit backend; use the gitfs backend for signed commits.")

		return errUnsupportedSigning
	}

	args := []string{"commit", "-S", "--no-verify", "-m", msg}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.fs.Path()
	// Honour the context's commit timestamp the same way the unsigned path
	// does, so a signed commit does not silently use the wall clock.
	if ctxutil.HasCommitTimestamp(ctx) {
		ts := strconv.FormatInt(ctxutil.GetCommitTimestamp(ctx).UTC().Unix(), 10)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_DATE="+ts,
			"GIT_COMMITTER_DATE="+ts,
		)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		debug.Log("git commit -S failed: %s: %s", err, string(out))

		return err
	}

	return nil
}

type signingError string

func (e signingError) Error() string { return string(e) }

const errUnsupportedSigning = signingError("ssh commit signing unsupported")
