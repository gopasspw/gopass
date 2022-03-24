package gpgconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGpgOpts(t *testing.T) { //nolint:paralleltest
	for _, vn := range []string{"GOPASS_GPG_OPTS", "PASSWORD_STORE_GPG_OPTS"} {
		for in, out := range map[string][]string{
			"": nil,
			"--decrypt --armor --recipient 0xDEADBEEF": {"--decrypt", "--armor", "--recipient", "0xDEADBEEF"},
		} {
			assert.NoError(t, os.Setenv(vn, in))
			assert.Equal(t, out, GPGOpts())
			assert.NoError(t, os.Unsetenv(vn))
		}
	}
}

func TestGPGConfig(t *testing.T) {
	t.Parallel()

	in := `
#-----------------------------
# default key
#-----------------------------
# The default key to sign with. If this option is not used, the default key is
# the first key found in the secret keyring
#default-key 0xD8692123C4065DEA5E0F3AB5249B39D24F25E3F6
#-----------------------------
# behavior
#-----------------------------
# Disable inclusion of the version string in ASCII armored output
no-emit-version
# Disable comment string in clear text signatures and ASCII armored messages
no-comments
# Display long key IDs
keyid-format 0xlong
# List all keys (or the specified ones) along with their fingerprints
with-fingerprint
# Display the calculated validity of user IDs during key listings
list-options show-uid-validity
verify-options show-uid-validity
# Try to use the GnuPG-Agent. With this option, GnuPG first tries to connect to
# the agent before it asks for a passphrase.
use-agent
#-----------------------------
# keyserver
#-----------------------------
# This is the server that --recv-keys, --send-keys, and --search-keys will
# communicate with to receive keys from, send keys to, and search for keys on
keyserver hkps://hkps.pool.sks-keyservers.net
# Provide a certificate store to override the system default
# Get this from https://sks-keyservers.net/sks-keyservers.netCA.pem
#keyserver-options ca-cert-file=/usr/local/etc/ssl/certs/hkps.pool.sks-keyservers.net.pem
# Set the proxy to use for HTTP and HKP keyservers - default to the standard
# local Tor socks proxy
# It is encouraged to use Tor for improved anonymity. Preferably use either a
# dedicated SOCKSPort for GnuPG and/or enable IsolateDestPort and
# IsolateDestAddr
#keyserver-options http-proxy=socks5-hostname://127.0.0.1:9050
# Don't leak DNS, see https://trac.torproject.org/projects/tor/ticket/2846
#keyserver-options no-try-dns-srv
# When using --refresh-keys, if the key in question has a preferred keyserver
# URL, then disable use of that preferred keyserver to refresh the key from
keyserver-options no-honor-keyserver-url
# When searching for a key with --search-keys, include keys that are marked on
# the keyserver as revoked
keyserver-options include-revoked
#-----------------------------
# algorithm and ciphers
#-----------------------------
# list of personal digest preferences. When multiple digests are supported by
# all recipients, choose the strongest one
personal-cipher-preferences AES256 AES192 AES CAST5
# list of personal digest preferences. When multiple ciphers are supported by
# all recipients, choose the strongest one
personal-digest-preferences SHA512 SHA384 SHA256 SHA224
# message digest algorithm used when signing a key
cert-digest-algo SHA512
# This preference list is used for new keys and becomes the default for
# "setpref" in the edit menu
default-preference-list SHA512 SHA384 SHA256 SHA224 AES256 AES192 AES CAST5 ZLIB BZIP2 ZIP Uncompressed

default-recipient-self
# GPGConf disabled this option here at Di 26 Aug 2008 14:22:45 CEST
# keyserver pgpkeys.pca.dfn.de
default-cert-check-level 3
no-mangle-dos-filenames
no-secmem-warning
use-agent

#throw-keyids

default-key  FEEDBEEF
utf8-strings
encrypt-to   DEADBEEF
`
	cfg, err := parseGpgConfig(bytes.NewReader([]byte(in)))
	require.NoError(t, err)
	assert.Equal(t, "SHA512", cfg["cert-digest-algo"])
	_, found := cfg["throw-keyids"]
	assert.False(t, found)
	assert.Equal(t, "DEADBEEF", cfg["encrypt-to"])
}
