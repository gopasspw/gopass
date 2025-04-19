// Package secrets provides the different secret types that gopass supports.
package secrets

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/caspr-io/yamlpath"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/set"
	yaml "gopkg.in/yaml.v3"
)

// make sure that YAML implements Secret.
var _ gopass.Secret = &YAML{}

// ErrNoYAML is returned when no YAML section is found.
var ErrNoYAML = fmt.Errorf("no YAML marker")

// ErrNotSupported is returned when a method is not supported.
var ErrNotSupported = fmt.Errorf("not supported")

// YAML is a gopass secret that contains a parsed YAML data structure.
// This is a legacy data type that is discouraged for new users as YAML
// is neither trivial nor intuitive for users manually editing secrets (e.g.
// unquoted phone numbers being parsed as octal and such).
//
// Format
// ------
// Line | Description
//
//	  0 | Password
//	1-n | Body
//	n+1 | Separator ("---")
//	n+2 | YAML content.
type YAML struct {
	password string
	data     map[string]any
	body     string
}

// Keys returns all keys.
func (y *YAML) Keys() []string {
	return set.SortedKeys(y.data)
}

// Get returns the first value of a single key.
func (y *YAML) Get(key string) (string, bool) {
	if y.data == nil {
		y.data = make(map[string]any)
	}

	if v, found := y.data[key]; found {
		return fmt.Sprintf("%v", v), found
	}

	if v, err := yamlpath.YamlPath(y.data, key); err == nil && v != nil {
		return fmt.Sprintf("%v", v), true
	}

	return "", false
}

// Values returns Get since as per YAML specification keys must be unique.
func (y *YAML) Values(key string) ([]string, bool) {
	data, found := y.Get(key)

	return []string{data}, found
}

// Set sets a key to a given value.
func (y *YAML) Set(key string, value any) error {
	if y.data == nil {
		y.data = make(map[string]any, 1)
	}

	y.data[key] = value

	return nil
}

// Add doesn't work since as per YAML specification keys must be unique.
func (y *YAML) Add(key string, value any) error {
	return ErrNotSupported
}

// Del removes a single key.
func (y *YAML) Del(key string) bool {
	_, found := y.data[key]

	delete(y.data, key)

	return found
}

// ParseYAML will try to parse a YAML secret.
func ParseYAML(in []byte) (*YAML, error) {
	y := &YAML{
		data: make(map[string]any, 10),
	}

	debug.V(3).Log("Parsing %q", out.Secret(in))

	r := bufio.NewReader(bytes.NewReader(in))

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}

	line = strings.TrimSpace(line)

	if line != "---" {
		y.password = line

		body, err := parseBody(r)
		if err != nil {
			return nil, fmt.Errorf("failed to parseBody: %w", err)
		}

		y.body = body
	}

	if err := yaml.NewDecoder(r).Decode(y.data); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to decode YAML secret: %w", err)
	}

	return y, nil
}

// Body returns the body.
func (y *YAML) Body() string {
	return y.body
}

// Password returns the password.
func (y *YAML) Password() string {
	return y.password
}

// SetPassword updates the password.
func (y *YAML) SetPassword(v string) {
	y.password = v
}

func parseBody(r *bufio.Reader) (string, error) {
	var sb strings.Builder

	for {
		nextLine, err := r.Peek(3)
		if err != nil {
			if err == io.EOF {
				break
			}

			return "", fmt.Errorf("failed to peek: %w", err)
		}

		if string(nextLine) == "---" {
			debug.V(2).Log("Beginning of YAML section detected")

			return sb.String(), nil
		}

		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return "", fmt.Errorf("failed to read line: %w", err)
		}

		_, _ = sb.WriteString(line)
	}

	return "", ErrNoYAML
}

// Bytes serialized this secret.
func (y *YAML) Bytes() []byte {
	defer func() {
		if r := recover(); r != nil {
			debug.Log("panic: %s", r)
		}
	}()

	buf := &bytes.Buffer{}
	buf.WriteString(y.password)

	if y.body != "" {
		buf.WriteString("\n")
		buf.WriteString(y.body)
	}

	if len(y.data) > 0 {
		if !strings.HasSuffix(y.body, "\n") {
			buf.WriteString("\n")
		}

		buf.WriteString("---\n")

		if err := yaml.NewEncoder(buf).Encode(y.data); err != nil {
			debug.Log("failed to encode YAML: %s", err)
		}
	}

	return buf.Bytes()
}

// Write appends the buffer to the secret's body.
func (y *YAML) Write(buf []byte) (int, error) {
	y.body += string(buf)

	return len(buf), nil
}

// SafeStr always returnes "(elided)".
func (y *YAML) SafeStr() string {
	return "(elided)"
}
