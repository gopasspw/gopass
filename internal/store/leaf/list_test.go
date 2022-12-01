package leaf

import (
	"bytes"
	"context"
	"os"
	"testing"

	plain "github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	for _, tc := range []struct { //nolint:paralleltest
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
				sec := secrets.New()
				sec.SetPassword("bar")

				return s.Set(ctx, "foo", sec)
			},
			out: []string{"foo"},
		},
		{
			name: "Multi-entry-single-level",
			prep: func(s *Store) error {
				for _, e := range []string{"foo", "bar", "baz"} {
					sec := secrets.New()
					sec.SetPassword("bar")
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
					sec := secrets.New()
					sec.SetPassword("bar")
					if err := s.Set(ctx, e, sec); err != nil {
						return err
					}
				}

				return nil
			},
			out: []string{"foo/bar", "foo/baz", "foo/zab"},
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			// common setup
			tempdir := t.TempDir()

			defer func() {
				obuf.Reset()
			}()

			s := &Store{
				alias:   "",
				path:    tempdir,
				crypto:  plain.New(),
				storage: fs.New(tempdir),
			}

			rs := recipients.New()
			rs.Add("john.doe")

			assert.NoError(t, s.saveRecipients(ctx, rs, "test"))

			// prepare store
			assert.NoError(t, tc.prep(s))
			obuf.Reset()

			// run test case
			out, err := s.List(ctx, "")
			require.NoError(t, err)
			assert.Equal(t, tc.out, out)
		})
	}
}
