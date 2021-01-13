package jsonapi

import (
	"context"
	"io"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/pkg/ctxutil"
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
	ctx = ctxutil.WithHidden(ctx, true)
	message, err := readMessage(api.Reader)
	if message == nil || err != nil {
		return err
	}

	return api.respondMessage(ctx, message)
}

// RespondError sends err as JSON response
func (api *API) RespondError(err error) error {
	return sendSerializedJSONMessage(errorResponse{
		Error: err.Error(),
	}, api.Writer)
}
