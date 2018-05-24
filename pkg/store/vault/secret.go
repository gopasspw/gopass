package vault

import (
	"bytes"
	"errors"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/pkg/store"
)

// Secret is a vault secret
type Secret struct {
	d map[string]interface{}
}

// Body always returns the empty string
func (s *Secret) Body() string {
	return ""
}

// Bytes returns a list serialized copy of this secret
func (s *Secret) Bytes() ([]byte, error) {
	if s.d == nil {
		return []byte{}, nil
	}

	buf := &bytes.Buffer{}
	if pw, found := s.d[passwordKey]; found {
		if sv, ok := pw.(string); ok {
			_, _ = buf.WriteString(sv)
		}
	}
	_, _ = buf.WriteString("\n")
	keys := make([]string, 0, len(s.d))
	for k := range s.d {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := s.d[k]
		if k == passwordKey {
			continue
		}
		_, _ = buf.WriteString(k)
		_, _ = buf.WriteString(": ")
		if sv, ok := v.(string); ok {
			_, _ = buf.WriteString(sv)
		}
		_, _ = buf.WriteString("\n")
	}
	return buf.Bytes(), nil
}

// Data returns the data map. Will never ne nil
func (s *Secret) Data() map[string]interface{} {
	if s.d == nil {
		s.d = make(map[string]interface{})
	}
	return s.d
}

// DeleteKey removes a single key
func (s *Secret) DeleteKey(key string) error {
	if s.d == nil {
		return nil
	}
	delete(s.d, key)
	return nil
}

// Equal returns true if two secrets match
func (s *Secret) Equal(other store.Secret) bool {
	b1, err := s.Bytes()
	if err != nil {
		return false
	}
	b2, err := other.Bytes()
	if err != nil {
		return false
	}
	return string(b1) == string(b2)
}

// Password returns the password
func (s *Secret) Password() string {
	v := s.d[passwordKey]
	if sv, ok := v.(string); ok {
		return sv
	}
	return ""
}

// SetBody is not supported
func (s *Secret) SetBody(string) error {
	return errors.New("not supported")
}

// SetPassword sets the password
func (s *Secret) SetPassword(pw string) {
	s.d[passwordKey] = pw
}

// SetValue sets a single key
func (s *Secret) SetValue(key string, value string) error {
	s.d[key] = value
	return nil
}

// String implement fmt.Stringer
func (s *Secret) String() string {
	var buf strings.Builder
	if sv, ok := s.d[passwordKey].(string); ok {
		_, _ = buf.WriteString(sv)
	}
	keys := make([]string, 0, len(s.d))
	for k := range s.d {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := s.d[key]
		if key == passwordKey {
			continue
		}
		_, _ = buf.WriteString(key)
		_, _ = buf.WriteString(": ")
		if sv, ok := value.(string); ok {
			_, _ = buf.WriteString(sv)
		}
		_, _ = buf.WriteString("\n")
	}
	return buf.String()
}

// Value returns a single value
func (s *Secret) Value(key string) (string, error) {
	v := s.d[key]
	if sv, ok := v.(string); ok {
		return sv, nil
	}
	return "", nil
}
