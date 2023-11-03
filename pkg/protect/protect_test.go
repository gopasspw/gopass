package protect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtect(t *testing.T) {
	t.Parallel()

	require.NoError(t, Pledge(""))
}
