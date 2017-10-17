package sub

import (
	"context"

	"github.com/blang/semver"
)

type giter interface {
	Add(context.Context, ...string) error
	Commit(context.Context, string) error
	Push(context.Context, string, string) error
	Init(context.Context, string, string, string) error
	InitConfig(context.Context, string, string, string) error
	Version(context.Context) semver.Version
	Cmd(context.Context, string, ...string) error
}

// GitInit initializes the the git repo in the store
func (s *Store) GitInit(ctx context.Context, sk, un, ue string) error {
	return s.git.Init(ctx, sk, un, ue)
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
func (s *Store) Git(ctx context.Context, args ...string) error {
	return s.git.Cmd(ctx, "Git", args...)
}

// GitPush performs a git push
func (s *Store) GitPush(ctx context.Context, origin, branch string) error {
	return s.git.Push(ctx, origin, branch)
}
