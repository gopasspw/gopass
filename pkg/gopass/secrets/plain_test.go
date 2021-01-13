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
			assert.Equal(t, tc.pw, sec.Password())
			assert.Equal(t, tc.body, sec.Body())
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
	v, ok := sec.Get("foozen")
	assert.False(t, ok)
	assert.Equal(t, "", v)
	assert.Equal(t, "foobar", sec.Password())
	sec.SetPassword("zab")
	assert.Equal(t, "zab", sec.Password())
}
