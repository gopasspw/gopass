package cui

import (
	"bytes"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskForPrivateKey(t *testing.T) {
	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	ctx := config.NewContextInMemory()

	ctx = ctxutil.WithAlwaysYes(ctx, true)
	key, err := AskForPrivateKey(ctx, plain.New(), "test")
	require.NoError(t, err)
	assert.Equal(t, "0xDEADBEEF", key)
	buf.Reset()
}

func TestAskForGitConfigUser(t *testing.T) {
	// necessary for setting up the env
	u := gptest.NewGUnitTester(t)
	assert.NotNil(t, u)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithTerminal(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	_, _, err := AskForGitConfigUser(ctx, plain.New())
	require.NoError(t, err)
}

type fakeMountPointer []string

func (f fakeMountPointer) MountPoints() []string {
	return f
}

func TestAskForStore(t *testing.T) {
	ctx := config.NewContextInMemory()

	// test non-interactive
	ctx = ctxutil.WithInteractive(ctx, false)
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{"foo", "bar"}))

	// test interactive
	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{"foo", "bar"}))

	// test zero mps
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{}))

	// test one mp
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{"foo"}))

	// test two mps
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{"foo", "bar"}))
}

func TestSorted(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"a", "b", "c"}, sorted([]string{"c", "a", "b"}))
}
