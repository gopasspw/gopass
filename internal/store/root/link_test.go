package root

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkTarget(t *testing.T) {
	ctx := config.NewContextInMemory()

	u := gptest.NewUnitTester(t)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NoError(t, rs.Link(ctx, "foo", "bar"))

	target, isLink, err := rs.LinkTarget(ctx, "bar")
	require.NoError(t, err)
	assert.True(t, isLink)
	assert.Equal(t, "foo", target)

	target, isLink, err = rs.LinkTarget(ctx, "foo")
	require.NoError(t, err)
	assert.False(t, isLink)
	assert.Empty(t, target)
}
