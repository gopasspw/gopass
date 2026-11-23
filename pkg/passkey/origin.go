package passkey

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

// Origin describes where a WebAuthn request comes from. It is established by
// the client, e.g. a browser extension, and never by the web page.
type Origin struct {
	// Origin is the serialized origin of the document that called the
	// WebAuthn API, as a browser reports it in location.origin, e.g.
	// "https://example.com" or "https://example.com:8443".
	Origin string
	// CrossOrigin is true if that document is embedded in a document of
	// another origin.
	CrossOrigin bool
	// TopOrigin is the serialized origin of the top-level document. It is
	// only used, and then required, if CrossOrigin is true.
	TopOrigin string
}

var (
	// errOpaqueOrigin is returned for origins that are not a tuple of a
	// scheme, a host and a port.
	errOpaqueOrigin = errors.New("origin is not a valid tuple origin")
	// errInsecureOrigin is returned for origins that are not a secure
	// context.
	errInsecureOrigin = errors.New("origin is not a secure context")
)

// parseOrigin validates a serialized origin. It returns the origin in the
// canonical form browsers use, i.e. with a lower case ASCII host and without
// the default port, together with its host. The host is the effective domain
// of the origin.
//
// A browser always reports an origin in that form. Other spellings of the
// scheme, the host and the port, and a single trailing slash, are accepted
// for the sake of origins that are typed by a user.
func parseOrigin(origin string) (string, string, error) {
	// A serialized origin has nothing but a scheme, a host and a port. The
	// URL parser is lenient about what surrounds them, so everything it
	// would ignore or decode is refused up front.
	if origin == "" || strings.ContainsAny(origin, "?#\\@%") || strings.ContainsFunc(origin, isSpaceOrControl) {
		return "", "", errOpaqueOrigin
	}

	u, err := url.Parse(origin)
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", errOpaqueOrigin, err)
	}

	if u.Opaque != "" || u.User != nil || (u.Path != "" && u.Path != "/") {
		return "", "", errOpaqueOrigin
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "http" {
		return "", "", fmt.Errorf("%w: unsupported scheme %q", errOpaqueOrigin, u.Scheme)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return "", "", errOpaqueOrigin
	}

	// The host is a domain or an IP address in its usual notation. IP
	// addresses are valid hosts of an origin. Where the effective domain is
	// required to be a domain they are refused later.
	var host string
	if addr, err := netip.ParseAddr(hostname); err == nil && addr.Zone() == "" {
		host = addr.String()
	} else if host, err = normalizeDomain(hostname); err != nil {
		return "", "", fmt.Errorf("%w: %w", errOpaqueOrigin, err)
	}

	// WebAuthn is limited to secure contexts. localhost and its subdomains
	// can be used over plain HTTP. The RP ID rules still apply separately.
	if scheme == "http" && host != "localhost" && !strings.HasSuffix(host, ".localhost") {
		return "", "", fmt.Errorf("%w: %q", errInsecureOrigin, origin)
	}

	serialized := scheme + "://" + host
	if strings.Contains(host, ":") {
		// IPv6 literal
		serialized = scheme + "://[" + host + "]"
	}

	if p := u.Port(); p != "" {
		port, err := strconv.ParseUint(p, 10, 16)
		if err != nil {
			return "", "", fmt.Errorf("%w: invalid port %q", errOpaqueOrigin, p)
		}

		if !isDefaultPort(scheme, port) {
			serialized += ":" + strconv.FormatUint(port, 10)
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return "", "", fmt.Errorf("%w: empty port", errOpaqueOrigin)
	}

	return serialized, host, nil
}

func isSpaceOrControl(r rune) bool {
	return r <= 0x20 || r == 0x7f
}

func isDefaultPort(scheme string, port uint64) bool {
	return (scheme == "https" && port == 443) || (scheme == "http" && port == 80)
}

// domainToASCII converts a domain name the way the host parser of a browser
// does: without the restrictions on hyphens and on characters other than
// letters and digits that only apply to names that are registered.
//
// See https://url.spec.whatwg.org/#concept-domain-to-ascii
var domainToASCII = idna.New(
	idna.MapForLookup(),
	idna.BidiRule(),
	idna.StrictDomainName(false),
	idna.CheckHyphens(false),
)

// normalizeDomain returns the ASCII, lower case form of a domain name. It
// fails for anything that is not a domain, in particular for IP addresses.
func normalizeDomain(host string) (string, error) {
	if host == "" {
		return "", errors.New("empty host")
	}

	if _, err := netip.ParseAddr(strings.Trim(host, "[]")); err == nil {
		return "", fmt.Errorf("%q is an IP address, not a domain", host)
	}

	ascii, err := domainToASCII.ToASCII(host)
	if err != nil {
		return "", fmt.Errorf("%q is not a valid domain: %w", host, err)
	}

	ascii = strings.ToLower(ascii)
	if len(ascii) > 253 {
		return "", fmt.Errorf("%q is not a valid domain: too long", host)
	}

	labels := strings.Split(ascii, ".")
	for _, label := range labels {
		if !isValidLabel(label) {
			return "", fmt.Errorf("%q is not a valid domain", host)
		}
	}

	// A host whose last label is a number is an IPv4 address to a browser,
	// even in forms like "127.1" or "0x7f.1" that are no valid addresses
	// here. It is never a domain.
	if isNumeric(labels[len(labels)-1]) {
		return "", fmt.Errorf("%q is an IP address, not a domain", host)
	}

	return ascii, nil
}

func isValidLabel(label string) bool {
	if label == "" || len(label) > 63 {
		return false
	}

	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}

	return true
}

