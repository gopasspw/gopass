package cache

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnDisk(t *testing.T) { //nolint:paralleltest
	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(td)
	}()

	t.Setenv("GOPASS_HOMEDIR", td)

	odc, err := NewOnDisk("test", time.Hour)
	assert.NoError(t, err)

	assert.NoError(t, odc.Set("foo", []string{"bar"}))
	res, err := odc.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar"}, res)
}
