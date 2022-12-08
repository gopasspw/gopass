package appdir

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserHome(t *testing.T) { //nolint:paralleltest
	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	assert.Equal(t, td, UserHome())
}
