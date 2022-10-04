package leaf

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tempdir := t.TempDir()
	t.Logf(tempdir)

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	sec := &secrets.Plain{}
	sec.SetPassword("foo")
	sec.WriteString("bar")
	require.NoError(t, s.Set(ctx, "zab/zab", sec))

	assert.NoError(t, s.Link(ctx, "zab/zab", "foo/123"))

	p, err := s.Get(ctx, "foo/123")
	require.NoError(t, err)
	assert.Equal(t, "foo", p.Password())
}
