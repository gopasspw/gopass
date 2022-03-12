package exit

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) { //nolint:paralleltest
	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, Error(Unknown, fmt.Errorf("test"), "test"))
	assert.NotContains(t, buf.String(), "Stacktrace")
}
