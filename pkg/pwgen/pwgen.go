package pwgen

import (
	"bytes"
	crand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	shellquote "github.com/kballard/go-shellquote"
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

func Init() {
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
	if c := os.Getenv("GOPASS_EXTERNAL_PWGEN"); c != "" {
		if pw, err := generateExternal(c); err == nil {
			return pw
		}
	}
	return GeneratePasswordCharsetCheck(length, chars)
}

func generateExternal(c string) (string, error) {
	cmdArgs, err := shellquote.Split(c)
	if err != nil {
		return "", err
	}
	if len(cmdArgs) < 1 {
		return "", fmt.Errorf("no command")
	}
	exe := cmdArgs[0]
	args := []string{}
	if len(cmdArgs) > 1 {
		args = cmdArgs[1:]
	}
	out, err := exec.Command(exe, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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

// GeneratePasswordWithAllClasses tries to enforce a password which
// contains all character classes instead of only enabling them.
// This is especially useful for broken (corporate) password policies
// that mandate the use of certain character classes for not good reason
func GeneratePasswordWithAllClasses(length int) (string, error) {
	pw := GeneratePasswordCharset(length, CharAll)
	for i := 0; i < 100; i++ {
		if containsAllClasses(pw, digits, upper, lower, syms) {
			return pw, nil
		}
		pw = GeneratePasswordCharset(length, CharAll)
	}
	return "", errors.New("failed to generate matching password after 100 rounds")
}

func containsAllClasses(pw string, classes ...string) bool {
CLASSES:
	for _, class := range classes {
		for _, ch := range class {
			if strings.Contains(pw, string(ch)) {
				continue CLASSES
			}
		}
		return false
	}
	return true
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
