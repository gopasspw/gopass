package fsutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUmask(t *testing.T) {
	for _, vn := range []string{"GOPASS_UMASK", "PASSWORD_STORE_UMASK"} {
		for in, out := range map[string]int{
			"002":      02,
			"0777":     0777,
			"000":      0,
			"07557575": 077,
		} {
			assert.NoError(t, os.Setenv(vn, in))
			assert.Equal(t, out, Umask())
			assert.NoError(t, os.Unsetenv(vn))
		}
	}
}
