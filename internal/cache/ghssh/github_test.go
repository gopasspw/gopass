package ghssh

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListKeys(t *testing.T) {
	// Set GOPASS_HOMEDIR to a temp directory
	tempDir := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", tempDir)

	// Mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/validuser.keys" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ssh-rsa AAAAB3Nza... validuser@github\n")) //nolint:errcheck
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	oURL := baseURL
	baseURL = mockServer.URL
	defer func() {
		baseURL = oURL
	}()

	cache, err := New()
	require.NoError(t, err)

	t.Run("valid user", func(t *testing.T) {
		keys, err := cache.ListKeys(context.Background(), "validuser")
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, "ssh-rsa AAAAB3Nza... validuser@github", keys[0])
	})

	t.Run("invalid user", func(t *testing.T) {
		keys, err := cache.ListKeys(context.Background(), "invaliduser")
		require.Error(t, err)
		assert.Nil(t, keys)
	})
}

func TestFetchKeys(t *testing.T) {
	// Set GOPASS_HOMEDIR to a temp directory
	tempDir := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", tempDir)

	// Mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/validuser.keys" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ssh-rsa AAAAB3Nza... validuser@github\n")) //nolint:errcheck
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	oURL := baseURL
	baseURL = mockServer.URL
	defer func() {
		baseURL = oURL
	}()

	cache, err := New()
	require.NoError(t, err)

	t.Run("valid user", func(t *testing.T) {
		keys, err := cache.fetchKeys(context.Background(), "validuser")
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, "ssh-rsa AAAAB3Nza... validuser@github", keys[0])
	})

	t.Run("invalid user", func(t *testing.T) {
		keys, err := cache.fetchKeys(context.Background(), "invaliduser")
		require.Error(t, err)
		assert.Nil(t, keys)
	})
}
