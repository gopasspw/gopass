package sub

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/backend/crypto/plain"
	"github.com/justwatchcom/gopass/pkg/backend/rcs/noop"
	"github.com/justwatchcom/gopass/pkg/backend/storage/fs"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store/secret"
	"github.com/stretchr/testify/assert"
)

func TestFsck(t *testing.T) {
	ctx := context.Background()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// common setup
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)

	s := &Store{
		alias:   "",
		url:     backend.FromPath(tempdir),
		crypto:  plain.New(),
		rcs:     noop.New(),
		storage: fs.New(tempdir),
	}
	assert.NoError(t, s.saveRecipients(ctx, []string{"john.doe"}, "test", false))

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
