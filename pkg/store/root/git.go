package root

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/store"

	"github.com/blang/semver"
)

// RCS returns the sync backend
func (r *Store) RCS(ctx context.Context, name string) backend.RCS {
	_, sub, _ := r.getStore(ctx, name)
	if sub == nil || !sub.Valid() {
		return nil
	}
	return sub.RCS()
}

// GitInit initializes the git repo
func (r *Store) GitInit(ctx context.Context, name, userName, userEmail string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.GitInit(ctx, userName, userEmail)
}

// GitInitConfig initializes the git repos local config
func (r *Store) GitInitConfig(ctx context.Context, name, userName, userEmail string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.RCS().InitConfig(ctx, userName, userEmail)
}

// GitVersion returns git version information
func (r *Store) GitVersion(ctx context.Context) semver.Version {
	return r.store.RCS().Version(ctx)
}

// GitAddRemote adds a git remote
func (r *Store) GitAddRemote(ctx context.Context, name, remote, url string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.RCS().AddRemote(ctx, remote, url)
}

// GitRemoveRemote removes a git remote
func (r *Store) GitRemoveRemote(ctx context.Context, name, remote string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.RCS().RemoveRemote(ctx, remote)
}

// GitPull performs a git pull
func (r *Store) GitPull(ctx context.Context, name, origin, remote string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.RCS().Pull(ctx, origin, remote)
}

// GitPush performs a git push
func (r *Store) GitPush(ctx context.Context, name, origin, remote string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.RCS().Push(ctx, origin, remote)
}

// ListRevisions will list all revisions for the named entity
func (r *Store) ListRevisions(ctx context.Context, name string) ([]backend.Revision, error) {
	ctx, store, name := r.getStore(ctx, name)
	return store.ListRevisions(ctx, name)
}

// GetRevision will try to retrieve the given revision from the sync backend
func (r *Store) GetRevision(ctx context.Context, name, revision string) (store.Secret, error) {
	ctx, store, name := r.getStore(ctx, name)
	return store.GetRevision(ctx, name, revision)
}
