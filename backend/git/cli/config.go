package cli

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/out"
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

	return g.fixConfigOSDep(ctx)
}

// InitConfig initialized and preparse the git config
func (g *Git) InitConfig(ctx context.Context, signKey, userName, userEmail string) error {
	// set commit identity
	if err := g.ConfigSet(ctx, "user.name", userName); err != nil {
		return errors.Wrapf(err, "failed to set git config user.name")
	}
	if err := g.ConfigSet(ctx, "user.email", userEmail); err != nil {
		return errors.Wrapf(err, "failed to set git config user.email")
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

	// set GPG signkey
	if err := g.SetSignKey(ctx, signKey); err != nil {
		color.Yellow("Failed to configure Git GPG Commit signing: %s\n", err)
	}

	return nil
}

// SetSignKey configures git to use the given sign key
func (g *Git) SetSignKey(ctx context.Context, sk string) error {
	if sk == "" {
		return errors.Errorf("SignKey not set")
	}

	if err := g.ConfigSet(ctx, "user.signingkey", sk); err != nil {
		return errors.Wrapf(err, "failed to set git sign key")
	}

	return g.ConfigSet(ctx, "commit.gpgsign", "true")
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

	buf := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, "git", "config", "--get", key)
	cmd.Dir = g.path
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "store.gitConfigValue: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
