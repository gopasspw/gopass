package vault

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/agent/client"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const (
	passwordKey = "__password__"
)

// Store is a vault backed store
type Store struct {
	api    *api.Client
	alias  string
	url    *backend.URL
	path   string
	agent  *client.Client
	cfgdir string
}

// New creates a new store
func New(ctx context.Context, alias string, url *backend.URL, cfgdir string, agent *client.Client) (*Store, error) {
	cfg := &api.Config{
		Address: fmt.Sprintf("%s://%s:%s", url.Scheme, url.Host, url.Port),
	}
	// configure TLS, if necessary
	if err := configureTLS(url.Query, cfg); err != nil {
		return nil, err
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	s := &Store{
		api:    client,
		alias:  alias,
		url:    url,
		path:   url.Path,
		cfgdir: cfgdir,
	}

	token := url.Query.Get("token")
	key := fmt.Sprintf("vault-%s-%s", alias, url.String())
	if token == "" {
		out.Debug(ctx, "Requesting token from secrets config: %s", key)
		t, err := s.loadSecret(ctx, key)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load token from secrets config: %s", err)
		}
		out.Debug(ctx, "Got token from secrets config: '%s'", t)
		token = t
	}
	out.Debug(ctx, "Vault-Token: '%s'", token)

	s.api.SetToken(token)

	// test connection and save token if it works
	if _, err := s.List(ctx, ""); err != nil {
		out.Debug(ctx, "Vault access not working. removing saved token")
		_ = s.eraseSecret
		return nil, err
	}
	out.Debug(ctx, "Vault access OK. saving token")
	if err := s.storeSecret(ctx, key, token); err != nil {
		return nil, err
	}

	return s, nil
}

func configureTLS(q url.Values, cfg *api.Config) error {
	// https://godoc.org/github.com/hashicorp/vault/api#TLSConfig
	if q == nil {
		return nil
	}
	if cfg == nil {
		return nil
	}

	tlscfg := api.TLSConfig{}
	if cc := q.Get("tls-cacert"); cc != "" {
		tlscfg.CACert = cc
	}
	if cc := q.Get("tls-capath"); cc != "" {
		tlscfg.CAPath = cc
	}
	if cc := q.Get("tls-clientcert"); cc != "" {
		tlscfg.ClientCert = cc
	}
	if cc := q.Get("tls-clientkey"); cc != "" {
		tlscfg.ClientKey = cc
	}
	if cc := q.Get("tls-servername"); cc != "" {
		tlscfg.TLSServerName = cc
	}
	if cc := q.Get("tls-insecure"); cc != "" {
		tlscfg.Insecure = true
	}

	defcfg := api.TLSConfig{}
	if tlscfg == defcfg {
		return nil
	}

	return cfg.ConfigureTLS(&tlscfg)
}

// String implement fmt.Stringer
func (s *Store) String() string {
	return fmt.Sprintf("VaultStore(Alias: %s, Path: %s)", s.alias, s.url.String())
}

// Path returns the path component
func (s *Store) Path() string {
	return s.path
}

// URL returns the full URL
func (s *Store) URL() string {
	return s.url.String()
}

// Alias returns the mount point of this store
func (s *Store) Alias() string {
	return s.alias
}

// Copy tries to copy one or more entries
func (s *Store) Copy(ctx context.Context, from string, to string) error {
	// recursive copy?
	if s.IsDir(ctx, from) {
		if s.Exists(ctx, to) {
			return errors.Errorf("Can not copy dir to file")
		}
		sf, err := s.List(ctx, "")
		if err != nil {
			return errors.Wrapf(err, "failed to list store")
		}
		destPrefix := to
		if s.IsDir(ctx, to) {
			destPrefix = filepath.Join(to, filepath.Base(from))
		}
		for _, e := range sf {
			if !strings.HasPrefix(e, strings.TrimSuffix(from, "/")+"/") {
				continue
			}
			et := filepath.Join(destPrefix, strings.TrimPrefix(e, from))
			if err := s.Copy(ctx, e, et); err != nil {
				out.Red(ctx, "Failed to copy '%s' to '%s': %s", e, et, err)
			}
		}
		return nil
	}

	content, err := s.Get(ctx, from)
	if err != nil {
		return errors.Wrapf(err, "failed to get '%s' from store", from)
	}
	if err := s.Set(ctx, to, content); err != nil {
		return errors.Wrapf(err, "failed to save '%s' to store", to)
	}
	return nil
}

// Delete removes a single entry
func (s *Store) Delete(ctx context.Context, path string) error {
	_, err := s.api.Logical().Delete(path)
	return err
}

