package protect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtect(t *testing.T) {
	assert.NoError(t, Pledge(""))
}
