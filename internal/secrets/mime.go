package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"strings"

	"github.com/gopasspw/gopass/pkg/gopass/secret"
)

// ParseLegacyMIME is a fallback parser for the transient MIME format
// TODO Unexport this
// TODO Add tests
func ParseLegacyMIME(buf []byte) (*KV, error) {
	var hdr textproto.MIMEHeader
	body := &bytes.Buffer{}
	var pw string

	r := bufio.NewReader(bytes.NewReader(buf))
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(line) != secret.Ident {
		return nil, fmt.Errorf("unknown secrets type: %s", line)
	}
	tpr := textproto.NewReader(r)
	hdr, err = tpr.ReadMIMEHeader()
	// we can reach EOF is there are no new line at the end of the secret file after the MIME header.
	if err != nil && err != io.EOF {
		return nil, &secret.PermanentError{Err: err}
	}
	if _, err := io.Copy(body, r); err != nil {
		return nil, &secret.PermanentError{Err: err}
	}

	if sv := hdr.Get("Password"); sv != "" {
		pw = sv
		hdr.Del("Password")
	}

	kv := &KV{
		password: pw,
		data:     make(map[string]string, len(hdr)),
		body:     body.String(),
		fromMime: true,
	}

	for k := range hdr {
		kv.data[strings.ToLower(k)] = hdr.Get(k)
	}

	return kv, nil
}
