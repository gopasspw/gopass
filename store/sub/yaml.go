package sub

import (
	"bytes"
	"fmt"

	"github.com/justwatchcom/gopass/store"
	yaml "gopkg.in/yaml.v2"
)

// GetKey returns a single key from a structured secret
func (s *Store) GetKey(name, key string) ([]byte, error) {
	content, err := s.Get(name)
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
func (s *Store) SetKey(name, key, value string) error {
	content, err := s.Get(name)
	if err != nil && err != store.ErrNotFound {
		return err
	}

	parts := bytes.Split(content, []byte("---\n"))

	d := make(map[string]interface{})
	if len(parts) > 1 {
		if err := yaml.Unmarshal(parts[1], &d); err != nil {
			return err
		}
	}

	d[key] = value

	buf, err := yaml.Marshal(d)
	if err != nil {
		return err
	}

	return s.SetConfirm(name, append(parts[0], append([]byte("\n---\n"), buf...)...), fmt.Sprintf("Updated key in %s", name), nil)
}

// DeleteKey will delete a single key in a YAML structured secret
func (s *Store) DeleteKey(name, key string) error {
	content, err := s.Get(name)
	if err != nil && err != store.ErrNotFound {
		return err
	}

	parts := bytes.Split(content, []byte("---\n"))

	d := make(map[string]interface{})
	if len(parts) > 1 {
		if err := yaml.Unmarshal(parts[1], &d); err != nil {
			return err
		}
	}

	delete(d, key)

	buf, err := yaml.Marshal(d)
	if err != nil {
		return err
	}

	return s.SetConfirm(name, append(parts[0], append([]byte("---\n"), buf...)...), fmt.Sprintf("Deleted key %s in %s", key, name), nil)
}
