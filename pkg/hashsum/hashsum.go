package hashsum

import (
	"crypto/sha1"
	"fmt"
)

// SHA1 returns an uppercase hex SHA1 sum of the input string
func SHA1(data string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(data))
	return fmt.Sprintf("%X", h.Sum(nil))
}
