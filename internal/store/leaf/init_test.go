package leaf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	ctx := context.Background()

	s, err := createSubStore(t)
	require.NoError(t, err)
	require.Error(t, s.Init(ctx, "", "0xDEADBEEF"))
}
