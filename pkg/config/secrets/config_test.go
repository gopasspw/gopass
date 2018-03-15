package secrets

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveLoad(t *testing.T) {
	d := map[string]string{
		"foo": "bar",
	}

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "config.sec")

	assert.NoError(t, save(fn, "foobar", d))
	data, err := load(fn, "foobar")
	assert.NoError(t, err)
	assert.Equal(t, d, data)
}
