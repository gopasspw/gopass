package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// make sure that Plain implements Secret.
var _ gopass.Secret = &Plain{}

// Plain is a fallback secret type that is used if none of the other secret
// parsers accept the input. This secret only contains a byte slice of the
// input data. We attempt to support retrieving and even writing the password
// by looking for the first line break. The body (everything after the first
// line break) can also be retrieved. Key-value operations are not supported.
//
// DO NOT use this, if possible.
type Plain struct {
	buf []byte
}

// ParsePlain never fails and always returns a Plain secret.
func ParsePlain(in []byte) *Plain {
	p := &Plain{
		buf: make([]byte, len(in)),
	}
	copy(p.buf, in)

	return p
}

// Bytes returns the complete secret.
func (p *Plain) Bytes() []byte {
	return p.buf
}

// Body contains everything but the first line.
func (p *Plain) Body() string {
	br := bufio.NewReader(bytes.NewReader(p.buf))
	_, _ = br.ReadString('\n')

	body := &bytes.Buffer{}
	_, _ = io.Copy(body, br)

	return body.String()
}

// Keys always returns nil.
func (p *Plain) Keys() []string {
	return nil
}

// Get returns the empty string for Plain secrets.
func (p *Plain) Get(key string) (string, bool) {
	debug.Log("Trying to access key %q on a Plain secret", key)

	return "", false
}

// Values returns the empty string for Plain secrets.
func (p *Plain) Values(key string) ([]string, bool) {
	debug.Log("Trying to access key %q on a Plain secret", key)

	return []string{""}, false
}

// Password returns the first line.
func (p *Plain) Password() string {
	br := bufio.NewReader(bytes.NewReader(p.buf))
	pw, _ := br.ReadString('\n')

	return strings.TrimSuffix(pw, "\n")
}

// Set does nothing.
func (p *Plain) Set(_ string, _ any) error {
	return fmt.Errorf("Set not supported for PLAIN: %w", ErrNotSupported)
}

// Add does nothing.
func (p *Plain) Add(_ string, _ any) error {
	return fmt.Errorf("Add not supported for PLAIN: %w", ErrNotSupported)
}

// SetPassword updates the first line.
func (p *Plain) SetPassword(value string) {
	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, value)
	br := bufio.NewReader(bytes.NewReader(p.buf))

	_, err := br.ReadString('\n')
	if err != nil {
		debug.Log("failed to discard password line: %s", err)
	}

	_, err = io.Copy(buf, br)
	if err != nil {
		debug.Log("failed to copy buffer: %s", err)
	}

	p.buf = buf.Bytes()
}

// Del does nothing.
func (p *Plain) Del(_ string) bool {
	return false
}

// Getbuf returns everything execpt the first line.
func (p *Plain) Getbuf() string {
	br := bufio.NewReader(bytes.NewReader(p.buf))
	if _, err := br.ReadString('\n'); err != nil {
		debug.Log("failed to discard password line: %s", err)

		return ""
	}

	buf := &bytes.Buffer{}
	_, _ = io.Copy(buf, br)

	return buf.String()
}

// Write appends to the internal buffer.
func (p *Plain) Write(buf []byte) (int, error) {
	p.buf = append(p.buf, buf...)

	return len(buf), nil
}

// WriteString append a string to the internal buffer.
func (p *Plain) WriteString(in string) {
	_, _ = p.Write([]byte(in))
}

// SafeStr always returnes "(elided)".
func (p *Plain) SafeStr() string {
	return "(elided)"
}
