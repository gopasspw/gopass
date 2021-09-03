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
	fmt.Println(GenerateMemorablePassword(12, false, false))
}

func TestPwgen(t *testing.T) {
	for _, sym := range []bool{true, false} {
		for i := 1; i < 50; i++ {
			Syms := CharAlphaNum
			if sym {
				Syms = CharAll
			}
			assert.Equal(t, i, len(GeneratePasswordCharset(i, Syms)))
		}
	}
}

func TestPwgenCharset(t *testing.T) {
	_ = os.Setenv("GOPASS_CHARACTER_SET", "a")
	assert.Equal(t, "aaaa", GeneratePassword(4, true))
	assert.Equal(t, "", GeneratePasswordCharsetCheck(4, "a"))
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
	os.Stderr = w
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
			classes: []string{Lower},
			ok:      true,
		},
		{
			pw:      "aB1$",
			classes: []string{Lower, Upper, Syms, Digits},
			ok:      true,
		},
		{
			pw:      "ab1$",
			classes: []string{Lower, Upper, Syms, Digits},
			ok:      false,
		},
	} {
		assert.Equal(t, tc.ok, containsAllClasses(tc.pw, tc.classes...))
	}
}

func TestGeneratePasswordWithAllClasses(t *testing.T) {
	pw, err := GeneratePasswordWithAllClasses(50, true)
	assert.NoError(t, err)
	assert.Equal(t, 50, len(pw))
}

func TestGenerateMemorablePassword(t *testing.T) {
	pw := GenerateMemorablePassword(20, false, false)
	assert.GreaterOrEqual(t, len(pw), 20)
}

func TestGenerateMemorablePasswordCapital(t *testing.T) {
	pw := GenerateMemorablePassword(20, false, true)
	assert.GreaterOrEqual(t, len(pw), 20)
}

func TestPrune(t *testing.T) {
	for _, tc := range []struct {
		In     string
		Cutset string
		Out    string
	}{
		{
			"abc",
			"b",
			"ac",
		},
		{
			"01lZO",
			"01lO",
			"Z",
		},
	} {
		assert.Equal(t, tc.Out, Prune(tc.In, tc.Cutset))
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
