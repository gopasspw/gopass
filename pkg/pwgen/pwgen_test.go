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
	"github.com/stretchr/testify/require"
)

func ExampleGenerateMemorablePassword() { //nolint:testableexamples
	fmt.Println(GenerateMemorablePassword(12, false, false))
}

func TestPwgen(t *testing.T) {
	t.Parallel()

	for _, sym := range []bool{true, false} {
		for i := 1; i < 50; i++ {
			Syms := CharAlphaNum
			if sym {
				Syms = CharAll
			}

			assert.Len(t, GeneratePasswordCharset(i, Syms), i)
		}
	}
}

func TestPwgenCharset(t *testing.T) {
	t.Setenv("GOPASS_CHARACTER_SET", "a")

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

	require.NoError(t, w.Close())

	os.Stdout = oldOut

	assert.Equal(t, 42, n)
	assert.Equal(t, "WARNING: No crypto/rand available. Falling back to PRNG\n", <-done)
}

func TestContainsAllClasses(t *testing.T) {
	t.Parallel()

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
		t.Run(tc.pw, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.ok, containsAllClasses(tc.pw, tc.classes...))
		})
	}
}

func TestGeneratePasswordWithAllClasses(t *testing.T) {
	t.Parallel()

	pw, err := GeneratePasswordWithAllClasses(50, true)
	require.NoError(t, err)
	assert.Len(t, pw, 50)
}

func TestGenerateMemorablePassword(t *testing.T) {
	t.Parallel()

	pw := GenerateMemorablePassword(20, false, false)
	assert.GreaterOrEqual(t, len(pw), 20)
	assert.Equal(t, pw, strings.ToLower(pw))
}

func TestGenerateMemorablePasswordCapital(t *testing.T) {
	t.Parallel()

	pw := GenerateMemorablePassword(20, false, true)
	assert.GreaterOrEqual(t, len(pw), 20)
	assert.NotEqual(t, pw, strings.ToLower(pw))
}

func TestPrune(t *testing.T) {
	t.Parallel()

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
	for n := 0; n < b.N; n++ { //nolint:intrange // b.N is evaluated at each iteration.
		GeneratePasswordCharset(24, CharAll)
	}
}

func BenchmarkPwgenCheck(b *testing.B) {
	for n := 0; n < b.N; n++ { //nolint:intrange // b.N is evaluated at each iteration.
		GeneratePasswordCharsetCheck(24, CharAll)
	}
}
