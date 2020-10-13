package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	goldenPlain = `GOPASS-SECRET-1.0
Content-Type: text/plain
Password: foobar

more content`
)

func TestMIME(t *testing.T) {
	msec, err := ParseMIME([]byte(goldenPlain))
	assert.NoError(t, err)
	assert.NotNil(t, msec)
	assert.Equal(t, "foobar", msec.Get("password"))
	msec.Set("password", "bar")
	assert.Equal(t, "bar", msec.Get("password"))
	assert.Equal(t, "more content", msec.GetBody())
	msec.Set("password", "foobar")
	assert.Equal(t, goldenPlain, string(msec.Bytes()))
	msec.Set("foo", "bar")
	// Let us add a second foo entry:
	msec.Add("foo", "zab")
	assert.Equal(t, []string{"Content-Type", "Password", "Foo", "Foo"}, msec.Keys())
	assert.Equal(t, []string{"bar", "zab"}, msec.Values("foo"))

	msec.Del("foo")
	assert.Equal(t, []string{"Content-Type", "Password"}, msec.Keys())
	add := "\nbar"
	msec.WriteString(add)
	assert.Equal(t, goldenPlain+add, string(msec.Bytes()))
}

func TestNewline(t *testing.T) {
	in := "GOPASS-SECRET-1.0\nFoo: bar\n\nbody"
	sec, err := ParseMIME([]byte(in))
	require.NoError(t, err)
	assert.Equal(t, "body", sec.GetBody())
	assert.Equal(t, "bar", sec.Get("Foo"))

	assert.Equal(t, in, string(sec.Bytes()))
}

func TestNoNewline(t *testing.T) {
	in := "GOPASS-SECRET-1.0\nFoo: bar"
	sec, err := ParseMIME([]byte(in))
	require.NoError(t, err)
	assert.Equal(t, "", sec.GetBody())
	assert.Equal(t, "bar", sec.Get("Foo"))

	assert.Equal(t, in+"\n", string(sec.Bytes()))
}

func TestEquals(t *testing.T) {
	for _, tc := range []struct {
		a  *MIME
		b  *MIME
		eq bool
	}{
		{
			a:  nil,
			b:  nil,
			eq: true,
		},
		{
			a:  &MIME{},
			b:  nil,
			eq: false,
		},
	} {
		assert.Equal(t, tc.eq, tc.a.Equals(tc.b))
	}
}
