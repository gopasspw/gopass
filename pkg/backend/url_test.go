package backend_test

import (
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	_ "github.com/gopasspw/gopass/pkg/backend/crypto"
	_ "github.com/gopasspw/gopass/pkg/backend/rcs"
	_ "github.com/gopasspw/gopass/pkg/backend/storage"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

/*

url = [scheme:][//[userinfo@]host][/]path[?query][#fragment]
letter = "a" ... "z" ;
digit = "0" ... "9" ;
backend = letter , { letter | digit } ;
backends = backend , "-" , backend , "-" , backend
path = backends , "+" , url

- format (all mandatory)
crypto-sync-store+url

- examples
gpgcli-gitcli-fs+file:///tmp/foo
xc-noop-consul+http://localhost:8500/v1/foo/bar
xc-noop-consul+https://localhost:8500/v1/foo/bar
file:///tmp/foo -> gpgcli, gitcli, fs (using defaults)
/tmp/foo -> gpgcli, gitcli, fs (using defaults)

*/

func TestURLString(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   *backend.URL
		out  string
	}{
		{
			name: "defaults",
			in:   &backend.URL{},
			out:  "plain-noop-fs+file:",
		},
	} {
		assert.Equal(t, tc.out, tc.in.String(), tc.name)
	}
}

func TestParseScheme(t *testing.T) {
	hd, err := homedir.Dir()
	require.NoError(t, err)
	for _, tc := range []struct {
		Name    string
		URL     string
		Crypto  backend.CryptoBackend
		RCS     backend.RCSBackend
		Storage backend.StorageBackend
		Path    string
	}{
		{
			Name:    "legacy file path",
			URL:     "/tmp/foo",
			Crypto:  backend.GPGCLI,
			RCS:     backend.GitCLI,
			Storage: backend.FS,
		},
		{
			Name:    "Homedir expansion",
			URL:     "gpgcli-gitcli-fs+file://~/.local/share/password-store",
			Crypto:  backend.GPGCLI,
			RCS:     backend.GitCLI,
			Storage: backend.FS,
			Path:    filepath.Join(hd, ".local", "share", "password-store"),
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

type testConfig struct {
	Path *backend.URL `yaml:"path"`
}

func TestMarshalYAML(t *testing.T) {
	out := `path: plain-noop-fs+file:///tmp/foo
`
	cfg := testConfig{
		Path: &backend.URL{
			Crypto:  backend.Plain,
			RCS:     backend.Noop,
			Storage: backend.FS,
			Path:    "/tmp/foo",
		},
	}
	buf, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, out, string(buf))
}

func TestFromPath(t *testing.T) {
	assert.Equal(t, "gpgcli-gitcli-fs+file:///tmp", backend.FromPath("/tmp").String())
}
