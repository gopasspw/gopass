package sub

import (
	"context"
	"fmt"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend"
	gitcli "github.com/justwatchcom/gopass/backend/rcs/git/cli"
	"github.com/justwatchcom/gopass/backend/rcs/git/gogit"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// RCS returns the sync backend
func (s *Store) RCS() backend.RCS {
	return s.rcs
}

// GitInit initializes the the git repo in the store
func (s *Store) GitInit(ctx context.Context, un, ue string) error {
	switch backend.GetRCSBackend(ctx) {
	case backend.GoGit:
		out.Cyan(ctx, "WARNING: Using experimental sync backend 'go-git'")
		git, err := gogit.Init(ctx, s.url.Path)
		if err != nil {
			return errors.Wrapf(err, "failed to init git: %s", err)
		}
		s.rcs = git
		return nil
	case backend.GitCLI:
		git, err := gitcli.Init(ctx, s.url.Path, un, ue)
		if err != nil {
			return errors.Wrapf(err, "failed to init git: %s", err)
		}
		s.rcs = git
		return nil
	case backend.Noop:
		out.Cyan(ctx, "WARNING: Initializing with no-op (mock) git backend")
		return nil
	default:
		return fmt.Errorf("Unknown Sync Backend: %d", backend.GetRCSBackend(ctx))
	}
}

// GitInitConfig (re-)intializes the git config in an existing repo
func (s *Store) GitInitConfig(ctx context.Context, un, ue string) error {
	return s.rcs.InitConfig(ctx, un, ue)
}

// GitVersion returns the git version
func (s *Store) GitVersion(ctx context.Context) semver.Version {
	return s.rcs.Version(ctx)
}

// GitAddRemote adds a new remote
func (s *Store) GitAddRemote(ctx context.Context, remote, url string) error {
	return s.rcs.AddRemote(ctx, remote, url)
}

// GitPull performs a git pull
func (s *Store) GitPull(ctx context.Context, origin, branch string) error {
	return s.rcs.Pull(ctx, origin, branch)
}

// GitPush performs a git push
func (s *Store) GitPush(ctx context.Context, origin, branch string) error {
	return s.rcs.Push(ctx, origin, branch)
}

// ListRevisions will list all revisions for a secret
func (s *Store) ListRevisions(ctx context.Context, name string) ([]backend.Revision, error) {
	p := s.passfile(name)
	return s.rcs.Revisions(ctx, p)
}

// GetRevision will retrieve a single revision from the backend
func (s *Store) GetRevision(ctx context.Context, name, revision string) (store.Secret, error) {
	p := s.passfile(name)
	ciphertext, err := s.rcs.GetRevision(ctx, p, revision)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ciphertext of '%s'@'%s'", name, revision)
	}

	content, err := s.crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		out.Debug(ctx, "Decryption failed: %s", err)
		return nil, store.ErrDecrypt
	}

	sec, err := secret.Parse(content)
	if err != nil {
		out.Debug(ctx, "Failed to parse YAML: %s", err)
	}
	return sec, nil
}
