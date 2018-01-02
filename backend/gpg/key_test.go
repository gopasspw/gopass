package gpg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	k := Key{
		KeyType:        "sec",
		KeyLength:      2048,
		Validity:       "u",
		CreationDate:   time.Now(),
		ExpirationDate: time.Now().Add(time.Hour),
		Ownertrust:     "ultimate",
		Fingerprint:    "25FF1614B8F87B52FFFF99B962AF4031C82E0039",
	}
	assert.Equal(t, k.IsUseable(), true)
}
