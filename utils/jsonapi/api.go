package jsonapi

import (
	"context"

	"io"
	"os"

	"github.com/justwatchcom/gopass/store/root"
	"github.com/urfave/cli"
)

// API type holding store and context
type API struct {
	Store      *root.Store
	CliContext *cli.Context
	Reader     io.Reader
	Writer     io.Writer
}

// ReadAndRespond a single message via stdin/stdout
func (api *API) ReadAndRespond(ctx context.Context) error {
	message, err := readMessage(os.Stdin)
	if message == nil || err != nil {
		return err
	}

	return api.respondMessage(ctx, message)
}

// RespondError sends err as JSON response via stdout
func (api *API) RespondError(err error) error {
	var response errorResponse
	response.Error = err.Error()

	return sendSerializedJSONMessage(response, api.Writer)
}
