package backend

import (
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
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
xc-gitmock-consul+http://localhost:8500/v1/foo/bar
xc-gitmock-consul+https://localhost:8500/v1/foo/bar
file:///tmp/foo -> gpgcli, gitcli, fs (using defaults)
/tmp/foo -> gpgcli, gitcli, fs (using defaults)

*/

func TestURLString(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   *URL
		out  string
	}{
		{
			name: "defaults",
			in:   &URL{},
			out:  "gpgmock-gitmock-fs+file:",
		},
		{
			name: "xc+gogit",
			in: &URL{
				Crypto: XC,
				Sync:   GoGit,
				Store:  FS,
				Path:   "/tmp/foo",
			},
			out: "xc-gogit-fs+file:///tmp/foo",
		},
	} {
		assert.Equal(t, tc.out, tc.in.String(), tc.name)
	}
}

func TestParseScheme(t *testing.T) {
	for _, tc := range []struct {
		Name   string
		URL    string
		Crypto CryptoBackend
		Sync   SyncBackend
		Store  StoreBackend
	}{
		{
			Name:   "legacy file path",
			URL:    "/tmp/foo",
			Crypto: GPGCLI,
			Sync:   GitCLI,
			Store:  FS,
		},
		{
			Name:   "XC+gogit+file",
			URL:    "xc-gogit-fs+file:///tmp/foo",
			Crypto: XC,
			Sync:   GoGit,
			Store:  FS,
		},
		//{
		//	URL:    "xc+consul://localhost:8500/api/v1/foo/bar?token=bla",
		//	Crypto: XC,
		//	Sync:   GitMock,
		//	Store:  Consul,
		//},
		//{
		//	URL: "vaults://localhost:9600/foo/bar",
		//	Crypto: GPGMock,
		//	Sync: GitMock,
		//	Store: Vault,
		//}
	} {
		u, err := ParseURL(tc.URL)
		assert.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, tc.Crypto, u.Crypto)
		assert.Equal(t, tc.Sync, u.Sync)
		assert.Equal(t, tc.Store, u.Store)
	}
}

type testConfig struct {
	Path *URL `yaml:"path"`
}

func TestUnmarshalYAML(t *testing.T) {
	in := `---
path: xc-gogit-fs+file:///tmp/foo
`
	cfg := testConfig{}
	assert.NoError(t, yaml.Unmarshal([]byte(in), &cfg))
	assert.Equal(t, "/tmp/foo", cfg.Path.Path)
}

func TestMarshalYAML(t *testing.T) {
	out := `path: xc-gogit-fs+file:///tmp/foo
`
	cfg := testConfig{
		Path: &URL{
			Crypto: XC,
			Sync:   GoGit,
			Store:  FS,
			Path:   "/tmp/foo",
		},
	}
	buf, err := yaml.Marshal(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, out, string(buf))
}
