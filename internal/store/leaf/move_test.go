package leaf

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	plain "github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/secrets"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithExportKeys(ctx, false)

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
					nsec := &secrets.Plain{}
					nsec.SetPassword("bar")
					assert.NoError(t, s.Set(ctx, "foo", nsec))
					assert.NoError(t, s.Copy(ctx, "foo", "bar"))
					sec, err := s.Get(ctx, "foo")
					require.NoError(t, err)
					assert.Equal(t, "bar", sec.Password())
					sec, err = s.Get(ctx, "bar")
					require.NoError(t, err)
					assert.Equal(t, "bar", sec.Password())
				}
			},
		},
		{
			name: "Recursive",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					sec := &secrets.Plain{}
					sec.SetPassword("baz")
					assert.NoError(t, s.Set(ctx, "foo/bar/baz", sec))
					sec.SetPassword("zab")
					assert.NoError(t, s.Set(ctx, "foo/bar/zab", sec))
					assert.Error(t, s.Copy(ctx, "foo", "bar"))
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		require.NoError(t, err)

		s := &Store{
			alias:   "",
			path:    tempdir,
			crypto:  plain.New(),
			storage: fs.New(tempdir),
		}

		assert.NoError(t, s.saveRecipients(ctx, []string{"john.doe"}, "test"))

		// run test case
		t.Run(tc.name, tc.tf(s))

		obuf.Reset()
		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestMove(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithExportKeys(ctx, false)

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
					nsec := &secrets.Plain{}
					nsec.SetPassword("bar")
					assert.NoError(t, s.Set(ctx, "foo", nsec))
					assert.NoError(t, s.Move(ctx, "foo", "bar"))
					_, err := s.Get(ctx, "foo")
					assert.Error(t, err)

					sec, err := s.Get(ctx, "bar")
					require.NoError(t, err)
					assert.Equal(t, "bar", sec.Password())
				}
			},
		},
		{
			name: "Recursive",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					sec := &secrets.Plain{}
					sec.SetPassword("baz")
					assert.NoError(t, s.Set(ctx, "foo/bar/baz", sec))
					sec.SetPassword("zab")
					assert.NoError(t, s.Set(ctx, "foo/bar/zab", sec))
					assert.Error(t, s.Move(ctx, "foo", "bar"))
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		require.NoError(t, err)

		s := &Store{
			alias:   "",
			path:    tempdir,
			crypto:  plain.New(),
			storage: fs.New(tempdir),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test")
		require.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		obuf.Reset()
		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithExportKeys(ctx, false)

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
					assert.Error(t, s.Delete(ctx, "foo"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					sec := &secrets.Plain{}
					sec.SetPassword("bar")
					assert.NoError(t, s.Set(ctx, "foo", sec))
					assert.NoError(t, s.Delete(ctx, "foo"))
					_, err := s.Get(ctx, "foo")
					assert.Error(t, err)
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		require.NoError(t, err)

		s := &Store{
			alias:   "",
			path:    tempdir,
			crypto:  plain.New(),
			storage: fs.New(tempdir),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test")
		require.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		obuf.Reset()
		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}

func TestPrune(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithExportKeys(ctx, false)

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
					assert.Error(t, s.Prune(ctx, "foo"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					sec := &secrets.Plain{}
					sec.SetPassword("bar")
					assert.NoError(t, s.Set(ctx, "foo", sec))
					assert.NoError(t, s.Prune(ctx, "foo"))

					_, err := s.Get(ctx, "foo")
					assert.Error(t, err)
				}
			},
		},
		{
			name: "Multi entry nested",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					sec := &secrets.Plain{}
					sec.SetPassword("bar")
					assert.NoError(t, s.Set(ctx, "foo/bar/baz", sec))
					assert.NoError(t, s.Set(ctx, "foo/bar/zab", sec))
					assert.NoError(t, s.Prune(ctx, "foo/bar"))

					_, err := s.Get(ctx, "foo/bar/baz")
					assert.Error(t, err)

					_, err = s.Get(ctx, "foo/bar/zab")
					assert.Error(t, err)

					// delete empty folder
					assert.NoError(t, s.Prune(ctx, "foo/"))
					assert.Error(t, s.Prune(ctx, "foo/"))
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		require.NoError(t, err)

		s := &Store{
			alias:   "",
			path:    tempdir,
			crypto:  plain.New(),
			storage: fs.New(tempdir),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test")
		assert.NoError(t, err)

		// run test case
		t.Run(tc.name, tc.tf(s))

		obuf.Reset()
		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}
