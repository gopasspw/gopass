package leaf

import (
	"bytes"
	"os"
	"testing"

	plain "github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

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
					t.Helper()
					require.Error(t, s.Copy(ctx, "foo", "bar"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					t.Helper()
					nsec := secrets.NewAKV()
					nsec.SetPassword("bar")
					require.NoError(t, s.Set(ctx, "foo", nsec))
					require.NoError(t, s.Copy(ctx, "foo", "bar"))
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
					t.Helper()
					sec := secrets.NewAKV()
					sec.SetPassword("baz")
					require.NoError(t, s.Set(ctx, "foo/bar/baz", sec))
					sec.SetPassword("zab")
					require.NoError(t, s.Set(ctx, "foo/bar/zab", sec))
					require.Error(t, s.Copy(ctx, "foo", "bar"))
				}
			},
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
			require.NoError(t, s.saveRecipients(ctx, rs, "test"))

			// run test case
			t.Run(tc.name, tc.tf(s))
		})
	}
}

func TestMove(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

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
					t.Helper()
					require.Error(t, s.Move(ctx, "foo", "bar"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					t.Helper()
					nsec := secrets.NewAKV()
					nsec.SetPassword("bar")
					require.NoError(t, s.Set(ctx, "foo", nsec))
					require.NoError(t, s.Move(ctx, "foo", "bar"))
					_, err := s.Get(ctx, "foo")
					require.Error(t, err)

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
					t.Helper()
					sec := secrets.NewAKV()
					sec.SetPassword("baz")
					require.NoError(t, s.Set(ctx, "foo/bar/baz", sec))
					sec.SetPassword("zab")
					require.NoError(t, s.Set(ctx, "foo/bar/zab", sec))
					require.Error(t, s.Move(ctx, "foo", "bar"))
				}
			},
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

			require.NoError(t, s.saveRecipients(ctx, rs, "test"))

			// run test case
			t.Run(tc.name, tc.tf(s))
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

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
					t.Helper()
					require.Error(t, s.Delete(ctx, "foo"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					t.Helper()
					sec := secrets.NewAKV()
					sec.SetPassword("bar")
					require.NoError(t, s.Set(ctx, "foo", sec))
					require.NoError(t, s.Delete(ctx, "foo"))
					_, err := s.Get(ctx, "foo")
					require.Error(t, err)
				}
			},
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

			require.NoError(t, s.saveRecipients(ctx, rs, "test"))

			// run test case
			t.Run(tc.name, tc.tf(s))
		})
	}
}

func TestPrune(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = config.NewInMemory().WithConfig(ctx)

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
					t.Helper()
					require.Error(t, s.Prune(ctx, "foo"))
				}
			},
		},
		{
			name: "Single entry",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					t.Helper()
					sec := secrets.NewAKV()
					sec.SetPassword("bar")
					require.NoError(t, s.Set(ctx, "foo", sec))
					require.NoError(t, s.Prune(ctx, "foo"))

					_, err := s.Get(ctx, "foo")
					require.Error(t, err)
				}
			},
		},
		{
			name: "Multi entry nested",
			tf: func(s *Store) func(t *testing.T) {
				return func(t *testing.T) {
					t.Helper()
					sec := secrets.NewAKV()
					sec.SetPassword("bar")
					require.NoError(t, s.Set(ctx, "foo/bar/baz", sec))
					require.NoError(t, s.Set(ctx, "foo/bar/zab", sec))
					require.NoError(t, s.Prune(ctx, "foo/bar"))

					_, err := s.Get(ctx, "foo/bar/baz")
					require.Error(t, err)

					_, err = s.Get(ctx, "foo/bar/zab")
					require.Error(t, err)

					// delete empty folder
					require.Error(t, s.Prune(ctx, "foo/"), "delete non-existing entry")
				}
			},
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

			require.NoError(t, s.saveRecipients(ctx, rs, "test"))

			// run test case
			t.Run(tc.name, tc.tf(s))
		})
	}
}
