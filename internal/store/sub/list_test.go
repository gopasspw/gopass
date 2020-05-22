package sub

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	plain "github.com/gopasspw/gopass/internal/backend/crypto/plain"
	noop "github.com/gopasspw/gopass/internal/backend/rcs/noop"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/secret"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	ctx := context.Background()
	ctx = WithExportKeys(ctx, false)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	for _, tc := range []struct {
		name string
		prep func(s *Store) error
		out  []string
	}{
		{
			name: "Empty store",
			prep: func(s *Store) error { return nil },
			out:  []string{},
		},
		{
			name: "Single entry",
			prep: func(s *Store) error {
				return s.Set(ctx, "foo", secret.New("bar", ""))
			},
			out: []string{"foo"},
		},
		{
			name: "Multi-entry-single-level",
			prep: func(s *Store) error {
				for _, e := range []string{"foo", "bar", "baz"} {
					if err := s.Set(ctx, e, secret.New("bar", "")); err != nil {
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
					if err := s.Set(ctx, e, secret.New("bar", "")); err != nil {
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
		require.NoError(t, err)

		s := &Store{
			alias:   "",
			url:     backend.FromPath(tempdir),
			crypto:  plain.New(),
			rcs:     noop.New(),
			storage: fs.New(tempdir),
			sc:      &fakeConfig{},
		}

		assert.NoError(t, s.saveRecipients(ctx, []string{"john.doe"}, "test"))

		// prepare store
		assert.NoError(t, tc.prep(s))
		obuf.Reset()

		// run test case
		out, err := s.List(ctx, "")
		require.NoError(t, err)
		assert.Equal(t, tc.out, out)
		obuf.Reset()

		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}
