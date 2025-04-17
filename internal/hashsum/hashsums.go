// Package hashsum provides hash functions for various algorithms.
package hashsum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"

	"github.com/zeebo/blake3"
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

func Blake3Hex(in string) string {
	return fmt.Sprintf("%x", blake3.Sum256([]byte(in)))
}
