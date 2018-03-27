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

type MockStore struct {
	alias   string
	storage backend.Storage
}

func New(alias string) *MockStore {
	return &MockStore{
		alias:   alias,
		storage: inmem.New(),
	}
}

func (m *MockStore) String() string {
	return "mockstore"
}

func (m *MockStore) GetTemplate(context.Context, string) ([]byte, error) {
	return []byte{}, nil
}

func (m *MockStore) HasTemplate(context.Context, string) bool {
	return false
}

func (m *MockStore) ListTemplates(context.Context, string) []string {
	return nil
}

func (m *MockStore) LookupTemplate(context.Context, string) ([]byte, bool) {
	return []byte{}, false
}

func (m *MockStore) RemoveTemplate(context.Context, string) error {
	return nil
}

func (m *MockStore) SetTemplate(context.Context, string, []byte) error {
	return nil
}

func (m *MockStore) TemplateTree(context.Context) (tree.Tree, error) {
	return nil, nil
}

func (m *MockStore) AddRecipient(context.Context, string) error {
	return nil
}

func (m *MockStore) GetRecipients(context.Context, string) ([]string, error) {
	return nil, nil
}

func (m *MockStore) RemoveRecipient(context.Context, string) error {
	return nil
}

func (m *MockStore) SaveRecipients(context.Context) error {
	return nil
}

func (m *MockStore) Recipients(context.Context) []string {
	return nil
}

func (m *MockStore) ImportMissingPublicKeys(context.Context) error {
	return nil
}

func (m *MockStore) ExportMissingPublicKeys(context.Context, []string) (bool, error) {
	return false, nil
}

func (m *MockStore) Fsck(context.Context, string) error {
	return nil
}

func (m *MockStore) Path() string {
	return ""
}

func (m *MockStore) URL() string {
	return "mockstore://"
}

func (m *MockStore) RCS() backend.RCS {
	return noop.New()
}

func (m *MockStore) Crypto() backend.Crypto {
	return plain.New()
}

func (m *MockStore) Storage() backend.Storage {
	return m.storage
}

func (m *MockStore) GitInit(context.Context, string, string) error {
	return nil
}

func (m *MockStore) Alias() string {
	return m.alias
}

func (m *MockStore) Copy(ctx context.Context, from string, to string) error {
	content, _ := m.storage.Get(ctx, from)
	m.storage.Set(ctx, to, content)
	return nil
}

func (m *MockStore) Delete(ctx context.Context, name string) error {
	return m.storage.Delete(ctx, name)
}

func (m *MockStore) Equals(other store.Store) bool {
	return false
}

func (m *MockStore) Exists(ctx context.Context, name string) bool {
	return m.storage.Exists(ctx, name)
}

func (m *MockStore) Get(ctx context.Context, name string) (store.Secret, error) {
	content, err := m.storage.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return secret.Parse(content)
}

func (m *MockStore) GetRevision(context.Context, string, string) (store.Secret, error) {
	return nil, fmt.Errorf("not supported")
}

func (m *MockStore) Init(context.Context, string, ...string) error {
	return nil
}

func (m *MockStore) Initialized(context.Context) bool {
	return true
}

func (m *MockStore) IsDir(ctx context.Context, name string) bool {
	return m.storage.IsDir(ctx, name)
}

func (m *MockStore) List(context.Context, string) ([]string, error) {
	return nil, nil
}

func (m *MockStore) ListRevisions(context.Context, string) ([]backend.Revision, error) {
	return nil, nil
}

func (m *MockStore) Move(ctx context.Context, from string, to string) error {
	content, _ := m.storage.Get(ctx, from)
	m.storage.Set(ctx, to, content)
	return m.storage.Delete(ctx, from)
}

func (m *MockStore) Set(ctx context.Context, name string, sec store.Secret) error {
	buf, err := sec.Bytes()
	if err != nil {
		return err
	}
	return m.storage.Set(ctx, name, buf)
}

func (m *MockStore) Prune(context.Context, string) error {
	return fmt.Errorf("not supported")
}

func (m *MockStore) Valid() bool {
	return true
}

func (m *MockStore) MountPoints() []string {
	return nil
}
