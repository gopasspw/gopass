package main

import (
	"context"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

func TestSetupApp(t *testing.T) {
	ctx := context.Background()
	app := setupApp(ctx, semver.Version{})
	assert.NotNil(t, app)
}
