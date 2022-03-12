package protect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtect(t *testing.T) {
	t.Parallel()

	assert.NoError(t, Pledge(""))
}
