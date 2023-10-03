package pwgen

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"
)

var gr *rand.Rand

func init() {
	gr = rand.New(rand.NewSource(time.Now().Unix() + int64(os.Getpid()+os.Getppid())))
}

func randomInteger(max int) int {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err == nil {
		return int(i.Int64())
	}
	fmt.Fprintln(os.Stderr, "WARNING: No crypto/rand available. Falling back to PRNG")

	return gr.Intn(max)
}
