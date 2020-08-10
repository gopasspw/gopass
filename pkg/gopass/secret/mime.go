package secret

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"sort"
	"strings"
)

const (
	// Ident is the Gopass MIME secret header
	Ident = "GOPASS-SECRET-1.0"
)

// PermanentError signal that parsing should not attempt other formats.
type PermanentError struct {
	Err error
}

func (p *PermanentError) Error() string {
	return p.Err.Error()
}

// MIME is a gopass MIME secret
type MIME struct {
	Header textproto.MIMEHeader
	body   *bytes.Buffer
}

// New creates a new MIME secret
func New() *MIME {
	m := &MIME{
		Header: textproto.MIMEHeader{},
		body:   &bytes.Buffer{},
	}
	return m
}

// MIME returns self
func (s *MIME) MIME() *MIME {
	return s
}

// Equals compare two secrets
func (s *MIME) Equals(other *MIME) bool {
	if s == nil {
		return other == nil
	}
	if other == nil {
		return false
	}

	return string(s.Bytes()) == string(other.Bytes())
}

// ParseMIME tries to parse a MIME secret
func ParseMIME(buf []byte) (*MIME, error) {
	m := &MIME{
		Header: textproto.MIMEHeader{},
		body:   &bytes.Buffer{},
	}
	r := bufio.NewReader(bytes.NewReader(buf))
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(line) != Ident {
		return nil, fmt.Errorf("unknown secrets type: %s", line)
	}
	tpr := textproto.NewReader(r)
	m.Header, err = tpr.ReadMIMEHeader()
	if err != nil {
		return nil, &PermanentError{Err: err}
	}
	if _, err := io.Copy(m.body, r); err != nil {
		return nil, &PermanentError{err}
	}
	return m, nil
}

// Bytes serializes the secret
func (s *MIME) Bytes() []byte {
	buf := &bytes.Buffer{}
	fmt.Fprint(buf, Ident)
	fmt.Fprint(buf, "\n")

	keys := make([]string, 0, len(s.Header))
	for k := range s.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := s.Header[k]
		sort.Strings(vs)
		for _, v := range vs {
			fmt.Fprint(buf, k)
			fmt.Fprint(buf, ": ")
			fmt.Fprint(buf, v)
			fmt.Fprint(buf, "\n")
		}
	}
	fmt.Fprint(buf, "\n")
	buf.Write(s.body.Bytes())
	return buf.Bytes()
}

// Keys returns all keys
func (s *MIME) Keys() []string {
	keys := make([]string, 0, len(s.Header))
	for k := range s.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Get returns the value of a single key
func (s *MIME) Get(key string) string {
	return s.Header.Get(key)
}

// Set sets a value of a key
func (s *MIME) Set(key, value string) {
	s.Header.Set(key, value)
}

// Del removes a key
func (s *MIME) Del(key string) {
	s.Header.Del(key)
}

// GetBody returns the body
func (s *MIME) GetBody() string {
	return s.body.String()
}

// WriteString appends a string to the buffer
func (s *MIME) WriteString(in string) (int, error) {
	return s.body.WriteString(in)
}

// Write implements io.Writer
func (s *MIME) Write(p []byte) (int, error) {
	return s.body.Write(p)
}
