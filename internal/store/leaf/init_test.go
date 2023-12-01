package leaf

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	ctx := config.NewContextInMemory()

	s, err := createSubStore(t)
	require.NoError(t, err)
	require.Error(t, s.Init(ctx, "", "0xDEADBEEF"))
}
