package backend_test

import (
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestUnmarshalYAMLWindows(t *testing.T) {
	in := `---
path: gpgcli-gitcli-fs+file:///C:\tmp\foo
`
	cfg := testConfig{}
	require.NoError(t, yaml.Unmarshal([]byte(in), &cfg))
	assert.Equal(t, "C:\\tmp\\foo", cfg.Path.Path)
}

func TestParseWindows(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		URL     string
		Crypto  backend.CryptoBackend
		RCS     backend.RCSBackend
		Storage backend.StorageBackend
		Path    string
	}{
		{
			Name:    "windows store path",
			URL:     `C:\Users\johndoe\.password-store-my-team`,
			Crypto:  backend.GPGCLI,
			RCS:     backend.GitCLI,
			Storage: backend.FS,
			Path:    `C:\Users\johndoe\.password-store-my-team`,
		},
		{
			Name:    "windows store path with whitespace",
			URL:     `C:\Users\johndoe\My Folder\.password-store`,
			Crypto:  backend.GPGCLI,
			RCS:     backend.GitCLI,
			Storage: backend.FS,
			Path:    `C:\Users\johndoe\My Folder\.password-store`,
		},
		{
			Name:    "file scheme and windows abs path",
			URL:     `file:///C:\Users\johndoe`,
			Crypto:  backend.GPGCLI,
			RCS:     backend.GitCLI,
			Storage: backend.FS,
			Path:    `C:\Users\johndoe`,
		},
	} {
		u, err := backend.ParseURL(tc.URL)
		require.NoError(t, err, tc.Name)
		require.NotNil(t, u)
		assert.NotNil(t, u, tc.Name)
		assert.Equal(t, tc.Crypto, u.Crypto, tc.Name)
		assert.Equal(t, tc.RCS, u.RCS, tc.Name)
		assert.Equal(t, tc.Storage, u.Storage, tc.Name)
		if tc.Path != "" {
			assert.Equal(t, tc.Path, u.Path, tc.Name)
		}
	}
}
