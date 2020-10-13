package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret"
)

// make sure that Plain implements Secret
var _ gopass.Secret = &Plain{}

// Plain is the fallback secret that only contains plain text
type Plain struct {
	Body []byte
}

// ParsePlain never fails and always returns a Plain secret
func ParsePlain(in []byte) *Plain {
	p := &Plain{
		Body: make([]byte, len(in)),
	}
	copy(p.Body, in)
	return p
}

// MIME converts this secret to a gopass MIME secret
func (p *Plain) MIME() *secret.MIME {
	m := secret.New()
	br := bufio.NewReader(bytes.NewReader(p.Body))
	pw, _ := br.ReadString('\n')
	m.Set("password", strings.TrimSuffix(pw, "\n"))
	io.Copy(m, br)
	return m
}

// Bytes returns the body
func (p *Plain) Bytes() []byte {
	return p.Body
}

// Keys always returns nil
func (p *Plain) Keys() []string {
	return nil
}

// Get returns the first line (for password) or the empty string
func (p *Plain) Get(key string) string {
	if strings.ToLower(key) != "password" {
		debug.Log("Plain secrets do not support key-values calls. Key ", key, " could not be used. Returning empty secret.")
		return ""
	}
	br := bufio.NewReader(bytes.NewReader(p.Body))
	pw, _ := br.ReadString('\n')
	return strings.TrimSuffix(pw, "\n")
}

// Values returns the first line (for password) or the empty string
func (p *Plain) Values(key string) []string {
	return []string{p.Get(key)}
}

// Set updates the first line (for password) or does nothing
func (p *Plain) Set(key, value string) {
	if strings.ToLower(key) != "password" {
		return
	}
	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, value)
	br := bufio.NewReader(bytes.NewReader(p.Body))
	_, err := br.ReadString('\n')
	if err != nil {
		debug.Log("failed to discard password line: %s", err)
	}
	_, err = io.Copy(buf, br)
	if err != nil {
		debug.Log("failed to copy buffer: %s", err)
	}
	p.Body = buf.Bytes()
}

// Del does nothing
func (p *Plain) Del(_ string) {}

// GetBody returns everything execpt the first line
func (p *Plain) GetBody() string {
	br := bufio.NewReader(bytes.NewReader(p.Body))
	_, err := br.ReadString('\n')
	if err != nil {
		debug.Log("failed to discard password line: %s", err)
		return ""
	}
	buf := &bytes.Buffer{}
	io.Copy(buf, br)
	return buf.String()
}
