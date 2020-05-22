package secret

import (
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/internal/store"

	yaml "gopkg.in/yaml.v2"
)

// Value returns the value of the given key if the body contained valid
// YAML
func (s *Secret) Value(key string) (string, error) {
	s.Lock()
	defer s.Unlock()

	if s.data == nil {
		if !strings.HasPrefix(s.body, "---\n") {
			return "", store.ErrYAMLNoMark
		}
		if err := s.decode(); err != nil {
			return "", err
		}
	}
	if v, found := s.data[key]; found {
		return fmt.Sprintf("%v", v), nil
	}
	return "", store.ErrYAMLNoKey
}

// SetValue sets a key to a given value. Will fail if an non-empty body exists
func (s *Secret) SetValue(key, value string) error {
	s.Lock()
	defer s.Unlock()

	if s.body == "" && s.data == nil {
		s.data = make(map[string]interface{}, 1)
	}
	if s.data == nil {
		return store.ErrYAMLNoMark
	}
	s.data[key] = value
	return s.encode()
}

// DeleteKey key will delete a single key from an decoded map
func (s *Secret) DeleteKey(key string) error {
	s.Lock()
	defer s.Unlock()

	if s.data == nil {
		return store.ErrYAMLNoMark
	}
	delete(s.data, key)
	return s.encode()
}

// decodeYAML attempts to decode an optional YAML part of a secret
func (s *Secret) decodeYAML() (bool, error) {
	if !strings.HasPrefix(s.body, "---\n") && s.password != "---" {
		return false, nil
	}
	d := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(s.body), &d)
	if err != nil {
		return true, err
	}
	s.data = d
	return true, nil
}

func (s *Secret) encodeYAML() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %s", r)
		}
	}()
	// update body
	yb, err := yaml.Marshal(s.data)
	if err != nil {
		return err
	}
	s.body = "---\n" + string(yb)
	return err
}
