package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/muesli/goprogressbar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAudit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	goprogressbar.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
		goprogressbar.Stdout = os.Stdout
	}()

	// add some entries
	assert.NoError(t, act.Store.Set(ctx, "bar", secret.New("123", "")))
	assert.NoError(t, act.Store.Set(ctx, "baz", secret.New("123", "")))

	assert.Error(t, act.Audit(clictx(ctx, t)))
	buf.Reset()

	// test with filter
	c := clictx(ctx, t, "foo")
	assert.Error(t, act.Audit(c))
	buf.Reset()

	// test empty store
	for _, v := range []string{"foo", "bar", "baz"} {
		assert.NoError(t, act.Store.Delete(ctx, v))
	}
	assert.NoError(t, act.Audit(clictx(ctx, t)))
	assert.Contains(t, "No secrets found", buf.String())
	buf.Reset()

}
