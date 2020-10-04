package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/caspr-io/yamlpath"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret"
	yaml "gopkg.in/yaml.v3"
)

// make sure that YAML implements Secret
var _ gopass.Secret = &YAML{}

// YAML is a YAML secret
type YAML struct {
	password string
	data     map[string]interface{}
	body     string
}

// Keys returns all keys
func (y *YAML) Keys() []string {
	keys := make([]string, 0, len(y.data)+1)
	for key := range y.data {
		keys = append(keys, key)
	}
	if _, found := y.data["password"]; !found {
		keys = append(keys, "password")
	}
	sort.Strings(keys)
	return keys
}

// Get returns the value of a single key
func (y *YAML) Get(key string) string {
	if y.data == nil {
		y.data = make(map[string]interface{})
	}
	if strings.ToLower(key) == "password" && y.password != "" {
		return y.password
	}
	if v, found := y.data[key]; found {
		return fmt.Sprintf("%v", v)
	}
	if v, err := yamlpath.YamlPath(y.data, key); err == nil && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// Set sets a key to a given value
func (y *YAML) Set(key, value string) {
	if strings.ToLower(key) == "password" {
		y.password = value
	}
	if y.data == nil {
		y.data = make(map[string]interface{}, 1)
	}
	y.data[key] = value
}

// Del removes a single key
func (y *YAML) Del(key string) {
	delete(y.data, key)
}

// ParseYAML will try to parse a YAML secret.
func ParseYAML(in []byte) (*YAML, error) {
	y := &YAML{
		data: make(map[string]interface{}, 10),
	}
	debug.Log("Parsing %s", string(in))
	r := bufio.NewReader(bytes.NewReader(in))
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line != "---" {
		y.password = line
		body, err := parseBody(r)
		if err != nil {
			return nil, err
		}
		y.body = body
	}

	if err := yaml.NewDecoder(r).Decode(y.data); err != nil && err != io.EOF {
		return nil, err
	}
	return y, nil
}

// GetBody returns the body
func (y *YAML) GetBody() string {
	return y.body
}

func parseBody(r *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		nextLine, err := r.Peek(3)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if string(nextLine) == "---" {
			debug.Log("Beginning of YAML section detected")
			return sb.String(), nil
		}
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		sb.WriteString(line)
	}
	return "", fmt.Errorf("no YAML marker")
}

// MIME converts this secret to a gopass MIME secret
func (y *YAML) MIME() *secret.MIME {
	m := secret.New()
	m.Set("password", y.password)
	m.Set("Content-Type", "text/yaml")
	m.WriteString(y.body)
	if err := yaml.NewEncoder(m).Encode(y.data); err != nil {
		debug.Log("failed to encode YAML: %s", err)
	}
	return m
}

// Bytes serialized this secret
func (y *YAML) Bytes() []byte {
	defer func() {
		if r := recover(); r != nil {
			debug.Log("panic: %s", r)
		}
	}()
	buf := &bytes.Buffer{}
	buf.WriteString(y.password)
	buf.WriteString("\n")
	if y.body != "" {
		buf.WriteString(y.body)
	}
	if !strings.HasSuffix(y.body, "\n") {
		buf.WriteString("\n")
	}
	if len(y.data) > 0 {
		buf.WriteString("---\n")
		if err := yaml.NewEncoder(buf).Encode(y.data); err != nil {
			debug.Log("failed to encode YAML: %s", err)
		}
	}
	return buf.Bytes()
}
