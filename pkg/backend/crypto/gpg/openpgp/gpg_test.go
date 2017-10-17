package openpgp

import (
	"testing"

	"github.com/justwatchcom/gopass/pkg/backend"
)

func TestInterface(t *testing.T) {
	var crypto backend.Crypto
	crypto = &GPG{}
	_ = crypto
}
