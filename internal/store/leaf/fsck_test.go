package leaf

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/backend/rcs/noop"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsck(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithExportKeys(ctx, false)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// common setup
	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		rcs:     noop.New(),
		storage: fs.New(tempdir),
	}
	assert.NoError(t, s.saveRecipients(ctx, []string{"john.doe"}, "test"))

	for _, e := range []string{"foo/bar", "foo/baz", "foo/zab"} {
		assert.NoError(t, s.Set(ctx, e, secret.New("bar", "")))
	}

	assert.NoError(t, s.Fsck(ctx, ""))
	obuf.Reset()

	// common tear down
	_ = os.RemoveAll(tempdir)
}
func TestCompareStringSlices(t *testing.T) {
	want := []string{"foo", "bar"}
	have := []string{"baz", "bar"}

	missing, extra := compareStringSlices(want, have)
	assert.Equal(t, []string{"foo"}, missing)
	assert.Equal(t, []string{"baz"}, extra)
}
