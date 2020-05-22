package pwgen

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleGenerateMemorablePassword() {
	fmt.Println(GenerateMemorablePassword(12, false))
}

func TestPwgen(t *testing.T) {
	for _, sym := range []bool{true, false} {
		for i := 0; i < 50; i++ {
			syms := CharAlphaNum
			if sym {
				syms = CharAll
			}
			assert.Equal(t, i, len(GeneratePasswordCharset(i, syms)))
		}
	}
}

func TestPwgenCharset(t *testing.T) {
	_ = os.Setenv("GOPASS_CHARACTER_SET", "a")
	assert.Equal(t, "aaaa", GeneratePassword(4, true))
}

func TestPwgenNoCrand(t *testing.T) {
	old := rand.Reader
	rand.Reader = strings.NewReader("")
	defer func() {
		rand.Reader = old
	}()
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	done := make(chan string)
	go func() {
		buf := &bytes.Buffer{}
		_, _ = io.Copy(buf, r)
		done <- buf.String()
	}()
	// if we seed math/rand with 1789, the first "random number" will be 42
	mrand.Seed(1789)
	n := randomInteger(1024)
	assert.NoError(t, w.Close())
	os.Stdout = oldOut
	assert.Equal(t, 42, n)
	assert.Equal(t, "WARNING: No crypto/rand available. Falling back to PRNG\n", <-done)
}

func TestContainsAllClasses(t *testing.T) {
	for _, tc := range []struct {
		pw      string
		classes []string
		ok      bool
	}{
		{
			pw:      "foobar",
			classes: []string{lower},
			ok:      true,
		},
		{
			pw:      "aB1$",
			classes: []string{lower, upper, syms, digits},
			ok:      true,
		},
		{
			pw:      "ab1$",
			classes: []string{lower, upper, syms, digits},
			ok:      false,
		},
	} {
		assert.Equal(t, tc.ok, containsAllClasses(tc.pw, tc.classes...))
	}
}

func BenchmarkPwgen(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GeneratePasswordCharset(24, CharAll)
	}
}

func BenchmarkPwgenCheck(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GeneratePasswordCharsetCheck(24, CharAll)
	}
}
