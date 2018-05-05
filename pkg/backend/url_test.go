package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		in   *URL
		out  string
	}{
		{
			name: "defaults",
			in:   &URL{},
			out:  "plain-noop-fs+file:",
		},
		{
			name: "xc+gogit",
			in: &URL{
				Crypto:  XC,
				RCS:     GoGit,
				Storage: FS,
				Path:    "/tmp/foo",
			},
			out: "xc-gogit-fs+file:///tmp/foo",
		},
		{
			name: "xc+consul",
			in: &URL{
				Crypto:  XC,
				RCS:     Noop,
				Storage: Consul,
				Scheme:  "http",
				Host:    "localhost",
				Port:    "8500",
				Path:    "/foo/bar",
			},
			out: "xc-noop-consul+http://localhost:8500/foo/bar",
		},
	} {
		assert.Equal(t, tc.out, tc.in.String(), tc.name)
	}
}

func TestParseScheme(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		URL     string
		Crypto  CryptoBackend
		RCS     RCSBackend
		Storage StorageBackend
	}{
		{
			Name:    "legacy file path",
			URL:     "/tmp/foo",
			Crypto:  GPGCLI,
			RCS:     GitCLI,
			Storage: FS,
		},
		{
			Name:    "XC+gogit+file",
			URL:     "xc-gogit-fs+file:///tmp/foo",
			Crypto:  XC,
			RCS:     GoGit,
			Storage: FS,
		},
		{
			Name:    "XC+consul+http",
			URL:     "xc-noop-consul+http://localhost:8500/api/v1/foo/bar?token=bla",
			Crypto:  XC,
			RCS:     Noop,
			Storage: Consul,
		},
		//{
		//	URL:     "plain+vault-http://localhost:9600/foo/bar",
		//	Crypto:  Plain,
		//	RCS:     Noop,
		//	Storage: Vault,
		//},
	} {
		u, err := ParseURL(tc.URL)
		assert.NoError(t, err, tc.Name)
		assert.NotNil(t, u, tc.Name)
		assert.Equal(t, tc.Crypto, u.Crypto, tc.Name)
		assert.Equal(t, tc.RCS, u.RCS, tc.Name)
		assert.Equal(t, tc.Storage, u.Storage, tc.Name)
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
			Crypto:  XC,
			RCS:     GoGit,
			Storage: FS,
			Path:    "/tmp/foo",
		},
	}
	buf, err := yaml.Marshal(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, out, string(buf))
}
