//go:build linux

package secretservice

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
)

// fakeStore is a minimal in-memory gopass.Store used to test gopassStore
// without GPG or a real password store.
type fakeStore struct {
	entries map[string]gopass.Secret
}

func newFakeStore() *fakeStore {
	return &fakeStore{entries: make(map[string]gopass.Secret)}
}

func (f *fakeStore) String() string { return "fake" }

func (f *fakeStore) List(context.Context) ([]string, error) {
	out := make([]string, 0, len(f.entries))
	for k := range f.entries {
		out = append(out, k)
	}

	return out, nil
}

func (f *fakeStore) AuditList(ctx context.Context) ([]string, error) { return f.List(ctx) }

func (f *fakeStore) Get(_ context.Context, name, _ string) (gopass.Secret, error) {
	sec, ok := f.entries[name]
	if !ok {
		return nil, store.ErrNotFound
	}

	return sec, nil
}

func (f *fakeStore) Set(_ context.Context, name string, sec gopass.Byter) error {
	s, ok := sec.(gopass.Secret)
	if !ok {
		return fmt.Errorf("not a secret: %T", sec)
	}
	// Round-trip through the line-oriented AKV format, exactly as the real
	// store does, so tests catch values that do not survive serialization.
	f.entries[name] = secrets.ParseAKV(s.Bytes())

	return nil
}

func (f *fakeStore) Revisions(context.Context, string) ([]string, error) { return nil, nil }

func (f *fakeStore) Remove(_ context.Context, name string) error {
	if _, ok := f.entries[name]; !ok {
		return fmt.Errorf("not found: %s", name)
	}
	delete(f.entries, name)

	return nil
}

func (f *fakeStore) RemoveAll(_ context.Context, prefix string) error {
	for k := range f.entries {
		if k == prefix || strings.HasPrefix(k, prefix+"/") {
			delete(f.entries, k)
		}
	}

	return nil
}

func (f *fakeStore) Rename(context.Context, string, string) error { return nil }

func (f *fakeStore) Sync(context.Context) error { return nil }

func (f *fakeStore) Close(context.Context) error { return nil }

func newTestStore() *gopassStore {
	return newGopassStoreWithBackend(newFakeStore(), "secret-service")
}

func TestStoreItemLifecycle(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	if err := s.CreateCollection(ctx, "default", "Default"); err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret:      []byte("s3cr3t"),
		Label:       "GitHub",
		ContentType: "text/plain",
		Attributes:  map[string]string{"username": "octocat", "service": "github.com"},
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if !strings.HasPrefix(id, "i") || strings.Contains(id, "-") {
		t.Fatalf("id = %q, want i-prefixed hyphen-free", id)
	}

	// GetSecret round-trips the password and attributes.
	got, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if string(got.Secret) != "s3cr3t" {
		t.Fatalf("secret = %q, want %q", got.Secret, "s3cr3t")
	}
	if got.Label != "GitHub" {
		t.Fatalf("label = %q, want %q", got.Label, "GitHub")
	}
	if got.Attributes["username"] != "octocat" {
		t.Fatalf("username = %q, want %q", got.Attributes["username"], "octocat")
	}
}

func TestStoreSearchAndCacheInvalidation(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	if err := s.CreateCollection(ctx, "default", "Default"); err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret:     []byte("pw"),
		Label:      "One",
		Attributes: map[string]string{"app": "mail"},
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	hits, err := s.SearchItems(ctx, "default", map[string]string{"app": "mail"})
	if err != nil {
		t.Fatalf("SearchItems: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != id {
		t.Fatalf("search hits = %+v, want the created item", hits)
	}

	// A non-matching query returns nothing.
	hits, _ = s.SearchItems(ctx, "default", map[string]string{"app": "chat"})
	if len(hits) != 0 {
		t.Fatalf("non-matching search returned %d hits, want 0", len(hits))
	}

	// Updating attributes must invalidate the metadata cache.
	upd, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	upd.Attributes = map[string]string{"app": "chat"}
	if err := s.UpdateItem(ctx, "default", id, upd); err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}

	hits, _ = s.SearchItems(ctx, "default", map[string]string{"app": "chat"})
	if len(hits) != 1 {
		t.Fatalf("post-update search returned %d hits, want 1", len(hits))
	}
	hits, _ = s.SearchItems(ctx, "default", map[string]string{"app": "mail"})
	if len(hits) != 0 {
		t.Fatalf("stale search returned %d hits, want 0", len(hits))
	}
}

