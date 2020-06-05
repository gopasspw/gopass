package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlain(t *testing.T) {
	for _, tc := range []struct {
		desc string
		in   string
		pw   string
		body string
	}{
		{
			desc: "only password",
			in:   "moar\n",
			pw:   "moar",
			body: "",
		},
		{
			desc: "only password (no line break)",
			in:   "moar",
			pw:   "moar",
			body: "",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			sec := ParsePlain([]byte(tc.in))
			t.Logf("Secret: %+v", sec)
			assert.Equal(t, tc.pw, sec.Get("password"))
			assert.Equal(t, tc.body, sec.GetBody())
		})
	}
}

func TestPlainMIME(t *testing.T) {
	in := `passw0rd
and some
more content
in the body`
	out := `GOPASS-SECRET-1.0
Password: passw0rd

and some
more content
in the body`
	sec := ParsePlain([]byte(in))
	msec := sec.MIME()
	assert.Equal(t, out, string(msec.Bytes()))
	assert.Equal(t, "and some\nmore content\nin the body", sec.GetBody())
}
