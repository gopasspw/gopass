//go:build !windows
// +build !windows

package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateIdentity(t *testing.T) {
	ctx := context.Background()

	g := &GPG{}
	g.binary = "true"

	assert.NoError(t, g.GenerateIdentity(ctx, "foo", "foo@bar.com", "bar"))
}
