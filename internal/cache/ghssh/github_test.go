package ghssh

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListKeys(t *testing.T) {
	// Set GOPASS_HOMEDIR to a temp directory
	tempDir := t.TempDir()
	os.Setenv("GOPASS_HOMEDIR", tempDir)
	defer os.Unsetenv("GOPASS_HOMEDIR")

	// Mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/validuser.keys" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ssh-rsa AAAAB3Nza... validuser@github\n"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Override the httpClient to use the mock server
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
			Proxy: func(req *http.Request) (*http.URL, error) {
				return http.ParseURL(mockServer.URL)
			},
		},
	}

	cache := &Cache{
		disk: NewDiskCache(tempDir),
	}

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
	os.Setenv("GOPASS_HOMEDIR", tempDir)
	defer os.Unsetenv("GOPASS_HOMEDIR")

	// Mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/validuser.keys" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ssh-rsa AAAAB3Nza... validuser@github\n"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Override the httpClient to use the mock server
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
			Proxy: func(req *http.Request) (*http.URL, error) {
				return http.ParseURL(mockServer.URL)
			},
		},
	}

	cache := &Cache{
		disk: NewDiskCache(tempDir),
	}

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
