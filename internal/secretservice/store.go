//go:build linux

package secretservice

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
)

// Reserved metadata keys. Keys prefixed with metaPrefix are internal; every
// other key on an item is a user-visible search attribute.
const (
	metaPrefix     = "_ss_"
	labelKey       = "_ss_label"
	createdKey     = "_ss_created"
	modifiedKey    = "_ss_modified"
	contentTypeKey = "_ss_content_type"

	collLabelKey    = "_ss_coll_label"
	collCreatedKey  = "_ss_coll_created"
	collModifiedKey = "_ss_coll_modified"
)

// ItemData is a Secret Service item backed by a gopass secret.
type ItemData struct {
	ID          string
	Secret      []byte
	Label       string
	Created     time.Time
	Modified    time.Time
	ContentType string
	Attributes  map[string]string
}

// CollectionData is a Secret Service collection.
type CollectionData struct {
	Name     string
	Label    string
	Created  time.Time
	Modified time.Time
	Locked   bool
}

// Store is the persistence backend used by the service. It is implemented by
// gopassStore (durable) and, for the volatile session collection, by memStore.
type Store interface {
	Collections(ctx context.Context) ([]string, error)
	GetCollection(ctx context.Context, name string) (*CollectionData, error)
	CreateCollection(ctx context.Context, name, label string) error
	DeleteCollection(ctx context.Context, name string) error
	SetCollectionLabel(ctx context.Context, name, label string) error

	Items(ctx context.Context, collection string) ([]string, error)
	GetItem(ctx context.Context, collection, id string) (*ItemData, error)
	CreateItem(ctx context.Context, collection string, item *ItemData) (string, error)
	UpdateItem(ctx context.Context, collection, id string, item *ItemData) error
	DeleteItem(ctx context.Context, collection, id string) error
	SearchItems(ctx context.Context, collection string, attrs map[string]string) ([]*ItemData, error)
	SearchAllItems(ctx context.Context, attrs map[string]string) (map[string][]*ItemData, error)

	LockCollection(ctx context.Context, name string) error
	UnlockCollection(ctx context.Context, name string) error

	GetAlias(ctx context.Context, alias string) (string, error)
	SetAlias(ctx context.Context, alias, collection string) error
	Aliases(ctx context.Context) (map[string]string, error)

	Close(ctx context.Context) error
}

// gopassStore implements Store on top of the public gopass Go API. It never
// shells out to the CLI.
type gopassStore struct {
	store  gopass.Store
	mapper mapper

	mu     sync.RWMutex
	locked map[string]bool

	// metaCache memoizes the decrypted *metadata* of an item, keyed by store
	// path. It never holds the secret payload: SearchItems only needs
	// attributes to match, and decryption dominates its latency, so caching
	// metadata turns a per-lookup O(N) decryption of a collection into a
	// single decryption per entry. Entries are invalidated on every local
	// mutation; there is no TTL, so out-of-process changes require a daemon
	// restart.
	cacheMu   sync.RWMutex
	metaCache map[string]map[string]string
}

// NewGopassStore creates a gopass-backed store under the given path prefix.
func NewGopassStore(ctx context.Context, prefix string) (*gopassStore, error) {
	gp, err := api.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize gopass: %w", err)
	}

	return newGopassStoreWithBackend(gp, prefix), nil
}

// newGopassStoreWithBackend is the dependency-injection seam used by tests.
func newGopassStoreWithBackend(backend gopass.Store, prefix string) *gopassStore {
	return &gopassStore{
		store:     backend,
		mapper:    mapper{prefix: prefix},
		locked:    make(map[string]bool),
		metaCache: make(map[string]map[string]string),
	}
}

// metaFromSecret extracts an item's searchable metadata. It copies only
// gopass Keys(), never Password() or Body(), so the result is safe to cache.
func metaFromSecret(sec gopass.Secret) map[string]string {
	keys := sec.Keys()
	m := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := sec.Get(k); ok {
			m[k] = v
		}
	}

	return m
}

// metaFor returns an item's cached metadata, decrypting and caching it on first
// access.
func (s *gopassStore) metaFor(ctx context.Context, p string) (map[string]string, error) {
	s.cacheMu.RLock()
	m, ok := s.metaCache[p]
	s.cacheMu.RUnlock()
	if ok {
		return m, nil
	}

	sec, err := s.store.Get(ctx, p, "latest")
	if err != nil {
		return nil, err
	}
	m = metaFromSecret(sec)

	s.cacheMu.Lock()
	s.metaCache[p] = m
	s.cacheMu.Unlock()

	return m, nil
}

