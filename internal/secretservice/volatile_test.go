//go:build linux

package secretservice

import (
	"context"
	"strings"
	"testing"
)

func TestVaultRoundTrip(t *testing.T) {
	// Exercises whichever vault the environment provides (kernel keyring or
	// the in-memory fallback).
	v := newVault()
	defer v.close()

	if err := v.put("i1", []byte("hunter2")); err != nil {
		t.Fatalf("put: %v", err)
	}

	got, err := v.get("i1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got) != "hunter2" {
		t.Fatalf("get = %q, want %q", got, "hunter2")
	}

	// Overwriting must replace the previous payload.
	if err := v.put("i1", []byte("new")); err != nil {
		t.Fatalf("put(overwrite): %v", err)
	}
	got, _ = v.get("i1")
	if string(got) != "new" {
		t.Fatalf("get after overwrite = %q, want %q", got, "new")
	}

	v.remove("i1")
	if _, err := v.get("i1"); err == nil {
		t.Fatal("get after remove succeeded, want error")
	}
}

func TestKeyringStoreItemLifecycle(t *testing.T) {
	ctx := context.Background()
	s := newKeyringStore()
	defer func() { _ = s.Close(ctx) }()

	if got, err := s.Collections(ctx); err != nil || len(got) != 1 || got[0] != SessionCollectionName {
		t.Fatalf("Collections = %v, %v; want [session], nil", got, err)
	}
	if _, err := s.GetCollection(ctx, SessionCollectionName); err != nil {
		t.Fatalf("GetCollection(session): %v", err)
	}

	id, err := s.CreateItem(ctx, SessionCollectionName, &ItemData{
		Secret:     []byte("transient"),
		Label:      "Temp",
		Attributes: map[string]string{"scope": "test"},
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if !strings.HasPrefix(id, "i") {
		t.Fatalf("id = %q, want i-prefixed", id)
	}

	got, err := s.GetItem(ctx, SessionCollectionName, id)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if string(got.Secret) != "transient" {
		t.Fatalf("secret = %q, want %q", got.Secret, "transient")
	}
	if got.ContentType != "text/plain" {
		t.Fatalf("content type = %q, want text/plain", got.ContentType)
	}

	hits, err := s.SearchItems(ctx, SessionCollectionName, map[string]string{"scope": "test"})
	if err != nil {
		t.Fatalf("SearchItems: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != id {
		t.Fatalf("search hits = %+v, want the created item", hits)
	}

	// Update replaces the payload and preserves the creation time.
	got.Secret = []byte("updated")
	if err := s.UpdateItem(ctx, SessionCollectionName, id, got); err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	got, _ = s.GetItem(ctx, SessionCollectionName, id)
	if string(got.Secret) != "updated" {
		t.Fatalf("secret after update = %q, want %q", got.Secret, "updated")
	}

	if err := s.DeleteItem(ctx, SessionCollectionName, id); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	if _, err := s.GetItem(ctx, SessionCollectionName, id); err == nil {
		t.Fatal("GetItem after delete succeeded, want error")
	}
}

func TestKeyringStoreImmutability(t *testing.T) {
	ctx := context.Background()
	s := newKeyringStore()
	defer func() { _ = s.Close(ctx) }()

	if err := s.CreateCollection(ctx, SessionCollectionName, "x"); err == nil {
		t.Fatal("CreateCollection(session) succeeded, want error")
	}
	if err := s.DeleteCollection(ctx, SessionCollectionName); err == nil {
		t.Fatal("DeleteCollection(session) succeeded, want error")
	}
	if err := s.SetAlias(ctx, "session", "other"); err == nil {
		t.Fatal("SetAlias(session) succeeded, want error")
	}
	if _, err := s.GetCollection(ctx, "other"); err == nil {
		t.Fatal("GetCollection(other) succeeded, want error")
	}
}
