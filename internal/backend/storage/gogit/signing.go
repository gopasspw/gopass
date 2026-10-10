package gogit

import (
	"context"
	"os/exec"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
)

// This file isolates commit-signing policy. go-git can sign with an in-process
// OpenPGP key, but it does not use gpg-agent, pinentry or smartcards, and it
// cannot use a passphrase-protected agent key. gopass therefore shells out to
// the git CLI for signed commits (hybrid mode) when the local config requests
// signing and a git binary is available, and otherwise creates an unsigned
// commit rather than failing the user's write.

// shouldSign reports whether the local config asks for signed commits.
func (g *GoGit) shouldSign() bool {
	if g.repo == nil {
		return false
	}
	cfg, err := g.repo.Config()
	if err != nil || cfg.Raw == nil {
		return false
	}

	section := cfg.Raw.Section("commit")
	if section == nil {
		return false
	}
	if section.Option("gpgsign") == "true" {
		return true
	}

	return false
}

// gpgFormat returns the configured gpg.format ("openpgp" or "ssh").
func (g *GoGit) gpgFormat() string {
	if g.repo == nil {
		return ""
	}
	cfg, err := g.repo.Config()
	if err != nil || cfg.Raw == nil {
		return ""
	}

	return cfg.Raw.Section("gpg").Option("format")
}

// commitSignedCLI stages the already-added changes and creates a signed commit
// by shelling out to the git CLI. It returns an error when git is unavailable or
// the commit fails, in which case the caller falls back to an unsigned commit.
func (g *GoGit) commitSignedCLI(ctx context.Context, msg string) error {
	if _, err := exec.LookPath("git"); err != nil {
		debug.Log("git binary not available for signed commit")

		return err
	}

	if g.gpgFormat() == "ssh" {
		out.Warningf(ctx, "SSH commit signing (gpg.format=ssh) is not supported by the gogit backend; use the gitfs backend for signed commits.")

		return errUnsupportedSigning
	}

	cmd := exec.CommandContext(ctx, "git", "commit", "-S", "-m", msg)
	cmd.Dir = g.fs.Path()
	if out, err := cmd.CombinedOutput(); err != nil {
		debug.Log("git commit -S failed: %s: %s", err, string(out))

		return err
	}

	return nil
}

type signingError string

func (e signingError) Error() string { return string(e) }

const errUnsupportedSigning = signingError("ssh commit signing unsupported")
