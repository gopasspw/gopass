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
password: somepasswd
url: http://www.test.com/
`
	s, err := ParseKV([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)

	assert.Equal(t, "", s.Get("Test / test.com"))

	t.Logf("Secret:\n%+v\n%s\n", s, string(s.Bytes()))

	mlOut := `somepasswd
password: somepasswd
url: http://www.test.com/
username: myuser@test.com
Test / test.com
`
	// read back the secret
	assert.Equal(t, mlOut, string(s.Bytes()))
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
Foo: bar
Password: passw0rd
Zab: 123

`
	sec, err := ParseKV([]byte(in))
	require.NoError(t, err)
	msec := sec.MIME()
	assert.Equal(t, out, string(msec.Bytes()))
}
