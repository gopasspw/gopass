//go:build linux

package secretservice

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gopasspw/gopass/pkg/gopass"
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
		return nil, fmt.Errorf("not found: %s", name)
	}

	return sec, nil
}

func (f *fakeStore) Set(_ context.Context, name string, sec gopass.Byter) error {
	s, ok := sec.(gopass.Secret)
	if !ok {
		return fmt.Errorf("not a secret: %T", sec)
	}
	f.entries[name] = s

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
