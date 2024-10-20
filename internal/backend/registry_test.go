package backend_test

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCryptoLoaderHighPrio struct{}

func (l fakeCryptoLoaderHighPrio) New(context.Context) (backend.Crypto, error) {
	return plain.New(), nil
}

func (l fakeCryptoLoaderHighPrio) String() string {
	return "fakeCryptoLoaderHighPrio"
}

func (l fakeCryptoLoaderHighPrio) Handles(context.Context, backend.Storage) error {
	return nil
}

func (l fakeCryptoLoaderHighPrio) Priority() int {
	return 2
}

type fakeCryptoLoaderLowPrio struct{}

func (l fakeCryptoLoaderLowPrio) New(context.Context) (backend.Crypto, error) {
	return plain.New(), nil
}

func (l fakeCryptoLoaderLowPrio) String() string {
	return "fakeCryptoLoaderLowPrio"
}

func (l fakeCryptoLoaderLowPrio) Handles(context.Context, backend.Storage) error {
	return nil
}

func (l fakeCryptoLoaderLowPrio) Priority() int {
	return 1
}

func TestCryptoLoader(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	backend.CryptoRegistry.Register(backend.Plain, "plain", fakeCryptoLoaderHighPrio{})
	c, err := backend.NewCrypto(ctx, backend.Plain)
	require.NoError(t, err)
	assert.Equal(t, "plain", c.Name())
}

func TestRegistry_BackendNames(t *testing.T) {
	t.Parallel()

	registry := backend.NewRegistry[backend.CryptoBackend, backend.CryptoLoader]()
	registry.Register(backend.Plain, "plain", fakeCryptoLoaderHighPrio{})
	registry.Register(backend.GPGCLI, "gpgcli", fakeCryptoLoaderHighPrio{})
	registry.Register(backend.Age, "age", fakeCryptoLoaderHighPrio{})

	expected := []string{"age", "gpgcli", "plain"}
	actual := registry.BackendNames()
	assert.Equal(t, expected, actual, "backend names should be sorted")
}

func TestRegistry_Backends(t *testing.T) {
	t.Parallel()

	registry := backend.NewRegistry[backend.CryptoBackend, backend.CryptoLoader]()
	registry.Register(backend.Plain, "plain", fakeCryptoLoaderHighPrio{})
	registry.Register(backend.GPGCLI, "gpgcli", fakeCryptoLoaderHighPrio{})
	registry.Register(backend.Age, "age", fakeCryptoLoaderLowPrio{})

	// iteration order of map is random, so it's hard to test the actual content
	assert.Len(t, registry.Backends(), 3, "should return all registered backend loaders")
}

func TestRegistry_Prioritized(t *testing.T) {
	t.Parallel()

	highPrio := fakeCryptoLoaderHighPrio{}
	lowPrio := fakeCryptoLoaderLowPrio{}

	registry := backend.NewRegistry[backend.CryptoBackend, backend.CryptoLoader]()
	registry.Register(backend.Plain, "plain", highPrio)
	registry.Register(backend.GPGCLI, "gpgcli", lowPrio)

	expected := []backend.CryptoLoader{lowPrio, highPrio}
	actual := registry.Prioritized()
	assert.Equal(t, expected, actual, "should return in ascending priority order")
}

func TestRegistry_Get(t *testing.T) {
	t.Parallel()

	loader := fakeCryptoLoaderHighPrio{}
	registry := backend.NewRegistry[backend.CryptoBackend, backend.CryptoLoader]()
	registry.Register(backend.Plain, "plain", loader)

	tests := map[string]struct {
		backend backend.CryptoBackend
		want    backend.CryptoLoader
		wantErr error
	}{
		"backend exists": {
			backend: backend.Plain,
			want:    loader,
			wantErr: nil,
		},
		"backend does not exist": {
			backend: backend.GPGCLI,
			want:    nil,
			wantErr: backend.ErrNotFound,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			v, err := registry.Get(tt.backend)
			assert.Equal(t, tt.want, v)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestRegistry_Backend(t *testing.T) {
	t.Parallel()

	loader := fakeCryptoLoaderHighPrio{}
	registry := backend.NewRegistry[backend.CryptoBackend, backend.CryptoLoader]()
	registry.Register(backend.GPGCLI, "gpgcli", loader)

	tests := map[string]struct {
		backendName string
		want        backend.CryptoBackend
		wantErr     error
	}{
		"backend name exists": {
			backendName: "gpgcli",
			want:        backend.GPGCLI,
			wantErr:     nil,
		},
		"backend name does not exist": {
			backendName: "fake",
			want:        0, // zero value
			wantErr:     backend.ErrNotFound,
		},
		`special case: "gpg" name should be handled as "gpgcli"`: {
			backendName: "gpg",
			want:        backend.GPGCLI,
			wantErr:     nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			v, err := registry.Backend(tt.backendName)
			assert.Equal(t, tt.want, v)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestRegistry_BackendName(t *testing.T) {
	t.Parallel()

	registry := backend.NewRegistry[backend.CryptoBackend, backend.CryptoLoader]()
	registry.Register(backend.Plain, "plain", fakeCryptoLoaderHighPrio{})

	tests := map[string]struct {
		backend backend.CryptoBackend
		want    string
		wantErr error
	}{
		"backend exists": {
			backend: backend.Plain,
			want:    "plain",
			wantErr: nil,
		},
		"backend does not exist": {
			backend: backend.GPGCLI,
			want:    "", // zero value
			wantErr: backend.ErrNotFound,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			v, err := registry.BackendName(tt.backend)
			assert.Equal(t, tt.want, v)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
