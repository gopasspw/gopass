package secret

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/justwatchcom/gopass/store"
	"github.com/stretchr/testify/assert"
)

const (
	yamlKey      = "bar"
	yamlValue    = "baz"
	yamlPassword = "zzz"
)

func TestYAMLKeyFromEmptySecret(t *testing.T) {
	t.Logf("Get Key from empty Secret")
	s := &Secret{}
	_, err := s.Value(yamlKey)
	assert.Error(t, err)
}

type inlineB struct {
	B       int
	inlineC `yaml:",inline"`
}

type inlineC struct {
	C int
}

func TestYAMLEncodingError(t *testing.T) {
	s := &Secret{
		data: map[string]interface{}{
			"foo": &struct {
				B       int
				inlineB `yaml:",inline"`
			}{1, inlineB{2, inlineC{3}}},
		},
	}
	assert.Error(t, s.encodeYAML())
}

func TestYAMLKeyToEmptySecret(t *testing.T) {
	t.Logf("Set Key to empty Secret")
	s := &Secret{}
	// write key
	err := s.SetValue(yamlKey, yamlValue)
	assert.NoError(t, err)

	// read back key
	content, err := s.Value(yamlKey)
	assert.NoError(t, err)
	assert.Equal(t, yamlValue, string(content))

	// read back whole entry
	buf, err := s.Bytes()
	assert.NoError(t, err)
	want := "\n---\n" + yamlKey + ": " + yamlValue + "\n"
	assert.Equal(t, want, string(buf))
}

func TestYAMLKeyFromPWOnlySecret(t *testing.T) {
	t.Logf("Get key from password-only secret")
	s, err := Parse([]byte(yamlPassword))
	assert.NoError(t, err)

	// read (non-existing) key
	_, err = s.Value(yamlKey)
	assert.Error(t, err)

	// read back whole entry
	content, err := s.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, string(yamlPassword), string(content))
}

func TestYAMLKeyToPWOnlySecret(t *testing.T) {
	t.Logf("Set key to password-only secret")
	s, err := Parse([]byte(yamlPassword))
	assert.NoError(t, err)

	// set new key
	assert.NoError(t, s.SetValue(yamlKey, yamlValue))

	// read back the password
	assert.Equal(t, yamlPassword, s.Password())

	// read back the key
	content, err := s.Value(yamlKey)
	assert.NoError(t, err)
	assert.Equal(t, yamlValue, string(content))

	// read back whole entry
	bv, err := s.Bytes()
	assert.NoError(t, err)
	want := yamlPassword + "\n---\nbar: baz\n"
	assert.Equal(t, want, string(bv))
}

func TestBareYAMLReadKey(t *testing.T) {
	t.Logf("Bare YAML - no document marker - read key")
	in := "bar: baz\nzab: 123\n"
	s, err := Parse([]byte(in))
	assert.NoError(t, err)

	// read back a key
	_, err = s.Value(yamlKey)
	if err != store.ErrYAMLNoMark {
		t.Fatalf("Should fail to read YAML without document marker")
	}
	// read back whole entry
	content, err := s.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, in, string(content)+"\n")
}

func TestYAMLSetMultipleKeys(t *testing.T) {
	t.Logf("Set multiple keys to a secret")
	s, err := Parse([]byte(yamlPassword))
	assert.NoError(t, err)

	want := yamlPassword + "\n---\n"
	numKey := 100
	for i := 0; i < numKey; i++ {
		// set key
		key := fmt.Sprintf("%s-%d", yamlKey, i)
		if err := s.SetValue(key, yamlValue); err != nil {
			t.Fatalf("Failed to write new key: %s", err)
			continue
		}
		want += key + ": " + yamlValue + "\n"
	}

	// read back the password
	assert.Equal(t, yamlPassword, s.Password())

	// read back the keys
	for i := 0; i < numKey; i++ {
		key := yamlKey + "-" + strconv.Itoa(i)
		content, err := s.Value(key)
		if err != nil {
			t.Fatalf("Failed to read key %s: %s", key, err)
		}
		assert.Equal(t, yamlValue, content)
	}

	// read back whole entry
	content, err := s.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, want, string(content))
}

func TestYAMLMultilineWithDashes(t *testing.T) {
	t.Logf("Get Multi-Line Value containing three dashes")
	mlValue := `-----BEGIN PGP PRIVATE KEY BLOCK-----
aaa
bbb
ccc
-----END PGP PRIVATE KEY BLOCK-----`
	s := &Secret{}
	// write key
	assert.NoError(t, s.SetValue(yamlKey, mlValue))

	// read back key
	content, err := s.Value(yamlKey)
	assert.NoError(t, err)
	assert.Equal(t, mlValue, string(content))
}

func TestYAMLContentFromInvalidYAML(t *testing.T) {
	t.Logf("Retrieve content from invalid YAML (#375)")
	mlValue := `somepasswd
---
Test / test.com
username: myuser@test.com
password: somepasswd
url: http://www.test.com/`
	s, err := Parse([]byte(mlValue))
	assert.Error(t, err)
	assert.NotNil(t, s)

	// read back key
	assert.Equal(t, mlValue, s.String())
}

func TestYAMLDocMarkerAsPW(t *testing.T) {
	t.Logf("Document Marker as Password (#398)")
	mlValue := `---`
	s, err := Parse([]byte(mlValue))
	if err != nil {
		t.Logf("%s", err)
	}
	assert.NotNil(t, s)

	// read back key
	assert.Equal(t, "---", s.Password())
}

func TestYAMLBodyWithoutPW(t *testing.T) {
	t.Logf("YAML Body without Password (#398)")
	mlValue := `---
username: myuser@test.com
password: somepasswd
url: http://www.test.com/`
	s, err := Parse([]byte(mlValue))
	if err != nil {
		t.Logf("%s", err)
	}
	assert.NotNil(t, s)
	t.Logf("Data: %+v", s.Data())

	// read back key
	val, err := s.Value("username")
	assert.NoError(t, err)
	assert.Equal(t, "myuser@test.com", val)
}
