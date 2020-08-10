package secparse

import (
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/secrets"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret"
)

// Parse tries to parse a secret. It will start with the most specific
// secrets type.
func Parse(in []byte) (gopass.Secret, error) {
	var s gopass.Secret
	var err error
	s, err = secret.ParseMIME(in)
	if err == nil {
		debug.Log("parsed as MIME: %+v", s)
		return s, nil
	}
	debug.Log("failed to parse as MIME: %s", err)
	if _, ok := err.(*secret.PermanentError); ok {
		return secrets.ParsePlain(in), err
	}
	s, err = secrets.ParseYAML(in)
	if err == nil {
		debug.Log("parsed as YAML: %+v", s)
		return s, nil
	}
	debug.Log("failed to parse as YAML: %s\n%s", err, string(in))
	s, err = secrets.ParseKV(in)
	if err == nil {
		debug.Log("parsed as KV: %+v", s)
		return s, nil
	}
	debug.Log("failed to parse as KV: %s", err)
	s = secrets.ParsePlain(in)
	debug.Log("parsed as plain: %s", string(s.Bytes()))
	return s, nil
}
