package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"

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
