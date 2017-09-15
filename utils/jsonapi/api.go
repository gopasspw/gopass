package jsonapi

import (
	"context"

	"github.com/justwatchcom/gopass/store/root"
	"github.com/urfave/cli"
)

// API type holding store and context
type API struct {
	Store      *root.Store
	Context    context.Context
	CliContext *cli.Context
}

// ReadAndRespond a single message via stdin/stdout
func (api *API) ReadAndRespond() error {
	message, err := readMessage()
	if message == nil || err != nil {
		return err
	}

	return api.respondMessage(message)
}

// RespondError sends err as JSON response via stdout
func (api *API) RespondError(err error) error {
	var response errorResponse
	response.Error = err.Error()

	return sendSerializedJSONMessage(response)
}
