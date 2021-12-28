package action

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/stretchr/testify/assert"
)

func TestExitError(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ExitError(ExitUnknown, fmt.Errorf("test"), "test"))
	assert.NotContains(t, buf.String(), "Stacktrace")
}
