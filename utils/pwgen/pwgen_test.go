package pwgen

import (
	"bytes"
	"crypto/rand"
	"io"
	mrand "math/rand"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwgen(t *testing.T) {
	for _, sym := range []bool{true, false} {
		for i := 0; i < 50; i++ {
			sec := GeneratePassword(i, sym)
			if len(sec) != i {
				t.Errorf("Length mismatch")
			}
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
