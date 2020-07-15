package ondisk

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
)

// Add is not supported / necessary
func (o *OnDisk) Add(ctx context.Context, args ...string) error {
	return nil
}

// Commit is not supported / necessary
func (o *OnDisk) Commit(ctx context.Context, msg string) error {
	return nil
}

// Push fetches the index from the remote, merges it and uploads the result
func (o *OnDisk) Push(ctx context.Context, remote, location string) error {
	o.mux.Lock()
	defer o.mux.Unlock()
	if err := o.uploadFiles(ctx); err != nil {
		return err
	}
	if err := o.syncIndex(ctx); err != nil {
		return err
	}
	return o.downloadFiles(ctx)
}

// Pull fetches the index from the remote but does not upload it.
func (o *OnDisk) Pull(ctx context.Context, remote, location string) error {
	o.mux.Lock()
	defer o.mux.Unlock()
	if err := o.syncIndex(ctx); err != nil {
		return err
	}
	return o.downloadFiles(ctx)
}

// InitConfig is not necessary
func (o *OnDisk) InitConfig(ctx context.Context, name, email string) error {
	return nil
}

// AddRemote sets the remote
func (o *OnDisk) AddRemote(ctx context.Context, _, location string) error {
	return o.SetRemote(ctx, location)
}

// RemoveRemote removes the remote
func (o *OnDisk) RemoveRemote(ctx context.Context, _ string) error {
	return o.SetRemote(ctx, "")
}

// Revisions returns a list of revisions for this entry
func (o *OnDisk) Revisions(ctx context.Context, name string) ([]backend.Revision, error) {
	if !o.Exists(ctx, name) {
		return nil, fmt.Errorf("not found")
	}
	e, err := o.getEntry(name)
	if err != nil {
		return nil, err
	}
	revs := make([]backend.Revision, 0, len(e.Revisions))
	for _, rev := range e.SortedRevisions() {
		revs = append(revs, backend.Revision{
			Hash:    rev.ID(),
			Subject: rev.Message,
			Date:    rev.Time(),
		})
	}
	return revs, nil
}

// GetRevision returns a single revision
func (o *OnDisk) GetRevision(ctx context.Context, name, revision string) ([]byte, error) {
	if !o.Exists(ctx, name) {
		return nil, fmt.Errorf("not found")
	}
	e, err := o.getEntry(name)
	if err != nil {
		return nil, err
	}
	for _, rev := range e.SortedRevisions() {
		if revision == rev.ID() {
			path := filepath.Join(o.dir, rev.GetFilename())
			return ioutil.ReadFile(path)
		}
	}
	return nil, fmt.Errorf("not found")
}

// Status is not necessary
func (o *OnDisk) Status(ctx context.Context) ([]byte, error) {
	return nil, nil
}
