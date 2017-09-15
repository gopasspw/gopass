package action

import (
	"context"

	"github.com/justwatchcom/gopass/utils/jsonapi"
	"github.com/urfave/cli"
)

// JSONAPI reads a json message on stdin and responds on stdout
func (s *Action) JSONAPI(ctx context.Context, c *cli.Context) error {
	api := jsonapi.API{Store: s.Store, Context: ctx, CliContext: c}
	if err := api.ReadAndRespond(); err != nil {
		return api.RespondError(err)
	}
	return nil
}
