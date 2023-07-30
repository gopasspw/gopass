package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gopasspw/gopass/internal/set"
	"golang.org/x/exp/maps"
)

var kvSep = ": "

// AKV is the new Key-Value implementation that will replace KV.
type AKV struct {
	password string
	kvp      map[string][]string

	raw strings.Builder

	fromMime bool
}

// NewAKV creates a new AKV instances.
func NewAKV() *AKV {
	a := &AKV{
		kvp: make(map[string][]string),
		raw: strings.Builder{},
	}

	return a
}

// NewKVWithData returns a new KV secret populated with data.
func NewAKVWithData(pw string, kvps map[string][]string, body string, converted bool) *AKV {
	kv := NewAKV()
	kv.SetPassword(pw)
	kv.fromMime = converted

	for _, k := range set.Sorted(maps.Keys(kvps)) {
		vs := kvps[k]
		for _, v := range vs {
			_ = kv.Add(k, v)
		}
	}

	kv.raw.WriteString(body)

	return kv
}

// Bytes returns the raw string as bytes.
func (a *AKV) Bytes() []byte {
	return []byte(a.raw.String())
}

// Keys returns all the parsed keys.
func (a *AKV) Keys() []string {
	return set.Sorted(maps.Keys(a.kvp))
}

// Get returns the value of the requested key, if found.
func (a *AKV) Get(key string) (string, bool) {
	if v, found := a.kvp[key]; found {
		return v[0], true
	}

	return "", false
}

// Values returns all values for that key.
func (a *AKV) Values(key string) ([]string, bool) {
	v, found := a.kvp[key]

	return v, found
}

// Set writes a single key.
func (a *AKV) Set(key string, value any) error {
	// if it's new key we can just append it at the end
	if _, found := a.kvp[key]; !found {
		return a.Add(key, value)
	}

	if len(a.kvp[key]) < 1 {
		a.kvp[key] = make([]string, 1)
	}
	a.kvp[key][0] = fmt.Sprintf("%s", value)

	// if the key does exist we must make sure to update only
	// the first instance and leave all others intact.

	if a.raw.Len() == 0 {
		a.raw.WriteString("\n")
	}
	s := bufio.NewScanner(strings.NewReader(a.raw.String()))
	a.raw = strings.Builder{}

	firstLine := true
	written := false
	for s.Scan() {
		line := s.Text()

		// pass through any non-key value pair and
		// always leave the password in place, even if it
		// might look like a kv pair. Also stop looking
		// for kv pairs after we updated the first instance.
		if written || firstLine || !strings.Contains(line, kvSep) {
			a.raw.WriteString(line)
			a.raw.WriteString("\n")
			firstLine = false

			continue
		}

		k, _, found := strings.Cut(strings.TrimSpace(line), ":")
		if !found {
			// should not happen
			a.raw.WriteString(line)
			a.raw.WriteString("\n")

			continue
		}

		k = strings.TrimSpace(k)
		if k == key {
			// update the key
			a.raw.WriteString(fmt.Sprintf("%s: %s\n", key, value))
			written = true

			continue
		}

		// keep all others
		a.raw.WriteString(line)
		a.raw.WriteString("\n")
	}

	return nil
}

// Add appends data to a given key.
func (a *AKV) Add(key string, value any) error {
	sv := fmt.Sprintf("%s", value)
	a.kvp[key] = append(a.kvp[key], sv)

	// we must not accidentially write the KVP into the password field
	if a.raw.Len() == 0 {
		a.raw.WriteString("\n")
	}
	a.raw.WriteString(fmt.Sprintf("%s: %s\n", key, sv))

	return nil
}

