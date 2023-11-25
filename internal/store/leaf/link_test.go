package leaf

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLink(t *testing.T) {
	ctx := config.NewNoWrites().WithConfig(context.Background())

	s, err := createSubStore(t)
	require.NoError(t, err)

	sec := secrets.NewAKV()
	sec.SetPassword("foo")
	_, err = sec.Write([]byte("bar"))
	require.NoError(t, err)
	require.NoError(t, s.Set(ctx, "zab/zab", sec))

	require.NoError(t, s.Link(ctx, "zab/zab", "foo/123"))

	p, err := s.Get(ctx, "foo/123")
	require.NoError(t, err)
	assert.Equal(t, "foo", p.Password())
}
