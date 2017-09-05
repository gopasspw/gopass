package sub

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/fsutil"
)

// Get returns the plaintext of a single key
func (s *Store) Get(ctx context.Context, name string) ([]byte, error) {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return []byte{}, store.ErrSneaky
	}

	if !fsutil.IsFile(p) {
		if s.debug {
			fmt.Printf("File %s not found\n", p)
		}
		return []byte{}, store.ErrNotFound
	}

	content, err := s.gpg.Decrypt(ctx, p)
	if err != nil {
		return []byte{}, store.ErrDecrypt
	}

	return content, nil
}

// GetFirstLine returns the first line of the plaintext of a single key
func (s *Store) GetFirstLine(ctx context.Context, name string) ([]byte, error) {
	content, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(content, []byte("\n"))
	if len(lines) < 1 {
		return nil, store.ErrNoPassword
	}

	return bytes.TrimSpace(lines[0]), nil
}

// GetBody returns everything but the first line
func (s *Store) GetBody(ctx context.Context, name string) ([]byte, error) {
	content, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	lines := bytes.SplitN(content, []byte("\n"), 2)
	if len(lines) < 2 || len(bytes.TrimSpace(lines[1])) < 1 {
		return nil, store.ErrNoBody
	}
	return lines[1], nil
}