func (s *gopassStore) invalidateMeta(p string) {
	s.cacheMu.Lock()
	delete(s.metaCache, p)
	s.cacheMu.Unlock()
}

func (s *gopassStore) invalidateMetaPrefix(prefix string) {
	s.cacheMu.Lock()
	for k := range s.metaCache {
		if k == prefix || strings.HasPrefix(k, prefix+"/") {
			delete(s.metaCache, k)
		}
	}
	s.cacheMu.Unlock()
}

// applyItemMeta fills an ItemData's metadata from a metadata map. It never
// touches ItemData.Secret.
func applyItemMeta(item *ItemData, meta map[string]string) {
	for key, val := range meta {
		switch key {
		case labelKey:
			item.Label = val
		case createdKey:
			if ts, err := time.Parse(time.RFC3339, val); err == nil {
				item.Created = ts
			}
		case modifiedKey:
			if ts, err := time.Parse(time.RFC3339, val); err == nil {
				item.Modified = ts
			}
		case contentTypeKey:
			item.ContentType = val
		default:
			if !strings.HasPrefix(key, metaPrefix) {
				item.Attributes[key] = val
			}
		}
	}
}

// Collections returns all collection names under the store prefix.
func (s *gopassStore) Collections(ctx context.Context) ([]string, error) {
	allPaths, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	for _, p := range allPaths {
		if !strings.HasPrefix(p, s.mapper.prefix+"/") {
			continue
		}
		rest := strings.TrimPrefix(p, s.mapper.prefix+"/")
		parts := strings.SplitN(rest, "/", 2)
		name := parts[0]
		if name == "" || strings.HasPrefix(name, "_") {
			continue
		}
		// Collection names become D-Bus object path elements, so only names
		// that are already path-safe can be served. Directories created out of
		// band with other characters are skipped.
		if !isPathSafe(name) {
			debug.Log("secret-service: skipping collection %q: not a valid D-Bus path element", name)

			continue
		}
		seen[name] = true
	}

	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)

	return out, nil
}

// GetCollection returns a collection's data.
func (s *gopassStore) GetCollection(ctx context.Context, name string) (*CollectionData, error) {
	name = SanitizeName(name)
	meta, err := s.metaFor(ctx, s.mapper.MetaPath(name))
	if err != nil {
		// No metadata entry: the collection still exists if it has any items.
		items, ierr := s.Items(ctx, name)
		if ierr != nil || len(items) == 0 {
			return nil, fmt.Errorf("collection not found: %s", name)
		}

		return &CollectionData{Name: name, Label: name, Created: time.Now(), Locked: s.isLocked(name)}, nil
	}

	data := &CollectionData{Name: name, Label: name, Locked: s.isLocked(name)}
	for key, val := range meta {
		switch key {
		case collLabelKey:
			data.Label = val
		case collCreatedKey:
			if ts, err := time.Parse(time.RFC3339, val); err == nil {
				data.Created = ts
			}
		case collModifiedKey:
			if ts, err := time.Parse(time.RFC3339, val); err == nil {
				data.Modified = ts
			}
		}
	}

	return data, nil
}

// CreateCollection writes a collection's metadata entry.
func (s *gopassStore) CreateCollection(ctx context.Context, name, label string) error {
	name = SanitizeName(name)
	now := time.Now().Format(time.RFC3339)

	sec := secrets.New()
	sec.SetPassword("collection-metadata")
	for _, kv := range []struct{ k, v string }{
		{collLabelKey, label},
		{collCreatedKey, now},
		{collModifiedKey, now},
	} {
		if err := sec.Set(kv.k, kv.v); err != nil {
			return fmt.Errorf("set %s: %w", kv.k, err)
		}
	}

	metaPath := s.mapper.MetaPath(name)
	if err := s.store.Set(ctx, metaPath, sec); err != nil {
		return err
	}
	s.invalidateMeta(metaPath)

	return nil
}

// DeleteCollection removes a collection and all of its items.
func (s *gopassStore) DeleteCollection(ctx context.Context, name string) error {
	collPath := s.mapper.CollectionPath(name)
	if err := s.store.RemoveAll(ctx, collPath); err != nil {
		return err
	}
	s.invalidateMetaPrefix(collPath)

	return nil
}

