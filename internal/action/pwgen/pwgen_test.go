package pwgen

import (
	"bytes"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPwgen(t *testing.T) {
	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	require.NoError(t, Pwgen(gptest.CliCtxWithFlags(ctx, t, map[string]string{"one-per-line": "true"}, "24", "1")))
	assert.GreaterOrEqual(t, len(buf.Bytes()), 24, string(buf.Bytes()))
}
