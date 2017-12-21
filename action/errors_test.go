package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
)

func TestExitError(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithDebug(ctx, true)
	buf := &bytes.Buffer{}
	out.Stdout = buf

	err := exitError(ctx, ExitUnknown, fmt.Errorf("test"), "test")
	if err == nil {
		t.Errorf("Should fail")
	}
	sv := buf.String()
	if !strings.Contains(sv, "Stacktrace") {
		t.Errorf("Should contain an stacktrace")
	}
	out.Stdout = os.Stdout
}
