package pwgen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwgenExternal(t *testing.T) {
	_ = os.Setenv("GOPASS_EXTERNAL_PWGEN", "powershell.exe -Command write-output 1234 #")
	ans, err := GenerateExternal(4)
	if err != nil {
		panic("Unable to generate using external generator")
	}
	assert.Equal(t, "1234", ans)
}
