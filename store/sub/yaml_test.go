package sub

import (
	"io/ioutil"
	"os"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/gpg/mock"
	"github.com/justwatchcom/gopass/store"
)

const (
	yamlSecret   = "foo"
	yamlKey      = "bar"
	yamlValue    = "baz"
	yamlPassword = "zzz"
)

/*
- empty doc (get/set)
- only pw (get/set)
- no pw (get/set)
- pw and yaml (get/set)
- no sep / sep (get/set)
*/
func TestYAML(t *testing.T) {
	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Get Key from empty Secret",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					_, err := s.GetKey(yamlSecret, yamlKey)
					if err == nil {
						t.Errorf("Should complain about missing YAML marker")
					}
				}
			},
		},
		{
			name: "Set Key to empty Secret",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					// write key
					err := s.SetKey(yamlSecret, yamlKey, yamlValue)
					if err != nil {
						t.Fatalf("%s", err)
					}
					// read back key
					content, err := s.GetKey(yamlSecret, yamlKey)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if string(content) != yamlValue {
						t.Errorf("Wrong value: %s", content)
					}
					// read back whole entry
					content, err = s.Get(yamlSecret)
					if err != nil {
						t.Fatalf("%s", err)
					}
					want := "\n---\n" + yamlKey + ": " + yamlValue + "\n"
					if string(content) != want {
						t.Errorf("Wrong value: '%s' != '%s'", content, want)
					}
				}
			},
		},
		{
			name: "Get key from password-only secret",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					// write password
					err := s.Set(yamlSecret, []byte(yamlPassword), "testing")
					if err != nil {
						t.Fatalf("%s", err)
					}
					// read (non-existing) key
					_, err = s.GetKey(yamlSecret, yamlKey)
					if err == nil {
						t.Errorf("Should complain about missing YAML marker")
					}
					// read back whole entry
					content, err := s.Get(yamlSecret)
					if err != nil {
						t.Fatalf("%s", err)
					}
					want := string(yamlPassword)
					if string(content) != want {
						t.Errorf("Wrong value: '%s' != '%s'", content, want)
					}
				}
			},
		},
		{
			name: "Set key to password-only secret",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					// write password
					err := s.Set(yamlSecret, []byte(yamlPassword), "testing")
					if err != nil {
						t.Fatalf("%s", err)
					}
					// set new key
					err = s.SetKey(yamlSecret, yamlKey, yamlValue)
					if err != nil {
						t.Fatalf("Failed to write new key: %s", err)
					}
					// read back the password
					pw, err := s.GetFirstLine(yamlSecret)
					if err != nil {
						t.Fatalf("Failed to read password: %s", err)
					}
					if string(pw) != yamlPassword {
						t.Errorf("Wrong password: %s", pw)
					}
					// read back the key
					content, err := s.GetKey(yamlSecret, yamlKey)
					if err != nil {
						t.Fatalf("Failed to read key %s: %s", yamlKey, err)
					}
					if string(content) != yamlValue {
						t.Errorf("Wrong value: %s", content)
					}
					// read back whole entry
					content, err = s.Get(yamlSecret)
					if err != nil {
						t.Fatalf("%s", err)
					}
					want := yamlPassword + "\n---\nbar: baz\n"
					if string(content) != want {
						t.Errorf("Wrong value: '%s' != '%s'", content, want)
					}
				}
			},
		},
		{
			name: "Bare YAML - no document marker - read key",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					secret := "bar: baz\nzab: 123\n"
					// write password
					err := s.Set(yamlSecret, []byte(secret), "testing")
					if err != nil {
						t.Fatalf("%s", err)
					}
					// read back a key
					_, err = s.GetKey(yamlSecret, yamlKey)
					if err != store.ErrYAMLNoMark {
						t.Fatalf("Should fail to read YAML without document marker")
					}
					// read back whole entry
					content, err := s.Get(yamlSecret)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if string(content) != secret {
						t.Errorf("Wrong value: '%s' != '%s'", content, secret)
					}
				}
			},
		},
		{
			name: "Set multiple keys to a secret",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					// write password
					err := s.Set(yamlSecret, []byte(yamlPassword), "testing")
					if err != nil {
						t.Fatalf("%s", err)
					}
					// set first key
					err = s.SetKey(yamlSecret, yamlKey+"-1", yamlValue)
					if err != nil {
						t.Fatalf("Failed to write new key: %s", err)
					}
					// set second key
					err = s.SetKey(yamlSecret, yamlKey+"-2", yamlValue)
					if err != nil {
						t.Fatalf("Failed to write new key: %s", err)
					}
					// set third key
					err = s.SetKey(yamlSecret, yamlKey+"-3", yamlValue)
					if err != nil {
						t.Fatalf("Failed to write new key: %s", err)
					}
					// read back the password
					pw, err := s.GetFirstLine(yamlSecret)
					if err != nil {
						t.Fatalf("Failed to read password: %s", err)
					}
					if string(pw) != yamlPassword {
						t.Errorf("Wrong password: %s", pw)
					}
					// read back the keys
					for _, k := range []string{"1", "2", "3"} {
						key := yamlKey + "-" + k
						content, err := s.GetKey(yamlSecret, key)
						if err != nil {
							t.Fatalf("Failed to read key %s: %s", key, err)
						}
						if string(content) != yamlValue {
							t.Errorf("Wrong value: %s", content)
						}
					}
					// read back whole entry
					content, err := s.Get(yamlSecret)
					if err != nil {
						t.Fatalf("%s", err)
					}
					want := yamlPassword + "\n---\nbar-1: baz\nbar-2: baz\nbar-3: baz\n"
					if string(content) != want {
						t.Errorf("Wrong value: '%s' != '%s'", content, want)
					}
				}
			},
		},
		{
			name: "Get Multi-Line Value containing three dashes",
			tf: func(s *Store) func(t *testing.T) {
				mlValue := `-----BEGIN PGP PRIVATE KEY BLOCK-----
aaa
bbb
ccc
-----END PGP PRIVATE KEY BLOCK-----`
				return func(t *testing.T) {
					// write key
					err := s.SetKey(yamlSecret, yamlKey, mlValue)
					if err != nil {
						t.Fatalf("%s", err)
					}
					// read back key
					content, err := s.GetKey(yamlSecret, yamlKey)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if string(content) != mlValue {
						t.Errorf("Wrong value: %s", content)
					}
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		if err != nil {
			t.Fatalf("Failed to create tempdir: %s", err)
		}

		s := &Store{
			alias:      "",
			path:       tempdir,
			gpg:        gpgmock.New(),
			recipients: []string{"john.doe"},
		}

		// run test case
		t.Run(tc.name, tc.tf(s))

		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}
