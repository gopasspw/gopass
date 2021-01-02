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

	v, found := s.Get("Test / test.com")
	assert.False(t, found)
	assert.Equal(t, "", v)

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
			"password": "bar",
			"url":      "http://www.test.com/",
			"username": "myuser@test.com",
		} {
			fv, found := s.Get(k)
			assert.True(t, found)
			assert.Equal(t, v, fv)
		}
		assert.Equal(t, "somepasswd", s.Password())
	})

	t.Run("remove a key", func(t *testing.T) {
		s.Set("foobar", "baz")
		v, ok := s.Get("foobar")
		assert.True(t, ok)
		assert.Equal(t, "baz", v)
		s.Del("foobar")
		v, ok = s.Get("foobar")
		assert.False(t, ok)
		assert.Equal(t, "", v)
	})

	t.Run("read the body", func(t *testing.T) {
		body := "Test / test.com\n"
		assert.Equal(t, body, s.Body())
		assert.Equal(t, body, s.Body())
		assert.Equal(t, body, s.Body())
	})
}

func TestKVNoNewLine(t *testing.T) {
	mlValue := `foobar
ab: cd`
	s, err := ParseKV([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)
	v, _ := s.Get("ab")
	assert.Equal(t, "cd", v)
}
