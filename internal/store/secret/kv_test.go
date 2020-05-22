package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLFromHereDoc(t *testing.T) {
	t.Logf("Parse K/V w/ HereDoc as YAML, not K/V")
	mlValue := `somepw
---
foo:  |
  bar
  baz
key: value
`
	s, err := Parse([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)
	v, err := s.Value("foo")
	require.NoError(t, err)
	assert.Equal(t, "bar\nbaz\n", v)
}

func TestKVContentFromInvalidYAML(t *testing.T) {
	t.Logf("Retrieve content from invalid YAML (#375)")
	mlValue := `somepasswd
---
Test / test.com
username: myuser@test.com
password: somepasswd
url: http://www.test.com/`
	s, err := Parse([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)
	v, err := s.Value("Test / test.com")
	require.NoError(t, err)
	assert.Equal(t, "", v)

	// read back key
	assert.Equal(t, mlValue, s.String())
}
