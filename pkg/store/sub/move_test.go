package sub

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	plain "github.com/gopasspw/gopass/pkg/backend/crypto/plain"
	noop "github.com/gopasspw/gopass/pkg/backend/rcs/noop"
	"github.com/gopasspw/gopass/pkg/backend/storage/fs"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"

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
					assert.Error(t, s.Copy(ctx, "foo", "bar"))
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		assert.NoError(t, err)

		s := &Store{
			alias:   "",
			url:     backend.FromPath(tempdir),
			crypto:  plain.New(),
			rcs:     noop.New(),
			storage: fs.New(tempdir),
			sc:      &fakeConfig{},
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
					assert.NoError(t, s.Set(ctx, "foo", secret.New("bar", "")))
					assert.NoError(t, s.Move(ctx, "foo", "bar"))
					_, err := s.Get(ctx, "foo")
					assert.Error(t, err)

					sec, err := s.Get(ctx, "bar")
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
					assert.Error(t, s.Move(ctx, "foo", "bar"))
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		assert.NoError(t, err)

		s := &Store{
			alias:   "",
			url:     backend.FromPath(tempdir),
			crypto:  plain.New(),
			rcs:     noop.New(),
			storage: fs.New(tempdir),
			sc:      &fakeConfig{},
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
					assert.Error(t, s.Delete(ctx, "foo"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					assert.NoError(t, s.Set(ctx, "foo", secret.New("bar", "")))
					assert.NoError(t, s.Delete(ctx, "foo"))
					_, err := s.Get(ctx, "foo")
					assert.Error(t, err)
				}
			},
		},
	} {
		// common setup
		tempdir, err := ioutil.TempDir("", "gopass-")
		assert.NoError(t, err)

		s := &Store{
			alias:   "",
			url:     backend.FromPath(tempdir),
			crypto:  plain.New(),
			rcs:     noop.New(),
			storage: fs.New(tempdir),
			sc:      &fakeConfig{},
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
					assert.Error(t, s.Prune(ctx, "foo"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					assert.NoError(t, s.Set(ctx, "foo", secret.New("bar", "")))
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
					assert.NoError(t, s.Set(ctx, "foo/bar/baz", secret.New("bar", "")))
					assert.NoError(t, s.Set(ctx, "foo/bar/zab", secret.New("bar", "")))
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
		assert.NoError(t, err)

		s := &Store{
			alias:   "",
			url:     backend.FromPath(tempdir),
			crypto:  plain.New(),
			rcs:     noop.New(),
			storage: fs.New(tempdir),
			sc:      &fakeConfig{},
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
