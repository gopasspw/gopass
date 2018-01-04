package secret

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/justwatchcom/gopass/store"

	yaml "gopkg.in/yaml.v2"
)

// Secret is a decoded secret
type Secret struct {
	sync.Mutex
	password string
	body     string
	data     map[string]interface{}
}

// New creates a new secret
func New(password, body string) *Secret {
	return &Secret{
		password: password,
		body:     body,
	}
}

// Parse decodes an secret. It will always return a valid secret. If decoding
// the body to YAML is may return an error which can be ignored.
func Parse(buf []byte) (*Secret, error) {
	s := &Secret{}
	lines := bytes.SplitN(buf, []byte("\n"), 2)
	if len(lines) > 0 {
		s.password = string(bytes.TrimSpace(lines[0]))
	}
	if len(lines) > 1 {
		s.body = string(bytes.TrimSpace(lines[1]))
	}
	if _, err := s.decodeYAML(); err != nil {
		return s, err
	}
	return s, nil
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

// Bytes encodes an secret
func (s *Secret) Bytes() ([]byte, error) {
	buf := &bytes.Buffer{}
	_, _ = buf.WriteString(s.password)
	_, _ = buf.WriteString("\n")
	_, _ = buf.WriteString(s.body)
	return buf.Bytes(), nil
}

// String encodes and returns a string representation of a secret
func (s *Secret) String() string {
	buf := &bytes.Buffer{}
	_, _ = buf.WriteString(s.password)
	_, _ = buf.WriteString("\n")
	_, _ = buf.WriteString(s.body)
	return buf.String()
}

// Password returns the first line from a secret
func (s *Secret) Password() string {
	s.Lock()
	defer s.Unlock()

	return s.password
}

// SetPassword sets a new password (i.e. the first line)
func (s *Secret) SetPassword(pw string) {
	s.Lock()
	defer s.Unlock()

	s.password = pw
}

// Body returns the body of a secret. If the body was valid YAML it returns an
// empty string
func (s *Secret) Body() string {
	s.Lock()
	defer s.Unlock()

	return s.body
}

// Data returns the data of a secret. Unless the body was valid YAML, it returns
// an map
func (s *Secret) Data() map[string]interface{} {
	s.Lock()
	defer s.Unlock()

	return s.data
}

// SetBody sets a new body possibly erasing an decoded YAML map
func (s *Secret) SetBody(b string) error {
	s.Lock()
	defer s.Unlock()

	s.body = b
	s.data = nil

	_, err := s.decodeYAML()
	return err
}

// Value returns the value of the given key if the body contained valid
// YAML
func (s *Secret) Value(key string) (string, error) {
	s.Lock()
	defer s.Unlock()

	if s.data == nil {
		if !strings.HasPrefix(s.body, "---\n") {
			return "", store.ErrYAMLNoMark
		}
		if _, err := s.decodeYAML(); err != nil {
			return "", err
		}
	}
	if v, found := s.data[key]; found {
		if sv, ok := v.(string); ok {
			return sv, nil
		}
		return "", store.ErrYAMLValueUnsupported
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
	return s.encodeYAML()
}

// DeleteKey key will delete a single key from an decoded map
func (s *Secret) DeleteKey(key string) error {
	s.Lock()
	defer s.Unlock()

	if s.data == nil {
		return store.ErrYAMLNoMark
	}
	delete(s.data, key)
	return s.encodeYAML()
}

// Equal returns true if two secrets are equal
func (s *Secret) Equal(other *Secret) bool {
	if s == nil && other == nil {
		return true
	}
	if s == nil || other == nil {
		return false
	}

	s.Lock()
	defer s.Unlock()

	if s.password != other.Password() {
		return false
	}

	if s.body != other.Body() {
		return false
	}

	buf, err := s.Bytes()
	if err != nil {
		return false
	}
	bufOther, err := other.Bytes()
	if err != nil {
		return false
	}

	if len(buf) != len(bufOther) {
		return false
	}

	return true
}
