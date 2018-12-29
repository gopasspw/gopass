// Package md5crypt provides functions to generate and validate {MD5-CRYPT} styled
// password schemes.
// The method used is compatible with libc crypt used in /etc/shadow
package md5crypt

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
)

const itoa64 = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var md5CryptSwaps = [16]int{12, 6, 0, 13, 7, 1, 14, 8, 2, 15, 9, 3, 5, 10, 4, 11}

var magic = []byte("$1$")

// ErrNotMd5cryptPassword occurs when Validate receives a non-SSHA hash
var ErrNotMd5cryptPassword = errors.New("string is not a MD5-CRYPT password")

// ErrNotMatching occurs when the given password and hash do not match
var ErrNotMatching = errors.New("hash does not match password")

// ErrSaltLengthInCorrect occurs when the given salt is not of the correct
// length
var ErrSaltLengthInCorrect = errors.New("salt length incorrect")

// Generate encrypts a password with a random salt of definable length and
// returns the {MD5-CRYPT} encoding of the password
func Generate(password string, length uint8) (string, error) {
	if length > 8 || length < 1 {
		return "", ErrSaltLengthInCorrect
	}
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("{MD5-CRYPT}%s", crypt([]byte(password), salt))

	return hash, nil
}

// Validate compares a given password with a {SSHA} encoded password
// Returns true is they match or an error otherwise
func Validate(password string, hash string) (bool, error) {
	if len(hash) < 15 || string(hash[0:14]) != "{MD5-CRYPT}$1$" {
		return false, ErrNotMd5cryptPassword
	}

	data := strings.Split(hash[14:], "$")

	newhash := crypt([]byte(password), []byte(data[0]))

	if string(newhash) == hash[11:] {
		return true, nil
	}

	return false, ErrNotMatching
}

func crypt(password, salt []byte) []byte {

	d := md5.New()

	d.Write(password)
	d.Write(magic)
	d.Write(salt)

	d2 := md5.New()
	d2.Write(password)
	d2.Write(salt)
	d2.Write(password)

	for i, mixin := 0, d2.Sum(nil); i < len(password); i++ {
		d.Write([]byte{mixin[i%16]})
	}

	for i := len(password); i != 0; i >>= 1 {
		if i&1 == 0 {
			d.Write([]byte{password[0]})
		} else {
			d.Write([]byte{0})
		}
	}

	final := d.Sum(nil)

	for i := 0; i < 1000; i++ {
		d2 := md5.New()
		if i&1 == 0 {
			d2.Write(final)
		} else {
			d2.Write(password)
		}

		if i%3 != 0 {
			d2.Write(salt)
		}

		if i%7 != 0 {
			d2.Write(password)
		}

		if i&1 == 0 {
			d2.Write(password)
		} else {
			d2.Write(final)
		}
		final = d2.Sum(nil)
	}

	result := make([]byte, 0, 22)
	v := uint(0)
	bits := uint(0)
	for _, i := range md5CryptSwaps {
		v |= (uint(final[i]) << bits)
		for bits = bits + 8; bits > 6; bits -= 6 {
			result = append(result, itoa64[v&0x3f])
			v >>= 6
		}
	}
	result = append(result, itoa64[v&0x3f])

	return append(append(append(magic, salt...), '$'), result...)
}