// isNumeric reports whether a label is a decimal or a hexadecimal number.
func isNumeric(label string) bool {
	digits := "0123456789"
	if rest, found := strings.CutPrefix(label, "0x"); found {
		label, digits = rest, "0123456789abcdef"
	}

	return strings.Trim(label, digits) == ""
}

// ResolveRPID determines the RP ID a request from the origin is scoped to.
// If rpID is nil the effective domain of the origin is used. Otherwise rpID
// must be equal to the effective domain or a registrable domain suffix of it:
// "login.example.com" may use "example.com", but neither "example.org" nor
// the public suffix "com".
//
// This check is what binds a passkey to a web site. It must be made with an
// origin that was determined by the client, never with one supplied by the
// web page.
//
// See https://www.w3.org/TR/webauthn-3/#rp-id and
// https://html.spec.whatwg.org/multipage/browsers.html#is-a-registrable-domain-suffix-of-or-is-equal-to
func ResolveRPID(origin string, rpID *string) (string, error) {
	_, host, err := parseOrigin(origin)
	if err != nil {
		if errors.Is(err, errInsecureOrigin) {
			return "", newError(SecurityError, err)
		}

		return "", newError(NotAllowedError, err)
	}

	domain, err := normalizeDomain(host)
	if err != nil {
		return "", errorf(SecurityError, "the effective domain of the origin is not a valid domain: %w", err)
	}

	if rpID == nil {
		return domain, nil
	}

	// The RP ID is hashed as it is. A different spelling of the same domain
	// would result in a credential no other client could use.
	suffix, err := normalizeDomain(*rpID)
	if err != nil {
		return "", errorf(SecurityError, "invalid RP ID: %w", err)
	}

	if suffix != *rpID {
		return "", errorf(SecurityError, "invalid RP ID: %q is not in its canonical form %q", *rpID, suffix)
	}

	if suffix == domain {
		return suffix, nil
	}

	if !strings.HasSuffix(domain, "."+suffix) {
		return "", errorf(SecurityError, "RP ID %q is not a registrable domain suffix of %q", *rpID, host)
	}

	// The RP ID must not be a public suffix like "com" or "co.uk", or a
	// part of one: every site below it could otherwise use the credential.
	if ps, _ := publicsuffix.PublicSuffix(suffix); ps == suffix {
		return "", errorf(SecurityError, "RP ID %q is a public suffix", *rpID)
	}

	if ps, _ := publicsuffix.PublicSuffix(domain); ps == suffix || strings.HasSuffix(ps, "."+suffix) {
		return "", errorf(SecurityError, "RP ID %q is a public suffix", *rpID)
	}

	return suffix, nil
}
