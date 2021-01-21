package root

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// RCSInit initializes the version control repo
func (r *Store) RCSInit(ctx context.Context, name, userName, userEmail string) error {
	ctx, store, _ := r.getStore(ctx, name)
	ctx = ctxutil.WithUsername(ctx, userName)
	ctx = ctxutil.WithEmail(ctx, userEmail)
	return store.GitInit(ctx)
}

// RCSInitConfig initializes the git repos local config
func (r *Store) RCSInitConfig(ctx context.Context, name, userName, userEmail string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.Storage().InitConfig(ctx, userName, userEmail)
}

// RCSAddRemote adds a git remote
func (r *Store) RCSAddRemote(ctx context.Context, name, remote, url string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.Storage().AddRemote(ctx, remote, url)
}

// RCSRemoveRemote removes a git remote
func (r *Store) RCSRemoveRemote(ctx context.Context, name, remote string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.Storage().RemoveRemote(ctx, remote)
}

// RCSPull performs a git pull
func (r *Store) RCSPull(ctx context.Context, name, origin, remote string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.Storage().Pull(ctx, origin, remote)
}

// RCSPush performs a git push
func (r *Store) RCSPush(ctx context.Context, name, origin, remote string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.Storage().Push(ctx, origin, remote)
}

// ListRevisions will list all revisions for the named entity
func (r *Store) ListRevisions(ctx context.Context, name string) ([]backend.Revision, error) {
	ctx, store, name := r.getStore(ctx, name)
	return store.ListRevisions(ctx, name)
}

// GetRevision will try to retrieve the given revision from the sync backend
func (r *Store) GetRevision(ctx context.Context, name, revision string) (context.Context, gopass.Secret, error) {
	ctx, store, name := r.getStore(ctx, name)
	sec, err := store.GetRevision(ctx, name, revision)
	return ctx, sec, err
}

// RCSStatus show the git status
func (r *Store) RCSStatus(ctx context.Context, name string) error {
	ctx, store, name := r.getStore(ctx, name)
	out.Print(ctx, "Store: %s", store.Path())
	return store.GitStatus(ctx, name)
}
