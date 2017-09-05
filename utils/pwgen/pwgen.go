package pwgen

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"
)

const (
	digits = "0123456789"
	upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower  = "abcdefghijklmnopqrstuvwxyz"
	syms   = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

func init() {
	// seed math/rand in case we have to fall back to using it
	rand.Seed(time.Now().Unix() + int64(os.Getpid()+os.Getppid()))
}

// GeneratePassword generates a random, hard to remember password
func GeneratePassword(length int, symbols bool) []byte {
	chars := digits + upper + lower
	if symbols {
		chars += syms
	}
	if c := os.Getenv("GOPASS_CHARACTER_SET"); c != "" {
		chars = c
	}
	pw := &bytes.Buffer{}
	for pw.Len() < length {
		_ = pw.WriteByte(chars[randomInteger(len(chars))])
	}

	return pw.Bytes()
}

func randomInteger(max int) int {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err == nil {
		return int(i.Int64())
	}
	fmt.Println("WARNING: No crypto/rand available. Falling back to PRNG")
	return rand.Intn(max)
}
