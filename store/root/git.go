package root

import (
	"context"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend"
)

// Sync returns the sync backend
func (r *Store) Sync(ctx context.Context, name string) backend.Sync {
	_, sub, _ := r.getStore(ctx, name)
	return sub.Sync()
}

// GitInit initializes the git repo
func (r *Store) GitInit(ctx context.Context, name, userName, userEmail string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.GitInit(ctx, userName, userEmail)
}

// GitInitConfig initializes the git repos local config
func (r *Store) GitInitConfig(ctx context.Context, name, userName, userEmail string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.GitInitConfig(ctx, userName, userEmail)
}

// GitVersion returns git version information
func (r *Store) GitVersion(ctx context.Context) semver.Version {
	return r.store.GitVersion(ctx)
}

// GitAddRemote adds a git remote
func (r *Store) GitAddRemote(ctx context.Context, name, remote, url string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.GitAddRemote(ctx, remote, url)
}

// GitPull performs a git pull
func (r *Store) GitPull(ctx context.Context, name, origin, remote string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.GitPush(ctx, origin, remote)
}

// GitPush performs a git push
func (r *Store) GitPush(ctx context.Context, name, origin, remote string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.GitPush(ctx, origin, remote)
}
