//go:build linux

package secretservice

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/gopasspw/gopass/pkg/debug"
	"golang.org/x/sys/unix"
)

// SessionCollectionName is the well-known name and alias of the volatile
// session collection. Its object path is
// /org/freedesktop/secrets/collection/session.
const SessionCollectionName = "session"

// vault stores secret payloads for the volatile session collection. The
// kernel-keyring implementation keeps payloads outside the Go heap so they
// never end up in a core dump or a heap profile; memVault is a fallback for
// environments where the keyctl syscalls are unavailable (e.g. some sandboxes).
type vault interface {
	put(id string, payload []byte) error
	get(id string) ([]byte, error)
	remove(id string)
	close()
}

// keyringVault stores payloads in the process kernel keyring. All keyring
// syscalls run on one dedicated, runtime.LockOSThread-pinned OS thread because
// Linux keeps keyring references in the per-task struct cred, so the syscalls
// must not migrate between threads.
type keyringVault struct {
	calls chan func()

	mu     sync.RWMutex
	closed bool
	once   sync.Once
}

// keyType is the keyring key type used for payloads.
const keyType = "user"

// newKeyringVault starts the pinned worker thread and ensures the process
// keyring exists.
func newKeyringVault() (*keyringVault, error) {
	v := &keyringVault{calls: make(chan func())}

	ready := make(chan error, 1)
	go func() {
		runtime.LockOSThread()

		// Create the process keyring if it does not exist yet.
		if _, err := unix.KeyctlGetKeyringID(unix.KEY_SPEC_PROCESS_KEYRING, true); err != nil {
			ready <- err

			return
		}
		ready <- nil

		for fn := range v.calls {
			fn()
		}
	}()

	if err := <-ready; err != nil {
		return nil, fmt.Errorf("kernel keyring unavailable: %w", err)
	}

	return v, nil
}

// keyDesc is the keyring description for a session item.
func keyDesc(id string) string {
	return "gopass-secret-service:" + id
}

// do runs fn on the pinned thread and returns its error.
func (v *keyringVault) do(fn func() error) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.closed {
		return fmt.Errorf("vault is closed")
	}

	done := make(chan error, 1)
	v.calls <- func() { done <- fn() }

	return <-done
}

// unlinkLocked removes any existing key for id. Must run on the pinned thread.
func (v *keyringVault) unlinkLocked(id string) {
	keyID, err := unix.KeyctlSearch(unix.KEY_SPEC_PROCESS_KEYRING, keyType, keyDesc(id), 0)
	if err != nil {
		return
	}

	// Unlink from the process keyring and then invalidate the key itself.
	_, _ = unix.KeyctlInt(unix.KEYCTL_UNLINK, keyID, unix.KEY_SPEC_PROCESS_KEYRING, 0, 0)
	_, _ = unix.KeyctlInt(unix.KEYCTL_INVALIDATE, keyID, 0, 0, 0)
}

func (v *keyringVault) put(id string, payload []byte) error {
	return v.do(func() error {
		// Replace any existing key: AddKey fails with EEXIST otherwise.
		v.unlinkLocked(id)

		if _, err := unix.AddKey(keyType, keyDesc(id), payload, unix.KEY_SPEC_PROCESS_KEYRING); err != nil {
			return fmt.Errorf("add key: %w", err)
		}

		return nil
	})
}