// Equals returns true if this and other are the same store
func (s *Store) Equals(other store.Store) bool {
	if other == nil {
		return false
	}
	return s.url.String() == other.URL()
}

// Exists checks if a given secret exists
func (s *Store) Exists(ctx context.Context, name string) bool {
	_, err := s.Get(ctx, name)
	return err == nil
}

// Get returns a secret
func (s *Store) Get(ctx context.Context, name string) (store.Secret, error) {
	key := path.Join(s.path, name)
	out.Debug(ctx, "Get(%s) %s", name, key)
	sec, err := s.api.Logical().Read(key)
	if err != nil {
		return nil, err
	}
	if sec == nil || sec.Data == nil {
		return nil, fmt.Errorf("not found")
	}
	return &Secret{d: sec.Data}, nil
}

// Init returns nil
func (s *Store) Init(context.Context, string, ...string) error {
	return nil
}

// Initialized returns true if the backend can communicate with Vault
func (s *Store) Initialized(ctx context.Context) bool {
	_, err := s.List(ctx, "")
	return err == nil
}

// IsDir returns true if the given name is a dir
func (s *Store) IsDir(ctx context.Context, name string) bool {
	ls, err := s.List(ctx, name)
	if err != nil {
		return false
	}
	return len(ls) > 1
}

func extractListKeys(d map[string]interface{}) []string {
	k, found := d["keys"]
	if !found {
		return nil
	}
	ki, ok := k.([]interface{})
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(ki))
	for _, e := range ki {
		if sv, ok := e.(string); ok {
			keys = append(keys, sv)
		}
	}
	return keys
}

// List returns a list of entries with the given prefix
func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	keys, err := s.list(ctx, prefix)
	if err != nil {
		return nil, err
	}
	for i, e := range keys {
		keys[i] = path.Join(s.alias, e)
	}
	return keys, nil
}

func (s *Store) list(ctx context.Context, prefix string) ([]string, error) {
	sec, err := s.api.Logical().List(path.Join(s.path, prefix))
	if err != nil {
		return nil, err
	}
	if sec == nil || sec.Data == nil {
		return nil, nil
	}
	dirents := extractListKeys(sec.Data)
	keys := make([]string, 0, len(dirents))
	for _, e := range dirents {
		if !strings.HasSuffix(e, "/") {
			keys = append(keys, e)
		}
		k, err := s.list(ctx, e)
		if err != nil {
			return nil, err
		}
		for _, key := range k {
			keys = append(keys, path.Join(e, key))
		}
	}
	return keys, nil
}

// Move moves one or many secrets
func (s *Store) Move(ctx context.Context, from string, to string) error {
	// recursive move?
	if s.IsDir(ctx, from) {
		if s.Exists(ctx, to) {
			return errors.Errorf("Can not move dir to file")
		}
		sf, err := s.List(ctx, "")
		if err != nil {
			return errors.Wrapf(err, "failed to list store")
		}
		destPrefix := to
		if s.IsDir(ctx, to) {
			destPrefix = filepath.Join(to, filepath.Base(from))
		}
		for _, e := range sf {
			if !strings.HasPrefix(e, strings.TrimSuffix(from, "/")+"/") {
				continue
			}
			et := filepath.Join(destPrefix, strings.TrimPrefix(e, from))
			if err := s.Move(ctx, e, et); err != nil {
				out.Red(ctx, "Failed to move '%s' to '%s': %s", e, et, err)
			}
		}
		return nil
	}

	content, err := s.Get(ctx, from)
	if err != nil {
		return errors.Wrapf(err, "failed to decrypt '%s'", from)
	}
	if err := s.Set(ctx, to, content); err != nil {
		return errors.Wrapf(err, "failed to write '%s'", to)
	}
	if err := s.Delete(ctx, from); err != nil {
		return errors.Wrapf(err, "failed to delete '%s'", from)
	}
	return nil
}

// Set writes a secret
func (s *Store) Set(ctx context.Context, name string, sec store.Secret) error {
	d := sec.Data()
	if d == nil {
		d = make(map[string]interface{}, 1)
	}
	d[passwordKey] = sec.Password()
	_, err := s.api.Logical().Write(path.Join(s.path, name), d)
	return err
}

// Prune removes a directory tree
func (s *Store) Prune(ctx context.Context, name string) error {
	ls, err := s.List(ctx, name)
	if err != nil {
		return err
	}
	for _, e := range ls {
		if err := s.Delete(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// Valid returns true if this store is not nil
func (s *Store) Valid() bool {
	return s != nil
}
