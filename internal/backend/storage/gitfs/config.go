package gitfs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	fileMode = 0o600
)

// fixConfig sets up the git config for the password store in a way to simplifies some of the quirks
// that git has. We'd prefer if that wasn't necessary but git has way too many modes of operation
// and we need it to behave a predicatable as possible.
func (g *Git) fixConfig(ctx context.Context) error {
	// set push default, to avoid issues with
	// "fatal: The current branch master has multiple upstream branches, refusing to push"
	// https://stackoverflow.com/questions/948354/default-behavior-of-git-push-without-a-branch-specified.
	if err := g.ConfigSet(ctx, "push.default", "matching"); err != nil {
		return fmt.Errorf("failed to set git config for push.default: %w", err)
	}

	if err := g.ConfigSet(ctx, "pull.rebase", "false"); err != nil {
		return fmt.Errorf("failed to set git config for pull.rebase: %w", err)
	}

	// setup for proper diffs.
	if err := g.ConfigSet(ctx, "diff.gpg.binary", "true"); err != nil {
		out.Errorf(ctx, "Error while initializing git: %s", err)
	}
	if err := g.ConfigSet(ctx, "diff.gpg.textconv", "gpg --no-tty --decrypt"); err != nil {
		out.Errorf(ctx, "Error while initializing git: %s", err)
	}

	// setup for persistent SSH connections.
	if sc := gitSSHCommand(); sc != "" {
		if err := g.ConfigSet(ctx, "core.sshCommand", sc); err != nil {
			out.Errorf(ctx, "Error while configuring persistent SSH connections: %s", err)
		}
	}

	return nil
}

// InitConfig initialized and preparse the git config.
func (g *Git) InitConfig(ctx context.Context, userName, userEmail string) error {
	// set commit identity.
	if userName != "" {
		if err := g.ConfigSet(ctx, "user.name", userName); err != nil {
			return fmt.Errorf("failed to set git config user.name: %w", err)
		}
	} else {
		out.Printf(ctx, "Git Username not set")
	}
	if userEmail != "" && strings.Contains(userEmail, "@") {
		if err := g.ConfigSet(ctx, "user.email", userEmail); err != nil {
			return fmt.Errorf("failed to set git config user.email: %w", err)
		}
	} else {
		out.Printf(ctx, "Git Email not set")
	}

	// ensure sane git config.
	if err := g.fixConfig(ctx); err != nil {
		return fmt.Errorf("failed to fix git config: %w", err)
	}

	if err := os.WriteFile(filepath.Join(g.fs.Path(), ".gitattributes"), []byte("*.gpg diff=gpg\n"), fileMode); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}
	if err := g.Add(ctx, g.fs.Path()+"/.gitattributes"); err != nil {
		out.Warningf(ctx, "Failed to add .gitattributes to git")
	}
	if err := g.Commit(ctx, "Configure git repository for gpg file diff."); err != nil {
		out.Warningf(ctx, "Failed to commit .gitattributes to git")
	}

	return nil
}

// ConfigSet sets a local config value.
func (g *Git) ConfigSet(ctx context.Context, key, value string) error {
	return g.Cmd(ctx, "gitConfigSet", "config", "--local", key, value)
}

// ConfigGet returns a given config value.
func (g *Git) ConfigGet(ctx context.Context, key string) (string, error) {
	if !g.IsInitialized() {
		return "", store.ErrGitNotInit
	}

	buf := &strings.Builder{}

	cmd := exec.CommandContext(ctx, "git", "config", "--get", key)
	cmd.Dir = g.fs.Path()
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// ConfigList returns all git config settings.
func (g *Git) ConfigList(ctx context.Context) (map[string]string, error) {
	if !g.IsInitialized() {
		return nil, store.ErrGitNotInit
	}

	buf := &strings.Builder{}

	cmd := exec.CommandContext(ctx, "git", "config", "--list")
	cmd.Dir = g.fs.Path()
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(buf.String(), "\n")
	kv := make(map[string]string, len(lines))
	for _, line := range lines {
		key, val, found := strings.Cut(strings.TrimSpace(line), "=")
		if !found {
			continue
		}
		kv[key] = val
	}

	return kv, nil
}
