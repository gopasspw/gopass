package pwgen

import (
	crand "crypto/rand"
	"math/big"
)

func randomInteger(maxVal int) int {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(maxVal)))
	if err != nil {
		panic(err)
	}

	return int(i.Int64())
}
