package sub

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/gpg/mock"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	ctx := context.Background()

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
				return s.Set(ctx, "foo", []byte("bar"), "test")
			},
			out: []string{"foo"},
		},
		{
			name: "Multi-entry-single-level",
			prep: func(s *Store) error {
				for _, e := range []string{"foo", "bar", "baz"} {
					if err := s.Set(ctx, e, []byte("bar"), "test"); err != nil {
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
					if err := s.Set(ctx, e, []byte("bar"), "test"); err != nil {
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
			alias: "",
			path:  tempdir,
			gpg:   gpgmock.New(),
		}

		err = s.saveRecipients(ctx, []string{"john.doe"}, "test", false)
		assert.NoError(t, err)

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