func TestStoreAliases(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	if got, err := s.GetAlias(ctx, "default"); err != nil || got != "default" {
		t.Fatalf("default alias = %q, %v; want default, nil", got, err)
	}

	if err := s.SetAlias(ctx, "login", "default"); err != nil {
		t.Fatalf("SetAlias: %v", err)
	}
	if got, err := s.GetAlias(ctx, "login"); err != nil || got != "default" {
		t.Fatalf("login alias = %q, %v; want default, nil", got, err)
	}

	if err := s.SetAlias(ctx, "login", ""); err != nil {
		t.Fatalf("SetAlias(remove): %v", err)
	}
	if _, err := s.GetAlias(ctx, "login"); err == nil {
		t.Fatal("login alias still resolves after removal")
	}
}

func TestStoreLockState(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	if err := s.CreateCollection(ctx, "default", "Default"); err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}
	if err := s.LockCollection(ctx, "default"); err != nil {
		t.Fatalf("LockCollection: %v", err)
	}
	if data, _ := s.GetCollection(ctx, "default"); !data.Locked {
		t.Fatal("collection not locked")
	}
	if err := s.UnlockCollection(ctx, "default"); err != nil {
		t.Fatalf("UnlockCollection: %v", err)
	}
	if data, _ := s.GetCollection(ctx, "default"); data.Locked {
		t.Fatal("collection still locked")
	}
}

func TestSanitizeName(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"default", "default"},
		{"a/b", "a_b"},
		{"a b", "a_b"},
		{".", "_"},
		{"..", "__"},
	} {
		if got := SanitizeName(tc.in); got != tc.want {
			t.Fatalf("SanitizeName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestStoreMetadataTimestamps(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	before := time.Now().Add(-time.Second)
	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret: []byte("pw"),
		Label:  "One",
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	got, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if got.Created.Before(before) {
		t.Fatalf("created = %v, want >= %v", got.Created, before)
	}
	if got.ContentType != "text/plain" {
		t.Fatalf("content type = %q, want text/plain", got.ContentType)
	}
}

func TestStoreBinarySafeSecret(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	// Arbitrary bytes, including newlines and a line that looks like a gopass
	// key/value pair, must survive the round-trip unchanged.
	payload := []byte("line1\nline2: looks-like-attribute\n\x00\xffbinary")

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret: payload,
		Label:  "Binary",
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	got, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if string(got.Secret) != string(payload) {
		t.Fatalf("secret = %q, want %q", got.Secret, payload)
	}
	if len(got.Attributes) != 0 {
		t.Fatalf("attributes = %v, want none", got.Attributes)
	}
}

func TestStoreSimpleSecretStaysCLIReadable(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret: []byte("s3cr3t-value"),
		Label:  "Simple",
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	// A simple single-line secret must live in the password field so that
	// `gopass show` displays it directly, not as a base64 blob.
	backend, ok := s.store.(*fakeStore)
	if !ok {
		t.Fatalf("unexpected backend type %T", s.store)
	}
	sec, err := backend.Get(ctx, s.mapper.ItemPath("default", id), "latest")
	if err != nil {
		t.Fatalf("backend Get: %v", err)
	}
	if sec.Password() != "s3cr3t-value" {
		t.Fatalf("password field = %q, want %q", sec.Password(), "s3cr3t-value")
	}
	if _, ok := sec.Get(secretKey); ok {
		t.Fatal("simple secret was base64 encoded unnecessarily")
	}
}

func TestStoreReservedAttributeEscaping(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret: []byte("pw"),
		Label:  "Real Label",
		Attributes: map[string]string{
			"_ss_label": "attacker",
			"service":   "example.com",
		},
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	got, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if got.Label != "Real Label" {
		t.Fatalf("label = %q, want %q (reserved attribute overwrote it)", got.Label, "Real Label")
	}
	if got.Attributes["_ss_label"] != "attacker" {
		t.Fatalf("_ss_label attribute = %q, want %q", got.Attributes["_ss_label"], "attacker")
	}
	if got.Attributes["service"] != "example.com" {
		t.Fatalf("service attribute = %q, want %q", got.Attributes["service"], "example.com")
	}
}

func TestStoreAliasConcurrentUpdates(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	const n = 20

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = s.SetAlias(ctx, fmt.Sprintf("alias%d", i), "default")
		}(i)
	}
	wg.Wait()

	aliases, err := s.Aliases(ctx)
	if err != nil {
		t.Fatalf("Aliases: %v", err)
	}
	for i := range n {
		key := fmt.Sprintf("alias%d", i)
		if aliases[key] != "default" {
			t.Fatalf("alias %q = %q, want default (lost update)", key, aliases[key])
		}
	}
}

