package leaf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tempdir := t.TempDir()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	assert.Error(t, s.Init(ctx, "", "0xDEADBEEF"))
}
