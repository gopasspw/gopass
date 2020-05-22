package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONAPI(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.NoError(t, act.JSONAPI(clictx(ctx, t)))
	buf.Reset()

	b, err := act.getBrowser(ctx, clictx(ctx, t))
	assert.NoError(t, err)
	assert.Equal(t, b, "chrome")
}
