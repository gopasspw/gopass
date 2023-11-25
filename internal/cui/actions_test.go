package cui

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCreateActions(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())
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
	require.Error(t, cas.Run(ctx, nil, 0))
	require.NoError(t, cas.Run(ctx, nil, 1))
	require.Error(t, cas.Run(ctx, nil, 2))
	require.Error(t, cas.Run(ctx, nil, 66))
}
