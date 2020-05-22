// +build windows
// +build xc
// +build gitcli

package backend

import (
	"path/filepath"
	"testing"

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
gpgcli-gitcli-fs+file:///C:\\Users\johndoe
xc-noop-consul+http://localhost:8500/v1/foo/bar
xc-noop-consul+https://localhost:8500/v1/foo/bar
file:///tmp/foo -> gpgcli, gitcli, fs (using defaults)
/tmp/foo -> gpgcli, gitcli, fs (using defaults)
*/

func TestURLStringXC(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   *URL
		out  string
	}{
		{
			name: "xc+gogit",
			in: &URL{
				Crypto:  XC,
				RCS:     GitCLI,
				Storage: FS,
				Path:    "C:\\tmp\\foo",
			},
			out: `xc-gogit-fs+file:///C:\tmp\foo`,
		},
	} {
		assert.Equal(t, tc.out, tc.in.String(), tc.name)
	}
}

func TestParseSchemeXC(t *testing.T) {
	hd, err := homedir.Dir()
	require.NoError(t, err)
	for _, tc := range []struct {
		Name    string
		URL     string
		Crypto  CryptoBackend
		RCS     RCSBackend
		Storage StorageBackend
		Path    string
	}{
		{
			Name:    "XC+gitcli+file",
			URL:     "xc-gitcli-fs+file:///C:\\tmp\\foo",
			Crypto:  XC,
			RCS:     gitcli,
			Storage: FS,
		},
		{
			Name:    "Homedir expansion",
			URL:     "gpgcli-gitcli-fs+file://~/.local/share/password-store",
			Crypto:  GPGCLI,
			RCS:     GitCLI,
			Storage: FS,
			Path:    filepath.Join(hd, ".local", "share", "password-store"),
		},
	} {
		u, err := ParseURL(tc.URL)
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
	Path *URL `yaml:"path"`
}

func TestUnmarshalYAMLXC(t *testing.T) {
	in := `---
path: xc-gitcli-fs+file:///C:\Users\johndoe
`
	cfg := testConfig{}
	require.NoError(t, yaml.Unmarshal([]byte(in), &cfg))
	assert.Equal(t, "C:\\Users\\johndoe", cfg.Path.Path)
}

func TestMarshalYAMLXC(t *testing.T) {
	out := `path: xc-gitcli-fs+file:///C:\Users\johndoe
`
	cfg := testConfig{
		Path: &URL{
			Crypto:  XC,
			RCS:     GitCLI,
			Storage: FS,
			Path:    "C:\\Users\\johndoe",
		},
	}
	buf, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, out, string(buf))
}
