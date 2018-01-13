package sub

import (
	"context"
	"os"

	"github.com/blang/semver"
	gitcli "github.com/justwatchcom/gopass/backend/sync/git/cli"
	"github.com/justwatchcom/gopass/backend/sync/git/gogit"
	"github.com/pkg/errors"
)

type giter interface {
	Add(context.Context, ...string) error
	AddRemote(context.Context, string, string) error
	Cmd(context.Context, string, ...string) error
	Commit(context.Context, string) error
	InitConfig(context.Context, string, string, string) error
	Pull(context.Context, string, string) error
	Push(context.Context, string, string) error
	Version(context.Context) semver.Version
}

// GitInit initializes the the git repo in the store
func (s *Store) GitInit(ctx context.Context, sk, un, ue string) error {
	if gg := os.Getenv("GOPASS_EXPERIMENTAL_GOGIT"); gg != "" {
		git, err := gogit.Init(ctx, s.path)
		if err != nil {
			return errors.Wrapf(err, "failed to init git: %s", err)
		}
		s.git = git
		return nil
	}

	git, err := gitcli.Init(ctx, s.path, s.gpg.Binary(), sk, un, ue)
	if err != nil {
		return errors.Wrapf(err, "failed to init git: %s", err)
	}
	s.git = git
	return nil
}

// GitInitConfig (re-)intializes the git config in an existing repo
func (s *Store) GitInitConfig(ctx context.Context, sk, un, ue string) error {
	return s.git.InitConfig(ctx, sk, un, ue)
}

// GitVersion returns the git version
func (s *Store) GitVersion(ctx context.Context) semver.Version {
	return s.git.Version(ctx)
}

// Git channels any git subcommand to git in the store
// TODO remove this command, doesn't work with gogit
func (s *Store) Git(ctx context.Context, args ...string) error {
	return s.git.Cmd(ctx, "Git", args...)
}

// GitAddRemote adds a new remote
func (s *Store) GitAddRemote(ctx context.Context, remote, url string) error {
	return s.git.AddRemote(ctx, remote, url)
}

// GitPull performs a git pull
func (s *Store) GitPull(ctx context.Context, origin, branch string) error {
	return s.git.Pull(ctx, origin, branch)
}

// GitPush performs a git push
func (s *Store) GitPush(ctx context.Context, origin, branch string) error {
	return s.git.Push(ctx, origin, branch)
}
