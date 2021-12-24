package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	aclip "github.com/atotto/clipboard"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	aclip.Unsupported = true

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithNotifications(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	act.cfg.ClipTimeout = 1

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// create
	c := gptest.CliCtx(ctx, t)

	assert.Error(t, act.Create(c))
	buf.Reset()
}
