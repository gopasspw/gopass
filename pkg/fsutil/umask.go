package fsutil

import (
	"os"
	"strconv"
)

// Umask extracts the umask from env
func Umask() int {
	for _, en := range []string{"GOPASS_UMASK", "PASSWORD_STORE_UMASK"} {
		if um := os.Getenv(en); um != "" {
			if iv, err := strconv.ParseInt(um, 8, 32); err == nil && iv >= 0 && iv <= 0777 {
				return int(iv)
			}
		}
	}
	return 077
}
