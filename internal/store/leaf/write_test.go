package leaf

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/secret"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	sec := secret.New()
	sec.Set("password", "foo")
	sec.WriteString("bar")
	require.NoError(t, s.Set(ctx, "zab/zab", sec))
	if runtime.GOOS != "windows" {
		assert.Error(t, s.Set(ctx, "../../../../../etc/passwd", sec))
	} else {
		assert.NoError(t, s.Set(ctx, "../../../../../etc/passwd", sec))
	}
	assert.NoError(t, s.Set(ctx, "zab", sec))
	assert.Error(t, s.Set(WithRecipientFunc(ctx, func(ctx context.Context, prompt string, list []string) ([]string, error) {
		return nil, fmt.Errorf("aborted")
	}), "zab/baz", sec))
}
