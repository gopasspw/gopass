package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/store/secret"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	assert.NoError(t, s.Set(ctx, "zab/zab", secret.New("foo", "bar")))
}
