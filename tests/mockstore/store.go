package mockstore

import (
	"context"
	"fmt"

	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/backend/crypto/plain"
	"github.com/justwatchcom/gopass/pkg/backend/rcs/noop"
	"github.com/justwatchcom/gopass/pkg/backend/storage/kv/inmem"
	"github.com/justwatchcom/gopass/pkg/store"
	"github.com/justwatchcom/gopass/pkg/store/secret"
	"github.com/justwatchcom/gopass/pkg/tree"
)

// MockStore is an mocked store
type MockStore struct {
	alias   string
	storage backend.Storage
}

// New creates a new mock store
func New(alias string) *MockStore {
	return &MockStore{
		alias:   alias,
		storage: inmem.New(),
	}
}

// String implements fmt.Stringer
func (m *MockStore) String() string {
	return "mockstore"
}

// GetTemplate returns nothing
func (m *MockStore) GetTemplate(context.Context, string) ([]byte, error) {
	return []byte{}, nil
}

// HasTemplate returns false
func (m *MockStore) HasTemplate(context.Context, string) bool {
	return false
}

// ListTemplates returns nothing
func (m *MockStore) ListTemplates(context.Context, string) []string {
	return nil
}

// LookupTemplate returns nothing
func (m *MockStore) LookupTemplate(context.Context, string) ([]byte, bool) {
	return []byte{}, false
}

// RemoveTemplate does nothing
func (m *MockStore) RemoveTemplate(context.Context, string) error {
	return nil
}

// SetTemplate does nothing
func (m *MockStore) SetTemplate(context.Context, string, []byte) error {
	return nil
}

// TemplateTree does nothing
func (m *MockStore) TemplateTree(context.Context) (tree.Tree, error) {
	return nil, nil
}

// AddRecipient does nothing
func (m *MockStore) AddRecipient(context.Context, string) error {
	return nil
}

// GetRecipients does nothing
func (m *MockStore) GetRecipients(context.Context, string) ([]string, error) {
	return nil, nil
}

// RemoveRecipient does nothing
func (m *MockStore) RemoveRecipient(context.Context, string) error {
	return nil
}

// SaveRecipients does nothing
func (m *MockStore) SaveRecipients(context.Context) error {
	return nil
}

// Recipients does nothing
func (m *MockStore) Recipients(context.Context) []string {
	return nil
}

// ImportMissingPublicKeys does nothing
func (m *MockStore) ImportMissingPublicKeys(context.Context) error {
	return nil
}

// ExportMissingPublicKeys does nothing
func (m *MockStore) ExportMissingPublicKeys(context.Context, []string) (bool, error) {
	return false, nil
}

// Fsck does nothing
func (m *MockStore) Fsck(context.Context, string) error {
	return nil
}

// Path does nothing
func (m *MockStore) Path() string {
	return ""
}

// URL does nothing
func (m *MockStore) URL() string {
	return "mockstore://"
}

// RCS does nothing
func (m *MockStore) RCS() backend.RCS {
	return noop.New()
}

// Crypto does nothing
func (m *MockStore) Crypto() backend.Crypto {
	return plain.New()
}

// Storage does nothing
func (m *MockStore) Storage() backend.Storage {
	return m.storage
}

// GitInit does nothing
func (m *MockStore) GitInit(context.Context, string, string) error {
	return nil
}

// Alias does nothing
func (m *MockStore) Alias() string {
	return m.alias
}

// Copy does nothing
func (m *MockStore) Copy(ctx context.Context, from string, to string) error {
	content, _ := m.storage.Get(ctx, from)
	m.storage.Set(ctx, to, content)
	return nil
}

// Delete does nothing
func (m *MockStore) Delete(ctx context.Context, name string) error {
	return m.storage.Delete(ctx, name)
}

// Equals does nothing
func (m *MockStore) Equals(other store.Store) bool {
	return false
}

// Exists does nothing
func (m *MockStore) Exists(ctx context.Context, name string) bool {
	return m.storage.Exists(ctx, name)
}

// Get does nothing
func (m *MockStore) Get(ctx context.Context, name string) (store.Secret, error) {
	content, err := m.storage.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return secret.Parse(content)
}

// GetRevision does nothing
func (m *MockStore) GetRevision(context.Context, string, string) (store.Secret, error) {
	return nil, fmt.Errorf("not supported")
}

// Init does nothing
func (m *MockStore) Init(context.Context, string, ...string) error {
	return nil
}

// Initialized does nothing
func (m *MockStore) Initialized(context.Context) bool {
	return true
}

// IsDir does nothing
func (m *MockStore) IsDir(ctx context.Context, name string) bool {
	return m.storage.IsDir(ctx, name)
}

// List does nothing
func (m *MockStore) List(context.Context, string) ([]string, error) {
	return nil, nil
}

// ListRevisions does nothing
func (m *MockStore) ListRevisions(context.Context, string) ([]backend.Revision, error) {
	return nil, nil
}

// Move does nothing
func (m *MockStore) Move(ctx context.Context, from string, to string) error {
	content, _ := m.storage.Get(ctx, from)
	m.storage.Set(ctx, to, content)
	return m.storage.Delete(ctx, from)
}

// Set does nothing
func (m *MockStore) Set(ctx context.Context, name string, sec store.Secret) error {
	buf, err := sec.Bytes()
	if err != nil {
		return err
	}
	return m.storage.Set(ctx, name, buf)
}

// Prune does nothing
func (m *MockStore) Prune(context.Context, string) error {
	return fmt.Errorf("not supported")
}

// Valid does nothing
func (m *MockStore) Valid() bool {
	return true
}

// MountPoints does nothing
func (m *MockStore) MountPoints() []string {
	return nil
}