// Del removes a given key and all of its values.
func (a *AKV) Del(key string) bool {
	_, found := a.kvp[key]
	if !found {
		return false
	}

	delete(a.kvp, key)

	s := bufio.NewScanner(strings.NewReader(a.raw.String()))
	a.raw = strings.Builder{}
	first := true
	for s.Scan() {
		line := s.Text()

		// pass through any non-key value pair and
		// always leave the password in place, even if it
		// might look like a kv pair.
		if first || !strings.Contains(line, kvSep) {
			a.raw.WriteString(line)
			a.raw.WriteString("\n")
			first = false

			continue
		}

		k, _, found := strings.Cut(strings.TrimSpace(line), kvSep)
		if !found {
			// should not happen
			a.raw.WriteString(line)
			a.raw.WriteString("\n")

			continue
		}

		k = strings.TrimSpace(k)
		if k == key {
			// skip the key we want to delete
			continue
		}

		// keep all others
		a.raw.WriteString(line)
		a.raw.WriteString("\n")
	}

	return true
}

// Password returns the password.
func (a *AKV) Password() string {
	return a.password
}

// SetPassword updates the password.
func (a *AKV) SetPassword(p string) {
	s := bufio.NewScanner(strings.NewReader(a.raw.String()))
	a.raw = strings.Builder{}

	// write the new password
	a.password = p
	a.raw.WriteString(p)
	a.raw.WriteString("\n")

	write := false
	for s.Scan() {
		// skip only the first line, that contains the new password and was already written
		if !write {
			write = true

			continue
		}

		a.raw.WriteString(s.Text())
		a.raw.WriteString("\n")
	}
}

// ParseAKV tries to parse an AKV secret.
func ParseAKV(in []byte) *AKV {
	a := NewAKV()
	a.raw = strings.Builder{}
	s := bufio.NewScanner(bytes.NewReader(in))

	first := true
	for s.Scan() {
		line := s.Text()
		a.raw.WriteString(line)
		a.raw.WriteString("\n")

		// handle the password that must be in the very first line
		if first {
			a.password = strings.TrimSpace(line)
			first = false

			continue
		}

		if !strings.Contains(line, kvSep) {
			continue
		}

		line = strings.TrimSpace(line)

		key, val, found := strings.Cut(line, kvSep)
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// we only store lower case keys for KV
		a.kvp[key] = append(a.kvp[key], val)
	}

	if a.raw.String() == "" {
		a.raw.WriteString("\n")
	}

	return a
}

// Body returns the body.
func (a *AKV) Body() string {
	out := strings.Builder{}

	s := bufio.NewScanner(strings.NewReader(a.raw.String()))
	first := true
	for s.Scan() {
		// skip over the password
		if first {
			first = false

			continue
		}

		line := s.Text()
		// ignore KV pairs
		if strings.Contains(line, kvSep) {
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
	}

	return out.String()
}

// Write appends the buffer to the secret's body.
func (a *AKV) Write(buf []byte) (int, error) {
	var w io.Writer
	w = &a.raw

	// If the body is empty before writing we must threat the first line
	// of the newly written content as the password or multi-line insert
	// and similar operations will fail. For regular command line usage
	// this is actually a non-issue since the gopass process will exit
	// after writing and when reading the secret it will be (correctly)
	// parsed. But for tests, library and maybe REPL usage this is an issue.
	if a.raw.Len() == 0 {
		w = io.MultiWriter(&a.raw, &pwWriter{a: a})
	}

	return w.Write(buf)
}

// FromMime returns whether this secret was converted from a Mime secret of not.
func (a *AKV) FromMime() bool {
	return a.fromMime
}

// SafeStr always returnes "(elided)".
func (a *AKV) SafeStr() string {
	return "(elided)"
}

// pwWriter is a io.Writer that will extract the first line of the input stream and
// then write it to the password field of the provided AKV. The first line can stretch
// multiple chunks but once the first line hass been completed any writes to this
// writer will be silently discarded.
type pwWriter struct {
	a       *AKV
	buf     strings.Builder
	written bool
}

func (p *pwWriter) Write(buf []byte) (int, error) {
	n := len(buf)
	if p.written || p.a == nil {
		return n, nil
	}

	i := bytes.Index(buf, []byte("\n"))
	if i > 0 {
		p.buf.Write(buf[0:i])
		p.a.password = p.buf.String()
		p.written = true

		return n, nil
	}
	p.buf.Write(buf)

	return n, nil
}
