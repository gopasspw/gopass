package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestGetCommands(t *testing.T) {
	ctx := context.Background()
	app := cli.NewApp()
	assert.Equal(t, 30, len(getCommands(ctx, nil, app)))
}
