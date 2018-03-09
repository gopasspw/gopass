package consul

import (
	"context"
	"strings"

	"github.com/blang/semver"
	api "github.com/hashicorp/consul/api"
	"github.com/justwatchcom/gopass/utils/out"
)

// Store is a consul-backed store
type Store struct {
	prefix string
	api    *api.Client
}

// New creates a new consul store
func New(host, prefix, datacenter, token string) (*Store, error) {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	if strings.HasPrefix(prefix, "/") {
		prefix = strings.TrimPrefix(prefix, "/")
	}
	client, err := api.NewClient(&api.Config{
		Address:    host,
		Datacenter: datacenter,
		Token:      token,
	})
	if err != nil {
		return nil, err
	}
	return &Store{
		api:    client,
		prefix: prefix,
	}, nil
}

// Get retrieves a single entry
func (s *Store) Get(ctx context.Context, name string) ([]byte, error) {
	name = s.prefix + name
	out.Debug(ctx, "consul.Get(%s)", name)
	p, _, err := s.api.KV().Get(name, nil)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Value == nil {
		return nil, nil
	}
	return p.Value, nil
}

// Set writes a single entry
func (s *Store) Set(ctx context.Context, name string, value []byte) error {
	name = s.prefix + name
	out.Debug(ctx, "consul.Set(%s)", name)
	p := &api.KVPair{
		Key:   name,
		Value: value,
	}
	_, err := s.api.KV().Put(p, nil)
	return err
}

// Delete removes a single entry
func (s *Store) Delete(ctx context.Context, name string) error {
	name = s.prefix + name
	out.Debug(ctx, "consul.Delete(%s)", name)
	_, err := s.api.KV().Delete(name, nil)
	return err
}

// Exists checks if a given entry exists
func (s *Store) Exists(ctx context.Context, name string) bool {
	out.Debug(ctx, "consul.Exists(%s)", name)
	v, err := s.Get(ctx, name)
	if err == nil && v != nil {
		return true
	}
	return false
}

// List lists all entries matching the given prefix
func (s *Store) List(ctx context.Context, _ string) ([]string, error) {
	prefix := s.prefix
	out.Debug(ctx, "consul.List(%s)", prefix)
	pairs, _, err := s.api.KV().List(prefix, nil)
	if err != nil {
		return nil, err
	}
	res := make([]string, len(pairs))
	for _, kvp := range pairs {
		res = append(res, strings.TrimPrefix(kvp.Key, s.prefix))
	}
	return res, nil
}

// IsDir checks if the given entry is a directory
func (s *Store) IsDir(ctx context.Context, name string) bool {
	name = s.prefix + name
	out.Debug(ctx, "consul.IsDir(%s)", name)
	count := 0
	ls, err := s.List(ctx, name)
	if err != nil {
		return false
	}
	for _, e := range ls {
		if strings.HasPrefix(e, name) {
			count++
		}
	}
	return count > 1
}

// Prune removes the given tree
func (s *Store) Prune(ctx context.Context, prefix string) error {
	prefix = s.prefix + prefix
	out.Debug(ctx, "consul.Prune(%s)", prefix)
	return s.Delete(ctx, prefix)
}

// Name returns consul
func (s *Store) Name() string {
	return "consul"
}

// Version returns 1.0.0
func (s *Store) Version() semver.Version {
	return semver.Version{Major: 1}
}
