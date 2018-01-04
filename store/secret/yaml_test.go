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
	if err != nil {
		t.Fatalf("%s", err)
	}
	// read back key
	content, err := s.Value(yamlKey)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(content) != yamlValue {
		t.Errorf("Wrong value: %s", content)
	}
	// read back whole entry
	buf, err := s.Bytes()
	if err != nil {
		t.Fatalf("%s", err)
	}
	want := "\n---\n" + yamlKey + ": " + yamlValue + "\n"
	if string(buf) != want {
		t.Errorf("Wrong value: '%s' != '%s'", content, want)
	}
}

func TestYAMLKeyFromPWOnlySecret(t *testing.T) {
	t.Logf("Get key from password-only secret")
	s, err := Parse([]byte(yamlPassword))
	if err != nil {
		t.Fatalf("faile to parse secret")
	}
	// read (non-existing) key
	_, err = s.Value(yamlKey)
	if err == nil {
		t.Errorf("Should complain about missing YAML marker")
	}
	// read back whole entry
	content, err := s.Bytes()
	if err != nil {
		t.Fatalf("%s", err)
	}
	want := string(yamlPassword)
	if string(content) != want+"\n" {
		t.Errorf("Wrong value: '%s' != '%s'", content, want)
	}
}

func TestYAMLKeyToPWOnlySecret(t *testing.T) {
	t.Logf("Set key to password-only secret")
	s, err := Parse([]byte(yamlPassword))
	if err != nil {
		t.Fatalf("%s", err)
	}
	// set new key
	err = s.SetValue(yamlKey, yamlValue)
	if err != nil {
		t.Fatalf("Failed to write new key: %s", err)
	}
	// read back the password
	if s.Password() != yamlPassword {
		t.Errorf("Wrong password: %s", s.Password())
	}
	// read back the key
	content, err := s.Value(yamlKey)
	if err != nil {
		t.Fatalf("Failed to read key %s: %s", yamlKey, err)
	}
	if string(content) != yamlValue {
		t.Errorf("Wrong value: %s", content)
	}
	// read back whole entry
	bv, err := s.Bytes()
	if err != nil {
		t.Fatalf("%s", err)
	}
	want := yamlPassword + "\n---\nbar: baz\n"
	if string(bv) != want {
		t.Errorf("Wrong value: '%s' != '%s'", content, want)
	}
}

func TestBareYAMLReadKey(t *testing.T) {
	t.Logf("Bare YAML - no document marker - read key")
	in := "bar: baz\nzab: 123\n"
	s, err := Parse([]byte(in))
	if err != nil {
		t.Fatalf("%s", err)
	}
	// read back a key
	_, err = s.Value(yamlKey)
	if err != store.ErrYAMLNoMark {
		t.Fatalf("Should fail to read YAML without document marker")
	}
	// read back whole entry
	content, err := s.Bytes()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(content)+"\n" != in {
		t.Errorf("Wrong value: '%s' != '%s'", content, in)
	}
}

func TestYAMLSetMultipleKeys(t *testing.T) {
	t.Logf("Set multiple keys to a secret")
	s, err := Parse([]byte(yamlPassword))
	if err != nil {
		t.Fatalf("%s", err)
	}
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
	if s.Password() != yamlPassword {
		t.Errorf("Wrong password: %s", s.Password())
	}
	// read back the keys
	for i := 0; i < numKey; i++ {
		key := yamlKey + "-" + strconv.Itoa(i)
		content, err := s.Value(key)
		if err != nil {
			t.Fatalf("Failed to read key %s: %s", key, err)
		}
		if content != yamlValue {
			t.Errorf("Wrong value: %s", content)
		}
	}
	// read back whole entry
	content, err := s.Bytes()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(content) != want {
		t.Errorf("Wrong value: '%s' != '%s'", content, want)
	}
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
	err := s.SetValue(yamlKey, mlValue)
	if err != nil {
		t.Fatalf("%s", err)
	}
	// read back key
	content, err := s.Value(yamlKey)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(content) != mlValue {
		t.Errorf("Wrong value: '%s' - Expected: '%s'", content, mlValue)
	}
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
	if err != nil {
		t.Logf("%s", err)
	}
	if s == nil {
		t.Fatalf("Secret is nil")
	}
	// read back key
	if s.String() != mlValue {
		t.Errorf("Decoded Secret does not match input")
	}
}
func TestYAMLDocMarkerAsPW(t *testing.T) {
	t.Logf("Document Marker as Password (#398)")
	mlValue := `---`
	s, err := Parse([]byte(mlValue))
	if err != nil {
		t.Logf("%s", err)
	}
	if s == nil {
		t.Fatalf("Secret is nil")
	}
	// read back key
	if s.Password() != "---" {
		t.Errorf("Secret does not match input")
	}
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
	if s == nil {
		t.Fatalf("Secret is nil")
	}
	t.Logf("Data: %+v", s.Data())
	// read back key
	val, err := s.Value("username")
	if err != nil {
		t.Fatalf("Failed to read username: %s", err)
	}
	if val != "myuser@test.com" {
		t.Errorf("Decoded Secret does not match input")
	}
}
