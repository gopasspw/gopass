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
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secret"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithExportKeys(ctx, false)

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
				sec := secret.New()
				sec.Set("password", "bar")
				return s.Set(ctx, "foo", sec)
			},
			out: []string{"foo"},
		},
		{
			name: "Multi-entry-single-level",
			prep: func(s *Store) error {
				for _, e := range []string{"foo", "bar", "baz"} {
					sec := secret.New()
					sec.Set("password", "bar")
					if err := s.Set(ctx, e, sec); err != nil {
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
					sec := secret.New()
					sec.Set("password", "bar")
					if err := s.Set(ctx, e, sec); err != nil {
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
			path:    tempdir,
			crypto:  plain.New(),
			storage: fs.New(tempdir),
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
