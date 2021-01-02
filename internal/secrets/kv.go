package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
)

var _ gopass.Secret = &KV{}

// NewKV creates a new KV secret
func NewKV() *KV {
	return &KV{
		data: make(map[string]string, 10),
	}
}

// KV is a simple key value secret
type KV struct {
	password string
	data     map[string]string
	body     string
}

// Bytes serializes
func (k *KV) Bytes() []byte {
	buf := &bytes.Buffer{}
	buf.WriteString(k.password)
	buf.WriteString("\n")
	for _, key := range k.Keys() {
		sv, ok := k.data[key]
		if !ok {
			continue
		}
		_, _ = buf.WriteString(key)
		_, _ = buf.WriteString(": ")
		_, _ = buf.WriteString(sv)
		_, _ = buf.WriteString("\n")
	}
	buf.WriteString(k.body)
	return buf.Bytes()
}

// Keys returns all keys
func (k *KV) Keys() []string {
	keys := make([]string, 0, len(k.data)+1)
	for key := range k.data {
		keys = append(keys, key)
	}
	if _, found := k.data["password"]; !found {
		keys = append(keys, "password")
	}
	sort.Strings(keys)
	return keys
}

// Get returns a single key
func (k *KV) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	v, found := k.data[key]
	return v, found
}

// Set writes a single key
func (k *KV) Set(key string, value interface{}) error {
	key = strings.ToLower(key)
	k.data[key] = fmt.Sprintf("%s", value)
	return nil
}

// Del removes a key
func (k *KV) Del(key string) bool {
	key = strings.ToLower(key)
	_, found := k.data[key]
	delete(k.data, key)
	return found
}

// Body returns the body
func (k *KV) Body() string {
	return k.body
}

// Password returns the password
func (k *KV) Password() string {
	return k.password
}

// SetPassword updates the password
func (k *KV) SetPassword(p string) {
	k.password = p
}

// ParseKV tries to parse a KV secret
func ParseKV(in []byte) (*KV, error) {
	k := &KV{
		data: make(map[string]string, 10),
	}
	r := bufio.NewReader(bytes.NewReader(in))
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	k.password = strings.TrimRight(line, "\n")

	var sb strings.Builder
	for {
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// append non KV pairs to the body
		if !strings.Contains(line, ": ") {
			sb.WriteString(line)
			continue
		}
		line = strings.TrimRight(line, "\n")

		parts := strings.SplitN(line, ":", 2)
		// should not happen
		if len(parts) < 1 {
			continue
		}
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		// preserve key only entries
		if len(parts) < 2 {
			k.data[parts[0]] = ""
			continue
		}
		k.data[parts[0]] = parts[1]
	}
	if len(k.data) < 1 {
		debug.Log("no KV entries")
		//return nil, fmt.Errorf("no KV entries")
	}
	k.body = sb.String()
	return k, nil
}

func (k *KV) Write(buf []byte) (int, error) {
	k.body += string(buf)
	return len(buf), nil
}
