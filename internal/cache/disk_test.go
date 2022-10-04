package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOnDisk(t *testing.T) { //nolint:paralleltest
	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	odc, err := NewOnDisk("test", time.Hour)
	assert.NoError(t, err)

	assert.NoError(t, odc.Set("foo", []string{"bar"}))
	res, err := odc.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar"}, res)
}
