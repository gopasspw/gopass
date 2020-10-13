package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKV(t *testing.T) {
	t.Logf("Retrieve content from invalid YAML (#375)")
	mlValue := `somepasswd
Test / test.com
username: myuser@test.com
url: http://www.test.com/
password: bar
`
	s, err := ParseKV([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)

	assert.Equal(t, "", s.Get("Test / test.com"))

	t.Logf("Secret:\n%+v\n%s\n", s, string(s.Bytes()))

	mlOut := `somepasswd
password: bar
url: http://www.test.com/
username: myuser@test.com
Test / test.com
`
	t.Run("read back the secret", func(t *testing.T) {
		assert.Equal(t, mlOut, string(s.Bytes()))
	})

	t.Run("no_duplicate_keys", func(t *testing.T) {
		assert.Equal(t, []string{"password", "url", "username"}, s.Keys())
	})

	t.Run("read some keys", func(t *testing.T) {
		for k, v := range map[string]string{
			"password": "somepasswd",
			"url":      "http://www.test.com/",
			"username": "myuser@test.com",
		} {
			assert.Equal(t, v, s.Get(k))
		}
	})

	t.Run("remove a key", func(t *testing.T) {
		s.Set("foobar", "baz")
		assert.Equal(t, "baz", s.Get("foobar"))
		s.Del("foobar")
		assert.Equal(t, "", s.Get("foobar"))
	})

	t.Run("read the body", func(t *testing.T) {
		body := "Test / test.com\n"
		assert.Equal(t, body, s.GetBody())
		assert.Equal(t, body, s.GetBody())
		assert.Equal(t, body, s.GetBody())
	})
}

func TestKVNoNewLine(t *testing.T) {
	mlValue := `foobar
ab: cd`
	s, err := ParseKV([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, "cd", s.Get("ab"))
}

func TestKVMIME(t *testing.T) {
	in := `passw0rd
foo: bar
zab: 123`
	out := `GOPASS-SECRET-1.0
Password: passw0rd
Foo: bar
Zab: 123
`
	sec, err := ParseKV([]byte(in))
	require.NoError(t, err)
	msec := sec.MIME()
	assert.Equal(t, out, string(msec.Bytes()))
}

// We currently do not support multiple keys in KV, and we order them in lexicographic order when converted to MIME
func TestMultiKeyKVMIME(t *testing.T) {
	in := `passw0rd
zab: 123
foo: baz
foo: bar
`
	out := `GOPASS-SECRET-1.0
Password: passw0rd
Foo: bar
Zab: 123
`
	sec, err := ParseKV([]byte(in))
	require.NoError(t, err)
	msec := sec.MIME()
	assert.Equal(t, out, string(msec.Bytes()))
}
