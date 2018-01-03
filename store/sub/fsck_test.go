package sub

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestFsck(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	_, err = s.Fsck(ctx, "")
	assert.NoError(t, err)
}
