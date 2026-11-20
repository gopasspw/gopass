package passkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRPID(t *testing.T) {
	t.Parallel()

	// none means that the Relying Party did not set an RP ID
	const none = "\x00"

	for _, tc := range []struct {
		origin string
		rpID   string
		want   string
	}{
		// the RP ID defaults to the effective domain of the origin
		{"https://example.com", none, "example.com"},
		{"https://login.example.com", none, "login.example.com"},
		{"https://Login.Example.COM", none, "login.example.com"},
		{"https://example.com:8443", none, "example.com"},
		{"https://example.com/", none, "example.com"},
		{"https://bücher.example", none, "xn--bcher-kva.example"},
		// names browsers accept although they could not be registered
		{"https://my_app.example.com", none, "my_app.example.com"},
		{"https://r3---sn-4g5e6n7s.example.com", "example.com", "example.com"},
		{"https://-dash-.example.com", none, "-dash-.example.com"},
		{"https://intranet", none, "intranet"},
		// and may be set to the effective domain
		{"https://example.com", "example.com", "example.com"},
		{"https://login.example.com", "login.example.com", "login.example.com"},
		{"https://xn--bcher-kva.example", "xn--bcher-kva.example", "xn--bcher-kva.example"},
		// or to a registrable domain suffix of it
		{"https://login.example.com", "example.com", "example.com"},
		{"https://login.example.com:1337", "example.com", "example.com"},
		{"https://a.b.example.co.uk", "example.co.uk", "example.co.uk"},
		{"https://a.b.example.co.uk", "b.example.co.uk", "b.example.co.uk"},
		{"https://app.corp.internal", "corp.internal", "corp.internal"},
		// wildcard and exception rules of the public suffix list:
		// "*.ck" is a public suffix, "www.ck" is not
		{"https://shop.example.co.ck", "example.co.ck", "example.co.ck"},
		{"https://shop.www.ck", "www.ck", "www.ck"},
		// plain HTTP is a secure context on localhost only
		{"http://localhost", none, "localhost"},
		{"http://localhost:8000", "localhost", "localhost"},
	} {
		var rpID *string
		if tc.rpID != none {
			rpID = &tc.rpID
		}

		got, err := ResolveRPID(tc.origin, rpID)
		require.NoError(t, err, "%s %s", tc.origin, tc.rpID)
		assert.Equal(t, tc.want, got, "%s %s", tc.origin, tc.rpID)
	}
}

