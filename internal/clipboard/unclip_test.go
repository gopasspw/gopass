package clipboard

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
)

func TestUnclip(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.EqualError(t, Clear(ctx, "", false), ErrNotSupported.Error())
}
