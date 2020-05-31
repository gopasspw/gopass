package cli

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"

	"github.com/pkg/errors"
)

const (
	fileMode = 0600
)

func (g *Git) fixConfig(ctx context.Context) error {
	// set push default, to avoid issues with
	// "fatal: The current branch master has multiple upstream branches, refusing to push"
	// https://stackoverflow.com/questions/948354/default-behavior-of-git-push-without-a-branch-specified
	if err := g.ConfigSet(ctx, "push.default", "matching"); err != nil {
		return errors.Wrapf(err, "failed to set git config for push.default")
	}

	// setup for proper diffs
	if err := g.ConfigSet(ctx, "diff.gpg.binary", "true"); err != nil {
		out.Yellow(ctx, "Error while initializing git: %s", err)
	}
	if err := g.ConfigSet(ctx, "diff.gpg.textconv", "gpg --no-tty --decrypt"); err != nil {
		out.Yellow(ctx, "Error while initializing git: %s", err)
	}

	return nil
}

// InitConfig initialized and preparse the git config
func (g *Git) InitConfig(ctx context.Context, userName, userEmail string) error {
	// set commit identity
	if userName != "" {
		if err := g.ConfigSet(ctx, "user.name", userName); err != nil {
			return errors.Wrapf(err, "failed to set git config user.name")
		}
	} else {
		out.Red(ctx, "Git Username not set")
	}
	if userEmail != "" && strings.Contains(userEmail, "@") {
		if err := g.ConfigSet(ctx, "user.email", userEmail); err != nil {
			return errors.Wrapf(err, "failed to set git config user.email")
		}
	} else {
		out.Red(ctx, "Git Email not set")
	}

	// ensure sane git config
	if err := g.fixConfig(ctx); err != nil {
		return errors.Wrapf(err, "failed to fix git config")
	}

	if err := ioutil.WriteFile(filepath.Join(g.path, ".gitattributes"), []byte("*.gpg diff=gpg\n"), fileMode); err != nil {
		return errors.Errorf("Failed to initialize git: %s", err)
	}
	if err := g.Add(ctx, g.path+"/.gitattributes"); err != nil {
		out.Yellow(ctx, "Warning: Failed to add .gitattributes to git")
	}
	if err := g.Commit(ctx, "Configure git repository for gpg file diff."); err != nil {
		out.Yellow(ctx, "Warning: Failed to commit .gitattributes to git")
	}

	return nil
}

// ConfigSet sets a local config value
func (g *Git) ConfigSet(ctx context.Context, key, value string) error {
	return g.Cmd(ctx, "gitConfigSet", "config", "--local", key, value)
}

// ConfigGet returns a given config value
func (g *Git) ConfigGet(ctx context.Context, key string) (string, error) {
	if !g.IsInitialized() {
		return "", store.ErrGitNotInit
	}

	buf := &strings.Builder{}

	cmd := exec.CommandContext(ctx, "git", "config", "--get", key)
	cmd.Dir = g.path
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// ConfigList returns all git config settings
func (g *Git) ConfigList(ctx context.Context) (map[string]string, error) {
	if !g.IsInitialized() {
		return nil, store.ErrGitNotInit
	}

	buf := &strings.Builder{}

	cmd := exec.CommandContext(ctx, "git", "config", "--list")
	cmd.Dir = g.path
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(buf.String(), "\n")
	kv := make(map[string]string, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		p := strings.SplitN(line, "=", 2)
		if len(p) < 2 {
			continue
		}
		kv[p[0]] = p[1]
	}
	return kv, nil
}
