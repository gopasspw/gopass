package fsutil

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
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
}

func TestTempFiler(t *testing.T) {
	ctx := context.Background()

	tf, err := TempFile(ctx, "gp-test-")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer func() {
		assert.NoError(t, tf.Close())
	}()
	t.Logf("Name: %s", tf.Name())
	if _, err := fmt.Fprintf(tf, "foobar"); err != nil {
		t.Errorf("failed to write: %s", err)
	}
}
