package sub

import (
	"context"
	"fmt"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend"
	gitcli "github.com/justwatchcom/gopass/backend/sync/git/cli"
	"github.com/justwatchcom/gopass/backend/sync/git/gogit"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// Sync returns the sync backend
func (s *Store) Sync() backend.Sync {
	return s.sync
}

// GitInit initializes the the git repo in the store
func (s *Store) GitInit(ctx context.Context, un, ue string) error {
	switch backend.GetSyncBackend(ctx) {
	case backend.GoGit:
		out.Cyan(ctx, "WARNING: Using experimental sync backend 'go-git'")
		git, err := gogit.Init(ctx, s.path)
		if err != nil {
			return errors.Wrapf(err, "failed to init git: %s", err)
		}
		s.sync = git
		return nil
	case backend.GitCLI:
		git, err := gitcli.Init(ctx, s.path, un, ue)
		if err != nil {
			return errors.Wrapf(err, "failed to init git: %s", err)
		}
		s.sync = git
		return nil
	case backend.GitMock:
		out.Cyan(ctx, "WARNING: Initializing with no-op (mock) git backend")
		return nil
	default:
		return fmt.Errorf("Unknown Sync Backend: %d", backend.GetSyncBackend(ctx))
	}
}

// GitInitConfig (re-)intializes the git config in an existing repo
func (s *Store) GitInitConfig(ctx context.Context, un, ue string) error {
	return s.sync.InitConfig(ctx, un, ue)
}

// GitVersion returns the git version
func (s *Store) GitVersion(ctx context.Context) semver.Version {
	return s.sync.Version(ctx)
}

// GitAddRemote adds a new remote
func (s *Store) GitAddRemote(ctx context.Context, remote, url string) error {
	return s.sync.AddRemote(ctx, remote, url)
}

// GitPull performs a git pull
func (s *Store) GitPull(ctx context.Context, origin, branch string) error {
	return s.sync.Pull(ctx, origin, branch)
}

// GitPush performs a git push
func (s *Store) GitPush(ctx context.Context, origin, branch string) error {
	return s.sync.Push(ctx, origin, branch)
}
