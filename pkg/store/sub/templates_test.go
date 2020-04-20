package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	ctx := context.Background()
	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	color.NoColor = true

	_, _, err = createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx = backend.WithCryptoBackendString(ctx, "plain")
	ctx = backend.WithRCSBackendString(ctx, "noop")
	s, err := New(
		ctx,
		nil,
		"",
		backend.FromPath(tempdir),
		tempdir,
		nil,
	)
	require.NoError(t, err)

	assert.Equal(t, 0, len(s.ListTemplates(ctx, "")))
	assert.NoError(t, s.SetTemplate(ctx, "foo", []byte("foobar")))
	assert.Equal(t, 1, len(s.ListTemplates(ctx, "")))

	tt, err := s.TemplateTree(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "gopass\n└── foo\n", tt.Format(0))

	assert.Equal(t, true, s.HasTemplate(ctx, "foo"))

	b, err := s.GetTemplate(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(b))

	_, b, found := s.LookupTemplate(ctx, "foo/bar")
	assert.Equal(t, true, found)
	assert.Equal(t, "foobar", string(b))

	assert.NoError(t, s.RemoveTemplate(ctx, "foo"))
	assert.Equal(t, 0, len(s.ListTemplates(ctx, "")))

	assert.Error(t, s.RemoveTemplate(ctx, "foo"))
}
