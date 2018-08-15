package backend

import (
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

// URL is a parsed backend URL
type URL struct {
	url *url.URL

	Crypto   CryptoBackend
	RCS      RCSBackend
	Storage  StorageBackend
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
		Crypto:  GPGCLI,
		RCS:     GitCLI,
		Storage: FS,
		Scheme:  "file",
		Path:    path,
	}
}

// ParseURL attempts to parse an backend URL
func ParseURL(us string) (*URL, error) {
	// if it's no URL build file URL and parse that
	nu, err := url.Parse(us)
	if err != nil || nu.Path == "" {
		if runtime.GOOS == "windows" {
			us = "/" + us
		}
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
	if runtime.GOOS == "windows" && strings.HasPrefix(nu.Path, "/") {
		// remove leading slash which comes from net/url parse as it expectes it for an correct URI
		u.Path = nu.Path[1:]
	}

	if nu.User != nil {
		u.Username = nu.User.Username()
		u.Password, _ = nu.User.Password()
	}
	if nu.Host == "~" {
		hd, err := homedir.Dir()
		if err != nil {
			return u, err
		}
		u.Path = filepath.Join(hd, u.Path)
		nu.Host = ""
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

	u.url.Path = u.Path
	if scheme == "file" && runtime.GOOS == "windows" && !strings.HasPrefix(u.url.Path, "///") {
		u.url.Path = "/" + u.Path
	}

	u.url.Scheme = fmt.Sprintf(
		"%s-%s-%s+%s",
		u.Crypto,
		u.RCS,
		u.Storage,
		scheme,
	)

	if u.Username != "" {
		u.url.User = url.UserPassword(u.Username, u.Password)
	}
	if u.Host != "" {
		u.url.Host = u.Host
		if u.Port != "" {
			u.url.Host += ":" + u.Port
		}
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
	u.RCS = rcsBackendFromName(sync)
	u.Storage = storageBackendFromName(store)
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
	if len(p) < 1 {
		return "", "", "", "", fmt.Errorf("invalid")
	}
	if len(p) < 3 {
		return p[0], "", "", scheme, nil
	}
	return p[0], p[1], p[2], scheme, nil
}
