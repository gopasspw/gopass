package leaf

import (
	"context"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	ctx := context.Background()

	s, err := createSubStore(t)
	require.NoError(t, err)

	sec := secrets.NewAKV()
	sec.SetPassword("foo")
	_, err = sec.Write([]byte("bar"))
	require.NoError(t, err)
	require.NoError(t, s.Set(ctx, "zab/zab", sec))

	if runtime.GOOS != "windows" {
		assert.Error(t, s.Set(ctx, "../../../../../etc/passwd", sec))
	} else {
		assert.NoError(t, s.Set(ctx, "../../../../../etc/passwd", sec))
	}

	assert.NoError(t, s.Set(ctx, "zab", sec))
}
