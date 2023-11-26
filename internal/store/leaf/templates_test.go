package leaf

import (
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextReadOnly()

	tempdir := t.TempDir()

	color.NoColor = true

	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx = backend.WithCryptoBackendString(ctx, "plain")
	ctx = backend.WithStorageBackendString(ctx, "fs")
	s, err := New(
		ctx,
		"",
		tempdir,
	)
	require.NoError(t, err)

	assert.Empty(t, s.ListTemplates(ctx, ""))
	require.NoError(t, s.SetTemplate(ctx, "foo", []byte("foobar")))
	assert.Len(t, s.ListTemplates(ctx, ""), 1)

	tt := s.TemplateTree(ctx)
	assert.Equal(t, "gopass\n└── foo\n", tt.Format(0))

	assert.True(t, s.HasTemplate(ctx, "foo"))

	b, err := s.GetTemplate(ctx, "foo")
	require.NoError(t, err)
	assert.Equal(t, "foobar", string(b))

	_, b, found := s.LookupTemplate(ctx, "foo/bar")
	assert.True(t, found)
	assert.Equal(t, "foobar", string(b))

	require.NoError(t, s.RemoveTemplate(ctx, "foo"))
	assert.Empty(t, s.ListTemplates(ctx, ""))

	require.Error(t, s.RemoveTemplate(ctx, "foo"))
}
