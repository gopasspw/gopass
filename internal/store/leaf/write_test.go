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
	t.Parallel()

	ctx := context.Background()

	tempdir := t.TempDir()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	sec := &secrets.Plain{}
	sec.SetPassword("foo")
	sec.WriteString("bar")
	require.NoError(t, s.Set(ctx, "zab/zab", sec))

	if runtime.GOOS != "windows" {
		assert.Error(t, s.Set(ctx, "../../../../../etc/passwd", sec))
	} else {
		assert.NoError(t, s.Set(ctx, "../../../../../etc/passwd", sec))
	}

	assert.NoError(t, s.Set(ctx, "zab", sec))
}
