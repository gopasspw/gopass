package leaf

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = WithCheckRecipients(ctx, false)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// common setup
	tempdir := t.TempDir()

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	rs := recipients.New()
	rs.Add("john.doe")

	require.NoError(t, s.saveRecipients(ctx, rs, "test"))

	for _, e := range []string{"foo/bar", "foo/baz", "foo/zab"} {
		sec := secrets.NewAKV()
		sec.SetPassword("bar")
		require.NoError(t, s.Set(ctx, e, sec))
	}

	require.NoError(t, s.Fsck(ctx, ""))
	obuf.Reset()
}

func TestCompareStringSlices(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		from    []string
		to      []string
		missing []string
		extra   []string
	}{
		{
			name:    "Add foo, remove baz",
			from:    []string{"foo", "bar"},
			to:      []string{"baz", "bar"},
			missing: []string{"foo"},
			extra:   []string{"baz"},
		},
		{
			name:    "Add foo, bar, baz, zab",
			from:    []string{"foo", "bar"},
			to:      []string{"foo", "bar", "bar", "baz", "zab"},
			missing: []string{},
			extra:   []string{"baz", "zab"},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			missing, extra := compareStringSlices(tc.from, tc.to)
			assert.Equal(t, tc.missing, missing)
			assert.Equal(t, tc.extra, extra)
		})
	}
}
