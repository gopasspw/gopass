package fsutil

import (
	"io/ioutil"
	"os"
	"testing"
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