func (v *keyringVault) get(id string) ([]byte, error) {
	var out []byte

	err := v.do(func() error {
		keyID, err := unix.KeyctlSearch(unix.KEY_SPEC_PROCESS_KEYRING, keyType, keyDesc(id), 0)
		if err != nil {
			return err
		}

		// A nil buffer makes KEYCTL_READ return the payload size.
		size, err := unix.KeyctlBuffer(unix.KEYCTL_READ, keyID, nil, 0)
		if err != nil {
			return err
		}
		if size <= 0 {
			out = []byte{}

			return nil
		}

		buf := make([]byte, size)
		n, err := unix.KeyctlBuffer(unix.KEYCTL_READ, keyID, buf, 0)
		if err != nil {
			return err
		}
		out = buf[:n]

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (v *keyringVault) remove(id string) {
	_ = v.do(func() error {
		v.unlinkLocked(id)

		return nil
	})
}

func (v *keyringVault) close() {
	v.once.Do(func() {
		v.mu.Lock()
		v.closed = true
		close(v.calls)
		v.mu.Unlock()
	})
}

// memVault is an in-memory fallback used when the kernel keyring is
// unavailable. Payloads live on the Go heap and are zeroed on removal.
type memVault struct {
	mu sync.Mutex
	m  map[string][]byte
}

func newMemVault() *memVault {
	return &memVault{m: make(map[string][]byte)}
}

func (v *memVault) put(id string, payload []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	cp := make([]byte, len(payload))
	copy(cp, payload)
	v.m[id] = cp

	return nil
}

func (v *memVault) get(id string) ([]byte, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	payload, ok := v.m[id]
	if !ok {
		return nil, fmt.Errorf("not found: %s", id)
	}

	cp := make([]byte, len(payload))
	copy(cp, payload)

	return cp, nil
}

func (v *memVault) remove(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if payload, ok := v.m[id]; ok {
		for i := range payload {
			payload[i] = 0
		}
		delete(v.m, id)
	}
}

func (v *memVault) close() {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, payload := range v.m {
		for i := range payload {
			payload[i] = 0
		}
	}
	v.m = make(map[string][]byte)
}

// newVault returns a kernel-keyring vault if the syscalls are available, or an
// in-memory fallback otherwise.
func newVault() vault {
	kv, err := newKeyringVault()
	if err != nil {
		debug.Log("secret-service: %s; falling back to in-memory vault", err)

		return newMemVault()
	}

	return kv
}

// volatileItem is the metadata of a session collection item. The secret payload
// itself is not held here; it lives in the vault.
type volatileItem struct {
	label       string
	created     time.Time
	modified    time.Time
	contentType string
	attributes  map[string]string
}

// keyringStore implements Store for the volatile session collection. It never
// touches the gopass store: payloads go to the vault and metadata is kept in
// memory only.
type keyringStore struct {
	vault   vault
	label   string
	created time.Time

	mu    sync.RWMutex
	items map[string]*volatileItem
}

// newKeyringStore creates the volatile store for the session collection.
func newKeyringStore() *keyringStore {
	return &keyringStore{
		vault:   newVault(),
		label:   "Session",
		created: time.Now(),
		items:   make(map[string]*volatileItem),
	}
}

// sessionOnly rejects any collection but the well-known session collection.
func sessionOnly(name string) error {
	if name != SessionCollectionName {
		return fmt.Errorf("volatile store only serves the session collection, got %q", name)
	}

	return nil
}

// Collections returns the single session collection.
func (s *keyringStore) Collections(context.Context) ([]string, error) {
	return []string{SessionCollectionName}, nil
}

// GetCollection returns the session collection's data.
func (s *keyringStore) GetCollection(_ context.Context, name string) (*CollectionData, error) {
	if err := sessionOnly(name); err != nil {
		return nil, err
	}

	return &CollectionData{
		Name:     SessionCollectionName,
		Label:    s.label,
		Created:  s.created,
		Modified: time.Now(),
		Locked:   false,
	}, nil
}

// CreateCollection is not supported: the session collection always exists.
func (s *keyringStore) CreateCollection(context.Context, string, string) error {
	return fmt.Errorf("the session collection cannot be created")
}

// DeleteCollection is not supported: the session collection cannot be deleted.
func (s *keyringStore) DeleteCollection(context.Context, string) error {
	return fmt.Errorf("the session collection cannot be deleted")
}

// SetCollectionLabel updates the session collection label.
func (s *keyringStore) SetCollectionLabel(ctx context.Context, name, label string) error {
	if err := sessionOnly(name); err != nil {
		return err
	}
	s.label = label

	return nil
}

// Items returns the IDs of all session items.
func (s *keyringStore) Items(_ context.Context, collection string) ([]string, error) {
	if err := sessionOnly(collection); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]string, 0, len(s.items))
	for id := range s.items {
		out = append(out, id)
	}
	sort.Strings(out)

	return out, nil
}

// GetItem reads an item and its payload from the vault.
func (s *keyringStore) GetItem(_ context.Context, collection, id string) (*ItemData, error) {
	if err := sessionOnly(collection); err != nil {
		return nil, err
	}

	s.mu.RLock()
	it, ok := s.items[id]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("item not found: %s", id)
	}

	payload, err := s.vault.get(id)
	if err != nil {
		return nil, err
	}

	attrs := make(map[string]string, len(it.attributes))
	for k, v := range it.attributes {
		attrs[k] = v
	}

	return &ItemData{
		ID:          id,
		Secret:      payload,
		Label:       it.label,
		Created:     it.created,
		Modified:    it.modified,
		ContentType: it.contentType,
		Attributes:  attrs,
	}, nil
}