func TestResolveRPIDRejected(t *testing.T) {
	t.Parallel()

	// none means that the Relying Party did not set an RP ID
	const none = "\x00"

	for _, tc := range []struct {
		origin string
		rpID   string
		want   ErrorName
	}{
		// another site
		{"https://evil.example", "example.com", SecurityError},
		{"https://example.com.evil.example", "example.com", SecurityError},
		{"https://evilexample.com", "example.com", SecurityError},
		{"https://example.com", "notexample.com", SecurityError},
		// a subdomain is not a suffix
		{"https://example.com", "login.example.com", SecurityError},
		{"https://login.example.com", "m.login.example.com", SecurityError},
		// public suffixes, including private ones, are shared by many sites
		{"https://example.com", "com", SecurityError},
		{"https://example.co.uk", "co.uk", SecurityError},
		{"https://example.co.uk", "uk", SecurityError},
		{"https://alice.github.io", "github.io", SecurityError},
		{"https://shop.example.co.ck", "co.ck", SecurityError},
		{"https://shop.example.co.ck", "ck", SecurityError},
		{"https://app.localhost", "localhost", SecurityError},
		// IP addresses are not domains, in whatever notation
		{"https://192.0.2.1", none, SecurityError},
		{"https://192.0.2.1", "192.0.2.1", SecurityError},
		{"https://[2001:db8::1]", none, SecurityError},
		{"https://127.1", none, NotAllowedError},
		{"https://2130706433", none, NotAllowedError},
		{"https://0x7f.1", none, NotAllowedError},
		{"https://test.0x", none, NotAllowedError},
		{"https://example.com", "192.0.2.1", SecurityError},
		{"https://sub.127.1", "127.1", NotAllowedError},
		// an empty RP ID is not the same as none
		{"https://example.com", "", SecurityError},
		// characters that are not part of a host
		{"https://exa mple.com", none, NotAllowedError},
		{"https://exa\u202emple.com", none, NotAllowedError},
		{"https://exa*mple.com", none, NotAllowedError},
		{"https://example..com", none, NotAllowedError},
		{"https://.example.com", none, NotAllowedError},
		{"https://[fe80::1%25eth0]", none, NotAllowedError},
		{"https://example.com", "exa mple.com", SecurityError},
		{"https://example.com", "example.com*", SecurityError},
		{"https://example.com", "xn--a.example.com", SecurityError},
		// RP IDs that are not the canonical form of a domain
		{"https://example.com", "EXAMPLE.com", SecurityError},
		{"https://xn--bcher-kva.example", "bücher.example", SecurityError},
		{"https://example.com", "example.com.", SecurityError},
		{"https://example.com", ".example.com", SecurityError},
		{"https://example.com", "example..com", SecurityError},
		{"https://example.com", " example.com", SecurityError},
		{"https://example.com", "https://example.com", SecurityError},
		{"https://example.com", "example.com/path", SecurityError},
		{"https://example.com", "example.com:443", SecurityError},
		{"https://example.com", "../example.com", SecurityError},
		{"https://example.com.", none, NotAllowedError},
		// not a secure context
		{"http://example.com", none, SecurityError},
		{"http://example.com", "example.com", SecurityError},
		{"http://dev.localhost", none, SecurityError},
		{"http://127.0.0.1", none, SecurityError},
		// not a tuple origin
		{"", "example.com", NotAllowedError},
		{"null", "example.com", NotAllowedError},
		{"example.com", "example.com", NotAllowedError},
		{"//example.com", none, NotAllowedError},
		{"file:///home/user/index.html", none, NotAllowedError},
		{"chrome-extension://abcdefghijklmnop", none, NotAllowedError},
		{"data:text/html,hello", none, NotAllowedError},
		{"wss://example.com", none, NotAllowedError},
		{"https://", none, NotAllowedError},
		{"https:example.com", none, NotAllowedError},
		// more than an origin, or not what a browser would serialize
		{"https://example.com/login", none, NotAllowedError},
		{"https://example.com//", none, NotAllowedError},
		{"https://example.com?x=1", none, NotAllowedError},
		{"https://example.com?", none, NotAllowedError},
		{"https://example.com#frag", none, NotAllowedError},
		{"https://example.com#", none, NotAllowedError},
		{"https://user@example.com", none, NotAllowedError},
		{"https://user:pass@example.com", "example.com", NotAllowedError},
		{"https://example.com@evil.example", "example.com", NotAllowedError},
		{"https://example.com\\@evil.example", "example.com", NotAllowedError},
		{"https://evil.example\\.example.com", "example.com", NotAllowedError},
		{"https://exa%6dple.com", none, NotAllowedError},
		{"https://example.com:65536", none, NotAllowedError},
		{"https://example.com:-1", none, NotAllowedError},
		{"https://example.com:", none, NotAllowedError},
		{"https://example.com:https", none, NotAllowedError},
		{" https://example.com", none, NotAllowedError},
		{"https://example.com ", none, NotAllowedError},
		{"https://example.com\n", none, NotAllowedError},
		{"https://exam\tple.com", none, NotAllowedError},
		{"https://example.com\x00.evil.example", none, NotAllowedError},
	} {
		var rpID *string
		if tc.rpID != none {
			rpID = &tc.rpID
		}

		got, err := ResolveRPID(tc.origin, rpID)
		require.Error(t, err, "%q %q", tc.origin, tc.rpID)
		assert.Empty(t, got)
		assert.Equal(t, tc.want, NameOf(err), "%q %q: %s", tc.origin, tc.rpID, err)
	}
}

func TestParseOriginRejected(t *testing.T) {
	t.Parallel()

	// The top-level origin of a request is shown to the user and signed. It
	// is held to the same standard as the origin of the request.
	for _, in := range []string{
		"",
		"null",
		"https://",
		"https://exa\u202emple.com",
		"https://exa mple.com",
		"https://example.com/path",
		"https://example.com:99999",
		"https://example..com",
		"https://127.1",
		"http://example.com",
		"ftp://example.com",
	} {
		got, host, err := parseOrigin(in)
		require.Error(t, err, "%q", in)
		assert.Empty(t, got)
		assert.Empty(t, host)
	}
}

func TestParseOrigin(t *testing.T) {
	t.Parallel()

	// The origin that ends up in the client data is the one a browser
	// reports, however it was written.
	for in, want := range map[string]string{
		"https://example.com":         "https://example.com",
		"https://example.com/":        "https://example.com",
		"HTTPS://Example.COM":         "https://example.com",
		"https://example.com:443":     "https://example.com",
		"https://example.com:00443":   "https://example.com",
		"https://example.com:8443":    "https://example.com:8443",
		"https://example.com:08443":   "https://example.com:8443",
		"https://example.com:80":      "https://example.com:80",
		"https://bücher.example:8443": "https://xn--bcher-kva.example:8443",
		"https://BÜCHER.example":      "https://xn--bcher-kva.example",
		"http://localhost:80":         "http://localhost",
		"http://localhost:8080":       "http://localhost:8080",
		"https://192.0.2.1":           "https://192.0.2.1",
		"https://[2001:DB8::1]:8443":  "https://[2001:db8::1]:8443",
	} {
		got, _, err := parseOrigin(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
}
