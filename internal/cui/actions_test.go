package cui

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCreateActions(t *testing.T) {
	ctx := context.Background()
	cas := Actions{
		{
			Name: "foo",
		},
		{
			Name: "bar",
			Fn: func(context.Context, *cli.Context) error {
				return nil
			},
		},
	}
	assert.Equal(t, []string{"foo", "bar"}, cas.Selection())
	assert.Error(t, cas.Run(ctx, nil, 0))
	assert.NoError(t, cas.Run(ctx, nil, 1))
	assert.Error(t, cas.Run(ctx, nil, 2))
	assert.Error(t, cas.Run(ctx, nil, 66))
}
