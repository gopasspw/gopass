package xkcdpwgen

// from https://github.com/justwatchcom/gopass/blob/master/utils/pwgen/pwgen.go

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
	rand.Seed(time.Now().Unix() + int64(os.Getpid()+os.Getppid()))
}

func randomInteger(max int) int {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err == nil {
		return int(i.Int64())
	}
	fmt.Println("WARNING: No crypto/rand available. Falling back to PRNG")
	return rand.Intn(max)
}
