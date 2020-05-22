package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
)

func TestExitError(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithDebug(ctx, true)
	buf := &bytes.Buffer{}
	out.Stdout = buf

	assert.Error(t, ExitError(ctx, ExitUnknown, fmt.Errorf("test"), "test"))
	sv := buf.String()
	if !strings.Contains(sv, "Stacktrace") {
		t.Errorf("Should contain an stacktrace")
	}
	out.Stdout = os.Stdout
}
