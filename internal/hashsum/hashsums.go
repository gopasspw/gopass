// Package hashsum provides hash functions for various algorithms.
package hashsum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"github.com/zeebo/blake3"
)

// Note: MD5 is cryptographically broken and should not be used for security purposes.
func MD5Hex(in string) string {
	hash := md5.Sum([]byte(in))
	return hex.EncodeToString(hash[:])
}

// Note: SHA-1 is cryptographically broken and should not be used for security purposes.
func SHA1Hex(in string) string {
	hash := sha1.Sum([]byte(in))
	return hex.EncodeToString(hash[:])
}

func SHA256Hex(in string) string {
	hash := sha256.Sum256([]byte(in))
	return hex.EncodeToString(hash[:])
}

func SHA512Hex(in string) string {
	hash := sha512.Sum512([]byte(in))
	return hex.EncodeToString(hash[:])
}

func Blake3Hex(in string) string {
	hash := blake3.Sum256([]byte(in))
	return hex.EncodeToString(hash[:])
}
