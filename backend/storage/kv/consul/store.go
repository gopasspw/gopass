package consul

import (
	"context"

	"github.com/blang/semver"
	api "github.com/hashicorp/consul/api"
)

// Store is a consul-backed store
type Store struct {
	api *api.Client
}

// New creates a new consul store
func New(host, datacenter, token string) (*Store, error) {
	client, err := api.NewClient(&api.Config{
		Address:    host,
		Datacenter: datacenter,
		Token:      token,
	})
	if err != nil {
		return nil, err
	}
	return &Store{
		api: client,
	}, nil
}

// Get retrieves a single entry
func (s *Store) Get(ctx context.Context, name string) ([]byte, error) {
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
	p := &api.KVPair{
		Key:   name,
		Value: value,
	}
	_, err := s.api.KV().Put(p, nil)
	return err
}

// Delete removes a single entry
func (s *Store) Delete(ctx context.Context, name string) error {
	_, err := s.api.KV().Delete(name, nil)
	return err
}

// Exists checks if a given entry exists
func (s *Store) Exists(ctx context.Context, name string) bool {
	v, err := s.Get(ctx, name)
	if err == nil && v != nil {
		return true
	}
	return false
}

// List lists all entries matching the given prefix
func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	pairs, _, err := s.api.KV().List(prefix, nil)
	if err != nil {
		return nil, err
	}
	res := make([]string, len(pairs))
	for _, kvp := range pairs {
		res = append(res, kvp.Key)
	}
	return res, nil
}

// IsDir checks if the given entry is a directory
func (s *Store) IsDir(ctx context.Context, name string) bool {
	ls, err := s.List(ctx, name)
	if err == nil && len(ls) > 1 {
		return true
	}
	return false
}

// Prune removes the given tree
func (s *Store) Prune(ctx context.Context, prefix string) error {
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
