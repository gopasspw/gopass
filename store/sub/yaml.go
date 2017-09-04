package sub

import (
	"bytes"
	"context"
	"fmt"

	"github.com/justwatchcom/gopass/store"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// GetKey returns a single key from a structured secret
func (s *Store) GetKey(ctx context.Context, name, key string) ([]byte, error) {
	content, err := s.Get(ctx, name)
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	parts := bytes.SplitN(content, []byte("---\n"), 2)
	if len(parts) < 2 {
		return nil, store.ErrYAMLNoMark
	}

	d := make(map[string]interface{})
	if err := yaml.Unmarshal(parts[1], &d); err != nil {
		return nil, err
	}

	if v, found := d[key]; found {
		if sv, ok := v.(string); ok {
			return []byte(sv), nil
		}
		return nil, store.ErrYAMLValueUnsupported
	}

	return nil, store.ErrYAMLNoKey
}

// SetKey will update a single key in a YAML structured secret
func (s *Store) SetKey(ctx context.Context, name, key, value string) error {
	content, err := s.Get(ctx, name)
	if err != nil && err != store.ErrNotFound {
		return errors.Wrapf(err, "failed to read secret '%s'", name)
	}

	parts := bytes.Split(content, []byte("---\n"))

	d := make(map[string]interface{})
	if len(parts) > 1 {
		if err := yaml.Unmarshal(parts[1], &d); err != nil {
			return errors.Wrapf(err, "failed to decode YAML from secret '%s'", name)
		}
	}

	d[key] = value

	buf, err := yaml.Marshal(d)
	if err != nil {
		return errors.Wrapf(err, "failed to encode YAML for secret '%s'", name)
	}

	return s.SetConfirm(ctx, name, append(bytes.TrimRight(parts[0], "\n"), append([]byte("\n---\n"), buf...)...), fmt.Sprintf("Updated key in %s", name), nil)
}

// DeleteKey will delete a single key in a YAML structured secret
func (s *Store) DeleteKey(ctx context.Context, name, key string) error {
	content, err := s.Get(ctx, name)
	if err != nil && err != store.ErrNotFound {
		return errors.Wrapf(err, "failed to read secret '%s'", name)
	}

	parts := bytes.Split(content, []byte("---\n"))

	d := make(map[string]interface{})
	if len(parts) > 1 {
		if err := yaml.Unmarshal(parts[1], &d); err != nil {
			return errors.Wrapf(err, "failed to decode YAML from secret '%s'", name)
		}
	}

	delete(d, key)

	buf, err := yaml.Marshal(d)
	if err != nil {
		return errors.Wrapf(err, "failed to encode YAML for secret '%s'", name)
	}

	return s.SetConfirm(ctx, name, append(parts[0], append([]byte("---\n"), buf...)...), fmt.Sprintf("Deleted key %s in %s", key, name), nil)
}
