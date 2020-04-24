// +build xc
// +build gitcli

package backend

import (
	"path/filepath"
	"runtime"
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
gpgcli-gitcli-fs+file:///tmp/foo
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
			name: "xc+gitcli",
			in: &URL{
				Crypto:  XC,
				RCS:     gitcli,
				Storage: FS,
				Path:    "/tmp/foo",
			},
			out: "xc-gitcli-fs+file:///tmp/foo",
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
			URL:     "xc-gitcli-fs+file:///tmp/foo",
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
		//{
		//	URL:     "plain+vault-http://localhost:9600/foo/bar",
		//	Crypto:  Plain,
		//	RCS:     Noop,
		//	Storage: Vault,
		//},
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
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	in := `---
path: xc-gitcli-fs+file:///tmp/foo
`
	cfg := testConfig{}
	require.NoError(t, yaml.Unmarshal([]byte(in), &cfg))
	assert.Equal(t, "/tmp/foo", cfg.Path.Path)
}

func TestMarshalYAMLXC(t *testing.T) {
	out := `path: xc-gitcli-fs+file:///tmp/foo
`
	cfg := testConfig{
		Path: &URL{
			Crypto:  XC,
			RCS:     GitCLI,
			Storage: FS,
			Path:    "/tmp/foo",
		},
	}
	buf, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, out, string(buf))
}
