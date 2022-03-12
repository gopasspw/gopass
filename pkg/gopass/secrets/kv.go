package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"golang.org/x/exp/maps"
)

var _ gopass.Secret = &KV{}

// ErrMultiKey is returned when a key is found multiple times.
var ErrMultiKey = fmt.Errorf("multiple identical keys not supported")

// NewKV creates a new KV secret.
func NewKV() *KV {
	return &KV{
		data: make(map[string][]string, 10),
	}
}

// NewKVWithData returns a new KV secret populated with data.
func NewKVWithData(pw string, kvps map[string][]string, body string, converted bool) *KV {
	kv := &KV{
		password: pw,
		data:     make(map[string][]string, len(kvps)),
		body:     body,
		fromMime: converted,
	}

	for k, v := range kvps {
		kv.data[k] = v
	}

	return kv
}

// KV is a secret that contains a password line (maybe empty), any number of
// lines of key-value pairs (defined as: contains a colon) and any number of
// free text lines. This is the default secret format gopass uses and encourages.
// It should be compatible with most other password store implementations and
// works well with our vanity features (e.g. accessing single entries in secret).
//
// Format
// ------
// Line | Description
// ---- | -----------
//    0 | Password. Must contain the "password" or be empty. Can not be omitted.
//  1-n | Key-Value pairs, e.g. "key: value". Can be omitted but the secret
//      | might get parsed as a "Plain" secret if zero key-value pairs are found.
//  n+1 | Body. Can contain any number of characters that will be parsed as
//      | UTF-8 and appended to an internal string. Note: Technically this can
//      | be any kind of binary data but we neither support nor test this with
//      | non-text data. Also we do not intent do support any kind of streaming
//      | access, i.e. this is not intended for huge files.
//
// Example
// -------
// Line | Content
// ---- | -------
//    0 | foobar
//    1 | hello: world
//    2 | gopass: secret
//    3 | Yo
//    4 | Hi
//
// This would be parsed as a KV secret that contains:
//   - password: "foobar"
//   - key-value pairs:
//     - "hello": "world"
//     - "gopass": "secret"
//   - body: "Yo\nHi"
type KV struct {
	password string
	data     map[string][]string
	body     string
	fromMime bool
}

// Bytes serializes.
func (k *KV) Bytes() []byte {
	buf := &bytes.Buffer{}
	buf.WriteString(k.password)
	buf.WriteString("\n")

	for ik, key := range k.Keys() {
		sv, ok := k.data[key]
		if !ok {
			continue
		}

		for iv, v := range sv {
			_, _ = buf.WriteString(key)
			_, _ = buf.WriteString(": ")
			_, _ = buf.WriteString(v)
			// the last one shouldn't add a newline, it's handled below
			if iv < len(sv)-1 {
				_, _ = buf.WriteString("\n")
			}
		}
		// we must only add a final newline if the body is non-empty
		if k.body != "" || ik < len(k.Keys())-1 {
			_, _ = buf.WriteString("\n")
		}
	}

	buf.WriteString(k.body)

	return buf.Bytes()
}

// Keys returns all keys.
func (k *KV) Keys() []string {
	keys := maps.Keys(k.data)

	sort.Strings(keys)

	return keys
}

// Get returns the first value of that key.
func (k *KV) Get(key string) (string, bool) {
	key = strings.ToLower(key)

	if v, found := k.data[key]; found {
		return v[0], true
	}

	return "", false
}

// Values returns all values for that key.
func (k *KV) Values(key string) ([]string, bool) {
	key = strings.ToLower(key)

	v, found := k.data[key]

	return v, found
}

// Set writes a single key.
func (k *KV) Set(key string, value any) error {
	key = strings.ToLower(key)
	if v, ok := k.data[key]; ok && len(v) > 1 {
		return fmt.Errorf("cannot set key %s: this entry contains multiple same keys. Please use 'gopass edit' instead: %w", key, ErrMultiKey)
	}

	k.data[key] = []string{fmt.Sprintf("%s", value)}

	return nil
}

// Add appends data to a given key.
func (k *KV) Add(key string, value any) error {
	key = strings.ToLower(key)
	k.data[key] = append(k.data[key], fmt.Sprintf("%s", value))

	return nil
}

// Del removes a given key and all of its values.
func (k *KV) Del(key string) bool {
	key = strings.ToLower(key)
	_, found := k.data[key]
	delete(k.data, key)

	return found
}

// Body returns the body.
func (k *KV) Body() string {
	return k.body
}

// Password returns the password.
func (k *KV) Password() string {
	return k.password
}

// SetPassword updates the password.
func (k *KV) SetPassword(p string) {
	k.password = p
}

// ParseKV tries to parse a KV secret.
func ParseKV(in []byte) (*KV, error) {
	k := &KV{
		data: make(map[string][]string, 10),
	}
	r := bufio.NewReader(bytes.NewReader(in))

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}

	k.password = strings.TrimRight(line, "\n")

	var sb strings.Builder

	for {
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("failed to read line: %w", err)
		}
		// append non KV pairs to the body
		if !strings.Contains(line, ":") {
			sb.WriteString(line)

			continue
		}

		line = strings.TrimRight(line, "\n")

		key, val, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// we only store lower case keys for KV
		key = strings.ToLower(key)
		k.data[key] = append(k.data[key], val)
	}

	if len(k.data) < 1 {
		debug.Log("no KV entries")
	}

	k.body = sb.String()

	return k, nil
}

// Write appends the buffer to the secret's body.
func (k *KV) Write(buf []byte) (int, error) {
	k.body += string(buf)

	return len(buf), nil
}

// FromMime returns whether this secret was converted from a Mime secret of not.
func (k *KV) FromMime() bool {
	return k.fromMime
}

// SafeStr always returnes "(elided)".
func (k *KV) SafeStr() string {
	return "(elided)"
}
