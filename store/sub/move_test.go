package sub

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	gitmock "github.com/justwatchcom/gopass/backend/git/mock"
	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	ctx := context.Background()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					assert.Error(t, s.Copy(ctx, "foo", "bar"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					assert.NoError(t, s.Set(ctx, "foo", secret.New("bar", "")))
					assert.NoError(t, s.Copy(ctx, "foo", "bar"))
					sec, err := s.Get(ctx, "foo")
					assert.NoError(t, err)
					assert.Equal(t, "bar", sec.Password())
					sec, err = s.Get(ctx, "bar")
					assert.NoError(t, err)
					assert.Equal(t, "bar", sec.Password())
				}
			},
		},
		{
			name: "Recursive",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					assert.NoError(t, s.Set(ctx, "foo/bar/baz", secret.New("baz", "")))
					assert.NoError(t, s.Set(ctx, "foo/bar/zab", secret.New("zab", "")))
					assert.NoError(t, s.Copy(ctx, "foo", "bar"))

					sec, err := s.Get(ctx, "bar/bar/baz")
					assert.NoError(t, err)
					assert.Equal(t, "baz", sec.Password())

					sec, err = s.Get(ctx, "bar/bar/zab")
					assert.NoError(t, err)
					assert.Equal(t, "zab", sec.Password())

					sec, err = s.Get(ctx, "foo/bar/baz")
					assert.NoError(t, err)
					assert.Equal(t, "baz", sec.Password())

					sec, err = s.Get(ctx, "foo/bar/zab")
					assert.NoError(t, err)
					assert.Equal(t, "zab", sec.Password())
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		assert.NoError(t, err)

		s := &Store{
			alias: "",
			path:  tempdir,
			gpg:   gpgmock.New(),
			git:   gitmock.New(),
		}

		assert.NoError(t, s.saveRecipients(ctx, []string{"john.doe"}, "test", false))

		// run test case
		t.Run(tc.name, tc.tf(s))

		obuf.Reset()
		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestMove(t *testing.T) {
	ctx := context.Background()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	for _, tc := range []struct {
		name string
		tf   func(s *Store) func(t *testing.T)
	}{
		{
			name: "Empty store",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					assert.Error(t, s.Move(ctx, "foo", "bar"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					if err := s.Set(ctx, "foo", secret.New("bar", "")); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Move(ctx, "foo", "bar"); err != nil {
						t.Errorf("Failed to copy 'foo' to 'bar': %s", err)
					}
					_, err := s.Get(ctx, "foo")
					if err == nil {
						t.Fatalf("Should fail to get 'foo': %s", err)
					}
					sec, err := s.Get(ctx, "bar")
					if err != nil {
						t.Fatalf("Failed to get 'bar': %s", err)
					}
					if sec.Password() != "bar" {
						t.Errorf("Wrong content in 'bar'")
					}
				}
			},
		},
		{
			name: "Recursive",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					assert.NoError(t, s.Set(ctx, "foo/bar/baz", secret.New("baz", "")))
					assert.NoError(t, s.Set(ctx, "foo/bar/zab", secret.New("zab", "")))
					assert.NoError(t, s.Move(ctx, "foo", "bar"))

					sec, err := s.Get(ctx, "bar/bar/baz")
					assert.NoError(t, err)
					assert.Equal(t, "baz", sec.Password())

					sec, err = s.Get(ctx, "bar/bar/zab")
					assert.NoError(t, err)
					assert.Equal(t, "zab", sec.Password())
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
			git:   gitmock.New(),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test", false)
		assert.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		obuf.Reset()
		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

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
					if err := s.Set(ctx, "foo", secret.New("bar", "")); err != nil {
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
			git:   gitmock.New(),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test", false)
		assert.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		obuf.Reset()
		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestPrune(t *testing.T) {
	ctx := context.Background()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

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
					if err := s.Set(ctx, "foo", secret.New("bar", "")); err != nil {
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
					if err := s.Set(ctx, "foo/bar/baz", secret.New("bar", "")); err != nil {
						t.Fatalf("Failed to insert test data: %s", err)
					}
					if err := s.Set(ctx, "foo/bar/zab", secret.New("bar", "")); err != nil {
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
			git:   gitmock.New(),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test", false)
		assert.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		obuf.Reset()
		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}
