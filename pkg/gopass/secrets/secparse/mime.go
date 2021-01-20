package secparse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"strings"

	"github.com/gopasspw/gopass/pkg/gopass/secrets"
)

// parseLegacyMIME is a fallback parser for the transient MIME format
// TODO Add tests
func parseLegacyMIME(buf []byte) (*secrets.KV, error) {
	var hdr textproto.MIMEHeader
	body := &bytes.Buffer{}
	var pw string

	r := bufio.NewReader(bytes.NewReader(buf))
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(line) != secrets.Ident {
		return nil, fmt.Errorf("unknown secrets type: %s", line)
	}
	tpr := textproto.NewReader(r)
	hdr, err = tpr.ReadMIMEHeader()
	// we can reach EOF is there are no new line at the end of the secret file after the MIME header.
	if err != nil && err != io.EOF {
		return nil, &secrets.PermanentError{Err: err}
	}
	if _, err := io.Copy(body, r); err != nil {
		return nil, &secrets.PermanentError{Err: err}
	}

	if sv := hdr.Get("Password"); sv != "" {
		pw = sv
		hdr.Del("Password")
	}

	data := make(map[string][]string, len(hdr))
	for k := range hdr {
		data[strings.ToLower(k)] = hdr.Values(k)
	}

	return secrets.NewKVWithData(pw, data, body.String(), true), nil
}
