package leaf

import (
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	ctx := context.Background()

	tempdir, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	color.NoColor = true

	_, _, err = createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx = backend.WithCryptoBackendString(ctx, "plain")
	ctx = backend.WithStorageBackendString(ctx, "fs")
	s, err := New(
		ctx,
		"",
		tempdir,
	)
	require.NoError(t, err)

	assert.Equal(t, 0, len(s.ListTemplates(ctx, "")))
	assert.NoError(t, s.SetTemplate(ctx, "foo", []byte("foobar")))
	assert.Equal(t, 1, len(s.ListTemplates(ctx, "")))

	tt := s.TemplateTree(ctx)
	assert.Equal(t, "gopass\n└── foo\n", tt.Format(0))

	assert.True(t, s.HasTemplate(ctx, "foo"))

	b, err := s.GetTemplate(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(b))

	_, b, found := s.LookupTemplate(ctx, "foo/bar")
	assert.True(t, found)
	assert.Equal(t, "foobar", string(b))

	assert.NoError(t, s.RemoveTemplate(ctx, "foo"))
	assert.Equal(t, 0, len(s.ListTemplates(ctx, "")))

	assert.Error(t, s.RemoveTemplate(ctx, "foo"))
}
