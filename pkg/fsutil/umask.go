package fsutil

import (
	"os"
	"strconv"
)

// Umask extracts the umask from the environment variables.
// It checks for GOPASS_UMASK and PASSWORD_STORE_UMASK.
// If neither is set, it returns the default umask of 0o77.
func Umask() int {
	for _, en := range []string{"GOPASS_UMASK", "PASSWORD_STORE_UMASK"} {
		um := os.Getenv(en)
		if um == "" {
			continue
		}

		iv, err := strconv.ParseInt(um, 8, 32)
		if err != nil {
			continue
		}

		if iv >= 0 && iv <= 0o777 {
			return int(iv)
		}
	}

	return 0o77
}
