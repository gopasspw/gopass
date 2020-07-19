package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret/secparse"
	"github.com/pkg/errors"
)

// GitInit initializes the git storage
func (s *Store) GitInit(ctx context.Context) error {
	storage, err := backend.InitStorage(ctx, backend.GetStorageBackend(ctx), s.path)
	if err != nil {
		return err
	}
	s.storage = storage
	return nil
}

// ListRevisions will list all revisions for a secret
func (s *Store) ListRevisions(ctx context.Context, name string) ([]backend.Revision, error) {
	p := s.passfile(name)
	return s.storage.Revisions(ctx, p)
}

// GetRevision will retrieve a single revision from the backend
func (s *Store) GetRevision(ctx context.Context, name, revision string) (gopass.Secret, error) {
	p := s.passfile(name)
	ciphertext, err := s.storage.GetRevision(ctx, p, revision)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ciphertext of '%s'@'%s'", name, revision)
	}

	content, err := s.crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		debug.Log("Decryption failed: %s", err)
		return nil, store.ErrDecrypt
	}

	sec, err := secparse.Parse(content)
	if err != nil {
		debug.Log("Failed to parse YAML: %s", err)
	}
	return sec, nil
}

// GitStatus shows the git status output
func (s *Store) GitStatus(ctx context.Context, _ string) error {
	buf, err := s.storage.Status(ctx)
	if err != nil {
		return err
	}
	out.Print(ctx, string(buf))
	return nil
}