// SetCollectionLabel updates a collection's label.
func (s *gopassStore) SetCollectionLabel(ctx context.Context, name, label string) error {
	existing, err := s.GetCollection(ctx, name)
	if err != nil {
		return err
	}

	sec := secrets.New()
	sec.SetPassword("collection-metadata")
	for _, kv := range []struct{ k, v string }{
		{collLabelKey, label},
		{collCreatedKey, existing.Created.Format(time.RFC3339)},
		{collModifiedKey, time.Now().Format(time.RFC3339)},
	} {
		if err := sec.Set(kv.k, kv.v); err != nil {
			return fmt.Errorf("set %s: %w", kv.k, err)
		}
	}

	metaPath := s.mapper.MetaPath(name)
	if err := s.store.Set(ctx, metaPath, sec); err != nil {
		return err
	}
	s.invalidateMeta(metaPath)

	return nil
}

// Items returns all item IDs within a collection.
func (s *gopassStore) Items(ctx context.Context, collection string) ([]string, error) {
	allPaths, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}

	collPath := s.mapper.CollectionPath(collection)

	var items []string
	for _, p := range allPaths {
		if !strings.HasPrefix(p, collPath+"/") {
			continue
		}
		id := strings.TrimPrefix(p, collPath+"/")
		if id == "" || strings.Contains(id, "/") || strings.HasPrefix(id, "_") {
			continue
		}
		items = append(items, id)
	}
	sort.Strings(items)

	return items, nil
}

// GetItem returns an item, decrypting its secret fresh (never cached).
func (s *gopassStore) GetItem(ctx context.Context, collection, id string) (*ItemData, error) {
	itemPath := s.mapper.ItemPath(collection, id)

	sec, err := s.store.Get(ctx, itemPath, "latest")
	if err != nil {
		return nil, fmt.Errorf("item not found: %s/%s", collection, id)
	}

	meta := metaFromSecret(sec)
	s.cacheMu.Lock()
	s.metaCache[itemPath] = meta
	s.cacheMu.Unlock()

	item := &ItemData{
		ID:          id,
		Secret:      []byte(sec.Password()),
		ContentType: "text/plain",
		Attributes:  make(map[string]string),
	}
	applyItemMeta(item, meta)

	return item, nil
}

// CreateItem writes a new item and returns its generated ID.
func (s *gopassStore) CreateItem(ctx context.Context, collection string, item *ItemData) (string, error) {
	if item.ID == "" {
		id, err := newID('i')
		if err != nil {
			return "", err
		}
		item.ID = id
	}

	if _, err := s.GetCollection(ctx, collection); err != nil {
		if cerr := s.CreateCollection(ctx, collection, collection); cerr != nil {
			return "", fmt.Errorf("create collection: %w", cerr)
		}
	}

	now := time.Now()
	if item.Created.IsZero() {
		item.Created = now
	}
	item.Modified = now
	if item.ContentType == "" {
		item.ContentType = "text/plain"
	}

	if err := s.writeItem(ctx, collection, item); err != nil {
		return "", err
	}

	return item.ID, nil
}

// UpdateItem overwrites an existing item, preserving its creation time.
func (s *gopassStore) UpdateItem(ctx context.Context, collection, id string, item *ItemData) error {
	existing, err := s.GetItem(ctx, collection, id)
	if err != nil {
		return err
	}

	item.ID = id
	item.Created = existing.Created
	item.Modified = time.Now()
	if item.ContentType == "" {
		item.ContentType = existing.ContentType
	}
	if item.ContentType == "" {
		item.ContentType = "text/plain"
	}

	return s.writeItem(ctx, collection, item)
}

// writeItem serializes an ItemData into a gopass secret.
func (s *gopassStore) writeItem(ctx context.Context, collection string, item *ItemData) error {
	sec := secrets.New()
	sec.SetPassword(string(item.Secret))
	for _, kv := range []struct{ k, v string }{
		{labelKey, item.Label},
		{createdKey, item.Created.Format(time.RFC3339)},
		{modifiedKey, item.Modified.Format(time.RFC3339)},
		{contentTypeKey, item.ContentType},
	} {
		if err := sec.Set(kv.k, kv.v); err != nil {
			return fmt.Errorf("set %s: %w", kv.k, err)
		}
	}

	keys := make([]string, 0, len(item.Attributes))
	for k := range item.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := sec.Set(k, item.Attributes[k]); err != nil {
			return fmt.Errorf("set attribute %s: %w", k, err)
		}
	}

	itemPath := s.mapper.ItemPath(collection, item.ID)
	if err := s.store.Set(ctx, itemPath, sec); err != nil {
		return err
	}
	s.invalidateMeta(itemPath)

	return nil
}