func TestStoreAttributeRoundTripLossless(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	// Attribute names and values are arbitrary D-Bus strings. Names containing
	// the AKV separator and values containing newlines or the separator must
	// survive the round-trip unchanged.
	attrs := map[string]string{
		"a: b":       "line1\nline2",
		"service":    "example.com",
		"multi: key": "value: with: colons",
		"empty":      "",
	}

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret:     []byte("pw"),
		Label:      "Lossless",
		Attributes: attrs,
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	got, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if len(got.Attributes) != len(attrs) {
		t.Fatalf("attributes = %v, want %v", got.Attributes, attrs)
	}
	for k, v := range attrs {
		if got.Attributes[k] != v {
			t.Fatalf("attribute %q = %q, want %q", k, got.Attributes[k], v)
		}
	}

	// Search must match on the decoded values.
	hits, err := s.SearchItems(ctx, "default", map[string]string{"a: b": "line1\nline2"})
	if err != nil {
		t.Fatalf("SearchItems: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("search returned %d hits, want 1", len(hits))
	}
}

func TestStoreCollectionModifiedUpdatesOnLabelChange(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	if err := s.CreateCollection(ctx, "default", "Default"); err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}

	before, err := s.GetCollection(ctx, "default")
	if err != nil {
		t.Fatalf("GetCollection: %v", err)
	}

	// Timestamps are stored with RFC3339 second precision, so wait long
	// enough for the label change to land in a later second.
	time.Sleep(1100 * time.Millisecond)
	if err := s.SetCollectionLabel(ctx, "default", "Renamed"); err != nil {
		t.Fatalf("SetCollectionLabel: %v", err)
	}

	after, err := s.GetCollection(ctx, "default")
	if err != nil {
		t.Fatalf("GetCollection: %v", err)
	}
	if after.Label != "Renamed" {
		t.Fatalf("label = %q, want Renamed", after.Label)
	}
	if !after.Modified.After(before.Modified) {
		t.Fatalf("modified = %v, want after %v", after.Modified, before.Modified)
	}
}

func TestStoreLabelRoundTripLossless(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	// Labels are arbitrary D-Bus strings. A label containing a line break must
	// not be truncated or inject reserved fields on reload.
	label := "line1\n_ss_secret: injected"

	if err := s.CreateCollection(ctx, "default", label); err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}
	coll, err := s.GetCollection(ctx, "default")
	if err != nil {
		t.Fatalf("GetCollection: %v", err)
	}
	if coll.Label != label {
		t.Fatalf("collection label = %q, want %q", coll.Label, label)
	}

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret: []byte("pw"),
		Label:  label,
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	item, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if item.Label != label {
		t.Fatalf("item label = %q, want %q", item.Label, label)
	}
	if string(item.Secret) != "pw" {
		t.Fatalf("secret = %q, want pw", item.Secret)
	}
}

