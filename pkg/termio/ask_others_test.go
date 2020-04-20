// +build !windows

package termio

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/stretchr/testify/assert"
)

func TestAskForPasswordNonInteractive(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)

	_, err := AskForPassword(ctx, "test")
	assert.Error(t, err)

	// provide value on redirected stdin
	input := `foo
foo
foobar
foobaz
`

	Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)
	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	sv, err := AskForPassword(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "foo", sv)

	sv, err = AskForPassword(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "", sv)
}
