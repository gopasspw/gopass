package sub

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/gpg/mock"
)

func createStore(dir string, recipients, entries []string) ([]string, []string, error) {
	if recipients == nil {
		recipients = []string{
			"0xDEADBEEF",
			"0xFEEDBEEF",
		}
	}
	if entries == nil {
		entries = []string{
			"foo/bar/baz",
			"baz/ing/a",
		}
	}
	sort.Strings(entries)
	for _, file := range entries {
		filename := filepath.Join(dir, file+".gpg")
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return recipients, entries, err
		}
		if err := ioutil.WriteFile(filename, []byte{}, 0644); err != nil {
			return recipients, entries, err
		}
	}
	err := ioutil.WriteFile(filepath.Join(dir, GPGID), []byte(strings.Join(recipients, "\n")), 0600)
	return recipients, entries, err
}

func TestList(t *testing.T) {
	for _, tc := range []struct {
		name string
		prep func(s *Store) error
		out  []string
	}{
		{
			name: "Empty store",
			prep: func(s *Store) error { return nil },
		},
		{
			name: "Single entry",
			prep: func(s *Store) error {
				if err := s.Set("foo", []byte("bar"), "test"); err != nil {
					return err
				}
				return nil
			},
			out: []string{"foo"},
		},
		{
			name: "Multi-entry-single-level",
			prep: func(s *Store) error {
				for _, e := range []string{"foo", "bar", "baz"} {
					if err := s.Set(e, []byte("bar"), "test"); err != nil {
						return err
					}
				}
				return nil
			},
			out: []string{"bar", "baz", "foo"},
		},
		{
			name: "Multi-entry-multi-level",
			prep: func(s *Store) error {
				for _, e := range []string{"foo/bar", "foo/baz", "foo/zab"} {
					if err := s.Set(e, []byte("bar"), "test"); err != nil {
						return err
					}
				}
				return nil
			},
			out: []string{"foo/bar", "foo/baz", "foo/zab"},
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

		// prepare store
		if err := tc.prep(s); err != nil {
			t.Fatalf("Failed to prepare store: %s", err)
		}

		// run test case
		out, err := s.List("")
		if err != nil {
			t.Fatalf("Failed to call List(): %s", err)
		}
		t.Logf("Output: %s", out)
		if strings.Join(out, "\n") != strings.Join(tc.out, "\n") {
			t.Errorf("Mismatched output: %+v vs. %+v", out, tc.out)
		}

		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestCopy(t *testing.T) {
	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Copy("foo", "bar"); err == nil {
						t.Errorf("Should fail to copy non-existing entries in empty store")
					}
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set("foo", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Copy("foo", "bar"); err != nil {
						t.Errorf("Failed to copy 'foo' to 'bar': %s", err)
					}
					content, err := s.Get("foo")
					if err != nil {
						t.Fatalf("Failed to get 'foo': %s", err)
					}
					if string(content) != "bar" {
						t.Errorf("Wrong content in 'foo'")
					}
					content, err = s.Get("bar")
					if err != nil {
						t.Fatalf("Failed to get 'bar': %s", err)
					}
					if string(content) != "bar" {
						t.Errorf("Wrong content in 'bar'")
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

func TestMove(t *testing.T) {
	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Move("foo", "bar"); err == nil {
						t.Errorf("Should fail to move non-existing entries in empty store")
					}
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set("foo", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Move("foo", "bar"); err != nil {
						t.Errorf("Failed to copy 'foo' to 'bar': %s", err)
					}
					_, err := s.Get("foo")
					if err == nil {
						t.Fatalf("Should fail to get 'foo': %s", err)
					}
					content, err := s.Get("bar")
					if err != nil {
						t.Fatalf("Failed to get 'bar': %s", err)
					}
					if string(content) != "bar" {
						t.Errorf("Wrong content in 'bar'")
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

func TestDelete(t *testing.T) {
	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Delete("foo"); err == nil {
						t.Errorf("Should fail to delete non-existing entries in empty store")
					}
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set("foo", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Delete("foo"); err != nil {
						t.Errorf("Failed to delete 'foo': %s", err)
					}
					_, err := s.Get("foo")
					if err == nil {
						t.Fatalf("Should fail to get 'foo': %s", err)
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

func TestPrune(t *testing.T) {
	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Prune("foo"); err == nil {
						t.Errorf("Should fail to delete non-existing entries in empty store")
					}
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set("foo", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Prune("foo"); err != nil {
						t.Errorf("Failed to delete 'foo': %s", err)
					}
					_, err := s.Get("foo")
					if err == nil {
						t.Fatalf("Should fail to get 'foo': %s", err)
					}
				}
			},
		},
		{
			name: "Multi entry nested",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set("foo/bar/baz", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Set("foo/bar/zab", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Prune("foo/bar"); err != nil {
						t.Errorf("Failed to delete 'foo': %s", err)
					}
					_, err := s.Get("foo/bar/baz")
					if err == nil {
						t.Fatalf("Should fail to get 'foo/bar/baz': %s", err)
					}
					_, err = s.Get("foo/bar/zab")
					if err == nil {
						t.Fatalf("Should fail to get 'foo/bar/zab': %s", err)
					}
					// delete empty folder
					if err := s.Prune("foo/"); err != nil {
						t.Errorf("Failed to delete 'foo': %s", err)
					}
					if err := s.Prune("foo/"); err == nil {
						t.Errorf("Should fail to delete 'foo' again")
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
