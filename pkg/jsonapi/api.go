package jsonapi

import (
	"context"
	"io"

	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/root"
)

// API type holding store and context
type API struct {
	Store  *root.Store
	Reader io.Reader
	Writer io.Writer
}

// ReadAndRespond a single message via stdin/stdout
func (api *API) ReadAndRespond(ctx context.Context) error {
	silentCtx := out.WithHidden(ctx, true)
	message, err := readMessage(api.Reader)
	if message == nil || err != nil {
		return err
	}

	return api.respondMessage(silentCtx, message)
}

// RespondError sends err as JSON response via stdout
func (api *API) RespondError(err error) error {
	var response errorResponse
	response.Error = err.Error()

	return sendSerializedJSONMessage(response, api.Writer)
}
