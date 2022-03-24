package fsutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUmask(t *testing.T) { //nolint:paralleltest
	for _, vn := range []string{"GOPASS_UMASK", "PASSWORD_STORE_UMASK"} {
		for in, out := range map[string]int{
			"002":      0o2,
			"0777":     0o777,
			"000":      0,
			"07557575": 0o77,
		} {
			assert.NoError(t, os.Setenv(vn, in))
			assert.Equal(t, out, Umask())
			assert.NoError(t, os.Unsetenv(vn))
		}
	}
}
