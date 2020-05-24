package jsonapi

import (
	"context"
	"io"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// API type holding store and context
type API struct {
	Store   gopass.Store
	Reader  io.Reader
	Writer  io.Writer
	Version semver.Version
}

// ReadAndRespond a single message
func (api *API) ReadAndRespond(ctx context.Context) error {
	silentCtx := out.WithHidden(ctx, true)
	message, err := readMessage(api.Reader)
	if message == nil || err != nil {
		return err
	}

	return api.respondMessage(silentCtx, message)
}

// RespondError sends err as JSON response
func (api *API) RespondError(err error) error {
	var response errorResponse
	response.Error = err.Error()

	return sendSerializedJSONMessage(response, api.Writer)
}
