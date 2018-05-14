package pwgen

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/muesli/crunchy"
)

const (
	digits = "0123456789"
	upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower  = "abcdefghijklmnopqrstuvwxyz"
	syms   = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	// CharAlphaNum is the class of alpha-numeric characters
	CharAlphaNum = digits + upper + lower
	// CharAll is the class of all characters
	CharAll = digits + upper + lower + syms
)

func init() {
	// seed math/rand in case we have to fall back to using it
	rand.Seed(time.Now().Unix() + int64(os.Getpid()+os.Getppid()))
}

// GeneratePassword generates a random, hard to remember password
func GeneratePassword(length int, symbols bool) string {
	chars := digits + upper + lower
	if symbols {
		chars += syms
	}
	if c := os.Getenv("GOPASS_CHARACTER_SET"); c != "" {
		chars = c
	}
	return GeneratePasswordCharsetCheck(length, chars)
}

// GeneratePasswordCharset generates a random password from a given
// set of characters
func GeneratePasswordCharset(length int, chars string) string {
	pw := &bytes.Buffer{}
	for pw.Len() < length {
		_ = pw.WriteByte(chars[randomInteger(len(chars))])
	}
	return pw.String()
}

// GeneratePasswordCharsetCheck generates a random password from a given
// set of characters and validates the generated password with crunchy
func GeneratePasswordCharsetCheck(length int, chars string) string {
	validator := crunchy.NewValidator()
	var password string

	for i := 0; i < 3; i++ {
		pw := &bytes.Buffer{}
		for pw.Len() < length {
			_ = pw.WriteByte(chars[randomInteger(len(chars))])
		}
		password = pw.String()

		if validator.Check(password) == nil {
			break
		}
	}

	return password
}

func randomInteger(max int) int {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err == nil {
		return int(i.Int64())
	}
	fmt.Println("WARNING: No crypto/rand available. Falling back to PRNG")
	return rand.Intn(max)
}
