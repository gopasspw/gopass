package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	_ "github.com/gopasspw/gopass/internal/backend/rcs" // register RCS backends
	"github.com/gopasspw/gopass/internal/backend/storage/kv/ondisk"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret/secparse"
	"github.com/pkg/errors"
)

func (s *Store) initRCSBackend(ctx context.Context) {
	if rcs, ok := s.storage.(*ondisk.OnDisk); ok {
		s.rcs = rcs
		return
	}
	rcs, err := backend.DetectRCS(ctx, s.path)
	if err != nil {
		debug.Log("Failed to initialized RCS backend: %s", err)
	}
	s.rcs = rcs
}

// RCS returns the sync backend
func (s *Store) RCS() backend.RCS {
	return s.rcs
}

// GitInit initializes the the git repo in the store
func (s *Store) GitInit(ctx context.Context) error {
	rcs, err := backend.InitRCS(ctx, backend.GetRCSBackend(ctx), s.path)
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
func (s *Store) GetRevision(ctx context.Context, name, revision string) (gopass.Secret, error) {
	p := s.passfile(name)
	ciphertext, err := s.rcs.GetRevision(ctx, p, revision)
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
	buf, err := s.rcs.Status(ctx)
	if err != nil {
		return err
	}
	out.Print(ctx, string(buf))
	return nil
}
