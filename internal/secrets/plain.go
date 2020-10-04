package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// make sure that Plain implements Secret
var _ gopass.Secret = &Plain{}

// Plain is the fallback secret that only contains plain text
type Plain struct {
	buf []byte
}

// ParsePlain never fails and always returns a Plain secret
func ParsePlain(in []byte) *Plain {
	p := &Plain{
		buf: make([]byte, len(in)),
	}
	copy(p.buf, in)
	return p
}

// Bytes returns the complete secret
func (p *Plain) Bytes() []byte {
	return p.buf
}

// Body contains everything but the first line
func (p *Plain) Body() string {
	br := bufio.NewReader(bytes.NewReader(p.buf))
	_, _ = br.ReadString('\n')
	body := &bytes.Buffer{}
	io.Copy(body, br)
	return body.String()
}

// Keys always returns nil
func (p *Plain) Keys() []string {
	return nil
}

// Get returns the first line (for password) or the empty string
func (p *Plain) Get(key string) (string, bool) {
	debug.Log("Trying to access key %q on a Plain secret", key)
	return "", false
}

// Password returns the first line
func (p *Plain) Password() string {
	br := bufio.NewReader(bytes.NewReader(p.buf))
	pw, _ := br.ReadString('\n')
	return strings.TrimSuffix(pw, "\n")
}

// Set does nothing
func (p *Plain) Set(_ string, _ interface{}) error {
	return fmt.Errorf("not supported for PLAIN")
}

// SetPassword updates the first line
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

// Del does nothing
func (p *Plain) Del(_ string) bool {
	return false
}

// Getbuf returns everything execpt the first line
func (p *Plain) Getbuf() string {
	br := bufio.NewReader(bytes.NewReader(p.buf))
	_, err := br.ReadString('\n')
	if err != nil {
		debug.Log("failed to discard password line: %s", err)
		return ""
	}
	buf := &bytes.Buffer{}
	io.Copy(buf, br)
	return buf.String()
}

// Write appends to the internal buffer
func (p *Plain) Write(buf []byte) (int, error) {
	p.buf = append(p.buf, buf...)
	return len(buf), nil
}

// WriteString append a string to the internal buffer
func (p *Plain) WriteString(in string) {
	p.Write([]byte(in))
}
