package secret

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/gopasspw/gopass/internal/store"
)

var debug bool

func init() {
	if gdb := os.Getenv("GOPASS_DEBUG"); gdb != "" {
		debug = true
	}
}

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
		s.password = string(lines[0])
	}
	if len(lines) > 1 {
		s.body = string(lines[1])
	}
	if err := s.decode(); err != nil {
		return s, err
	}
	return s, nil
}

// Bytes encodes an secret
func (s *Secret) Bytes() ([]byte, error) {
	buf := &bytes.Buffer{}
	_, _ = buf.WriteString(s.password)
	if s.body != "" {
		_, _ = buf.WriteString("\n")
		_, _ = buf.WriteString(s.body)
	}
	return buf.Bytes(), nil
}

// String encodes and returns a string representation of a secret
func (s *Secret) String() string {
	var buf strings.Builder
	_, _ = buf.WriteString(s.password)

	if s.body != "" {
		_, _ = buf.WriteString("\n")
		_, _ = buf.WriteString(s.body)
	}

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

	err := s.decode()
	return err
}

// Equal returns true if two secrets are equal
func (s *Secret) Equal(other store.Secret) bool {
	if s == nil && (other == nil || reflect.ValueOf(other).IsNil()) {
		return true
	}
	if s == nil || other == nil || reflect.ValueOf(other).IsNil() {
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

	return true
}

func (s *Secret) encode() error {
	return s.encodeKV()
}

func (s *Secret) decode() error {
	return s.decodeKV()
}
