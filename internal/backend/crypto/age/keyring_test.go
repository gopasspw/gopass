package age

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/mock"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte("test-password"), nil
	})

	storage := mock.New()
	storage.Set("test/.age-ids", []byte("test-id"))
	storage.Set(filepath.Join(appdir.UserConfig(), "age-keyring.age"), []byte("test-keyring"))

	err := migrate(ctx, storage)
	assert.NoError(t, err)

	_, err = storage.Get("test/.age-ids")
	assert.Error(t, err, "old ID file should be removed")

	_, err = storage.Get(filepath.Join(appdir.UserConfig(), "age-keyring.age"))
	assert.Error(t, err, "old keyring file should be removed")
}

func TestLoadIdentitiesFromKeyring(t *testing.T) {
	ctx := context.Background()
	a := &Age{
		identity: filepath.Join(appdir.UserConfig(), "age-keyring.age"),
	}

	// Create a mock keyring file
	keyring := Keyring{
		{Name: "test", Email: "test@example.com", Identity: "test-identity"},
	}
	data, err := json.Marshal(keyring)
	assert.NoError(t, err)

	err = os.WriteFile(a.identity, data, 0o600)
	assert.NoError(t, err)
	defer os.Remove(a.identity)

	ids, err := a.loadIdentitiesFromKeyring(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ids))
	assert.Equal(t, "test-identity", ids[0])
}
