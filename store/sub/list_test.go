package sub

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/backend/crypto/gpg/mock"
	"github.com/justwatchcom/gopass/backend/store/fs"
	gitmock "github.com/justwatchcom/gopass/backend/sync/git/mock"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	ctx := context.Background()

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
		assert.NoError(t, err)

		s := &Store{
			alias:  "",
			path:   tempdir,
			crypto: gpgmock.New(),
			sync:   gitmock.New(),
			store:  fs.New(tempdir),
		}

		assert.NoError(t, s.saveRecipients(ctx, []string{"john.doe"}, "test", false))

		// prepare store
		assert.NoError(t, tc.prep(s))
		obuf.Reset()

		// run test case
		out, err := s.List(ctx, "")
		assert.NoError(t, err)
		assert.Equal(t, tc.out, out)
		obuf.Reset()

		// common tear down
		_ = os.RemoveAll(tempdir)
	}
}
