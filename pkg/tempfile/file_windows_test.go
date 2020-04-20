package tempfile

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalPrefix(t *testing.T) {
	assertPrefix := func(file *File, prefix string) {
		// windows specific: use os.TempDir() rather than tempdirBase()
		// to compare the prefixes
		requirePrefix := filepath.Join(os.TempDir(), prefix)
		assert.True(t, strings.HasPrefix(file.Name(), requirePrefix))
	}
	ctx := context.Background()
	assert.Equal(t, "", globalPrefix)

	// without global prefix
	withoutGlobalPrefix, err := New(ctx, "some-prefix")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, withoutGlobalPrefix.Close())
	}()
	assertPrefix(withoutGlobalPrefix, "some-prefix")

	// with global prefix
	globalPrefix = "global-prefix."
	withGlobalPrefix, err := New(ctx, "some-prefix")
	assert.NoError(t, err)
	defer func() {
		globalPrefix = ""
		assert.NoError(t, withGlobalPrefix.Close())
	}()
	assertPrefix(withGlobalPrefix, "global-prefix.some-prefix")
}
