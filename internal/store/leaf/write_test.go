package leaf

import (
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	ctx := config.NewContextReadOnly()

	s, err := createSubStore(t)
	require.NoError(t, err)

	sec := secrets.NewAKV()
	sec.SetPassword("foo")
	_, err = sec.Write([]byte("bar"))
	require.NoError(t, err)
	require.NoError(t, s.Set(ctx, "zab/zab", sec))

	if runtime.GOOS != "windows" {
		require.Error(t, s.Set(ctx, "../../../../../etc/passwd", sec))
	} else {
		require.NoError(t, s.Set(ctx, "../../../../../etc/passwd", sec))
	}

	require.NoError(t, s.Set(ctx, "zab", sec))
}