// CreateItem stores a new session item.
func (s *keyringStore) CreateItem(_ context.Context, collection string, item *ItemData) (string, error) {
	if err := sessionOnly(collection); err != nil {
		return "", err
	}

	if item.ID == "" {
		id, err := newID('i')
		if err != nil {
			return "", err
		}
		item.ID = id
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[item.ID]; ok {
		return "", fmt.Errorf("item already exists: %s", item.ID)
	}

	if err := s.vault.put(item.ID, item.Secret); err != nil {
		return "", err
	}

	now := time.Now()
	if item.Created.IsZero() {
		item.Created = now
	}
	item.Modified = now
	if item.ContentType == "" {
		item.ContentType = "text/plain"
	}

	s.items[item.ID] = cloneVolatile(item)

	return item.ID, nil
}

// UpdateItem overwrites an existing session item.
func (s *keyringStore) UpdateItem(_ context.Context, collection, id string, item *ItemData) error {
	if err := sessionOnly(collection); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.items[id]
	if !ok {
		return fmt.Errorf("item not found: %s", id)
	}

	if err := s.vault.put(id, item.Secret); err != nil {
		return err
	}

	item.ID = id
	item.Created = existing.created
	item.Modified = time.Now()
	if item.ContentType == "" {
		item.ContentType = existing.contentType
	}
	if item.ContentType == "" {
		item.ContentType = "text/plain"
	}

	s.items[id] = cloneVolatile(item)

	return nil
}

// DeleteItem removes a session item.
func (s *keyringStore) DeleteItem(_ context.Context, collection, id string) error {
	if err := sessionOnly(collection); err != nil {
		return err
	}

	s.mu.Lock()
	_, ok := s.items[id]
	delete(s.items, id)
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("item not found: %s", id)
	}
	s.vault.remove(id)

	return nil
}

// SearchItems returns session items matching all given attributes.
func (s *keyringStore) SearchItems(ctx context.Context, collection string, attrs map[string]string) ([]*ItemData, error) {
	ids, err := s.Items(ctx, collection)
	if err != nil {
		return nil, err
	}

	var out []*ItemData
	for _, id := range ids {
		item, err := s.GetItem(ctx, collection, id)
		if err != nil {
			continue
		}
		if matchesAttributes(item, attrs) {
			out = append(out, item)
		}
	}

	return out, nil
}

// SearchAllItems searches the session collection only.
func (s *keyringStore) SearchAllItems(ctx context.Context, attrs map[string]string) (map[string][]*ItemData, error) {
	items, err := s.SearchItems(ctx, SessionCollectionName, attrs)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return map[string][]*ItemData{}, nil
	}

	return map[string][]*ItemData{SessionCollectionName: items}, nil
}

// LockCollection is a no-op: the session collection is always unlocked while
// the session exists.
func (s *keyringStore) LockCollection(context.Context, string) error { return nil }

// UnlockCollection is a no-op.
func (s *keyringStore) UnlockCollection(context.Context, string) error { return nil }

// GetAlias resolves the session alias.
func (s *keyringStore) GetAlias(_ context.Context, alias string) (string, error) {
	if alias != SessionCollectionName {
		return "", fmt.Errorf("alias not found: %s", alias)
	}

	return SessionCollectionName, nil
}

// SetAlias rejects changes to the session alias.
func (s *keyringStore) SetAlias(context.Context, string, string) error {
	return fmt.Errorf("the session alias is read-only")
}

// Aliases returns the well-known session alias.
func (s *keyringStore) Aliases(context.Context) (map[string]string, error) {
	return map[string]string{SessionCollectionName: SessionCollectionName}, nil
}

// Close wipes the vault.
func (s *keyringStore) Close(context.Context) error {
	s.vault.close()

	return nil
}

// cloneVolatile makes a private copy of an item's metadata (never its payload,
// which is held by the vault).
func cloneVolatile(item *ItemData) *volatileItem {
	attrs := make(map[string]string, len(item.Attributes))
	for k, v := range item.Attributes {
		attrs[k] = v
	}

	return &volatileItem{
		label:       item.Label,
		created:     item.Created,
		modified:    item.Modified,
		contentType: item.ContentType,
		attributes:  attrs,
	}
}
