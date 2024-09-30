package pwgen

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"
)

func init() {
	// seed math/rand in case we have to fall back to using it
	randFallback = rand.New(rand.NewSource(time.Now().Unix() + int64(os.Getpid()+os.Getppid())))
}

var randFallback *rand.Rand

func randomInteger(max int) int {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err == nil {
		return int(i.Int64())
	}

	fmt.Fprintln(os.Stderr, "WARNING: No crypto/rand available. Falling back to PRNG")

	return randFallback.Intn(max)
}
