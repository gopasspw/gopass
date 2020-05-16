package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/muesli/goprogressbar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipients(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	goprogressbar.Stdout = buf
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
		goprogressbar.Stdout = os.Stdout
	}()

	// RecipientsPrint
	assert.NoError(t, act.RecipientsPrint(clictx(ctx, t)))
	want := `Hint: run 'gopass sync' to import any missing public keys
gopass
└── 0xDEADBEEF

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// RecipientsComplete
	act.RecipientsComplete(clictx(ctx, t))
	want = "0xDEADBEEF\n"
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// RecipientsAdd
	assert.Error(t, act.RecipientsAdd(clictx(ctx, t)))
	buf.Reset()

	// RecipientsRemove
	assert.Error(t, act.RecipientsRemove(clictx(ctx, t)))
	buf.Reset()

	// RecipientsAdd 0xFEEDBEEF
	assert.NoError(t, act.RecipientsAdd(clictx(ctx, t, "0xFEEDBEEF")))
	buf.Reset()

	// RecipientsAdd 0xBEEFFEED
	assert.Error(t, act.RecipientsAdd(clictx(ctx, t, "0xBEEFFEED")))
	buf.Reset()

	// RecipientsRemove 0xDEADBEEF
	assert.NoError(t, act.RecipientsRemove(clictx(ctx, t, "0xDEADBEEF")))
	buf.Reset()
}
