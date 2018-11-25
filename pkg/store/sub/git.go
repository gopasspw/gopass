package sub

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	_ "github.com/gopasspw/gopass/pkg/backend/rcs" // register RCS backends
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store"
	"github.com/gopasspw/gopass/pkg/store/secret"

	"github.com/pkg/errors"
)

// RCS returns the sync backend
func (s *Store) RCS() backend.RCS {
	return s.rcs
}

// GitInit initializes the the git repo in the store
func (s *Store) GitInit(ctx context.Context, un, ue string) error {
	rcs, err := backend.InitRCS(ctx, backend.GetRCSBackend(ctx), s.url.Path, un, ue)
	if err != nil {
		return err
	}
	s.rcs = rcs
	return nil
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
