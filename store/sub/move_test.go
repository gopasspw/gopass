package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/gpg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Copy(ctx, "foo", "bar"); err == nil {
						t.Errorf("Should fail to copy non-existing entries in empty store")
					}
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set(ctx, "foo", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Copy(ctx, "foo", "bar"); err != nil {
						t.Errorf("Failed to copy 'foo' to 'bar': %s", err)
					}
					content, err := s.Get(ctx, "foo")
					if err != nil {
						t.Fatalf("Failed to get 'foo': %s", err)
					}
					if string(content) != "bar" {
						t.Errorf("Wrong content in 'foo'")
					}
					content, err = s.Get(ctx, "bar")
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
			alias: "",
			path:  tempdir,
			gpg:   gpgmock.New(),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test", false)
		assert.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestMove(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Move(ctx, "foo", "bar"); err == nil {
						t.Errorf("Should fail to move non-existing entries in empty store")
					}
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set(ctx, "foo", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Move(ctx, "foo", "bar"); err != nil {
						t.Errorf("Failed to copy 'foo' to 'bar': %s", err)
					}
					_, err := s.Get(ctx, "foo")
					if err == nil {
						t.Fatalf("Should fail to get 'foo': %s", err)
					}
					content, err := s.Get(ctx, "bar")
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
			alias: "",
			path:  tempdir,
			gpg:   gpgmock.New(),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test", false)
		assert.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Delete(ctx, "foo"); err == nil {
						t.Errorf("Should fail to delete non-existing entries in empty store")
					}
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set(ctx, "foo", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Delete(ctx, "foo"); err != nil {
						t.Errorf("Failed to delete 'foo': %s", err)
					}
					_, err := s.Get(ctx, "foo")
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
			alias: "",
			path:  tempdir,
			gpg:   gpgmock.New(),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test", false)
		assert.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestPrune(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Prune(ctx, "foo"); err == nil {
						t.Errorf("Should fail to delete non-existing entries in empty store")
					}
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set(ctx, "foo", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Prune(ctx, "foo"); err != nil {
						t.Errorf("Failed to delete 'foo': %s", err)
					}
					_, err := s.Get(ctx, "foo")
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
					if err := s.Set(ctx, "foo/bar/baz", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Set(ctx, "foo/bar/zab", []byte("bar"), "test"); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Prune(ctx, "foo/bar"); err != nil {
						t.Errorf("Failed to delete 'foo': %s", err)
					}
					_, err := s.Get(ctx, "foo/bar/baz")
					if err == nil {
						t.Fatalf("Should fail to get 'foo/bar/baz': %s", err)
					}
					_, err = s.Get(ctx, "foo/bar/zab")
					if err == nil {
						t.Fatalf("Should fail to get 'foo/bar/zab': %s", err)
					}
					// delete empty folder
					if err := s.Prune(ctx, "foo/"); err != nil {
						t.Errorf("Failed to delete 'foo': %s", err)
					}
					if err := s.Prune(ctx, "foo/"); err == nil {
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
			alias: "",
			path:  tempdir,
			gpg:   gpgmock.New(),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test", false)
		assert.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}
