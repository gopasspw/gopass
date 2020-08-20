package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret"
)

var _ gopass.Secret = &KV{}

// KV is a simple key value secret
type KV struct {
	password string
	data     map[string]string
	body     string
}

// MIME converts this secret to a gopass MIME secret
func (k *KV) MIME() *secret.MIME {
	m := secret.New()
	m.Set("password", k.password)
	for k, v := range k.data {
		if strings.ToLower(k) == "password" {
			continue
		}
		m.Set(k, v)
	}
	m.WriteString(k.body)
	return m
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
	keys = append(keys, "password")
	sort.Strings(keys)
	return keys
}

// Get returns a single key
func (k *KV) Get(key string) string {
	if strings.ToLower(key) == "password" {
		return k.password
	}
	return k.data[key]
}

// Set writes a single key
func (k *KV) Set(key, value string) {
	key = strings.ToLower(key)
	if key == "password" {
		k.password = value
		return
	}
	k.data[key] = value
}

// Del removes a key
func (k *KV) Del(key string) {
	key = strings.ToLower(key)
	delete(k.data, key)
}

// GetBody returns the body
func (k *KV) GetBody() string {
	return k.body
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
		line = strings.TrimRight(line, "\n")
		// append non KV pairs to the body
		if !strings.Contains(line, ": ") {
			sb.WriteString(line)
			sb.WriteString("\n")
			continue
		}

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
		return nil, fmt.Errorf("no KV entries")
	}
	k.body = sb.String()
	return k, nil
}
