package secrets

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	yamlKey      = "bar"
	yamlValue    = "baz"
	yamlPassword = "zzz"
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
	s, err := ParseYAML([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)
	v, ok := s.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar\nbaz\n", v)
}

func TestYAMLKeyFromEmptySecret(t *testing.T) {
	t.Logf("Get Key from empty Secret")
	s := &YAML{}
	v, ok := s.Get(yamlKey)
	assert.False(t, ok)
	assert.Equal(t, "", v)
}

type inlineB struct {
	B       int
	inlineC `yaml:",inline"`
}

type inlineC struct {
	C int
}

func TestYAMLEncodingError(t *testing.T) {
	s := &YAML{
		data: map[string]interface{}{
			"foo": &struct {
				B       int
				inlineB `yaml:",inline"`
			}{1, inlineB{2, inlineC{3}}},
		},
	}
	assert.Equal(t, "", string(s.Bytes()))
}

func TestYAMLKeyToEmptySecret(t *testing.T) {
	t.Logf("Set Key to empty Secret")
	s := &YAML{}
	// write key
	s.Set(yamlKey, yamlValue)

	// read back key
	v, ok := s.Get(yamlKey)
	assert.True(t, ok)
	assert.Equal(t, yamlValue, v)

	// read back whole entry
	want := "\n---\n" + yamlKey + ": " + yamlValue + "\n"
	assert.Equal(t, want, string(s.Bytes()))
}

func TestYAMLKeyFromPWOnlySecret(t *testing.T) {
	t.Logf("Get key from password-only secret")
	_, err := ParseYAML([]byte(yamlPassword))
	require.Error(t, err)
}

func TestYAMLKeyToPWOnlySecret(t *testing.T) {
	t.Logf("Set key to password-only secret")
	_, err := ParseYAML([]byte(yamlPassword))
	require.Error(t, err)
}

func TestBareYAMLReadKey(t *testing.T) {
	t.Logf("Bare YAML - no document marker - read key")
	in := "\nbar: baz\nzab: 123\n"
	_, err := ParseYAML([]byte(in))
	require.Error(t, err)
}

func TestYAMLSetMultipleKeys(t *testing.T) {
	t.Logf("Set multiple keys to a secret")
	s := &YAML{
		password: yamlPassword,
	}

	var b strings.Builder
	_, _ = b.WriteString(yamlPassword)
	_, _ = b.WriteString("\n")
	numKey := 100
	keys := make([]string, 0, numKey)
	for i := 0; i < numKey; i++ {
		// set key
		key := fmt.Sprintf("%s-%04d", yamlKey, i)
		s.Set(key, yamlValue)
		keys = append(keys, key)
	}
	_, _ = b.WriteString("---\n")
	sort.Strings(keys)
	for _, key := range keys {
		_, _ = b.WriteString(key)
		_, _ = b.WriteString(": ")
		_, _ = b.WriteString(yamlValue)
		_, _ = b.WriteString("\n")
	}

	// read back the password
	assert.Equal(t, yamlPassword, s.Password())

	// read back the keys
	for _, key := range keys {
		v, ok := s.Get(key)
		assert.True(t, ok)
		assert.Equal(t, yamlValue, v)
	}

	// read back whole entry
	assert.Equal(t, b.String(), string(s.Bytes()))
}

func TestYAMLMultilineWithDashes(t *testing.T) {
	t.Logf("Get Multi-Line Value containing three dashes")
	mlValue := `-----BEGIN PGP PRIVATE KEY BLOCK-----
aaa
bbb
ccc
-----END PGP PRIVATE KEY BLOCK-----`
	s := &YAML{}
	// write key
	s.Set(yamlKey, mlValue)

	// read back key
	v, ok := s.Get(yamlKey)
	assert.True(t, ok)
	assert.Equal(t, mlValue, v)
}

func TestYAMLDocMarkerAsPW(t *testing.T) {
	t.Logf("Document Marker as Password (#398)")
	mlValue := `---`
	_, err := ParseYAML([]byte(mlValue))
	require.Error(t, err)
}

func TestYAMLBodyWithoutPW(t *testing.T) {
	t.Logf("YAML Body without Password (#398)")
	mlValue := `---
username: myuser@test.com
password: somepasswd
url: http://www.test.com/`
	s, err := ParseYAML([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)

	t.Logf("Secret: \n%+v\n%s", s, string(s.Bytes()))

	// read back key
	v, ok := s.Get("username")
	assert.True(t, ok)
	assert.Equal(t, "myuser@test.com", v)
}

func TestYAMLBodyWithPW(t *testing.T) {
	t.Logf("YAML Body with Password (#1607)")
	mlValue := `password
---
username: myuser@test.com
password: somepasswd
url: http://www.test.com/`
	s, err := ParseYAML([]byte(mlValue))
	require.NoError(t, err)
	assert.NotNil(t, s)

	t.Logf("Secret: \n%+v\n%s", s, string(s.Bytes()))

	// read back key
	assert.Equal(t, []string{"password", "url", "username"}, s.Keys())
}
func TestYAMLValues(t *testing.T) {
	s := &YAML{
		data: map[string]interface{}{
			"string": "string",
			"int":    int(32),
			"float":  2.3,
			"slice":  []int{1, 2, 3},
			"map":    map[string]string{"a": "b"},
		},
	}

	get := func(k string) string {
		v, _ := s.Get(k)
		return v
	}

	assert.Equal(t, "string", get("string"))
	assert.Equal(t, "32", get("int"))
	assert.Equal(t, "2.3", get("float"))
	assert.Equal(t, "[1 2 3]", get("slice"))
	assert.Equal(t, "map[a:b]", get("map"))
}

func TestYAMLComplex(t *testing.T) {
	in := `20
---
login: hallo
number: 42
sub:
  subentry: 123
`
	s, err := ParseYAML([]byte(in))
	require.NoError(t, err)
	assert.NotNil(t, s)

	get := func(k string) string {
		v, _ := s.Get(k)
		return v
	}

	assert.Equal(t, "hallo", get("login"))
	assert.Equal(t, "42", get("number"))
	assert.Equal(t, "map[subentry:123]", get("sub"))
	assert.Equal(t, []string{"login", "number", "sub"}, s.Keys())
}