func TestStoreAliasRoundTripLossless(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	// Alias names are arbitrary D-Bus strings but are used as AKV keys, which
	// cannot contain the separator or a line break.
	alias := "a: b\nc"

	if err := s.SetAlias(ctx, alias, "default"); err != nil {
		t.Fatalf("SetAlias: %v", err)
	}
	got, err := s.GetAlias(ctx, alias)
	if err != nil {
		t.Fatalf("GetAlias: %v", err)
	}
	if got != "default" {
		t.Fatalf("alias = %q, want default", got)
	}

	aliases, err := s.Aliases(ctx)
	if err != nil {
		t.Fatalf("Aliases: %v", err)
	}
	if aliases[alias] != "default" {
		t.Fatalf("aliases[%q] = %q, want default", alias, aliases[alias])
	}
}

func TestStoreMarkerPrefixedValueRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	// A value that already starts with the base64 marker must be encoded too,
	// otherwise it would be decoded as if it were encoded data.
	value := attrValueEncodedPrefix + "YQ=="

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret:     []byte("pw"),
		Label:      "Marker",
		Attributes: map[string]string{"key": value},
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	got, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if got.Attributes["key"] != value {
		t.Fatalf("attribute = %q, want %q", got.Attributes["key"], value)
	}
}

func TestStoreFieldSpecificUpdatesPreserveOtherFields(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret:      []byte("original"),
		Label:       "Original",
		ContentType: "text/plain",
		Attributes:  map[string]string{"service": "example.com"},
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	// Updating the label must not disturb the secret or attributes.
	if err := s.SetItemLabel(ctx, "default", id, "Renamed"); err != nil {
		t.Fatalf("SetItemLabel: %v", err)
	}
	got, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if got.Label != "Renamed" {
		t.Fatalf("label = %q, want Renamed", got.Label)
	}
	if string(got.Secret) != "original" {
		t.Fatalf("secret = %q, want original", got.Secret)
	}
	if got.Attributes["service"] != "example.com" {
		t.Fatalf("attributes = %v, want service=example.com", got.Attributes)
	}

	// Updating the attributes must not disturb the label or secret.
	if err := s.SetItemAttributes(ctx, "default", id, map[string]string{"service": "other.com"}); err != nil {
		t.Fatalf("SetItemAttributes: %v", err)
	}
	got, err = s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if got.Label != "Renamed" {
		t.Fatalf("label = %q, want Renamed", got.Label)
	}
	if string(got.Secret) != "original" {
		t.Fatalf("secret = %q, want original", got.Secret)
	}
	if got.Attributes["service"] != "other.com" {
		t.Fatalf("attributes = %v, want service=other.com", got.Attributes)
	}

	// Updating the secret must not disturb the label or attributes.
	if err := s.SetItemSecret(ctx, "default", id, []byte("updated"), "text/plain"); err != nil {
		t.Fatalf("SetItemSecret: %v", err)
	}
	got, err = s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if got.Label != "Renamed" {
		t.Fatalf("label = %q, want Renamed", got.Label)
	}
	if string(got.Secret) != "updated" {
		t.Fatalf("secret = %q, want updated", got.Secret)
	}
	if got.Attributes["service"] != "other.com" {
		t.Fatalf("attributes = %v, want service=other.com", got.Attributes)
	}
}

func TestStoreConcurrentFieldUpdatesDoNotLoseData(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	id, err := s.CreateItem(ctx, "default", &ItemData{
		Secret:     []byte("original"),
		Label:      "Original",
		Attributes: map[string]string{"service": "example.com"},
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	// A label update and an attribute update racing on the same item must both
	// survive: a full read-modify-write of the whole item would revert one of
	// them.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = s.SetItemLabel(ctx, "default", id, "Renamed")
	}()
	go func() {
		defer wg.Done()
		_ = s.SetItemAttributes(ctx, "default", id, map[string]string{"service": "other.com"})
	}()
	wg.Wait()

	got, err := s.GetItem(ctx, "default", id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if got.Label != "Renamed" {
		t.Fatalf("label = %q, want Renamed (lost update)", got.Label)
	}
	if got.Attributes["service"] != "other.com" {
		t.Fatalf("attributes = %v, want service=other.com (lost update)", got.Attributes)
	}
	if string(got.Secret) != "original" {
		t.Fatalf("secret = %q, want original", got.Secret)
	}
}
