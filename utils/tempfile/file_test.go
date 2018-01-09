package tempfile

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTempdirBase(t *testing.T) {
	tempdir, err := ioutil.TempDir(tempdirBase(), "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
}

func TestTempFiler(t *testing.T) {
	ctx := context.Background()

	// regular tempfile
	tf, err := New(ctx, "gp-test-")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, tf.Close())
	}()

	t.Logf("Name: %s", tf.Name())
	_, err = fmt.Fprintf(tf, "foobar")
	assert.NoError(t, err)

	// unintialized tempfile
	utf := File{}
	assert.Equal(t, utf.Name(), "")
	_, err = utf.Write([]byte("foo"))
	assert.Error(t, err)
	assert.NoError(t, utf.Remove(ctx))
	assert.NoError(t, utf.Close())
}
