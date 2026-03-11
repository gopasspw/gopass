package pwgen

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
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
	assert.Empty(t, GeneratePasswordCharsetCheck(4, "a"))
}

func TestPwgenNoCrandFallback(t *testing.T) {
	oldFallback := randFallback
	oldReader := crand.Reader
	crand.Reader = strings.NewReader("")

	defer func() {
		crand.Reader = oldReader
		randFallback = oldFallback
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
	randFallback = rand.New(rand.NewSource(1789))

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

func TestGeneratePasswordCharsetStrict(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		length  int
		charset string
		wantErr bool
	}{
		{
			name:    "all character classes",
			length:  20,
			charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%&*",
			wantErr: false,
		},
		{
			name:    "only digits and lowercase",
			length:  10,
			charset: "abcdefghijklmnopqrstuvwxyz0123456789",
			wantErr: false,
		},
		{
			name:    "only uppercase",
			length:  10,
			charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			wantErr: false,
		},
		{
			name:    "digits only",
			length:  6,
			charset: "0123456789",
			wantErr: false,
		},
		{
			name:    "symbols and digits",
			length:  15,
			charset: "0123456789!@#$%^&*()",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pw, err := GeneratePasswordCharsetStrict(tc.length, tc.charset)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Len(t, pw, tc.length)

			// Verify all detected character classes are present
			if strings.ContainsAny(tc.charset, Digits) {
				assert.True(t, containsAllClasses(pw, Digits), "password should contain at least one digit")
			}
			if strings.ContainsAny(tc.charset, Upper) {
				assert.True(t, containsAllClasses(pw, Upper), "password should contain at least one uppercase letter")
			}
			if strings.ContainsAny(tc.charset, Lower) {
				assert.True(t, containsAllClasses(pw, Lower), "password should contain at least one lowercase letter")
			}
			if strings.ContainsAny(tc.charset, Syms) {
				assert.True(t, containsAllClasses(pw, Syms), "password should contain at least one symbol")
			}

			// Verify all characters are from the charset
			for _, c := range pw {
				assert.Contains(t, tc.charset, string(c), "password should only contain characters from charset")
			}
		})
	}
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