// DeleteItem removes an item.
func (s *gopassStore) DeleteItem(ctx context.Context, collection, id string) error {
	itemPath := s.mapper.ItemPath(collection, id)
	if err := s.store.Remove(ctx, itemPath); err != nil {
		return err
	}
	s.invalidateMeta(itemPath)

	return nil
}

// SearchItems returns items in a collection matching all given attributes.
func (s *gopassStore) SearchItems(ctx context.Context, collection string, attrs map[string]string) ([]*ItemData, error) {
	ids, err := s.Items(ctx, collection)
	if err != nil {
		return nil, err
	}

	var results []*ItemData
	for _, id := range ids {
		meta, err := s.metaFor(ctx, s.mapper.ItemPath(collection, id))
		if err != nil {
			debug.Log("secret-service: skipping item %s/%s: %s", collection, id, err)

			continue
		}
		item := &ItemData{ID: id, ContentType: "text/plain", Attributes: make(map[string]string)}
		applyItemMeta(item, meta)
		if matchesAttributes(item, attrs) {
			results = append(results, item)
		}
	}

	debug.Log("secret-service: search %s for %v matched %d/%d", collection, attrs, len(results), len(ids))

	return results, nil
}

// SearchAllItems searches every collection for matching items.
func (s *gopassStore) SearchAllItems(ctx context.Context, attrs map[string]string) (map[string][]*ItemData, error) {
	collections, err := s.Collections(ctx)
	if err != nil {
		return nil, err
	}

	results := make(map[string][]*ItemData)
	for _, coll := range collections {
		items, err := s.SearchItems(ctx, coll, attrs)
		if err != nil {
			continue
		}
		if len(items) > 0 {
			results[coll] = items
		}
	}

	return results, nil
}

// LockCollection marks a collection locked in memory.
func (s *gopassStore) LockCollection(_ context.Context, name string) error {
	s.mu.Lock()
	s.locked[name] = true
	s.mu.Unlock()

	return nil
}

// UnlockCollection marks a collection unlocked in memory.
func (s *gopassStore) UnlockCollection(_ context.Context, name string) error {
	s.mu.Lock()
	s.locked[name] = false
	s.mu.Unlock()

	return nil
}

func (s *gopassStore) isLocked(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.locked[name]
}

// GetAlias resolves a collection alias. The "default" alias always exists.
func (s *gopassStore) GetAlias(ctx context.Context, alias string) (string, error) {
	sec, err := s.store.Get(ctx, s.mapper.AliasesPath(), "latest")
	if err != nil {
		if alias == "default" {
			return "default", nil
		}

		return "", fmt.Errorf("alias not found: %s", alias)
	}

	if val, ok := sec.Get(alias); ok && val != "" {
		return val, nil
	}
	if alias == "default" {
		return "default", nil
	}

	return "", fmt.Errorf("alias not found: %s", alias)
}

// SetAlias sets or removes (empty collection) a collection alias.
func (s *gopassStore) SetAlias(ctx context.Context, alias, collection string) error {
	aliases, err := s.Aliases(ctx)
	if err != nil {
		aliases = make(map[string]string)
	}

	if collection == "" {
		delete(aliases, alias)
	} else {
		aliases[alias] = collection
	}

	sec := secrets.New()
	sec.SetPassword("aliases")
	for k, v := range aliases {
		if err := sec.Set(k, v); err != nil {
			return fmt.Errorf("set alias %q: %w", k, err)
		}
	}

	return s.store.Set(ctx, s.mapper.AliasesPath(), sec)
}

// Aliases returns the full alias map. The "default" alias is always present.
func (s *gopassStore) Aliases(ctx context.Context) (map[string]string, error) {
	aliases := make(map[string]string)

	sec, err := s.store.Get(ctx, s.mapper.AliasesPath(), "latest")
	if err == nil {
		for _, key := range sec.Keys() {
			if val, ok := sec.Get(key); ok && val != "" {
				aliases[key] = val
			}
		}
	}

	if _, ok := aliases["default"]; !ok {
		aliases["default"] = "default"
	}

	return aliases, nil
}

// Close closes the underlying store.
func (s *gopassStore) Close(ctx context.Context) error {
	return s.store.Close(ctx)
}

// matchesAttributes reports whether item has every key/value in attrs.
func matchesAttributes(item *ItemData, attrs map[string]string) bool {
	for k, v := range attrs {
		if item.Attributes[k] != v {
			return false
		}
	}

	return true
}
