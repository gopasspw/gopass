package pwgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwgenExternal(t *testing.T) {
	t.Setenv("GOPASS_EXTERNAL_PWGEN", "powershell.exe -Command write-output 1234 #")
	ans, err := GenerateExternal(4)
	if err != nil {
		panic("Unable to generate using external generator")
	}
	assert.Equal(t, "1234", ans)
}
