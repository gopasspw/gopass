package hashsum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

func MD5Hex(in string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(in)))
}

func SHA1Hex(in string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(in)))
}

func SHA256Hex(in string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(in)))
}

func SHA512Hex(in string) string {
	return fmt.Sprintf("%x", sha512.Sum512([]byte(in)))
}
