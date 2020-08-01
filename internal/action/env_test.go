package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvLeafHappyPath(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	// Command-line would be: "gopass env foo env", where "foo" is an existing
	// secret with value "secret". We expect to see the key/value in the output
	// of the /usr/bin/env utility in the form "FOO=secret".
	//
	// TODO(@dominikschulz): consider populating foo with a long, random password to
	// absolutely ensure that the correct secret is displayed.
	assert.NoError(t, act.Env(gptest.CliCtx(ctx, t, "foo", "env")))
	assert.Contains(t, buf.String(), "FOO=secret\n")
}

func TestEnvSecretNotFound(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	// Command-line would be: "gopass env non-existing true".
	assert.EqualError(t, act.Env(gptest.CliCtx(ctx, t, "non-existing", "true")),
		"Secret non-existing not found")
}
