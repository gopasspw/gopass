package sub

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/store/secret"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	assert.NoError(t, s.Set(ctx, "zab/zab", secret.New("foo", "bar")))
	assert.Error(t, s.Set(ctx, "../../../../../etc/passwd", secret.New("foo", "bar")))
	assert.Error(t, s.Set(ctx, "zab", secret.New("foo", "bar")))
	assert.Error(t, s.Set(WithRecipientFunc(ctx, func(ctx context.Context, prompt string, list []string) ([]string, error) {
		return nil, fmt.Errorf("aborted")
	}), "zab/baz", secret.New("foo", "bar")))
}
