package backend

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// URL is a parsed backend URL
type URL struct {
	url *url.URL

	Crypto   CryptoBackend
	Sync     SyncBackend
	Store    StoreBackend
	Scheme   string
	Host     string
	Port     string
	Path     string
	Username string
	Password string
	Query    url.Values
}

// FromPath returns a new backend URL with the given path
// and default backends (GitCLI, GPGCLI, FS)
func FromPath(path string) *URL {
	return &URL{
		Crypto: GPGCLI,
		Sync:   GitCLI,
		Store:  FS,
		Scheme: "file",
		Path:   path,
	}
}

// ParseURL attempts to parse an backend URL
func ParseURL(us string) (*URL, error) {
	// if it's no URL build file URL and parse that
	nu, err := url.Parse(us)
	if err != nil {
		nu, err = url.Parse("gpgcli-gitcli-fs+file://" + us)
		if err != nil {
			return nil, err
		}
	}
	u := &URL{
		url: nu,
	}
	if err := u.parseScheme(); err != nil {
		return u, err
	}
	u.Path = nu.Path
	if nu.User != nil {
		u.Username = nu.User.Username()
		u.Password, _ = nu.User.Password()
	}
	u.Query = nu.Query()
	if nu.Host != "" {
		h, p, err := net.SplitHostPort(nu.Host)
		if err == nil {
			u.Host = h
			u.Port = p
		}
	}
	return u, nil
}

// String implements fmt.Stringer
func (u *URL) String() string {
	if u.url == nil {
		u.url = &url.URL{}
	}

	scheme := u.Scheme
	if scheme == "" {
		scheme = "file"
	}
	u.url.Scheme = fmt.Sprintf(
		"%s-%s-%s+%s",
		u.Crypto,
		u.Sync,
		u.Store,
		scheme,
	)
	u.url.Path = u.Path
	if u.Username != "" {
		u.url.User = url.UserPassword(u.Username, u.Password)
	}
	u.url.RawQuery = u.Query.Encode()
	return u.url.String()
}

func (u *URL) parseScheme() error {
	crypto, sync, store, scheme, err := splitBackends(u.url.Scheme)
	if err != nil {
		return err
	}

	u.Crypto = cryptoBackendFromName(crypto)
	u.Sync = syncBackendFromName(sync)
	u.Store = storeBackendFromName(store)
	u.Scheme = scheme

	return nil
}

// MarshalYAML implements yaml.Marshaler
func (u *URL) MarshalYAML() (interface{}, error) {
	return u.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (u *URL) UnmarshalYAML(umf func(interface{}) error) error {
	path := ""
	if err := umf(&path); err != nil {
		return err
	}
	um, err := ParseURL(path)
	if err != nil {
		return err
	}
	*u = *um
	return nil
}

func splitBackends(in string) (string, string, string, string, error) {
	p := strings.Split(in, "+")
	if len(p) < 2 {
		return "gpgcli", "gitcli", "fs", "file", nil
	}
	backends := p[0]
	scheme := p[1]
	p = strings.Split(backends, "-")
	if len(p) < 3 {
		return "", "", "", "", fmt.Errorf("invalid")
	}
	return p[0], p[1], p[2], scheme, nil
}
