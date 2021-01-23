package cache

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnDisk(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(td)
	}()

	ogh := os.Getenv("GOPASS_HOMEDIR")
	os.Setenv("GOPASS_HOMEDIR", td)
	defer func() {
		os.Setenv("GOPASS_HOMEDIR", ogh)
	}()

	odc, err := NewOnDisk("test", time.Hour)
	assert.NoError(t, err)

	assert.NoError(t, odc.Set("foo", []string{"bar"}))
	res, err := odc.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar"}, res)
}
