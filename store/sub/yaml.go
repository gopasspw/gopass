package sub

import (
	"fmt"

	"github.com/justwatchcom/gopass/store"
	yaml "gopkg.in/yaml.v2"
)

// GetKey returns a single key from a structured secret
func (s *Store) GetKey(name, key string) ([]byte, error) {
	content, err := s.GetBody(name)
	if err != nil {
		return nil, err
	}

	d := make(map[string]string)
	if err := yaml.Unmarshal(content, &d); err != nil {
		return nil, err
	}

	if v, found := d[key]; found {
		return []byte(v), nil
	}

	return nil, fmt.Errorf("key not found")
}

// SetKey will update a single key in a YAML structured secret
func (s *Store) SetKey(name, key, value string) error {
	var err error
	first, err := s.GetFirstLine(name)
	if err != nil {
		first = []byte("")
	}
	first = append(first, '\n')
	body, err := s.GetBody(name)
	if err != nil && err != store.ErrNotFound && err != store.ErrNoBody {
		return err
	}

	d := make(map[string]string)
	if err := yaml.Unmarshal(body, &d); err != nil {
		return err
	}

	d[key] = value

	buf, err := yaml.Marshal(d)
	if err != nil {
		return err
	}

	return s.SetConfirm(name, append(first, buf...), fmt.Sprintf("Updated key %s in %s", key, name), nil)
}

// DeleteKey will delete a single key in a YAML structured secret
func (s *Store) DeleteKey(name, key string) error {
	var err error
	first, err := s.GetFirstLine(name)
	if err != nil {
		first = []byte("\n")
	}
	body, err := s.GetBody(name)
	if err != nil && err != store.ErrNotFound {
		return err
	}

	d := make(map[string]string)
	if err := yaml.Unmarshal(body, &d); err != nil {
		return err
	}

	delete(d, key)

	buf, err := yaml.Marshal(d)
	if err != nil {
		return err
	}

	return s.SetConfirm(name, append(first, buf...), fmt.Sprintf("Deleted key %s in %s", key, name), nil)
}
