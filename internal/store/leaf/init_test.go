package leaf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	ctx := context.Background()

	s, err := createSubStore(t)
	assert.NoError(t, err)

	assert.Error(t, s.Init(ctx, "", "0xDEADBEEF"))
}
