package fsutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUmask(t *testing.T) {
	for _, vn := range []string{"GOPASS_UMASK", "PASSWORD_STORE_UMASK"} {
		for in, out := range map[string]int{
			"002":      0o2,
			"0777":     0o777,
			"000":      0,
			"07557575": 0o77,
		} {
			t.Run(vn, func(t *testing.T) {
				t.Setenv(vn, in)
				assert.Equal(t, out, Umask())
			})
		}
	}
}
