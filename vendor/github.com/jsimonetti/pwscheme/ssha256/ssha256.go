// Package ssha256 provides functions to generate and validate {SSHA256} styled
// password schemes.
// The method used is defined in RFC 2307 and uses a salted SHA256 secure hashing
// algorithm
package ssha256

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

// ErrNotSshaPassword occurs when Validate receives a non-SSHA256 hash
var ErrNotSshaPassword = errors.New("string is not a SSHA256 hashed password")

// ErrBase64DecodeFailed occurs when the given hash cannot be decode
var ErrBase64DecodeFailed = errors.New("base64 decode of hash failed")

// ErrNotMatching occurs when the given password and hash do not match
var ErrNotMatching = errors.New("hash does not match password")

// Generate encrypts a password with a random salt of definable length and
// returns the {SSHA256} encoding of the password
func Generate(password string, length uint8) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := createHash(password, salt)
	ret := fmt.Sprintf("{SSHA256}%s", base64.StdEncoding.EncodeToString(hash))
	return ret, nil
}

// Validate compares a given password with a {SSHA256} encoded password
// Returns true is they match or an error otherwise
func Validate(password string, hash string) (bool, error) {
	if len(hash) < 10 || string(hash[0:9]) != "{SSHA256}" {
		return false, ErrNotSshaPassword
	}
	data, err := base64.StdEncoding.DecodeString(hash[9:])
	if len(data) < 33 || err != nil {
		return false, ErrBase64DecodeFailed
	}

	newhash := createHash(password, data[32:])
	hashedpw := base64.StdEncoding.EncodeToString(newhash)

	if hashedpw == hash[9:] {
		return true, nil
	}

	return false, ErrNotMatching
}

// createHash appends password and salt together to a byte array
func createHash(password string, salt []byte) []byte {
	pass := []byte(password)
	str := append(pass[:], salt[:]...)
	sum := sha256.Sum256(str)
	result := append(sum[:], salt[:]...)
	return result
}
