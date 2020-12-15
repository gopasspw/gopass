package secret

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
)

const (
	// Ident is the Gopass MIME secret header
	Ident = "GOPASS-SECRET-1.0"
)

var (
	// WriteMIME can be disabled to disable writing the new secrets format.
	// Use this to ensure secrets written by gopass can be correctly consumed
	// by other Password Store implementations, too.
	WriteMIME = true
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
	// we can reach EOF is there are no new line at the end of the secret file after the MIME header.
	if err != nil && err != io.EOF {
		return nil, &PermanentError{Err: err}
	}
	if _, err := io.Copy(m.body, r); err != nil {
		return nil, &PermanentError{err}
	}

	// let us try to store the keys' order:
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	// we skip the magic Ident
	scanner.Scan()
	var order []string
	for scanner.Scan() {
		p := strings.Split(scanner.Text(), ":")
		// we stop once we're past the header
		if p[0] == "" {
			break
		}
		order = append(order, p[0])
	}
	m.Set("Gopass-Key-Order", strings.Join(order, ","))

	return m, nil
}

// bytesCompat writes a pass compatible representation
// of the secret.
func (s *MIME) bytesCompat() []byte {
	buf := &bytes.Buffer{}
	if len(s.Header) > 0 {
		fmt.Fprint(buf, s.Header.Get("Password"))
		fmt.Fprintln(buf)

		preserveOrder := make(map[string]int)
		// then the header (containing typically an entry 'password')
		for _, k := range s.Keys() {
			if key := textproto.CanonicalMIMEHeaderKey(k); key == "Password" || key == "Gopass-Key-Order" {
				debug.Log("Skipping key", key, "in bytesCompat mode")
				continue
			}
			currentIndex := preserveOrder[k]
			v := s.Values(k)[currentIndex]
			preserveOrder[k]++
			fmt.Fprint(buf, k)
			fmt.Fprint(buf, ": ")
			fmt.Fprint(buf, v)
			fmt.Fprint(buf, "\n")
		}
	}

	if body := s.body.Bytes(); s.body != nil && len(body) > 0 {
		if len(s.Header) > 0 {
			fmt.Fprint(buf, "\n")
		}
		buf.Write(s.body.Bytes())
	}

	return buf.Bytes()
}

// Bytes serializes the secret
func (s *MIME) Bytes() []byte {
	if !WriteMIME {
		debug.Log("WriteMIME set to false, falling back to bytesCompat()")
		return s.bytesCompat()
	}
	buf := &bytes.Buffer{}
	// We first have the Mime magic
	fmt.Fprint(buf, Ident)
	fmt.Fprint(buf, "\n")

	preserveOrder := make(map[string]int)
	// then the header (containing typically an entry 'password')
	for _, k := range s.Keys() {
		currentIndex := preserveOrder[k]
		v := s.Values(k)[currentIndex]
		preserveOrder[k]++
		fmt.Fprint(buf, k)
		fmt.Fprint(buf, ": ")
		fmt.Fprint(buf, v)
		fmt.Fprint(buf, "\n")
	}

	// finally the body if any
	if body := s.body.Bytes(); s.body != nil && len(body) > 0 {
		fmt.Fprint(buf, "\n")
		buf.Write(body)
	}
	return buf.Bytes()
}

// Keys returns all keys
func (s *MIME) Keys() []string {
	keys := make([]string, 0, len(s.Header))

	order := s.Get("Gopass-Key-Order")
	if order != "" {
		keys = strings.Split(order, ",")
	} else {
		for k := range s.Header {
			for i := 0; i < len(s.Header[k]); i++ {
				keys = append(keys, k)
			}
		}
		// we need to sort the keys to be deterministic since maps aren't.
		sort.Strings(keys)
	}
	return keys
}

// Get returns the values of a given key
func (s *MIME) Get(key string) string {
	return s.Header.Get(key)
}

// Values returns the values of a given key
func (s *MIME) Values(key string) []string {
	return s.Header.Values(key)
}

// Set sets a value of a key
func (s *MIME) Set(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	s.Header.Set(key, value)
	if key == "Gopass-Key-Order" {
		return
	}
	if keys := s.Get("Gopass-Key-Order"); !strings.Contains(keys, key) {
		if keys != "" {
			keys += ","
		}
		// we add the key after the others
		keys += key
		s.Header.Set("Gopass-Key-Order", keys)
	}
}

// Add adds a value for a key, it appends to any existing values associated with key
func (s *MIME) Add(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	// check if we are to preserve the order
	if keys := s.Get("Gopass-Key-Order"); keys != "" {
		// we add the key after the others
		keys += "," + key
		s.Header.Set("Gopass-Key-Order", keys)
	}
	s.Header.Add(key, value)
}

// Del removes a key
func (s *MIME) Del(key string) {
	s.Header.Del(key)
	key = textproto.CanonicalMIMEHeaderKey(key)
	// check if we are to preserve the order
	if keys := s.Get("Gopass-Key-Order"); keys != "" {
		// we delete all the keys
		keys = strings.ReplaceAll(keys, ","+key, "")
		s.Set("Gopass-Key-Order", keys)
	}
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
