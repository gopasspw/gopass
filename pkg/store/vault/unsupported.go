package vault

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/backend/crypto/plain"
	"github.com/gopasspw/gopass/pkg/backend/rcs/noop"
	"github.com/gopasspw/gopass/pkg/backend/storage/kv/inmem"
	"github.com/gopasspw/gopass/pkg/store"
	"github.com/gopasspw/gopass/pkg/tree"
)

// GetTemplate is unsupported
func (s *Store) GetTemplate(context.Context, string) ([]byte, error) {
	return nil, fmt.Errorf("not supported")
}

// HasTemplate is unsupported
func (s *Store) HasTemplate(context.Context, string) bool {
	return false
}

// ListTemplates is unsupported
func (s *Store) ListTemplates(context.Context, string) []string {
	return nil
}

// LookupTemplate is unsupported
func (s *Store) LookupTemplate(context.Context, string) ([]byte, bool) {
	return nil, false
}

// RemoveTemplate is unsupported
func (s *Store) RemoveTemplate(context.Context, string) error {
	return fmt.Errorf("not supported")
}

// SetTemplate is unsupported
func (s *Store) SetTemplate(context.Context, string, []byte) error {
	return fmt.Errorf("not supported")
}

// TemplateTree is unsupported
func (s *Store) TemplateTree(context.Context) (tree.Tree, error) {
	return nil, fmt.Errorf("not supported")
}

// AddRecipient is unsupported
func (s *Store) AddRecipient(context.Context, string) error {
	return fmt.Errorf("not supported")
}

// GetRecipients is unsupported
func (s *Store) GetRecipients(context.Context, string) ([]string, error) {
	return nil, fmt.Errorf("not supported")
}

// RemoveRecipient is unsupported
func (s *Store) RemoveRecipient(context.Context, string) error {
	return fmt.Errorf("not supported")
}

// SaveRecipients is unsupported
func (s *Store) SaveRecipients(context.Context) error {
	return fmt.Errorf("not supported")
}

// SetRecipients is unsupported
func (s *Store) SetRecipients(context.Context, []string) error {
	return fmt.Errorf("not supported")
}

// Recipients is unsupported
func (s *Store) Recipients(context.Context) []string {
	return nil
}

// ImportMissingPublicKeys is unsupported
func (s *Store) ImportMissingPublicKeys(context.Context) error {
	return fmt.Errorf("not supported")
}

// ExportMissingPublicKeys is unsupported
func (s *Store) ExportMissingPublicKeys(context.Context, []string) (bool, error) {
	return false, fmt.Errorf("not supported")
}

// RCS is unsupported
func (s *Store) RCS() backend.RCS {
	return noop.New()
}

// Crypto is unsupported
func (s *Store) Crypto() backend.Crypto {
	return plain.New()
}

// Storage is unsupported
func (s *Store) Storage() backend.Storage {
	return inmem.New()
}

// GitInit is unsupported
func (s *Store) GitInit(context.Context, string, string) error {
	return fmt.Errorf("not supported")
}

// GetRevision is unsupported
func (s *Store) GetRevision(context.Context, string, string) (store.Secret, error) {
	return nil, fmt.Errorf("not supported")
}

// ListRevisions is unsupported
func (s *Store) ListRevisions(context.Context, string) ([]backend.Revision, error) {
	return nil, fmt.Errorf("not supported")
}

// Fsck is unsupported
func (s *Store) Fsck(ctx context.Context, prefix string) error {
	return nil
}
