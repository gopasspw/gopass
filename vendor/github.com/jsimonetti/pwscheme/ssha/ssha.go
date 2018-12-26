// Package ssha provides functions to generate and validate {SSHA} styled
// password schemes.
// The method used is defined in RFC 2307 and uses a salted SHA1 secure hashing
// algorithm
package ssha

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
)

// ErrNotSshaPassword occurs when Validate receives a non-SSHA hash
var ErrNotSshaPassword = errors.New("string is not a SSHA hashed password")

// ErrBase64DecodeFailed occurs when the given hash cannot be decode
var ErrBase64DecodeFailed = errors.New("base64 decode of hash failed")

// ErrNotMatching occurs when the given password and hash do not match
var ErrNotMatching = errors.New("hash does not match password")

// Generate encrypts a password with a random salt of definable length and
// returns the {SSHA} encoding of the password
func Generate(password string, length uint8) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := createHash(password, salt)
	ret := fmt.Sprintf("{SSHA}%s", base64.StdEncoding.EncodeToString(hash))
	return ret, nil
}

// Validate compares a given password with a {SSHA} encoded password
// Returns true is they match or an error otherwise
func Validate(password string, hash string) (bool, error) {
	if len(hash) < 7 || string(hash[0:6]) != "{SSHA}" {
		return false, ErrNotSshaPassword
	}
	data, err := base64.StdEncoding.DecodeString(hash[6:])
	if len(data) < 21 || err != nil {
		return false, ErrBase64DecodeFailed
	}

	newhash := createHash(password, data[20:])
	hashedpw := base64.StdEncoding.EncodeToString(newhash)

	if hashedpw == hash[6:] {
		return true, nil
	}

	return false, ErrNotMatching
}

// createHash appends password and salt together to a byte array
func createHash(password string, salt []byte) []byte {
	pass := []byte(password)
	str := append(pass[:], salt[:]...)
	sum := sha1.Sum(str)
	result := append(sum[:], salt[:]...)
	return result
}
