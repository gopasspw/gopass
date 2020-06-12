package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestPlainModify(t *testing.T) {
	in := "foobar\nhello world\nhow are you?"
	sec := ParsePlain([]byte(in))
	require.NotNil(t, sec)
	assert.Equal(t, in, string(sec.Bytes()))
	assert.Equal(t, 0, len(sec.Keys()))
	sec.Set("foozen", "zab")
	assert.Equal(t, "", sec.Get("foozen"))
	assert.Equal(t, "foobar", sec.Get("password"))
	sec.Set("password", "zab")
	assert.Equal(t, "zab", sec.Get("password"))
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
